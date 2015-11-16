package taxi

import (
	"net/http"
	"log"
	"fmt"
	"encoding/json"
	"msngr/utils"
	"math"
	"errors"
	"regexp"
	"strings"

	s "msngr/taxi/set"
	u "msngr/utils"
	c "msngr/configuration"
)

/*
Коды состояния

Поле status в объекте ответа на запрос содержит данные о состоянии запроса и может содержать отладочную информацию, позволяющую выяснить причину сбоя. В поле status могут быть указаны следующие значения:

OK – ошибок нет, место обнаружено, и получен хотя бы один результат.
ZERO_RESULTS – поиск выполнен, результатов не найдено. Такое может произойти, если для поиска установлены координаты latlng отдаленного места.
OVER_QUERY_LIMIT – превышена квота.
REQUEST_DENIED – означает, что запрос отклонен, как правило, из-за отсутствия или неверного значения параметра key.
INVALID_REQUEST – как правило, отсутствует обязательный параметр запроса (location или radius).
*/

var cc_reg = regexp.MustCompilePOSIX("(ул(ица|\\.| )|пр(\\.|оспект|\\-кт)?|пер(\\.|еулок| )|г(ород|\\.|ор\\.| )|обл(асть|\\.| )|р(айон|\\-н )|^с )?")

const GOOGLE_API_URL = "https://maps.googleapis.com/maps/api"


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
	AddressSupplier

	key                     string
	orbit                   c.TaxiGeoOrbit

	cache                   map[string]*AddressF
	cache_dests             map[string]*GoogleDetailPlaceResult

	ExternalAddressSupplier AddressSupplier
}


func NewGoogleAddressHandler(key string, orbit c.TaxiGeoOrbit, external AddressSupplier) *GoogleAddressHandler {
	result := GoogleAddressHandler{key:key, orbit:orbit}
	result.cache = make(map[string]*AddressF)
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

func (ah *GoogleAddressHandler)IsHere(place_id string) bool {
	addr_details, ok := ah.cache_dests[place_id]
	if !ok {
		var err error
		addr_details, err = ah.GetDetailPlace(place_id)
		if err != nil || addr_details == nil {
			return true
		}
		ah.cache_dests[place_id] = addr_details
	}
	point := addr_details.Result.Geometry.Location
	distance := Distance(point.Lat, point.Lon, ah.orbit.Lat, ah.orbit.Lon)

	return distance < ah.orbit.Radius
}

func (ah *GoogleAddressHandler) GetStreetInfo(place_id string) (*AddressF, error) {
	street_id, ok := ah.cache[place_id]
	if ok {
		return street_id, nil
	}
	var err error
	addr_details, ok := ah.cache_dests[place_id]
	if !ok {
		addr_details, err = ah.GetDetailPlace(place_id)
		if err != nil || addr_details == nil || addr_details.Status != "OK" {
			log.Printf("ERROR GetStreetId IN get place %+v", addr_details)
			return nil, err
		}
	}
	address_components := addr_details.Result.AddressComponents
	query, google_set := _process_address_components(address_components)

	if query == ""{
		return nil, errors.New("GetStreetId: Can not imply query for external adress supplier")
	}
	if !ah.ExternalAddressSupplier.IsConnected() {
		return nil, errors.New("GetStreetId: External service is not avaliable")
	}

	rows := ah.ExternalAddressSupplier.AddressesSearch(query).Rows
	if rows == nil {
		return nil, errors.New("GetStreetId: no results at external")
	}
	ext_rows := *rows
	for i := len(ext_rows) -1; i>= 0; i-- {
		external_set := s.NewSet()
		nitem := ext_rows[i]
		_add_to_set(external_set, nitem.Name)
		_add_to_set(external_set, nitem.Region)
		_add_to_set(external_set, nitem.City)
		_add_to_set(external_set, nitem.District)
		_add_to_set(external_set, nitem.Place)

//		log.Printf("GetStreetId [%v]:\n %+v <=> %+v", query, external_set, google_set)
		if google_set.IsSuperset(external_set) {
			log.Printf("GetStreetId: [%+v] \nat %v", place_id, nitem.FullName)
			ah.cache[place_id] = &nitem
			return &nitem, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("No any results for [%v] address in external source", query))
}


func (ah *GoogleAddressHandler) AddressesSearch(q string) AddressPackage {
	rows := []AddressF{}
	result := AddressPackage{Rows:&rows}
	suff := "/place/autocomplete/json"
	url := GOOGLE_API_URL + suff
	log.Printf(fmt.Sprintf("location= %v,%v", ah.orbit.Lat, ah.orbit.Lon))
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

	result = _to_fast_address(address_result)
	return result
}

func (ah *GoogleAddressHandler) IsConnected() bool {
	return true
}



func StreetsSearchController(w http.ResponseWriter, r *http.Request, i AddressSupplier) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)

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
			log.Printf("was returned: %v rows", len(*rows))
			for _, nitem := range *rows {
				var item DictItem

				var key string
				if nitem.GID != "" {
					key = nitem.GID
				}else {
					key_raw, err := json.Marshal(nitem)
					key = string(key_raw)
					if err != nil {
						log.Printf("SSC: ERROR At unmarshal:%+v", err)
					}

				}
				item.Key = string(key)
				if nitem.ShortName != "" {
					item.Title = fmt.Sprintf("%v %v", nitem.Name, nitem.ShortName)
				}else {
					item.Title = nitem.Name
				}
				item.SubTitle = fmt.Sprintf("%v", utils.FirstOf(nitem.Place, nitem.District, nitem.City, nitem.Region))
				results = append(results, item)
			}
		}
		ans, err := json.Marshal(results)
		if err != nil {
			log.Printf("SSC: ERROR At unmarshal:%+v", err)
		}
		fmt.Fprintf(w, "%s", string(ans))
	}
}

