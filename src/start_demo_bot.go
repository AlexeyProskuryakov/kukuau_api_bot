package main

import (
	"fmt"
	"log"
	m "msngr"
	ia "msngr/infinity"
	"net/http"
)

func main() {
	conf := m.ReadConfig()

	url := &m.DictUrl
	*url = conf.Main.DictUrl
	infApi := ia.GetInfinityAPI(conf.Infinity, conf.Main.Test)
	im := ia.InfinityMixin{API: infApi}

	db := m.NewDbHandler(conf.Database.ConnString, conf.Database.Name)
	db.Users.SetUserPassword("test", "123")

	taxi_controller := m.FormBotController(m.FormTaxiCommands(im, *db))
	shop_controller := m.FormBotController(m.FormShopCommands(*db))

	http.HandleFunc("/taxi", taxi_controller)
	http.HandleFunc("/shop", shop_controller)

	realInfApi := ia.GetRealInfinityAPI(conf.Infinity)
	http.HandleFunc("/_streets", func(w http.ResponseWriter, r *http.Request) {
		ia.StreetsSearchController(w, r, realInfApi)
	})

	addr := fmt.Sprintf(":%v", conf.Main.Port)

	log.Printf("\nStart listen and serving at: %v\n", addr)
	serv := &http.Server{
		Addr: addr,
	}

	//////////////////////////////////////////////////////////////////
	n_taxi := m.NewNotifier(conf.Main.CallbackAddr, conf.Main.TaxiKey)
	carsCache := ia.NewCarsCache(realInfApi)
	go m.TaxiOrderWatch(*db, im, carsCache, n_taxi)

	log.Fatal(serv.ListenAndServe())
}
