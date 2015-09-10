package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	m "msngr"
	"net/http"
)

func main() {
	addr := ":9876"

	http.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("ths < method: %+v", r.Method)

		log.Printf("ths < headers %+v", r.Header)

		body, err := ioutil.ReadAll(r.Body)
		log.Printf("ths < data: %+v", string(body))

		var pkg m.OutPkg
		err = json.Unmarshal(body, &pkg)
		if err != nil {
			log.Printf("ths err: %+v", err)
		}
		log.Printf("ths data parsed: %+v", pkg)

	})

	serv := &http.Server{
		Addr: addr,
	}

	log.Fatal(serv.ListenAndServe())
}
