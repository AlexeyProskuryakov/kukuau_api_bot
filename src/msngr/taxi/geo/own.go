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
	m "msngr"

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
	result := OwnAddressHandler{client:client, connect_string:conn_str}
	result.orbit = orbit
	result.ExternalAddressSupplier = external
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

func get_own_result(client *elastic.Client, t_query elastic.TermQuery, filter elastic.Filter) []t.AddressF {
	rows := []t.AddressF{}
	s_result, err := client.Search().Index("autocomplete").Query(t_query).PostFilter(filter).Do()
	if err != nil {
		log.Printf("error in own address handler search at search in elastic %v", err)
	}
	var oae OsmAutocompleteEntity
	for _, osm_hit := range s_result.Each(reflect.TypeOf(oae)) {
		if entity, ok := osm_hit.(OsmAutocompleteEntity); ok {
			rows = append(rows, t.AddressF{OSM_ID:entity.OSM_ID, Name:entity.Name, City:entity.City})
		}
	}
	return rows
}

func (oh *OwnAddressHandler) AddressesAutocomplete(q string) t.AddressPackage {
	rows := []t.AddressF{}
	result := t.AddressPackage{Rows:&rows}
	t_query := elastic.NewTermQuery("name", q)
	filter := elastic.NewGeoDistanceFilter("location_filter")
	filter.Distance("50km")
	filter.Lat(oh.orbit.Lat)
	filter.Lon(oh.orbit.Lon)
	rows = get_own_result(oh.client, t_query, filter)
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
			if m.DEBUG {
				log.Printf("OWN GEI name == _name ? %v", name == _name)
			}
			rows := oh.ExternalAddressSupplier.AddressesAutocomplete(_name).Rows
			if rows == nil {
				return nil, errors.New("GetStreetId: no results at external")
			}
			ext_rows := *rows

			for i := len(ext_rows) - 1; i >= 0; i-- {
				nitem := ext_rows[i]
				ext_set := GetSetOfAddressF(nitem)
				if ext_set.IsSuperset(local_set) || local_set.IsSuperset(ext_set) {
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



