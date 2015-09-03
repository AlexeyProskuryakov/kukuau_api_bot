package main

import (
	"fmt"
	// "gopkg.in/yaml.v2"
	"encoding/json"
	"io/ioutil"
	"log"
	m "msngr"
	ia "msngr/infinity"
	"net/http"
)

type config struct {
	Infinity ia.InfinityApiParams `json:"infinity"`
	Main     struct {
		Port         int    `json:"port"`
		CallbackAddr string `json:"callback_addr"`
	} `json:"main"`
}

func _check(e error) {
	if e != nil {
		panic(e)
	}
}

func read_config() config {
	cdata, _ := ioutil.ReadFile("config.json")
	conf := config{}
	err := json.Unmarshal(cdata, &conf)
	_check(err)
	return conf
}

func form_taxi_commands(im ia.InfinityMixin) (map[string]m.RequestCommandProcessor, map[string]m.MessageCommandProcessor) {
	var TaxiRequestCommands = map[string]m.RequestCommandProcessor{
		"commands": m.TaxiCommandsHandler{},
	}

	var TaxiMessageCommands = map[string]m.MessageCommandProcessor{
		"information":     m.TaxiInformationHandler{},
		"new_order":       m.TaxiNewOrderHandler{InfinityMixin: im},
		"cancel_order":    m.TaxiCancelOrderHandler{InfinityMixin: im},
		"calculate_price": m.TaxiCalculatePriceHandler{InfinityMixin: im},
	}
	return TaxiRequestCommands, TaxiMessageCommands
}

func main() {
	conf := read_config()
	infApi := ia.GetInfinityAPI(conf.Infinity)
	im := ia.InfinityMixin{API: infApi}

	taxi_controller_handler := m.FormBotControllerHandler(form_taxi_commands(im))
	shop_controller_handler := m.FormBotControllerHandler(m.ShopRequestCommands, m.ShopMessageCommands)

	http.HandleFunc("/taxi", taxi_controller_handler)
	http.HandleFunc("/shop", shop_controller_handler)

	http.HandleFunc("/_streets", func(w http.ResponseWriter, r *http.Request) {
		ia.StreetsSearchHandler(w, r, *infApi)
	})

	addr := fmt.Sprintf(":%v", conf.Main.Port)

	log.Printf("\nStart listen and serving at: %v\n", addr)
	serv := &http.Server{
		Addr: addr,
	}

	log.Fatal(serv.ListenAndServe())
}
