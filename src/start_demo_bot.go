package main

import (
	"log"
	m "msngr"
	ia "msngr/infinity"
	"net/http"
)

var (
	i_login       = "test1"
	i_password    = "test1"
	i_conn_string = "http://109.202.25.248:8080/WebAPITaxi/"
	i_host        = "109.202.25.248:8080"
)

func main() {
	taxi_controller_handler := m.FormBotControllerHandler(m.TaxiRequestCommands, m.TaxiMessageCommands)
	shop_controller_handler := m.FormBotControllerHandler(m.ShopRequestCommands, m.ShopMessageCommands)

	//prepare Infinity API
	var infApi = ia.InfinityAPI
	infApi.ConnString = i_conn_string
	infApi.Host = i_host
	status := infApi.Login(i_password, i_password)
	if status {
		log.Println("Установлено соединение с Infinity")
	}

	http.HandleFunc("/taxi", taxi_controller_handler)
	http.HandleFunc("/shop", shop_controller_handler)
	http.HandleFunc("/_streets", func(w http.ResponseWriter, r *http.Request) {
		ia.StreetsSearchHandler(w, r, infApi)
	})

	addr := ":8080"

	log.Printf("\nStart listen and serving at: %v\n", addr)
	serv := &http.Server{
		Addr: addr,
	}

	log.Fatal(serv.ListenAndServe())

}
