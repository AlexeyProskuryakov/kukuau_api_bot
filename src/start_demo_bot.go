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
	"net/http"
	"time"
	"errors"
	"flag"
)

func startAfter(check s.CheckFunc, what func()) {
	for {
		if message, ok := check(); ok {
			break
		}else {
			log.Printf("wait %v", message)
			time.Sleep(5 * time.Second)
		}
	}
	go what()
}

func GetAPIInstruments(params t.ApiParams) (t.TaxiInterface, t.AddressSupplier, error) {
	switch api_name := params.Name; api_name{
	case "infinity":
		return i.GetInfinityAPI(params), i.GetInfinityAddressSupplier(params), nil
	case "fake":
		return t.GetFakeInfinityAPI(params), i.GetInfinityAddressSupplier(params), nil
	}
	return nil, nil, errors.New("Not imply name of api")
}

func InsertTestUser(conf m.Configuration, user, pwd *string) {
	db := d.NewDbHandler(conf.Database.ConnString, conf.Database.Name)
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

func main() {
	conf := m.ReadConfig()
	var test = flag.Bool("test", false, "go in test use?")
	flag.Parse()

	d.DELETE_DB = *test
	log.Printf("%+v %+v", *test, d.DELETE_DB)
	if d.DELETE_DB {
		log.Println("!start at test mode!")
		conf.Database.Name = conf.Database.Name + "_test"
	}

	for _, taxi_conf := range conf.Taxis {
		external_api, external_address_supplier, err := GetAPIInstruments(taxi_conf.Api)

		if err != nil {
			log.Printf("Skip this taxi api [%+v]\nBecause: %v", taxi_conf.Api, err)
			continue
		}

		apiMixin := t.ExternalApiMixin{API: external_api}
		db := d.NewDbHandler(conf.Database.ConnString, conf.Database.Name)
		carsCache := t.NewCarsCache(external_api)
		notifier := n.NewNotifier(conf.Main.CallbackAddr, taxi_conf.Key)

		google_address_handler := t.NewGoogleAddressHandler(conf.Main.GoogleKey, taxi_conf.GeoOrbit, external_address_supplier)

		botContext := t.FormTaxiBotContext(&apiMixin, db, taxi_conf, google_address_handler)
		taxiContext := t.TaxiContext{API:external_api, DataBase:db, Cars:carsCache, Notifier:notifier}

		controller := m.FormBotController(botContext)

		http.HandleFunc(fmt.Sprintf("/taxi/%v", taxi_conf.Name), controller)
		startAfter(botContext.Check, func() {
			t.TaxiOrderWatch(&taxiContext, botContext)
		})

		http.HandleFunc(fmt.Sprintf("/taxi/%v/streets", taxi_conf.Name), func(w http.ResponseWriter, r *http.Request) {
			t.StreetsSearchController(w, r, google_address_handler)
		})
	}

	for _, shop_conf := range conf.Shops {
		db := d.NewDbHandler(conf.Database.ConnString, conf.Database.Name)
		bot_context := sh.FormShopCommands(db, &shop_conf)
		shop_controller := m.FormBotController(bot_context)
		http.HandleFunc(fmt.Sprintf("/shop/%v", shop_conf.Name), shop_controller)

	}

	user, pwd := "test", "test"
	InsertTestUser(conf, &user, &pwd)

	server_address := fmt.Sprintf(":%v", conf.Main.Port)
	log.Printf("\nStart listen and serving at: %v\n", server_address)
	server := &http.Server{
		Addr: server_address,
	}

	log.Fatal(server.ListenAndServe())
}
