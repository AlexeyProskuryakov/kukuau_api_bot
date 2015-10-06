package taxi

import (
	"net/http"
	"log"
	"fmt"
	"encoding/json"
	"msngr/utils"
	"io/ioutil"
)


const GOOGLE_API_URL = "https://maps.googleapis.com/maps/api"

type TaxiGeoOrbit struct {
	Lat    float32 `json:"lat"`
	Lon    float32 `json:"lon"`
	Radius int16 `json:"radius"`
}

type GoogleAddressHandler struct {
	AddressSupplier

	key   string
	orbit TaxiGeoOrbit

	// todo change to redis
	cache map[string]string
}


func NewGoogleAddressHandler(key string, orbit TaxiGeoOrbit) *GoogleAddressHandler {
	result := GoogleAddressHandler{key:key, orbit:orbit}
	return &result
}

func (ah *GoogleAddressHandler) GetExternalAddress(intern_adr string) (externalAddress *string, err error) {

	found, ok := ah.cache[intern_adr]
	if ok {
		return &found, nil
	}
	//todo realise request	
	return nil, nil
}

func (ah *GoogleAddressHandler) SetExternalAddress(ex_adr, intern_adr string) {
	ah.cache[intern_adr] = ex_adr
}

/*
"predictions":[
	{
		"description": "Лесосечная улица, Новосибирск, Новосибирская область, Россия",
		"id": "901e0f84483aea379b62f1a84612925968cefce4",
		"matched_substrings":[{"length": 7, "offset": 0 }],
		"place_id": "EnDQm9C10YHQvtGB0LXRh9C90LDRjyDRg9C70LjRhtCwLCDQndC-0LLQvtGB0LjQsdC40YDRgdC6LCDQndC-0LLQvtGB0LjQsdC40YDRgdC60LDRjyDQvtCx0LvQsNGB0YLRjCwg0KDQvtGB0YHQuNGP",
		"reference": "CqQBnwAAAKUKwB6aRUaud_UJSfxHkGuYDLmNuylCy8lw5S93tWzsJZjxE-d9RKpcIXKqgwqukQrpUpmbP4dlAntc5wIDVHGuJiqhq55OTd18VqpB3RuBt78EfJspbVftgQZB8I3pTH6tiuHzdd8Q5DmM6gSQ-zoSTu9Zbw-jMk7SRpkbKtSXPfUy91uvR4p-6SpTp4HIaDhDUbTTNhB0FAp-gi2py5sSEMIVlPKKp0qZHuj9JS9JrqwaFPWtEDLtuBvJkRCoFpZnl_--1uJI",
		"terms":[
		{"offset": 0, "value": "Лесосечная улица"},
		{"offset": 18, "value": "Новосибирск"},
		{"offset": 31, "value": "Новосибирская область"},
		{"offset": 54, "value": "Россия"}
		],
		"types":[
		"route",
		"geocode"
		]
},

*/
type GoogleTerm struct {
	Offset int16 `json:"offset"`
	Value  string `json:"value"`
}

type GooglePrediction struct {
	Description string `json:"description"`
	PlaceId     string `json:"place_id"`
	Terms       []GoogleTerm `json:"terms"`
}
type GoogleResultAddress struct {
	Predictions []GooglePrediction `json:"predictions"`
	Status      string `json:"status"`
}
func _to_fast_address(input GoogleResultAddress) FastAddress {
	rows := []FastAddressRow{}
	for _, prediction := range input.Predictions {
		row := FastAddressRow{}
		terms_len := len(prediction.Terms)
		if terms_len > 0 {
			row.Name = prediction.Terms[0].Value
		}
		if terms_len > 1 {
			row.City = prediction.Terms[1].Value
		}
		if terms_len > 2 {
			row.Region = prediction.Terms[2].Value
		}
		row.GID = prediction.PlaceId
		rows = append(rows, row)
	}
	result := FastAddress{Rows:&rows}
	return result
}
func GET(url string, params *map[string]string) (*[]byte, error) {
	log.Println("GET > ", url, " | ", params)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("ERROR! GAS With reqest [%v] ", url)
		return nil, err
	}

	if params != nil {
		values := req.URL.Query()

		for k, v := range *params {
			values.Add(k, v)
		}
		req.URL.RawQuery = values.Encode()
	}

	client := &http.Client{}
	log.Println("DDD", req)
	res, err := client.Do(req)
	if res == nil || err != nil {
		log.Println("ERROR! GAS response is: ", res, "; error is:", err, ". I will reconnect and will retrieve data again after 3s.")
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	return &body, err
}

