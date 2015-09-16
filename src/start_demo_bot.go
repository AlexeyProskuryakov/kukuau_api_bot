package main

import (
	"fmt"
	"log"
	m "msngr"
	taxi "msngr/taxi"
	"net/http"
	"os"

	"time"
)

func main() {
	conf := m.ReadConfig()
	if conf.Main.LoggingFile != "" {
		f, err := os.OpenFile("demo_bot.log", os.O_RDWR | os.O_CREATE | os.O_TRUNC, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()

		log.SetOutput(f)
		log.Println("This is a test log entry")
	}

	url := &m.DictUrl
	*url = conf.Main.DictUrl
	infApi := taxi.GetInfinityAPI(conf.Infinity, conf.Main.Test)

	im := taxi.InfinityMixin{API: infApi}

	db := m.NewDbHandler(conf.Database.ConnString, conf.Database.Name)
	db.Users.SetUserPassword("test", "123")

	taxi_controller := m.FormBotController(m.FormTaxiCommands(&im, *db))
	shop_controller := m.FormBotController(m.FormShopCommands(*db))

	http.HandleFunc("/taxi", taxi_controller)
	http.HandleFunc("/shop", shop_controller)

	realInfApi := taxi.GetRealInfinityAPI(conf.Infinity)

	http.HandleFunc("/_streets", func(w http.ResponseWriter, r *http.Request) {
		taxi.StreetsSearchController(w, r, realInfApi)
	})

	//////////////////////////////////////////////////////////////////
	go func() {
		for {
			if realInfApi.IsConnected(){
				break
			} else {
				time.Sleep(5*time.Second)
			}
		}
		n_taxi := m.NewNotifier(conf.Main.CallbackAddr, conf.Main.TaxiKey)
		carsCache := taxi.NewCarsCache(realInfApi)
		go m.TaxiOrderWatch(*db, im, carsCache, n_taxi)
	}()



	addr := fmt.Sprintf(":%v", conf.Main.Port)

	log.Printf("\nStart listen and serving at: %v\n", addr)
	serv := &http.Server{
		Addr: addr,
	}

	log.Fatal(serv.ListenAndServe())
}
