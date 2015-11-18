package main
import (
	"net/http"
	"log"
	"io/ioutil"
)

func main() {
	http.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("error at reading: %q \n", err)
		}
		log.Printf("RECEIVED: \n%s\n", string(body))

	})

	server := &http.Server{
		Addr: ":9876",
	}

	server.ListenAndServe()
}