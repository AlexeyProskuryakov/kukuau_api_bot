package main

import (
	"log"
	m "msngr"
	"net/http"
)

func main() {
	taxi_controller_handler := m.FormBotControllerHandler(m.TaxiRequestCommands, m.TaxiMessageCommands)
	shop_controller_handler := m.FormBotControllerHandler(m.ShopRequestCommands, m.ShopMessageCommands)

	http.HandleFunc("/taxi", taxi_controller_handler)
	http.HandleFunc("/shop", shop_controller_handler)

	addr := ":8080"

	log.Printf("\nStart listen and serving at: %v\n", addr)
	serv := &http.Server{
		Addr: addr,
	}

	log.Fatal(serv.ListenAndServe())

}
