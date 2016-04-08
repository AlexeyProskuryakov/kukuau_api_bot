package init

import (
	"fmt"
	"log"
	"net/http"
	"errors"

	m "msngr"
	d "msngr/db"
	c "msngr/configuration"
	cnsl "msngr/console"
	n "msngr/notify"
	rp "msngr/ruposts"
	sh "msngr/shop"
	q "msngr/quests"

	t "msngr/taxi"
	i "msngr/taxi/infinity"
	"msngr/taxi/sedi"
	"msngr/taxi/geo"

	v "msngr/voting"
	"msngr/chat"
	"msngr/utils"
	"msngr/users"
)

func GetTaxiAPI(params c.TaxiApiParams, for_name string) (t.TaxiInterface, error) {
	switch api_name := params.Name; api_name {
	case i.INFINITY:
		return i.GetInfinityAPI(params, for_name), nil
	case t.FAKE:
		return t.GetFakeAPI(params), nil
	case sedi.SEDI:
		sedi_api := sedi.NewSediAPI(params)
		return sedi_api, nil
	}
	return nil, errors.New("Not imply name of api")
}

func GetTaxiAPIInstruments(params c.TaxiApiParams, for_name string) (t.TaxiInterface, t.AddressSupplier, error) {
	switch api_name := params.Name; api_name {
	case i.INFINITY:
		return i.GetInfinityAPI(params, for_name), i.GetInfinityAddressSupplier(params, for_name), nil
	case t.FAKE:
		return t.GetFakeAPI(params), i.GetInfinityAddressSupplier(params, for_name), nil
	case sedi.SEDI:
		sedi_api := sedi.NewSediAPI(params)
		return sedi_api, sedi_api, nil
	}
	return nil, nil, errors.New("Not imply name of api")
}

func GetAddressInstruments(c c.Configuration, taxi_name string, external_supplier t.AddressSupplier) (t.AddressHandler, t.AddressSupplier) {
	if c.Taxis[taxi_name].Api.Name == sedi.SEDI {
		log.Printf("[ADDRESSES ENGINE] For %v Will use SEDI address supplier no any address handler", taxi_name)
		return nil, external_supplier
	}
	own := geo.NewOwnAddressHandler(c.Main.ElasticConn, c.Taxis[taxi_name].Api.GeoOrbit, external_supplier)
	if own == nil {
		google := geo.NewGoogleAddressHandler(c.Main.GoogleKey, c.Taxis[taxi_name].Api.GeoOrbit, external_supplier)
		if google == nil {
			log.Printf("[ADDRESSES ENGINE] For %v Will use external address supplier and no any address handler", taxi_name)
			return nil, external_supplier
		}
		log.Printf("[ADDRESSES ENGINE] For %v Will use google addresses", taxi_name)
		return google, google
	}
	log.Printf("For %v Will use own addresses", taxi_name)
	return own, own
}

