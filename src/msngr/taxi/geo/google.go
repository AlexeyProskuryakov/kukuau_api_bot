package geo

import (

	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	t "msngr/taxi"
	u "msngr/utils"
	c "msngr/configuration"
	s "msngr/taxi/set"
	"msngr/utils"

)

var NOT_IMPLY_TYPES = []string{"country"}

const (

	GOOGLE_API_URL = "https://maps.googleapis.com/maps/api"
	ERROR_EXTERNAL_MESSAGE = "служба такси не доступна."
)



type GoogleTerm struct {
	Offset int16 `json:"offset"`
	Value  string `json:"value"`
}
type GooglePrediction struct {
	Description string `json:"description"`
	PlaceId     string `json:"place_id"`
	Terms       []GoogleTerm `json:"terms"`
	Types       []string `json:"types"`
}

func (gp GooglePrediction) String() string {
	return fmt.Sprintf("GP: %s\n%s\nTerms:%+v\nTypes:%+v\n", gp.Description, gp.PlaceId, gp.Terms, gp.Types)
}

type GoogleResultAddress struct {
	Predictions []GooglePrediction `json:"predictions"`
	Status      string `json:"status"`
}

func (input GoogleResultAddress) ToFastAddress() t.AddressPackage {
	rows := []t.AddressF{}
	for _, prediction := range input.Predictions {
		if utils.InS("route", prediction.Types) {
			row := t.AddressF{}
			terms_len := len(prediction.Terms)
			if terms_len > 0 {
				row.Name, row.ShortName = GetStreetNameAndShortName(prediction.Terms[0].Value)
			}
			if terms_len > 1 {
				row.City = prediction.Terms[1].Value
			}
			if terms_len > 2 {
				row.Region = prediction.Terms[2].Value
			}
			row.GID = prediction.PlaceId
			rows = append(rows, row)
		} else {
			log.Printf("Adress is not route :( \n%+v", prediction)
		}
	}
	result := t.AddressPackage{Rows:&rows}
	return result
}

type GoogleAddressComponent struct {
	LongName  string `json:"long_name"`
	ShortName string `json:"short_name"`
	Types     []string `json:"types"`
}
type GoogleDetailPlaceResult struct {
	Result struct {
			   AddressComponents []GoogleAddressComponent `json:"address_components"`
			   Geometry          struct {
									 Location GooglePoint `json:"location"`
								 } `json:"geometry"`
			   FormattedAddress  string `json:"formatted_address"`
			   PlaceId           string `json:"place_id"`
			   Name              string `json:"name"`
		   }`json:"result"`
	Status string `json:"status"`
}
type GooglePoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lng"`
}

type GoogleAddressHandler struct {
	t.AddressSupplier
	t.AddressHandler

	key                     string

	cache                   map[string]*t.AddressF
	cache_dests             map[string]*GoogleDetailPlaceResult

	orbit                   c.TaxiGeoOrbit
	ExternalAddressSupplier t.AddressSupplier
}

func NewGoogleAddressHandler(key string, orbit c.TaxiGeoOrbit, external t.AddressSupplier) *GoogleAddressHandler {
	if key == ""{
		return nil
	}
	result := GoogleAddressHandler{key:key, orbit:orbit}
	result.cache = make(map[string]*t.AddressF)
	result.cache_dests = make(map[string]*GoogleDetailPlaceResult)
	result.ExternalAddressSupplier = external
	return &result
}

func (ah *GoogleAddressHandler) GetDetailPlace(place_id string) (*GoogleDetailPlaceResult, error) {
	from_info, err := u.GET(GOOGLE_API_URL + "/place/details/json", &map[string]string{
		"placeid":place_id,
		"key":ah.key,
		"language":"ru",
	})
	if err != nil || from_info == nil {
		log.Printf("ERROR! GetDetailPlace IN GET: %v", err)
		return nil, err
	}

	addr_details := GoogleDetailPlaceResult{}

	err = json.Unmarshal(*from_info, &addr_details)
	if err != nil {
		log.Printf("ERROR! GetDetailPlace IN UNMARSHALL: %v", err)
		return nil, err
	}
	if addr_details.Status != "OK" {
		log.Printf("ERROR! GetDetailPlace GOOGLE STATUS: %v", addr_details.Status)
		return nil, errors.New(addr_details.Status)

	}
	return &addr_details, nil
}

func (ah *GoogleAddressHandler) IsHere(key string) bool {
	addr_details, ok := ah.cache_dests[key]
	if !ok {
		var err error
		addr_details, err = ah.GetDetailPlace(key)
		if err != nil || addr_details == nil {
			return false
		}
		ah.cache_dests[key] = addr_details
	}
	point := addr_details.Result.Geometry.Location
	distance := Distance(point.Lat, point.Lon, ah.orbit.Lat, ah.orbit.Lon)

	return distance < ah.orbit.Radius
}

