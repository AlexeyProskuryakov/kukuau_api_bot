package msngr

import (
	"encoding/json"
	"io/ioutil"
	"log"
	t "msngr/taxi"
	s "msngr/shop"
	"os"
)




type shop_config struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type config struct {
	Main     struct {
				 Port         int    `json:"port"`
				 CallbackAddr string `json:"callback_addr"`
				 LoggingFile  string `json:"log_file"`
			 } `json:"main"`

	Database struct {
				 ConnString string `json:"connection_string"`
				 Name       string `json:"name"`
			 } `json:"database"`

	Taxis    []t.TaxiConfig `json:"taxis"`
	Shops    []s.ShopConfig `json:"shops"`
}


func ReadConfig() config {
	cdata, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Printf("error reading config")
		os.Exit(-1)
	}
	log.Println("config data: ", string(cdata))
	conf := config{}
	err = json.Unmarshal(cdata, &conf)
	if err != nil {
		log.Printf("error decoding configuration file")
		os.Exit(-1)
	}

	if conf.Main.LoggingFile != "" {
		f, err := os.OpenFile("demo_bot.log", os.O_RDWR | os.O_CREATE | os.O_TRUNC, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()

		log.SetOutput(f)
		log.Println("This is a test log entry")
	}

	return conf
}
