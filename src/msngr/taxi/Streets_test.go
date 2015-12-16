package taxi


import (
	"net/http"
	"fmt"
	"log"
	"encoding/json"

	c "msngr/configuration"
	d "msngr/db"
	i "msngr/taxi/infinity"
	u "msngr/utils"
	"testing"
	"os"
)

func start_serv(conf c.Configuration, threaded bool) (string, *GoogleAddressHandler) {
	taxi_conf := conf.Taxis["fake"]

	address_supplier := NewGoogleAddressHandler(conf.Main.GoogleKey, taxi_conf.GeoOrbit, i.GetInfinityAddressSupplier(taxi_conf.Api))

	streets_address := fmt.Sprintf("/taxi/%v/streets", taxi_conf.Name)

	http.HandleFunc(streets_address, func(w http.ResponseWriter, r *http.Request) {
		StreetsSearchController(w, r, address_supplier)
	})

	server_address := fmt.Sprintf(":%v", conf.Main.Port)

	server := &http.Server{
		Addr: server_address,
	}
	test_url := "http://localhost" + server_address + streets_address
	log.Printf("start server... send tests to: %v", test_url)

	if threaded {
		go server.ListenAndServe()
	} else {
		server.ListenAndServe()
	}
	return test_url, address_supplier
}


func TestStreetsGlobal(t *testing.T) {
	conf := c.ReadConfig()

	d.DELETE_DB = true
	if d.DELETE_DB {
		log.Println("!start at test mode!")
		conf.Main.Database.Name = conf.Main.Database.Name + "_test"
	}

	test_url, address_supplier := start_serv(conf, true)
	taxi_conf := conf.Taxis["fake"]

	last_result := DictItem{}
	//test is next:
	for _, q := range []string{"Никола"} {
		body, err := u.GET(test_url, &map[string]string{"q":q})
		if body != nil {
			var results []DictItem
			err = json.Unmarshal(*body, &results)
			for _, val := range results {
				log.Printf("KEY: %v, TITLE: %v, SUBTITLE: %v", val.Key, val.Title, val.SubTitle)
			}
			last_result = results[0]
		}
		if err != nil {
			log.Printf("!!!ERRRR!!! %+v", err)
		}
	}
	log.Printf("LAST RESULT: %#v", last_result)

	is_here := address_supplier.IsHere(last_result.Key)
	log.Print("is here: ", is_here)

	external_suppier := i.GetInfinityAddressSupplier(taxi_conf.Api)
	address_supplier.ExternalAddressSupplier = external_suppier

	street_id, err := address_supplier.GetExternalInfo(last_result.Key, last_result.Title)
	log.Printf("address err?: %v\n street_id: %#v", err, street_id)
	os.Exit(0)
}
