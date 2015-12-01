package main

import (
	"fmt"
	"log"
	m "msngr"
	t "msngr/taxi"
	i "msngr/taxi/infinity"
	sh "msngr/shop"
	d "msngr/db"
	n "msngr/notify"
	s "msngr/structs"
	cnsl "msngr/console"
	rp "msngr/ruposts"
	c "msngr/configuration"
	"net/http"
	"time"
	"errors"
	"flag"

)


func GetTaxiAPIInstruments(params c.TaxiApiParams) (t.TaxiInterface, t.AddressSupplier, error) {
	switch api_name := params.Name; api_name{
	case "infinity":
		return i.GetInfinityAPI(params), i.GetInfinityAddressSupplier(params), nil
	case "fake":
		return t.GetFakeAPI(params), i.GetInfinityAddressSupplier(params), nil
	}
	return nil, nil, errors.New("Not imply name of api")
}

func InsertTestUser(db *d.DbHandlerMixin, user, pwd *string) {
	err := db.Users.SetUserPassword(user, pwd)
	if err != nil {
		go func() {
			for err == nil {
				time.Sleep(1 * time.Second)
				err = db.Users.SetUserPassword(user, pwd)
				log.Printf("trying add user for test shops... now we have err:%+v", err)
			}
		}()
	}
}

func get_address_instruments(c c.Configuration, taxi_name string, external_supplier t.AddressSupplier) (t.AddressHandler, t.AddressSupplier) {
	own := t.NewOwnAddressHandler(c.Main.ElasticConn, c.Taxis[taxi_name].GeoOrbit, external_supplier)
	if own == nil {

		google := t.NewGoogleAddressHandler(c.Main.GoogleKey, c.Taxis[taxi_name].GeoOrbit, external_supplier)
		if google == nil {
			log.Printf("Will use external address receiver")
			return nil, external_supplier
		}
		log.Printf("Will use google addresses")
		return google, google
	}
	log.Printf("Will use own addresses")
	return own,own
}
func main() {
	conf := c.ReadConfig()
	var test = flag.Bool("test", false, "go in test use?")
	flag.Parse()

	d.DELETE_DB = *test
	m.DEBUG = *test

	log.Printf("configuration for db:\nconnection string: %+v\ndatabase name: %+v", conf.Main.Database.ConnString, conf.Main.Database.Name)
	db := d.NewDbHandler(conf.Main.Database.ConnString, conf.Main.Database.Name)

	log.Printf("Is delete DB? [%+v] Is debug? [%v]", d.DELETE_DB, m.DEBUG)
	if d.DELETE_DB {
		log.Println("!!!!!!!!!!start at test mode!!!!!!!!!!!!!")
		conf.Main.Database.Name = conf.Main.Database.Name + "_test"
		db.Session.DB(conf.Main.Database.Name).DropDatabase()

	}

	for taxi_name, taxi_conf := range conf.Taxis {
		log.Printf("taxi api configuration for %+v: \nconnection str: %+v\nhost: %+v\nid_service: %+v\nlogin: %+v\npassword: %+v", taxi_conf.Name, taxi_conf.Api.GetConnectionStrings(), taxi_conf.Api.GetHost(), taxi_conf.Api.GetIdService(), taxi_conf.Api.GetLogin(), taxi_conf.Api.GetPassword())
		external_api, external_address_supplier, err := GetTaxiAPIInstruments(taxi_conf.Api)

		if err != nil {
			log.Printf("Skip this taxi api [%+v]\nBecause: %v", taxi_conf.Api, err)
			continue
		}

		apiMixin := t.ExternalApiMixin{API: external_api}

		carsCache := t.NewCarsCache(external_api)
		notifier := n.NewNotifier(conf.Main.CallbackAddr, taxi_conf.Key)

		address_handler, address_supplier := get_address_instruments(conf, taxi_name, external_address_supplier)

		botContext := t.FormTaxiBotContext(&apiMixin, db, taxi_conf, address_handler, carsCache)
		log.Printf("Was create bot context: %+v", botContext)
		taxiContext := t.TaxiContext{API:external_api, DataBase:db, Cars:carsCache, Notifier:notifier}
		controller := m.FormBotController(botContext)

		http.HandleFunc(fmt.Sprintf("/taxi/%v", taxi_conf.Name), controller)

		s.StartAfter(botContext.Check, func() {
			log.Printf("Will start order watcher for [%v]", botContext.Name)
			t.TaxiOrderWatch(&taxiContext, botContext)
		})

		http.HandleFunc(fmt.Sprintf("/taxi/%v/streets", taxi_conf.Name), func(w http.ResponseWriter, r *http.Request) {
			//			t.StreetsSearchController(w, r, google_address_handler)
			t.StreetsSearchController(w, r, address_supplier)
		})
	}

	for _, shop_conf := range conf.Shops {
		bot_context := sh.FormShopCommands(db, &shop_conf)
		shop_controller := m.FormBotController(bot_context)
		http.HandleFunc(fmt.Sprintf("/shop/%v", shop_conf.Name), shop_controller)

	}

	user, pwd := "test", "test"
	InsertTestUser(db, &user, &pwd)

	if conf.RuPost.WorkUrl != "" {
		log.Printf("will start ru post controller at: %v and will send requests to: %v", conf.RuPost.WorkUrl, conf.RuPost.ExternalUrl)
		rp_bot_context := rp.FormRPBotContext(conf)
		rp_controller := m.FormBotController(rp_bot_context)
		http.HandleFunc(conf.RuPost.WorkUrl, rp_controller)
	}

	server_address := fmt.Sprintf(":%v", conf.Main.Port)
	log.Printf("\nStart listen and serving at: %v\n", server_address)
	server := &http.Server{
		Addr: server_address,
	}

	go cnsl.Run(conf, db)
	log.Fatal(server.ListenAndServe())
}
