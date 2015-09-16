package main

import (
	"fmt"
	"log"
	m "msngr"
	t "msngr/taxi"
	sh "msngr/shop"
	d "msngr/db"
	n "msngr/notify"
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

	url := &t.DictUrl
	*url = conf.Main.DictUrl
	infApi := t.GetInfinityAPI(conf.Infinity, conf.Main.Test)

	im := t.InfinityMixin{API: infApi}

	db := d.NewDbHandler(conf.Database.ConnString, conf.Database.Name)
	db.Users.SetUserPassword("test", "123")

	taxi_controller := m.FormBotController(t.FormTaxiCommands(&im, *db))
	shop_controller := m.FormBotController(sh.FormShopCommands(*db))

	http.HandleFunc("/taxi", taxi_controller)
	http.HandleFunc("/shop", shop_controller)

	realInfApi := t.GetRealInfinityAPI(conf.Infinity)

	http.HandleFunc("/_streets", func(w http.ResponseWriter, r *http.Request) {
		t.StreetsSearchController(w, r, realInfApi)
	})

	n_taxi := n.NewNotifier(conf.Main.CallbackAddr, conf.Main.TaxiKey)

	//start watching
	go func() {
		for {
			if realInfApi.IsConnected(){
				break
			} else {
				time.Sleep(5*time.Second)
			}
		}

		carsCache := t.NewCarsCache(realInfApi)
		go t.TaxiOrderWatch(*db, im, carsCache, n_taxi)
	}()

	server_address := fmt.Sprintf(":%v", conf.Main.Port)
	log.Printf("\nStart listen and serving at: %v\n", server_address)
	server := &http.Server{
		Addr: server_address,
	}

	log.Fatal(server.ListenAndServe())
}
