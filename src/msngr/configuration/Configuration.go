package configuration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	u "msngr/utils"
	"path"
)

type Transformation struct {
	Field     string `json:"field"`
	RegexCode string `json:"regex_code"`
	To        string `json:"to"`
}

type ApiData struct {
	Host               string `json:"host"`
	Login              string `json:"login"`
	Password           string `json:"password"`
	ConnectionsStrings []string `json:"connection_strings"`
	IdService          int64 `json:"id_service"`

	BearerToken        string `json:"bearer_token"`

	AppKey             string `json:"app_key"`
	ApiKey             string `json:"api_key"`
	City               string `json:"city"`
	Phone              string `json:"phone"`
	Name               string `json:"name"`
	SaleKeyword        string `json:"sale_kw"`
}

type TaxiApiParams struct {
	Name            string `json:"name"`
	Data            ApiData `json:"data"`
	GeoOrbit        TaxiGeoOrbit `json:"geo_orbit"`
	NotSendPrice    bool `json:"not_send_price"`
	Transformations []Transformation `json:"transformations"`
	Fake            struct {
				SendedStates []int `json:"sended_states"`
				SleepTime    int `json:"sleep_time"`
			} `json:"fake"`
}

func (api TaxiApiParams) String() string {
	return fmt.Sprintf("API %s\nAPI data: %+v\nFake?:%+v\nNotSendPrice?:%v\nGeoOrbit:%v", api.Name, api.Data, api.Fake, api.NotSendPrice, api.GeoOrbit)
}

func (api TaxiApiParams) GetHost() string {
	return api.Data.Host
}
func (api TaxiApiParams) GetConnectionStrings() []string {
	return api.Data.ConnectionsStrings
}
func (api TaxiApiParams) GetLogin() string {
	return api.Data.Login
}
func (api TaxiApiParams) GetPassword() string {
	return api.Data.Password
}
func (api TaxiApiParams) GetIdService() int64 {
	return api.Data.IdService
}
func (api TaxiApiParams) GetAPIData() ApiData {
	return api.Data
}
func (api TaxiApiParams) GetTransformations() []Transformation {
	return api.Transformations
}

type TaxiGeoOrbit struct {
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Radius float64 `json:"radius"`
}

type TaxiConfig struct {
	Api               TaxiApiParams `json:"api"`
	DictUrl           string `json:"dict_url"`
	Key               string `json:"key"`
	Name              string `json:"name"`
	Information       struct {
				  Phone string `json:"phone"`
				  Text  string `json:"text"`
			  } `json:"information"`
	Markups           *[]string `json:"markups,omitempty"`
	AvailableCommands map[string][]string `json:"available_commands"`
}

type ShopConfig struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Info string `json:"information"`
}

type QuestConfig struct {
	AcceptPhrase string `json:"accept_phrase"`
	RejectPhrase string `json:"reject_phrase"`
	ErrorPhrase  string `json:"error_phrase"`
	Info         string `json:"information"`
	WebPort      string `json:"web_port"`
	Key          string `json:"key"`
}

type ConsoleConfig struct {
	WebPort     string `json:"web_port"`
	Key         string `json:"key"`
	Information string `json:"information"`
}

type Configuration struct {
	Main    struct {
			Port         int    `json:"port"`
			CallbackAddr string `json:"callback_addr"`
			LoggingFile  string `json:"log_file"`
			GoogleKey    string `json:"google_key"`
			ElasticConn  string `json:"elastic_conn"`
			Database     struct {
					     ConnString string `json:"connection_string"`
					     Name       string `json:"name"`
				     } `json:"database"`
		} `json:"main"`
	Console ConsoleConfig  `json:"console"`
	Taxis   map[string]TaxiConfig `json:"taxis"`
	Shops   map[string]ShopConfig `json:"shops"`
	Quests  map[string]QuestConfig `json:"quests"`
	RuPost  struct {
			ExternalUrl string `json:"external_url"`
			WorkUrl     string `json:"work_url"`
		} `json:"ru_post"`
}

func (conf *Configuration) SetLogFile(fn string) {
	f, err := os.OpenFile(fn, os.O_RDWR | os.O_CREATE | os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	log.SetOutput(f)
	log.Println("Logging file is setted to %v", fn)
}

func UnmarshallConfig(cdata []byte) Configuration {
	log.Println("config data: ", string(cdata))
	conf := Configuration{}
	err := json.Unmarshal(cdata, &conf)
	if err != nil {
		log.Printf("error decoding configuration file", err)
		os.Exit(-1)
	}

	if conf.Main.LoggingFile != "" {
		f, err := os.OpenFile(conf.Main.LoggingFile, os.O_RDWR | os.O_CREATE | os.O_TRUNC, 0666)
		if err != nil {
			log.Fatalf("error opening log file: %v", err)
		}
		log.SetOutput(f)
		log.Println("Logging file is setted to %v", conf.Main.LoggingFile)
	}
	return conf
}

func ReadConfigInRecursive() Configuration {
	log.Printf("Path sep: %+v", os.PathSeparator)
	fn := u.FoundFile("config.json")
	if fn == nil {
		log.Printf("can not find config.json file :(")
		os.Exit(-1)
	}
	cdata, err := ioutil.ReadFile(*fn)
	if err != nil {
		log.Printf("error reading config %v", err)
		os.Exit(-1)
	}
	return UnmarshallConfig(cdata)
}

func ReadTestConfigInRecursive() Configuration {
	log.Printf("Path sep: %+v", os.PathSeparator)
	fn := u.FoundFile("config.test.json")
	if fn == nil {
		log.Printf("can not find config.json file :(")
		os.Exit(-1)
	}
	cdata, err := ioutil.ReadFile(*fn)
	if err != nil {
		log.Printf("error reading config %v", err)
		os.Exit(-1)
	}
	return UnmarshallConfig(cdata)
}

func ReadConfig() Configuration {
	//log.Printf("Path sep: %s", RuneToAscii(os.PathSeparator))
	//fn := u.FoundFile("config.json")
	//if fn == nil {
	//	log.Printf("can not find config.json file :(")
	//	os.Exit(-1)
	//}
	//cdata, err := ioutil.ReadFile(*fn)
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("ca not recognise current dir %v", err)
	}
	fn := path.Join(dir, "config.json")
	cdata, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Printf("error reading config %v", err)
		os.Exit(-1)
	}
	return UnmarshallConfig(cdata)
}