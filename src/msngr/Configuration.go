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

type Configuration struct {
	Main     struct {
				 Port         int    `json:"port"`
				 CallbackAddr string `json:"callback_addr"`
				 LoggingFile  string `json:"log_file"`
				 GoogleKey    string `json:"google_key"`
			 } `json:"main"`

	Database struct {
				 ConnString string `json:"connection_string"`
				 Name       string `json:"name"`
			 } `json:"database"`

	Taxis    map[string]t.TaxiConfig `json:"taxis"`
	Shops    map[string]s.ShopConfig `json:"shops"`
}


func ReadConfig() Configuration {
	cdata, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Printf("error reading config")
		os.Exit(-1)
	}
	log.Println("config data: ", string(cdata))
	conf := Configuration{}
	err = json.Unmarshal(cdata, &conf)
	if err != nil {
		log.Printf("error decoding configuration file", err)
		os.Exit(-1)
	}

	if conf.Main.LoggingFile != "" {
		f, err := os.OpenFile("demo_bot.log", os.O_RDWR | os.O_CREATE | os.O_TRUNC, 0666)
		if err != nil {
			log.Fatalf("error opening log file: %v", err)
		}

		log.SetOutput(f)
		log.Println("Logging file is setted here...")
	}

	return conf
}
