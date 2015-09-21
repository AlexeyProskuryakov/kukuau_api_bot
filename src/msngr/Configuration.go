package msngr

import (
	"encoding/json"
	"io/ioutil"
	"log"
	ia "msngr/taxi"
	"os"
)

type taxi_config struct {
	Infinity ia.InfinityApiParams `json:"infinity"`
	DictUrl  string `json:"dict_url"`
	Key string `json:"key"`
	Collection string `json:"collection"`
}

type fake_taxi_config struct {
	Collection string `json:"collection"`
	Key string `json:"key"`
}

type shop_config struct {
	Key string `json:"key"`
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

	Taxi taxi_config `json:"taxi"`
	Shop shop_config `json:"shop"`
	FakeTaxi taxi_config `json:"fake_taxi"`

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
