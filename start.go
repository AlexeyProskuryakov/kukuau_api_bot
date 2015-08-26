package main

import (
	"msngr"
	"net/http"
)

func main() {
	f := msngr.BotControlHandler
	http.HandleFunc("/", f)
	http.ListenAndServe(":8080", nil)
}