func StartBot(db *d.MainDb, result chan string) c.Configuration {
	conf := c.ReadConfig()
	log.Printf("configuration for db:\nconnection string: %+v\ndatabase name: %+v", conf.Main.Database.ConnString, conf.Main.Database.Name)

	for taxi_name, taxi_conf := range conf.Taxis {
		log.Printf("taxi api configuration for %+v:\n%v", taxi_conf.Name, taxi_conf.Api)
		external_api, external_address_supplier, err := GetTaxiAPIInstruments(taxi_conf.Api, taxi_name)

		if err != nil {
			log.Printf("Skip this taxi api [%+v]\nBecause: %v", taxi_conf.Api, err)
			continue
		}

		apiMixin := t.ExternalApiMixin{API: external_api}

		carsCache := t.NewCarsCache(external_api)
		notifier := n.NewNotifier(conf.Main.CallbackAddr, taxi_conf.Key, db)

		address_handler, address_supplier := GetAddressInstruments(conf, taxi_name, external_address_supplier)

		botContext := t.FormTaxiBotContext(&apiMixin, db, taxi_conf, address_handler, carsCache)
		controller := m.FormBotController(botContext, db)

		log.Printf("Was create bot context: %+v\n", botContext)
		http.HandleFunc(fmt.Sprintf("/taxi/%v", taxi_conf.Name), controller)

		go func() {
			api, err := GetTaxiAPI(taxi_conf.Api, taxi_name + "_watch")
			if err != nil {
				log.Printf("Error at get api: %v for %v, will not use order watching", err, taxi_name)
			}
			cc := t.NewCarsCache(api)
			taxiContext := t.TaxiContext{API: api, DataBase: db, Cars: cc, Notifier: notifier}
			log.Printf("Will start order watcher for [%v]", botContext.Name)
			t.TaxiOrderWatch(&taxiContext, botContext)
		}()

		http.HandleFunc(fmt.Sprintf("/taxi/%v/streets", taxi_conf.Name), func(w http.ResponseWriter, r *http.Request) {
			geo.StreetsSearchController(w, r, address_supplier)
		})
		if m.TEST {
			http.HandleFunc(fmt.Sprintf("/taxi/%v/streets/ext", taxi_conf.Name), func(w http.ResponseWriter, r *http.Request) {
				geo.StreetsSearchController(w, r, external_address_supplier)
			})
		}

		result <- fmt.Sprintf("taxi_%v", taxi_name)
	}

	for _, shop_conf := range conf.Shops {
		bot_context := sh.FormShopCommands(db, &shop_conf)
		shop_controller := m.FormBotController(bot_context, db)
		http.HandleFunc(fmt.Sprintf("/shop/%v", shop_conf.Name), shop_controller)
		result <- fmt.Sprintf("shops_%v", shop_conf.Name)
	}

	if conf.RuPost.WorkUrl != "" {
		log.Printf("will start ru post controller at: %v and will send requests to: %v", conf.RuPost.WorkUrl, conf.RuPost.ExternalUrl)
		rp_bot_context := rp.FormRPBotContext(conf)
		rp_controller := m.FormBotController(rp_bot_context, db)
		http.HandleFunc(conf.RuPost.WorkUrl, rp_controller)
		result <- "rupost"
	}

	cs := c.NewConfigurationStorage(conf.Main.Database.ConnString, conf.Main.Database.Name)
	qs := q.NewQuestStorage(conf.Main.Database.ConnString, conf.Main.Database.Name)

	for q_name, _ := range conf.Quests {
		log.Printf("Will handling quests controller for quest: %v", q_name)
		qb_controller := q.FormQuestBotContext(conf, q_name, cs, qs, db)
		q_controller := m.FormBotController(qb_controller, db)
		http.HandleFunc(fmt.Sprintf("/quest/%v", q_name), q_controller)
		result <- fmt.Sprintf("quest_%v", q_name)
	}

	if conf.Console.WebPort != "" && conf.Console.Key != "" {
		log.Printf("Will handling requests from /console")
		bc := cnsl.FormConsoleBotContext(conf, db, cs)
		cc := m.FormBotController(bc, db)
		http.HandleFunc("/console", cc)
		result <- "console"
	}

	if conf.Vote.DictUrl != "" {
		vdh, _ := v.NewVotingHandler(conf.Main.Database.ConnString, conf.Main.Database.Name)
		http.HandleFunc("/vote/autocomplete/name", func(w http.ResponseWriter, r *http.Request) {
			v.AutocompleteController(w, r, vdh, "name", []string{})
		})
		http.HandleFunc("/vote/autocomplete/city", func(w http.ResponseWriter, r *http.Request) {
			v.AutocompleteController(w, r, vdh, "city", conf.Vote.Cities)
		})
		http.HandleFunc("/vote/autocomplete/service", func(w http.ResponseWriter, r *http.Request) {
			v.AutocompleteController(w, r, vdh, "service", conf.Vote.Services)
		})
		http.HandleFunc("/vote/autocomplete/role", func(w http.ResponseWriter, r *http.Request) {
			v.AutocompleteController(w, r, vdh, "vote.voters.role", conf.Vote.Roles)
		})

		voteBot := v.FormVoteBotContext(conf, db)
		voteBotController := m.FormBotController(voteBot, db)
		log.Printf("Will handling requests for /vote")
		http.HandleFunc("/vote", voteBotController)
		result <- "vote"
	}

	if len(conf.Chats) > 0 {
		fs := http.FileServer(http.Dir("static"))
		http.Handle("/static/", http.StripPrefix("/static/", fs))

		for _, chat_conf := range conf.Chats {
			chatBotContext := chat.FormChatBotContext(chat_conf, db)
			chatBotController := m.FormBotController(chatBotContext, db)
			route := fmt.Sprintf("/bot/chat/%v", chat_conf.CompanyId)
			http.HandleFunc(route, chatBotController)
			log.Printf("I will serving message for chat bot at : [%v]", route)

			notifier := n.NewNotifier(conf.Main.CallbackAddr, chat_conf.Key, db)
			notifier.SetFrom(chat_conf.CompanyId)

			if chat_conf.AutoAnswer.Enable {
				go chat.Watch(db.Messages, notifier, chat_conf)
			}
			var salt string
			if chat_conf.UrlSalt != "" {
				salt = fmt.Sprintf("%v-%v", chat_conf.CompanyId, chat_conf.UrlSalt)
			} else {
				salt = chat_conf.CompanyId
			}
			webRoute := fmt.Sprintf("/web/chat/%v", salt)
			http.Handle(webRoute, chat.GetChatMainHandler(webRoute, notifier, db, chat_conf))

			sr := func(s string) string {
				return fmt.Sprintf("%v%v", webRoute, s)
			}
			http.Handle(sr("/send"), chat.GetChatSendHandler(sr("/send"), notifier, db, chat_conf, chat.NewChatStorage(db)))
			http.Handle(sr("/messages"), chat.GetChatMessagesHandler(sr("/messages"), notifier, db, chat_conf))
			http.Handle(sr("/messages_read"), chat.GetChatMessageReadHandler(sr("/messages_read"), notifier, db, chat_conf))
			http.Handle(sr("/contacts"), chat.GetChatContactsHandler(sr("/contacts"), notifier, db, chat_conf))
			http.Handle(sr("/contacts_change"), chat.GetChatContactsChangeHandler(sr("/contacts_change"), notifier, db, chat_conf))
			http.Handle(sr("/config"), chat.GetChatConfigHandler(sr("/config"), webRoute, db, chat_conf))
			http.Handle(sr("/delete_messages"), chat.GetChatDeleteMessagesHandler(sr("/delete_messages"), db, chat_conf))

			log.Printf("I will handling web requests for chat at : [%v]", webRoute)

			db.Users.AddOrUpdateUserObject(d.UserWrapper{UserName:chat_conf.User, Password:utils.PHash(chat_conf.Password), Role:users.MANAGER, UserId:chat_conf.User})
		}
	}

	server_address := fmt.Sprintf(":%v", conf.Main.Port)
	log.Printf("\nStart listen and serving at: %v\n", server_address)
	server := &http.Server{
		Addr: server_address,
	}
	result <- "listen"
	log.Fatal(server.ListenAndServe())
	return conf
}
