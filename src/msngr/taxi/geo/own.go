package geo

import (
	"log"
	"fmt"
	"errors"
	"math"
	"reflect"
	"sort"
	"strings"

	"gopkg.in/olivere/elastic.v2"

	t "msngr/taxi"
	s "msngr/utils/set"
	c "msngr/configuration"
)

type OwnAddressHandler struct {
	t.AddressSupplier
	t.AddressHandler

	ExternalAddressSupplier t.AddressSupplier
	orbit                   c.TaxiGeoOrbit
	client                  *elastic.Client
	connect_string          string
	city_handler            *CityHandler
}

func CountsOfCities(index string, client *elastic.Client) map[string]int {

	cities := map[string]int{}

	//	agg := elastic.NewSumAggregation().Field("city")
	var eet OsmAutocompleteEntity
	result, err := client.Search().Index(index).Query(elastic.NewMatchAllQuery()).Size(math.MaxInt32).Do()
	if err != nil {
		log.Printf("elastic err: %v", err)
		return cities
	}
	for _, hit := range result.Each(reflect.TypeOf(eet)) {
		if entity, ok := hit.(OsmAutocompleteEntity); ok {
			if strings.TrimSpace(entity.City) == "" {
				cities[entity.City] = 0
			}
			if val, ok := cities[entity.City]; ok {
				cities[entity.City] = val + 1
			}else {
				cities[entity.City] = 1
			}
		}
	}
	return cities
}

func NewOwnAddressHandler(conn_str string, orbit c.TaxiGeoOrbit, external t.AddressSupplier) *OwnAddressHandler {
	if conn_str == "" {
		return nil
	}
	client, err := elastic.NewClient(elastic.SetURL(conn_str))
	if err != nil {
		log.Printf("OWN can not connect to elastic index")
		return nil
	}
	result := OwnAddressHandler{
		client:client,
		connect_string:conn_str,
		orbit:orbit,
		ExternalAddressSupplier:external,
		city_handler:&CityHandler{city_weights:CountsOfCities("autocomplete", client)},
	}
	return &result
}

func (oh *OwnAddressHandler) IsConnected() bool {
	result, _, err := oh.client.Ping().Do()
	if err != nil {
		return false
	}
	if result != nil {
		return true
	}
	return false
}

func (oh *OwnAddressHandler) Connect() {
	client, err := elastic.NewClient(elastic.SetURL(oh.connect_string))
	if err != nil {
		log.Printf("OWN can not connect to elastic index")
		return
	}
	oh.client = client
}

type OsmAutocompleteEntity struct {
	Name        string `json:"name"`
	OSM_ID      int64 `json:"osm_id"`
	City        string `json:"city"`
	Coordinates Coordinates `json:"coordinates"`
}

type OsmName struct {
	Default string `json:"default"`
	Ru      string `json:"ru"`
}

func (on OsmName) String() string {
	return fmt.Sprintf("%v (%v)", on.Ru, on.Default)
}

func (on OsmName) Is(s string) bool {
	return on.Ru == s || on.Default == s
}

func (on OsmName) GetAny() string {
	if on.Ru != "" {
		return on.Ru
	}else {
		return on.Default
	}
}


type PhotonEntity struct {
	Coordinates Coordinates `json:"coordinate"`
	State       OsmName `json:"state"`
	City        OsmName `json:"city"`
	Name        OsmName `json:"name"`
	Street      OsmName `json:"street"`
	OSM_ID      int64 `json:"osm_id"`
	OSM_key     string `json:"osm_key"`
	Importance  float64 `json:"importance"`
}

func (e PhotonEntity) String() string {
	return fmt.Sprintf("[%v] {%v} Name: %v |%v| %v", e.OSM_ID, e.OSM_key, e.Name, e.Importance, e.City)
}


type CityHandler struct {
	city_weights map[string]int
}

func (ch *CityHandler) GetWeight(city string) (int, bool) {
	w, ok := ch.city_weights[city]
	return w, ok
}

func ByCitySizeWithCityCounts(data []t.AddressF, city_handler *CityHandler) ByCitySize {
	return ByCitySize{data:data, cities:city_handler}
}

type ByCitySize struct {
	data   []t.AddressF
	cities *CityHandler
}

func (s ByCitySize) Len() int {
	return len(s.data)
}
func (s ByCitySize) Swap(i, j int) {
	s.data[i], s.data[j] = s.data[j], s.data[i]
}
func (s ByCitySize) Less(i, j int) bool {
	if s_c_i, ok_i := s.cities.GetWeight(s.data[i].City); ok_i {
		if s_c_j, ok_j := s.cities.GetWeight(s.data[j].City); ok_j {
			return s_c_i > s_c_j
		}
	}
	return false
}


func ByWeightsOnOSM(data []t.AddressF, weights map[int64]float64) ByWeights {
	return ByWeights{weights:weights, data:data}
}

type ByWeights  struct {
	weights map[int64]float64
	data    []t.AddressF
}

