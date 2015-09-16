package msngr

import (
	"encoding/json"
	"io/ioutil"
	"log"
	ia "msngr/taxi"
)

type config struct {
	Infinity ia.InfinityApiParams `json:"infinity"`

	Main     struct {
				 Port         int    `json:"port"`
				 CallbackAddr string `json:"callback_addr"`
				 DictUrl      string `json:"dict_url"`
				 Test         bool   `json:"test"`
				 TaxiKey      string `json:"taxi_key"`
				 LoggingFile  string `json:"log_file"`
			 } `json:"main"`


	Database struct {
				 ConnString string `json:"connection_string"`
				 Name       string `json:"name"`
			 } `json:"database"`


}

func ReadConfig() config {
	cdata, _ := ioutil.ReadFile("config.json")
	log.Println("config data: ", string(cdata))
	conf := config{}
	err := json.Unmarshal(cdata, &conf)
	_check(err)
	return conf
}
