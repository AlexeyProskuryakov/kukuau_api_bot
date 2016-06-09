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
	"msngr/coffee"

	"msngr/utils"
	"msngr/users"
	"msngr/web"
	"msngr/autoanswers"
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
	log.Printf("configuration for conf db:\nconnection string: %+v\ndatabase name: %+v", conf.Main.ConfigDatabase.ConnString, conf.Main.ConfigDatabase.Name)
	db.Users.AddOrUpdateUserObject(d.UserData{
		UserId:"alesha",
		UserName:"alesha",
		Password:utils.PHash("sederfes100500"),
		Role:users.MANAGER,
		BelongsTo:"klichat",
	})
	configStorage := d.NewConfigurationStorage(conf.Main.ConfigDatabase)
	questStorage := q.NewQuestStorage(conf.Main.Database.ConnString, conf.Main.Database.Name)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.Handle("/", web.NewSessionAuthorisationHandler(db))

	for taxi_name, taxi_conf := range conf.Taxis {
		log.Printf("taxi api configuration for %+v:\n%v", taxi_conf.Name, taxi_conf.Api)
		external_api, external_address_supplier, err := GetTaxiAPIInstruments(taxi_conf.Api, taxi_name)

		if err != nil {
			log.Printf("Skip this taxi api [%+v]\nBecause: %v", taxi_conf.Api, err)
			continue
		}

		apiMixin := t.ExternalApiMixin{API: external_api}

		carsCache := t.NewCarsCache(external_api)
		notifier := n.NewNotifier(conf.Main.CallbackAddr, taxi_conf.Chat.Key, db)

		address_handler, address_supplier := GetAddressInstruments(conf, taxi_name, external_address_supplier)

		botContext := t.FormTaxiBotContext(&apiMixin, db, configStorage, taxi_conf, address_handler, carsCache)
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

		configStorage.SetChatConfig(taxi_conf.Chat, false)

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

	for _, qConf := range conf.Quests {
		q_name := qConf.Chat.CompanyId
		log.Printf("Will handling quests controller for quest: %v", q_name)
		qb_controller := q.FormQuestBotContext(conf, qConf, questStorage, db, configStorage)
		q_controller := m.FormBotController(qb_controller, db)
		http.HandleFunc(fmt.Sprintf("/quest/%v", q_name), q_controller)

		configStorage.SetChatConfig(qConf.Chat, false)
		questStorage.SetMessageConfiguration(q.QuestMessageConfiguration{CompanyId:qConf.Chat.CompanyId}, false)

		result <- fmt.Sprintf("quest_%v", q_name)
	}
	if conf.Console.WebPort != "" && conf.Console.Chat.Key != "" {
		log.Printf("Will handling requests from /console at %v", conf.Console.WebPort)
		bc := cnsl.FormConsoleBotContext(conf, db, configStorage)
		cc := m.FormBotController(bc, db)
		http.HandleFunc("/console", cc)

		web.DefaultUrlMap.AddAccessory("klichat", "/klichat")

		configStorage.SetChatConfig(conf.Console.Chat, false)

		result <- "console"
	}

	if conf.Vote.DictUrl != "" {
		vdh, _ := v.NewVotingHandler(conf.Main.Database.ConnString, conf.Main.Database.Name)
		http.HandleFunc("/autocomplete/vote/name", func(w http.ResponseWriter, r *http.Request) {
			v.AutocompleteController(w, r, vdh, "name", []string{})
		})
		http.HandleFunc("/autocomplete/vote/city", func(w http.ResponseWriter, r *http.Request) {
			v.AutocompleteController(w, r, vdh, "city", conf.Vote.Cities)
		})
		http.HandleFunc("/autocomplete/vote/service", func(w http.ResponseWriter, r *http.Request) {
			v.AutocompleteController(w, r, vdh, "service", conf.Vote.Services)
		})
		http.HandleFunc("/autocomplete/vote/role", func(w http.ResponseWriter, r *http.Request) {
			v.AutocompleteController(w, r, vdh, "vote.voters.role", conf.Vote.Roles)
		})

		voteBot := v.FormVoteBotContext(conf, db)
		voteBotController := m.FormBotController(voteBot, db)
		log.Printf("Will handling requests for /vote")
		http.HandleFunc("/vote", voteBotController)
		result <- "vote"
	}

	if len(conf.Coffee) > 0 {
		for _, coffee_conf := range conf.Coffee {
			c_store := coffee.NewCoffeeConfigHandler(db)
			c_store.LoadFromConfig(coffee_conf)
			coffeeHouseConfiguration := coffee.NewCHCFromConfig(coffee_conf)

			cbc := coffee.FormBotCoffeeContext(coffee_conf, db, coffeeHouseConfiguration, configStorage)
			cntrl := m.FormBotController(cbc, db)
			route := fmt.Sprintf("/bot/coffee/%v", coffee_conf.Chat.CompanyId)
			http.HandleFunc(route, cntrl)
			log.Printf("will handling bot messages for coffee %v at %v", coffee_conf.Chat.CompanyId, route)

			http.HandleFunc(fmt.Sprintf("/autocomplete/coffee/%v/drink", coffee_conf.Chat.CompanyId), func(w http.ResponseWriter, r *http.Request) {
				coffee.AutocompleteController(w, r, c_store, "drinks", coffee_conf.Chat.CompanyId)
			})

			http.HandleFunc(fmt.Sprintf("/autocomplete/coffee/%v/volume", coffee_conf.Chat.CompanyId), func(w http.ResponseWriter, r *http.Request) {
				coffee.AutocompleteController(w, r, c_store, "volumes", coffee_conf.Chat.CompanyId)
			})

			http.HandleFunc(fmt.Sprintf("/autocomplete/coffee/%v/bake", coffee_conf.Chat.CompanyId), func(w http.ResponseWriter, r *http.Request) {
				coffee.AutocompleteController(w, r, c_store, "bakes", coffee_conf.Chat.CompanyId)
			})

			http.HandleFunc(fmt.Sprintf("/autocomplete/coffee/%v/additive", coffee_conf.Chat.CompanyId), func(w http.ResponseWriter, r *http.Request) {
				coffee.AutocompleteController(w, r, c_store, "additives", coffee_conf.Chat.CompanyId)
			})
			var salt string
			if coffee_conf.Chat.UrlSalt != "" {
				salt = fmt.Sprintf("%v-%v", coffee_conf.Chat.CompanyId, coffee_conf.Chat.UrlSalt)
			} else {
				salt = coffee_conf.Chat.CompanyId
			}

			notifier := n.NewNotifier(conf.Main.CallbackAddr, coffee_conf.Chat.Key, db)
			notifier.SetFrom(coffee_conf.Chat.CompanyId)

			webRoute := fmt.Sprintf("/web/coffee/%v", salt)
			http.Handle(webRoute, coffee.GetChatMainHandler(webRoute, notifier, db, coffee_conf.Chat))
			web.DefaultUrlMap.AddAccessory(coffee_conf.Chat.CompanyId, webRoute)

			sr := func(s string) string {
				return fmt.Sprintf("%v%v", webRoute, s)
			}
			http.Handle(sr("/send"), chat.GetChatSendHandler(sr("/send"), notifier, db, coffee_conf.Chat, chat.NewChatStorage(db)))
			http.Handle(sr("/unread_messages"), chat.GetChatUnreadMessagesHandler(sr("/unread_messages"), notifier, db, coffee_conf.Chat))
			http.Handle(sr("/messages_read"), chat.GetChatMessageReadHandler(sr("/messages_read"), notifier, db, coffee_conf.Chat))
			http.Handle(sr("/contacts_change"), chat.GetChatContactsChangeHandler(sr("/contacts_change"), notifier, db, coffee_conf.Chat))
			http.Handle(sr("/delete_messages"), chat.GetChatDeleteMessagesHandler(sr("/delete_messages"), db, coffee_conf.Chat))

			http.Handle(sr("/config"), coffee.GetChatConfigHandler(sr("/config"), webRoute, db, coffee_conf.Chat))
			http.Handle(sr("/contacts"), coffee.GetChatContactsHandler(sr("/contacts"), notifier, db, coffee_conf.Chat))
			http.Handle(sr("/message_function"), coffee.GetMessageAdditionalFunctionsHandler(sr("/message_function"), notifier, db, coffee_conf.Chat, coffeeHouseConfiguration))
			//http.Handle(sr("/order_page"), coffee.GetOrdersPageFunctionHandler(sr("/order_page"), webRoute, db, coffee_conf.Chat, coffee_conf.Chat.CompanyId))
			//http.Handle(sr("/order_page_supply"), coffee.GetOrdersPageSupplierFunctionHandler(sr("/order_page_supply"), webRoute, db, coffee_conf.Chat, coffee_conf.Chat.CompanyId))
			http.Handle(sr("/logout"), coffee.GetChatLogoutHandler(sr("/logout"), webRoute, db, coffee_conf.Chat))

			log.Printf("I will handling web requests for coffee %v at : [%v]", coffee_conf.Chat.CompanyId, webRoute)
			db.Users.AddOrUpdateUserObject(d.UserData{
				UserId:coffee_conf.Chat.User,
				UserName:coffee_conf.Chat.User,
				ShowedName:coffee_conf.Chat.Name,
				Password:utils.PHash(coffee_conf.Chat.Password),
				BelongsTo:coffee_conf.Chat.CompanyId,
				ReadRights:[]string{coffee_conf.Chat.CompanyId},
				WriteRights:[]string{coffee_conf.Chat.CompanyId},
			})

			configStorage.SetChatConfig(coffee_conf.Chat, false)
		}
		result <- "coffee"
	}
	if len(conf.Chats) > 0 {
		for _, chat_conf := range conf.Chats {

			chatBotContext := chat.FormChatBotContext(db, configStorage, chat_conf.CompanyId)
			chatBotController := m.FormBotController(chatBotContext, db)
			route := fmt.Sprintf("/bot/chat/%v", chat_conf.CompanyId)
			http.HandleFunc(route, chatBotController)
			log.Printf("I will serving message for chat bot at : [%v]", route)

			notifier := n.NewNotifier(conf.Main.CallbackAddr, chat_conf.Key, db)
			notifier.SetFrom(chat_conf.CompanyId)

			var salt string
			if chat_conf.UrlSalt != "" {
				salt = fmt.Sprintf("%v-%v", chat_conf.CompanyId, chat_conf.UrlSalt)
			} else {
				salt = chat_conf.CompanyId
			}
			webRoute := fmt.Sprintf("/web/chat/%v", salt)
			http.Handle(webRoute, chat.GetChatMainHandler(webRoute, notifier, db, chat_conf))
			web.DefaultUrlMap.AddAccessory(chat_conf.CompanyId, webRoute)

			sr := func(s string) string {
				return fmt.Sprintf("%v%v", webRoute, s)
			}
			http.Handle(sr("/send"), chat.GetChatSendHandler(sr("/send"), notifier, db, chat_conf, chat.NewChatStorage(db)))
			http.Handle(sr("/unread_messages"), chat.GetChatUnreadMessagesHandler(sr("/unread_messages"), notifier, db, chat_conf))
			http.Handle(sr("/messages_read"), chat.GetChatMessageReadHandler(sr("/messages_read"), notifier, db, chat_conf))
			http.Handle(sr("/contacts"), chat.GetChatContactsHandler(sr("/contacts"), notifier, db, chat_conf))
			http.Handle(sr("/contacts_change"), chat.GetChatContactsChangeHandler(sr("/contacts_change"), notifier, db, chat_conf))
			http.Handle(sr("/config"), chat.GetChatConfigHandler(sr("/config"), webRoute, db, chat_conf))
			http.Handle(sr("/delete_messages"), chat.GetChatDeleteMessagesHandler(sr("/delete_messages"), db, chat_conf))
			http.Handle(sr("/logout"), chat.GetChatLogoutHandler(sr("/logout"), webRoute, db, chat_conf))

			log.Printf("I will handling web requests for chat at : [%v]", webRoute)

			db.Users.AddOrUpdateUserObject(d.UserData{
				UserName:chat_conf.User,
				Password:utils.PHash(chat_conf.Password),
				UserId:chat_conf.User,
				BelongsTo:chat_conf.CompanyId,
				ReadRights:[]string{chat_conf.CompanyId},
				WriteRights:[]string{chat_conf.CompanyId},
			})

			configStorage.SetChatConfig(chat_conf, false)
		}
		result <- "chat"
	}

	go n.WatchUnreadMessages(db, configStorage, conf.Main.CallbackAddrMembers)
	go autoanswers.WatchNotAnsweredMessages(db, configStorage, conf.Main.CallbackAddr)

	server_address := fmt.Sprintf(":%v", conf.Main.Port)
	log.Printf("\nStart listen and serving at: %v\n", server_address)
	server := &http.Server{
		Addr: server_address,
	}
	result <- "listen"
	log.Fatal(server.ListenAndServe())
	return conf
}