func _clear_address_string(element string) (string){
	result := strings.ToLower(element)
	result_raw := cc_reg.ReplaceAllString(result, "")
	result = string(result_raw)
	result = strings.TrimSpace(result)
	return result
}

func _add_to_set(set s.Set, element string) (string, error) {
	result := _clear_address_string(element)
	if result != "" {
		set.Add(result)
		return result, nil
	}
	return element, errors.New(fmt.Sprintf("can not imply %+v ==> %+v", element, result))
}

func _process_address_components(components []GoogleAddressComponent) (string, s.Set) {
	var route string
	google_set := s.NewSet()
	for _, component := range components {
		long_name, err := _add_to_set(google_set, component.LongName)
		if err != nil{
			log.Printf("WARN AT PROCESSING ADRESS COMPONENTS: %v", err)
			continue
		}
		if utils.InS("route", component.Types) {
			route = long_name
		}
	}
	return route, google_set
}

func _to_fast_address(input GoogleResultAddress) AddressPackage {
	rows := []AddressF{}
	for _, prediction := range input.Predictions {
		row := AddressF{}
		terms_len := len(prediction.Terms)
		if terms_len > 0 {
			row.Name, row.ShortName = _get_street_name_shortname(prediction.Terms[0].Value)
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
	result := AddressPackage{Rows:&rows}
	return result
}

func _get_street_name_shortname(input string) (string, string) {
	addr_split := strings.Split(input, " ")
	var street_type, street_name string
	for _, sn_part := range addr_split {
		if u.InS(sn_part, []string{"улица", "проспект", "площадь", "переулок", "шоссе", "магистраль"}) {
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


func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta / 2), 2)
}
// Distance function returns the distance (in meters) between two points of
// a given longitude and latitude relatively accurately (using a spherical
// approximation of the Earth) through the Haversin Distance Formula for
// great arc distance on a sphere with accuracy for small distances
// point coordinates are supplied in degrees and converted into rad. in the func
// distance returned is METERS!!!!!!
// http://en.wikipedia.org/wiki/Haversine_formula
func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2 - la1) + math.Cos(la1) * math.Cos(la2) * hsin(lo2 - lo1)
	result := 2 * r * math.Asin(math.Sqrt(h))
	return result
}

