package main

import (
	"encoding/json"
	"gopkg.in/mgo.v2"
	"io/ioutil"
	"log"
)

type config struct {
	Main struct {
		Port         int    `json:"port"`
		CallbackAddr string `json:"callback_addr"`
		DictUrl      string `json:"dict_url"`
		Test         bool   `json:"test"`
		TaxiKey      string `json:"taxi_key"`
	} `json:"main"`
	Database struct {
		ConnString string `json:"connection_string"`
		Name       string `json:"name"`
	} `json:"database"`
}

func read_config() config {
	cdata, _ := ioutil.ReadFile("config.json")
	log.Println("config data: ", string(cdata))
	conf := config{}
	err := json.Unmarshal(cdata, &conf)
	if err != nil {
		panic(err)
	}
	return conf
}

func main() {
	conf := read_config()

	session, err := mgo.Dial(conf.Database.ConnString)
	if err != nil {
		panic(err)
	}

	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	err = session.DB(conf.Database.Name).DropDatabase()
	if err != nil {
		panic(err)
	}
}