func (s ByWeights) Len() int {
	return len(s.data)
}
func (s ByWeights) Swap(i, j int) {
	s.data[i], s.data[j] = s.data[j], s.data[i]
}
func (s ByWeights) Less(i, j int) bool {
	if s_c_i, ok_i := s.weights[s.data[i].OSM_ID]; ok_i {
		if s_c_j, ok_j := s.weights[s.data[j].OSM_ID]; ok_j {
			return s_c_i > s_c_j
		}
	}
	return false
}

func (oh *OwnAddressHandler) form_own_autocomplete_result(query elastic.Query, sort_by elastic.Sorter) []t.AddressF {
	rows := []t.AddressF{}
	s_result, err := oh.client.Search().Index("autocomplete").Query(query).Size(1000).SortBy(sort_by).Pretty(true).Do()
	if err != nil {
		log.Printf("error in own address handler search at search in elastic: \n%v", err)
		return rows
	}
	log.Printf("OWN Found %v in index", s_result.TotalHits())

	var oae OsmAutocompleteEntity
	name_city_set := s.NewSet()
	for _, osm_hit := range s_result.Each(reflect.TypeOf(oae)) {
		if entity, ok := osm_hit.(OsmAutocompleteEntity); ok {
			street_name, street_type := GetStreetNameAndShortName(entity.Name)
			entity_hash := fmt.Sprintf("%v%v%v", street_name, street_type, entity.City)
			if !name_city_set.Contains(entity_hash) && street_type != "" {
				addr := t.AddressF{
					Name:street_name,
					ShortName:street_type,
					OSM_ID:entity.OSM_ID,
					City:entity.City,
				}
				rows = append(rows, addr)
				name_city_set.Add(entity_hash)
			}
		}
	}
	return rows
}

func _get_address_slice(s s.Set, info map[string]t.AddressF) []t.AddressF {
	result := []t.AddressF{}
	iter_chan := s.Iter()
	for el := range iter_chan {
		a_h := el.(string)
		if address, ok := info[a_h]; ok {
			result = append(result, address)
		}
	}
	return result
}

func _addr_hash(el t.AddressF) string {
	return fmt.Sprintf("%v%v%v", el.Name, el.ShortName, el.City)
}

func _get_address_set(sl []t.AddressF, info map[string]t.AddressF) s.Set {
	result := s.NewSet()
	for _, el := range sl {
		ah := _addr_hash(el)
		info[ah] = el
		result.Add(ah)
	}
	return result
}

func _sort_addresses(sl []t.AddressF, weights map[int64]float64) []t.AddressF {
	result_sort := ByWeightsOnOSM(sl, weights)
	sort.Sort(result_sort)

	sorted_result := result_sort.data
	if len(sorted_result) > 20 {
		return sorted_result[:20]
	}
	return sorted_result
}


func _correct_io(in string) string {
	if IO_REGEXP.MatchString(in) {
		return IO_REGEXP.ReplaceAllString(in, "е")
	}
	return in
}

func (oh *OwnAddressHandler) autocomplete_rows(q string) []t.AddressF {
	result_sets := map[string]s.Set{}
	weghts := map[int64]float64{}
	info := map[string]t.AddressF{}
	all_found_addresses := []t.AddressF{}

	query := _correct_io(q)
	q_words := SPLIT_REGEXP.Split(query, -1)

	for _, word := range q_words {
		t_query := elastic.NewTermQuery("name", word)
		filter := elastic.NewGeoDistanceFilter("location").Distance("75km").Lat(oh.orbit.Lat).Lon(oh.orbit.Lon)
		query := elastic.NewFilteredQuery(t_query).Filter(filter)
		q_sort := elastic.NewGeoDistanceSort("location").Order(true).Point(oh.orbit.Lat, oh.orbit.Lon).Unit("km").SortMode("min").Asc()
		q_result := oh.form_own_autocomplete_result(query, q_sort)

		result_sets[word] = _get_address_set(q_result, info)
		all_found_addresses = append(all_found_addresses, q_result...)
	}
	for _, word := range q_words {
		for _, addr := range all_found_addresses {
			new_weight := oh.get_weight(addr, word)
			if weight, ok := weghts[addr.OSM_ID]; ok {
				new_weight += weight
			}
			weghts[addr.OSM_ID] = new_weight
		}
	}
	result := []t.AddressF{}
	if len(result_sets) == 1 {
		if set, ok := result_sets[query]; ok {
			result = _get_address_slice(set, info)
		}
	} else {
		intersect := s.NewSet()
		for _, set := range result_sets {
			if set.Cardinality() == 0 {continue}
			if intersect.Cardinality() == 0 {intersect = set}
			intersect = set.Intersect(intersect)
		}
		result = _get_address_slice(intersect, info)
	}
	return _sort_addresses(result, weghts)
}