func (ah *GoogleAddressHandler) GetExternalInfo(key, name string) (*t.AddressF, error) {
	street_id, ok := ah.cache[key]
	if ok {
		return street_id, nil
	}
	var err error
	addr_details, ok := ah.cache_dests[key]
	if !ok {
		addr_details, err = ah.GetDetailPlace(key)
		if err != nil || addr_details == nil || addr_details.Status != "OK" {
			log.Printf("ERROR GetStreetId IN get place %+v %v", addr_details, err)
			return nil, err
		}
		ah.cache_dests[key] = addr_details
	}
	address_components := addr_details.Result.AddressComponents
	//	log.Printf(">>> [%v]\n%+v", key, address_components)
	query, google_set := _process_address_components(address_components)

	if query == "" {
		query = addr_details.Result.Name
	}
	log.Printf("GOOGLE query: [%v]\nGoogle set: %+v", query, google_set)
	if !ah.ExternalAddressSupplier.IsConnected() {
		return nil, errors.New(ERROR_EXTERNAL_MESSAGE)
	}

	rows := ah.ExternalAddressSupplier.AddressesAutocomplete(query).Rows
	if rows == nil {
		return nil, errors.New(ERROR_EXTERNAL_MESSAGE)
	}
	ext_rows := *rows

	for i := len(ext_rows) - 1; i >= 0; i-- {
		nitem := ext_rows[i]
		external_set := GetSetOfAddressF(nitem)

		log.Printf("GOOGLE query [%v]:\n external set: %+v < ? > google set: %+v", query, external_set, google_set)
		if google_set.IsSuperset(external_set) || external_set.IsSuperset(google_set) {
			log.Printf("GOOGLE: OK! [%+v] \nat %v", key, nitem)
			ah.cache[key] = &nitem
			return &nitem, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("No any results for [%v] address in external source", query))
}

func (ah *GoogleAddressHandler) AddressesAutocomplete(q string) t.AddressPackage {
	rows := []t.AddressF{}
	result := t.AddressPackage{Rows:&rows}
	suff := "/place/autocomplete/json"
	url := GOOGLE_API_URL + suff
	//	log.Printf(fmt.Sprintf("location= %v,%v", ah.orbit.Lat, ah.orbit.Lon))
	address_result := GoogleResultAddress{}
	params := map[string]string{
		"components": "country:ru",
		"language": "ru",
		"location": fmt.Sprintf("%v,%v", ah.orbit.Lat, ah.orbit.Lon),
		"radius": fmt.Sprintf("%v", ah.orbit.Radius),
		"types": "address",
		"input": q,
		"key":ah.key,
	}
	body, err := u.GET(url, &params)
	err = json.Unmarshal(*body, &address_result)
	if err != nil {
		log.Printf("ERROR! Google Adress Supplier unmarshal error [%+v]", string(*body))
		return result
	}

	result = address_result.ToFastAddress()
	return result
}

func (ah *GoogleAddressHandler) IsConnected() bool {
	return true
}



func _process_address_components(components []GoogleAddressComponent) (string, s.Set) {
	var route string
	google_set := s.NewSet()
	for _, component := range components {
		if u.IntersectionS(NOT_IMPLY_TYPES, component.Types) {
			//			log.Printf("component type %+v \ncontains not imply types: %v", component, component.Types)
			continue
		} else {
			long_name, err := AddStringToSet(google_set, component.LongName)
			if err != nil {
				log.Printf("WARN AT PROCESSING ADRESS COMPONENTS: %v", err)
				continue
			}
			if u.InS("route", component.Types) {
				route = long_name
			}
		}
	}
	return route, google_set
}

func GetStreetNameAndShortName(input string) (string, string) {
	addr_split := strings.Split(input, " ")
	var street_type, street_name string
	for _, sn_part := range addr_split {
		if u.InS(sn_part, []string{"улица", "проспект", "площадь", "переулок", "шоссе", "магистраль", "бульвар", "проезд"}) {
			street_type = _shorten_street_type(sn_part)
		} else {
			if street_name == "" {
				street_name += sn_part
			}else {
				street_name += " "
				street_name += sn_part
			}
		}
	}
	return street_name, street_type
}

func _shorten_street_type(input string) string {
	runes_array := []rune(input)
	if u.InS(input, []string{"улица", "проспект", "площадь"}) {
		return string(runes_array[:2]) + "."
	}else if u.InS(input, []string{"переулок", "шоссе", "магистраль"}) {
		return string(runes_array[:3]) + "."
	}
	return string(runes_array)
}

