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
	"time"
)

func main() {
	conf := m.ReadConfig()

	url := &t.DictUrl
	*url = conf.Taxi.DictUrl
	realInfinity := t.GetRealInfinityAPI(conf.Taxi.Infinity)
	realMixin := t.InfinityMixin{API: realInfinity}

	fakeInfinity := t.GetFakeInfinityAPI()
	fakeMixin := t.InfinityMixin{API: fakeInfinity}

	db := d.NewDbHandler(conf.Database.ConnString, conf.Database.Name)
	fake_db := d.NewDbHandler(conf.Database.ConnString, fmt.Sprintf("%v_fake", conf.Database.Name))
	db.Users.SetUserPassword("test", "123")

	taxi_controller := m.FormBotController(t.FormTaxiCommands(&realMixin, db))
	fake_taxi_controller := m.FormBotController(t.FormTaxiCommands(&fakeMixin, fake_db))
	shop_controller := m.FormBotController(sh.FormShopCommands(db))

	http.HandleFunc("/taxi", taxi_controller)
	http.HandleFunc("/shop", shop_controller)
	http.HandleFunc("/taxi_fake", fake_taxi_controller)

	infinity := t.GetInfinity(conf.Taxi.Infinity)

	http.HandleFunc("/_streets", func(w http.ResponseWriter, r *http.Request) {
		t.StreetsSearchController(w, r, infinity)
	})

	notifier_real_taxi := n.NewNotifier(conf.Main.CallbackAddr, conf.Taxi.Key)
	notifier_fake_taxi := n.NewNotifier(conf.Main.CallbackAddr, conf.FakeTaxi.Key)


	//start watching
	go func() {
		for {
			if infinity.IsConnected() {
				break
			} else {
				time.Sleep(5 * time.Second)
			}
		}
		carsCache := t.NewCarsCache(infinity)
		go t.TaxiOrderWatch(db, &realMixin, carsCache, notifier_real_taxi)
		go t.TaxiOrderWatch(fake_db, &fakeMixin, carsCache, notifier_fake_taxi)
	}()


	server_address := fmt.Sprintf(":%v", conf.Main.Port)
	log.Printf("\nStart listen and serving at: %v\n", server_address)
	server := &http.Server{
		Addr: server_address,
	}

	log.Fatal(server.ListenAndServe())
}