func (oh *OwnAddressHandler) get_weight(a_addr t.AddressF, q string) float64 {
	if w, ok := oh.city_handler.GetWeight(a_addr.City); ok {
		addr_weight := float64(w)
		q = strings.TrimSpace(strings.ToLower(q))
		adr_name := strings.TrimSpace(strings.ToLower(_correct_io(a_addr.Name)))
		q_len := float64(len([]rune(q)))
		an_len := float64(len([]rune(adr_name)))
		var koef float64
		if strings.HasPrefix(adr_name, q) {
			if strings.EqualFold(adr_name, q) {
				koef = q_len
			} else {
				koef = (q_len / math.Abs(an_len - q_len)) + 1.0
			}
		}  else if strings.Contains(adr_name, q) {
			koef = (an_len - q_len) / (an_len + q_len)
		} else {
			return 0.0
		}
		return koef * addr_weight
	}
	return 0.0
}


func (oh *OwnAddressHandler) AddressesAutocomplete(q string) t.AddressPackage {

	a_rows := oh.autocomplete_rows(q)
	result := t.AddressPackage{Rows:&a_rows}
	return result
}

func (oh *OwnAddressHandler) IsHere(key string) bool {
	coords := oh.GetCoordinates(key)
	if coords != nil {
		coordinates := *coords
		distance := Distance(coordinates.Lat, coordinates.Lon, oh.orbit.Lat, oh.orbit.Lon)
		return distance < oh.orbit.Radius
	}
	return false
}

type Coordinates struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type OwnGeoCodeResult struct {
	Coordinates Coordinates `json:"coordinate"`
	State       OsmName `json:"state"`
	City        OsmName `json:"city"`
	Name        OsmName `json:"name"`
	Street      OsmName `json:"street"`
	OSM_ID      int64 `json:"osm_id"`
}

func (oh *OwnAddressHandler) GetCoordinates(key string) *Coordinates {
	t_query := elastic.NewTermQuery("osm_id", key)
	s_result, err := oh.client.Search().Index("photon").Query(t_query).Do()
	if err != nil {
		log.Printf("error in own address handler search at search in elastic %v", err)
		return nil
	}
	var ogcr OwnGeoCodeResult
	for _, osm_hit := range s_result.Each(reflect.TypeOf(ogcr)) {
		if entity, ok := osm_hit.(OwnGeoCodeResult); ok {
			return &entity.Coordinates
		}
	}
	return nil
}

func (oh *OwnAddressHandler) GetExternalInfo(key, name string) (*t.AddressF, error) {
	log.Printf("OWN Will getting external info of %v [%v]", name, key)
	t_query := elastic.NewTermQuery("osm_id", key)
	s_result, err := oh.client.Search().Index("photon").Query(t_query).Do()
	if err != nil {
		log.Printf("error in own address handler search at search in elastic %v", err)
	}
	var ogcr OwnGeoCodeResult
	for _, osm_hit := range s_result.Each(reflect.TypeOf(ogcr)) {
		if entity, ok := osm_hit.(OwnGeoCodeResult); ok {
			local_set := s.NewSet()
			_name := clear_address_string(entity.Name.GetAny())
			add_to_set(local_set, _name)
			add_to_set(local_set, clear_address_string(entity.City.GetAny()))

			log.Printf("OWN Query to external: |%v| \nlocal set: %+v", _name, local_set)

			rows := oh.ExternalAddressSupplier.AddressesAutocomplete(_name).Rows
			if IO_REGEXP.MatchString(_name) {
				rows_io := oh.ExternalAddressSupplier.AddressesAutocomplete(_correct_io(_name)).Rows
				if rows_io != nil {
					if rows == nil {
						rows = rows_io
					} else {
						e_rows_io := *rows_io
						e_rows := *rows
						e_rows = append(e_rows, e_rows_io...)
						rows = &e_rows
					}
					local_set.Remove(_name)
					local_set.Add(_correct_io(_name))
				}
			}

			if rows == nil {
				return nil, errors.New(fmt.Sprintf("система такси не знает местонахождение %v", _name))
			}
			ext_rows := *rows

			for i := len(ext_rows) - 1; i >= 0; i-- {
				nitem := ext_rows[i]
				ext_set := GetSetOfAddressF(nitem)
				log.Printf("OWN external set: %+v < ? > Local set %+v ", ext_set, local_set)
				if ext_set.IsSuperset(local_set) || local_set.IsSuperset(ext_set) {
					return &nitem, nil
				}
			}
		}
	}
	return nil, errors.New(fmt.Sprintf("не найдено ничего похожее на %v", name))

}

func clear_address_string(element string) (string) {
	result := strings.ToLower(element)
	result_raw := CC_REGEXP.ReplaceAllString(result, "")
	result = string(result_raw)
	result = strings.TrimSpace(result)
	return result
}

func add_to_set(set s.Set, element string) (string, error) {
	result := clear_address_string(element)
	if result != "" {
		set.Add(result)
		return result, nil
	}
	return element, errors.New(fmt.Sprintf("can not imply %+v ==> %+v", element, result))
}




