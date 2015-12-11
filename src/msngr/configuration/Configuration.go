package configuration
import (
	"log"
	"os"
	"io/ioutil"
	"encoding/json"
	u "msngr/utils"
	"fmt"
)
type ApiData struct {
	Host               string `json:"host"`
	Login              string `json:"login"`
	Password           string `json:"password"`
	ConnectionsStrings []string `json:"connection_strings"`
	IdService          string `json:"id_service"`

	BearerToken        string `json:"bearer_token"`

	AppKey             string `json:"app_key"`
	ApiKey             string `json:"api_key"`
	City               string `json:"city"`
	Phone              string `json:"phone"`
	Name               string `json:"name"`
	SaleKeyword        string `json:"sale_kw"`

}


type TaxiApiParams struct {
	Name         string `json:"name"`
	Data         ApiData `json:"data"`
	Fake         struct {
					 SendedStates []int `json:"sended_states"`
					 SleepTime    int `json:"sleep_time"`
				 } `json:"fake"`

	NotSendPrice bool `json:"not_send_price"`
}

func (api TaxiApiParams) String() string {
	return fmt.Sprintf("API %s\nAPI data: %+v\nFake?:%+v\nNotSendPrice?:%v", api.Name, api.Data, api.Fake, api.NotSendPrice)
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
func (api TaxiApiParams) GetIdService() string {
	return api.Data.IdService
}

func (api TaxiApiParams) GetAPIData() ApiData {
	return api.Data
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
	GeoOrbit          TaxiGeoOrbit `json:"geo_orbit"`
	Markups           *[]string `json:"markups,omitempty"`
	AvailableCommands map[string][]string `json:"available_commands"`
}


type ShopConfig struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Info string `json:"information"`
}


type Configuration struct {
	Main   struct {
			   Port         int    `json:"port"`
			   CallbackAddr string `json:"callback_addr"`
			   ConsoleAddr  string `json:"console_addr"`
			   LoggingFile  string `json:"log_file"`
			   GoogleKey    string `json:"google_key"`
			   ElasticConn  string `json:"elastic_conn"`
			   Database     struct {
								ConnString string `json:"connection_string"`
								Name       string `json:"name"`
							} `json:"database"`

		   } `json:"main"`
	Taxis  map[string]TaxiConfig `json:"taxis"`
	Shops  map[string]ShopConfig `json:"shops"`
	RuPost struct {
			   ExternalUrl string `json:"external_url"`
			   WorkUrl     string `json:"work_url"`
		   } `json:"ru_post"`
}


func ReadConfig() Configuration {
	fn := u.FoundFile("config.json")
	if fn == nil {
		log.Printf("can not find config.json file :(")
		os.Exit(-1)
	}
	cdata, err := ioutil.ReadFile(*fn)
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