func (ah *GoogleAddressHandler) AddressesSearch(q string) FastAddress {
	rows := []FastAddressRow{}
	result := FastAddress{Rows:&rows}
	suff := "/place/autocomplete/json"
	url := GOOGLE_API_URL + suff


	//maps.googleapis.com/maps/api/place/autocomplete/json?key=AIzaSyBkmvXK-SqfQcyj2XlXgTx-r_B18TJb-vY&components=country:ru&language=ru&location=54.890909,83.084399&radius=50000&types=address&input=лесосеч
	tmp := GoogleResultAddress{}
	params := map[string]string{
		"components": "country:ru",
		"language": "ru",
		"location": fmt.Sprintf("%+v,%+v", ah.orbit.Lat, ah.orbit.Lon),
		"radius": string(ah.orbit.Radius),
		"types": "address",
		"input": q,
		"key":ah.key,
	}
	body, err := GET(url, &params)
	err = json.Unmarshal(*body, &tmp)
	if err != nil {
		log.Printf("ERROR! GAS unmarshal error [%+v]", string(*body))
		return result
	}
	log.Printf("GA! loaded addr: [%+v]", tmp)
	result = _to_fast_address(tmp)
	log.Printf("GA! Formed to fast address:[%+v]", result)

	return result
}
func (ah *GoogleAddressHandler) IsConnected() bool {
	return true
}

func StreetsSearchController(w http.ResponseWriter, r *http.Request, i AddressSupplier) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	log.Println("Searching address..\n<<<.", r.Method, r.URL.Path, r.URL.Query())
	if r.Method == "GET" {

		params := r.URL.Query()
		query := params.Get("q")

		var results []DictItem
		if query != "" {

			if !i.IsConnected() {
				ans, _ := json.Marshal(map[string]string{"error":"true", "details":"service is not avaliable"})
				fmt.Fprintf(w, "%s", string(ans))
				return
			}
			log.Printf("connected. All ok. Start querying for: %+v", query)
			rows := i.AddressesSearch(query).Rows
			if rows == nil {
				return
			}
			log.Printf("was returned some data...")
			for _, nitem := range *rows {
				var item DictItem

				var key string
				if nitem.GID != ""{
					key = nitem.GID
				}else{
					key_raw, err := json.Marshal(nitem)
					key = string(key_raw)
					utils.CheckErr(err)
				}
				item.Key = string(key)
				item.Title = fmt.Sprintf("%v %v", nitem.Name, nitem.ShortName)
				item.SubTitle = fmt.Sprintf("%v", utils.FirstOf(nitem.Place, nitem.District, nitem.City, nitem.Region))
				results = append(results, item)
				log.Printf("interested: %+v", item)
			}
		}
		ans, err := json.Marshal(results)
		utils.CheckErr(err)
		log.Printf(">>> %q", string(ans))
		fmt.Fprintf(w, "%s", string(ans))

	}
}

type DictItem struct {
	Key      string `json:"key"`
	Title    string `json:"title"`
	SubTitle string `json:"subtitle"`
}


type InPlace struct {
	StreetId   int64 `json:"ID"`
	RegionId   int64 `json:"IDRegion"`
	DistrictId *int64 `json:"IDDistrict"`
	CityId     *int64 `json:"IDCity"`
	PlaceId    *int64 `json:"IDPlace"`
}

//helpers for forming
// destination and delivery on infinity results after street search request
func GetDeliveryHelper(info string, house string, entrance *string) Delivery {
	log.Printf("0 NO delivery marshalled: %+v\n and parameters: house: %+v, entrance: %q", info, house, entrance)
	in := InPlace{}
	err := json.Unmarshal([]byte(info), &in)
	utils.CheckErr(err)
	result := Delivery{IdStreet:in.StreetId, IdRegion:in.RegionId, House:house, Entrance:entrance,
		IdCity:in.CityId,
		IdPlace:in.PlaceId,
		IdDistrict:in.DistrictId,
	}
	log.Printf("1 NO delivery: %+v", result)
	return result
}

func GetDestinationHelper(info string, house string) Destination {
	log.Printf("0 NO destination marshalled: %+v", info)
	in := InPlace{}
	err := json.Unmarshal([]byte(info), &in)
	utils.CheckErr(err)
	result := Destination{IdStreet:in.StreetId, IdRegion:in.RegionId, House:house,
		IdCity:in.CityId,
		IdPlace:in.PlaceId,
		IdDistrict:in.DistrictId,
	}
	log.Printf("1 NO destination: %+v", result)
	return result
}
