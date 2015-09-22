package msngr

import (
	"encoding/json"
	"io/ioutil"
	"log"
	ia "msngr/taxi"
	"os"
)


type taxi_config struct {
	Api 	ia.ApiParams `json:"api"`
	DictUrl  string `json:"dict_url"`
	Key      string `json:"key"`
	Name     string `json:"name"`
}

type shop_config struct {
	Key string `json:"key"`
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

	Taxis    []taxi_config `json:"taxis"`
	Shops     []shop_config `json:"shops"`
}


func ReadConfig() config {
	cdata, _ := ioutil.ReadFile("config.json")
	log.Println("config data: ", string(cdata))
	conf := config{}
	err := json.Unmarshal(cdata, &conf)
	_check(err)

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
