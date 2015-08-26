package main

import (
	"log"
	"msngr"
	"net/http"
)

func main() {
	f := msngr.BotControlHandler
	http.HandleFunc("/", f)
	addr := ":8080"

	log.Printf("\nStart listen and serving at: %v\nuse handler: %+q", addr, f)
	serv := &http.Server{
		Addr: addr,
	}
	log.Fatal(serv.ListenAndServe())

}
