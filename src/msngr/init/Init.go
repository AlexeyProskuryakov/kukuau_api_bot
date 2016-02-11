package init

import (
	"fmt"
	"log"
	"net/http"
	"errors"
	m "msngr"
	s "msngr/structs"
//tm "msngr/text_messages"
	d "msngr/db"
	c "msngr/configuration"
	i "msngr/taxi/infinity"
	cnsl "msngr/console"
	n "msngr/notify"
	rp "msngr/ruposts"
	sh "msngr/shop"
	q "msngr/quests"
	t "msngr/taxi"
	"msngr/taxi/geo"
	sedi "msngr/taxi/sedi"
)

func GetTaxiAPIInstruments(params c.TaxiApiParams) (t.TaxiInterface, t.AddressSupplier, error) {

	switch api_name := params.Name; api_name {
	case i.INFINITY:
		return i.GetInfinityAPI(params), i.GetInfinityAddressSupplier(params), nil
	case t.FAKE:
		return t.GetFakeAPI(params), i.GetInfinityAddressSupplier(params), nil

	case sedi.SEDI:
		sedi_api := sedi.NewSediAPI(params)
		return sedi_api, sedi_api, nil
	}
	return nil, nil, errors.New("Not imply name of api")
}

func get_address_instruments(c c.Configuration, taxi_name string, external_supplier t.AddressSupplier) (t.AddressHandler, t.AddressSupplier) {
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
		external_api, external_address_supplier, err := GetTaxiAPIInstruments(taxi_conf.Api)

		if err != nil {
			log.Printf("Skip this taxi api [%+v]\nBecause: %v", taxi_conf.Api, err)
			continue
		}

		apiMixin := t.ExternalApiMixin{API: external_api}

		carsCache := t.NewCarsCache(external_api)
		notifier := n.NewNotifier(conf.Main.CallbackAddr, taxi_conf.Key, db)

		address_handler, address_supplier := get_address_instruments(conf, taxi_name, external_address_supplier)

		botContext := t.FormTaxiBotContext(&apiMixin, db, taxi_conf, address_handler, carsCache)
		log.Printf("Was create bot context: %+v\n", botContext)
		taxiContext := t.TaxiContext{API: external_api, DataBase: db, Cars: carsCache, Notifier: notifier}
		controller := m.FormBotController(botContext, db)

		http.HandleFunc(fmt.Sprintf("/taxi/%v", taxi_conf.Name), controller)

		s.StartAfter(botContext.Check, func() {
			log.Printf("Will start order watcher for [%v]", botContext.Name)
			t.TaxiOrderWatch(&taxiContext, botContext)
		})

		http.HandleFunc(fmt.Sprintf("/taxi/%v/streets", taxi_conf.Name), func(w http.ResponseWriter, r *http.Request) {
			geo.StreetsSearchController(w, r, address_supplier)
		})

		http.HandleFunc(fmt.Sprintf("/taxi/%v/streets/ext", taxi_conf.Name), func(w http.ResponseWriter, r *http.Request) {
			geo.StreetsSearchController(w, r, external_address_supplier)
		})
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

	server_address := fmt.Sprintf(":%v", conf.Main.Port)
	log.Printf("\nStart listen and serving at: %v\n", server_address)
	server := &http.Server{
		Addr: server_address,
	}

	if conf.Console.WebPort != "" && conf.Console.Key != ""  {
		log.Printf("Will handling requests from /console")
		bc := cnsl.FormConsoleBotContext(conf, db,cs)
		cc := m.FormBotController(bc,db)
		http.HandleFunc("/console", cc)
		result <- "console"
	}

	result <- "listen"
	log.Fatal(server.ListenAndServe())
	return conf
}
