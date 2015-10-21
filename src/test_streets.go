package main
import (
	"net/http"
	"fmt"
	"log"
	t "msngr/taxi"
	m "msngr"
	d "msngr/db"
	i "msngr/taxi/infinity"
	"encoding/json"
)

func start_serv(conf m.Configuration, threaded bool) (string, *t.GoogleAddressHandler) {
	taxi_conf := conf.Taxis["fake"]

	address_supplier := t.NewGoogleAddressHandler(conf.Main.GoogleKey, taxi_conf.GeoOrbit, i.GetInfinityAddressSupplier(taxi_conf.Api))

	streets_address := fmt.Sprintf("/taxi/%v/streets", taxi_conf.Name)

	http.HandleFunc(streets_address, func(w http.ResponseWriter, r *http.Request) {
		t.StreetsSearchController(w, r, address_supplier)
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


func test_serv() {
	conf := m.ReadConfig()

	d.DELETE_DB = true
	if d.DELETE_DB {
		log.Println("!start at test mode!")
		conf.Database.Name = conf.Database.Name + "_test"
	}

	start_serv(conf, false)
}

func test_all() {
	conf := m.ReadConfig()

	d.DELETE_DB = true
	if d.DELETE_DB {
		log.Println("!start at test mode!")
		conf.Database.Name = conf.Database.Name + "_test"
	}

	test_url, address_supplier := start_serv(conf, true)
	taxi_conf := conf.Taxis["fake"]

	last_result := t.DictItem{}
	//test is next:
	for _, q := range []string{"Никола"} {
		log.Printf(">>> %v", q)
		body, err := t.GET(test_url, &map[string]string{"q":q})
		if body != nil {
			log.Printf("<<< %q", string(*body))
			var results []t.DictItem
			err = json.Unmarshal(*body, &results)
			//			log.Printf("err: %v \nunmarshaled:%+v", err, results)
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

	street_id, err := address_supplier.GetStreetId(last_result.Key)
	log.Printf("address err?: %v\n street_id: %#v", err, street_id)


}
func main() {
	test_all()
}
