package geo

import (
	"log"
	"fmt"
	"errors"
	"strings"
	"reflect"

	"gopkg.in/olivere/elastic.v2"

	t "msngr/taxi"
	s "msngr/taxi/set"
	u "msngr/utils"
	c "msngr/configuration"
)
/*
Open street map and elastic search handler
*/
type OwnAddressHandler struct {
	t.AddressSupplier
	t.AddressHandler

	ExternalAddressSupplier t.AddressSupplier
	orbit                   c.TaxiGeoOrbit
	client                  *elastic.Client
	connect_string          string
}

func NewOwnAddressHandler(conn_str string, orbit c.TaxiGeoOrbit, external t.AddressSupplier) *OwnAddressHandler {
	client, err := elastic.NewClient(elastic.SetURL(conn_str))
	if err != nil {
		log.Printf("Error at connect to elastic")
		return nil
	}
	result := OwnAddressHandler{
		client:client,
		connect_string:conn_str,
		orbit:orbit,
		ExternalAddressSupplier:external}
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

type OsmAutocompleteEntity struct {
	Name   string `json:"name"`
	OSM_ID int64 `json:"osm_id"`
	City   string `json:"city"`
}

func get_own_result(client *elastic.Client, query elastic.Query, sort elastic.Sorter) []t.AddressF {
	rows := []t.AddressF{}
	s_result, err := client.Search().Index("autocomplete").Query(query).SortBy(sort).Pretty(true).Do()
	if err != nil {
		log.Printf("error in own address handler search at search in elastic: \n%v", err)
		return rows
	}
	var oae OsmAutocompleteEntity
	name_city_set := s.NewSet()
	for _, osm_hit := range s_result.Each(reflect.TypeOf(oae)) {
		if entity, ok := osm_hit.(OsmAutocompleteEntity); ok {
			name, short_name := GetStreetNameAndShortName(entity.Name)
			entity_hash := fmt.Sprintf("%v%v%v", name, short_name, entity.City)
			if !name_city_set.Contains(entity_hash) {
				addr := t.AddressF{}
				addr.Name, addr.ShortName = name, short_name
				addr.City = entity.City
				rows = append(rows, addr)
				log.Printf("OWN ADDR adding to result: %+v", addr)
				name_city_set.Add(entity_hash)
			}
		}
	}
	return rows
}

func (oh *OwnAddressHandler) AddressesAutocomplete(q string) t.AddressPackage {
	rows := []t.AddressF{}
	result := t.AddressPackage{Rows:&rows}

	t_query := elastic.NewTermQuery("name", q)
	filter := elastic.NewGeoDistanceFilter("location").Distance("50km").Lat(oh.orbit.Lat).Lon(oh.orbit.Lon)
	query := elastic.NewFilteredQuery(t_query).Filter(filter)
	sort := elastic.NewGeoDistanceSort("location").
	Order(true).
	Point(oh.orbit.Lat, oh.orbit.Lon).
	Unit("km").
	SortMode("min").
	Asc()

	rows = get_own_result(oh.client, query, sort)
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

type OsmName struct {
	Default string `json:"default"`
	Ru      string `json:"ru"`
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
	t_query := elastic.NewTermQuery("osm_id", key)
	s_result, err := oh.client.Search().Index("photon").Query(t_query).Do()
	if err != nil {
		log.Printf("error in own address handler search at search in elastic %v", err)
	}
	var ogcr OwnGeoCodeResult
	for _, osm_hit := range s_result.Each(reflect.TypeOf(ogcr)) {
		if entity, ok := osm_hit.(OwnGeoCodeResult); ok {
			local_set := s.NewSet()
			_name := clear_address_string(u.FirstOf(entity.Name.Ru, entity.Name.Default).(string))
			add_to_set(local_set, _name)
			add_to_set(local_set, clear_address_string(u.FirstOf(entity.City.Ru, entity.City.Default).(string)))

			log.Printf("OWN predict name: %v (input name: %v)\nset: %+v", _name, name, local_set)

			rows := oh.ExternalAddressSupplier.AddressesAutocomplete(_name).Rows
			if rows == nil {
				return nil, errors.New("GetStreetId: no results at external")
			}
			ext_rows := *rows

			for i := len(ext_rows) - 1; i >= 0; i-- {
				nitem := ext_rows[i]
				ext_set := GetSetOfAddressF(nitem)
				log.Printf("OWN External set: \n%+v", ext_set)
				if ext_set.IsSuperset(local_set) || local_set.IsSuperset(ext_set) {
					log.Printf("OWN result of comparing: %+v", nitem )
					return &nitem, nil
				}
			}
		}
	}
	return nil, errors.New(fmt.Sprintf("No any results for [%v] address in external source", key))
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




