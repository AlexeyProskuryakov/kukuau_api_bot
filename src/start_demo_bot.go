package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	m "msngr"
	ia "msngr/infinity"
	"net/http"
	"time"
)

type config struct {
	Infinity ia.InfinityApiParams `json:"infinity"`
	Main     struct {
		Port         int    `json:"port"`
		CallbackAddr string `json:"callback_addr"`
		DictUrl      string `json:"dict_url"`
		Test         bool   `json:"test"`
		TaxiKey      string `json:"taxi_key"`
	} `json:"main"`
	Database struct {
		ConnString string `json:"connection_string"`
		Name       string `json:"name"`
	} `json:"database"`
}

func _check(e error) {
	if e != nil {
		panic(e)
	}
}

func read_config() config {
	cdata, _ := ioutil.ReadFile("config.json")
	log.Println("config data: ", string(cdata))
	conf := config{}
	err := json.Unmarshal(cdata, &conf)
	_check(err)
	return conf
}

func form_taxi_commands(im ia.InfinityMixin, oh m.OrderHandlerMixin) (map[string]m.RequestCommandProcessor, map[string]m.MessageCommandProcessor) {
	var TaxiRequestCommands = map[string]m.RequestCommandProcessor{
		"commands": m.TaxiCommandsHandler{},
	}

	var TaxiMessageCommands = map[string]m.MessageCommandProcessor{
		"information":     m.TaxiInformationHandler{},
		"new_order":       m.TaxiNewOrderHandler{InfinityMixin: im, OrderHandlerMixin: oh},
		"cancel_order":    m.TaxiCancelOrderHandler{InfinityMixin: im, OrderHandlerMixin: oh},
		"calculate_price": m.TaxiCalculatePriceHandler{InfinityMixin: im},
	}
	return TaxiRequestCommands, TaxiMessageCommands
}

func order_watch(ohm m.OrderHandlerMixin, im ia.InfinityMixin, carsCache *ia.CarsCache, n *m.Notifier) {
	for {
		api_orders := im.API.Orders()
		// log.Printf("OW api have %v orders", len(api_orders))
		for _, order := range api_orders {
			order_state := ohm.Orders.GetState(order.ID)
			// log.Printf("state of %+v is: %v\n", order, order_state)

			if order_state == -1 {
				log.Printf("order %+v is not present in system :(\n", order)
				continue
			}
			if order.State != order_state {
				log.Printf("state of %v will persist", order)
				ohm.Orders.SetState(order.ID, order.State, order)
				n.Notify(m.FormNotification(order.ID, order.State, ohm, carsCache))
			}
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func main() {
	conf := read_config()
	url := &m.DictUrl
	*url = conf.Main.DictUrl
	infApi := ia.GetInfinityAPI(conf.Infinity, conf.Main.Test)
	im := ia.InfinityMixin{API: infApi}

	orderHandler := m.NewOrderHandler(conf.Database.ConnString, conf.Database.Name)
	ohm := m.OrderHandlerMixin{Orders: orderHandler}

	taxi_controller_handler := m.FormBotControllerHandler(form_taxi_commands(im, ohm))
	shop_controller_handler := m.FormBotControllerHandler(m.ShopRequestCommands, m.ShopMessageCommands)

	http.HandleFunc("/taxi", taxi_controller_handler)
	http.HandleFunc("/shop", shop_controller_handler)

	realInfApi := ia.GetRealInfinityAPI(conf.Infinity)
	http.HandleFunc("/_streets", func(w http.ResponseWriter, r *http.Request) {
		ia.StreetsSearchHandler(w, r, realInfApi)
	})

	addr := fmt.Sprintf(":%v", conf.Main.Port)

	log.Printf("\nStart listen and serving at: %v\n", addr)
	serv := &http.Server{
		Addr: addr,
	}

	//////////////////////////////////////////////////////////////////
	n_taxi := m.NewNotifier(conf.Main.CallbackAddr, conf.Main.TaxiKey)
	carsCache := ia.NewCarsCache(realInfApi)
	go order_watch(ohm, im, carsCache, n_taxi)

	log.Fatal(serv.ListenAndServe())
}
