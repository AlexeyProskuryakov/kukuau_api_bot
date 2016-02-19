package main
import (
	"gopkg.in/olivere/elastic.v2"
	"log"
	"reflect"
	"fmt"
	"math"
//	"strings"
//	"msngr/taxi/set"
	"regexp"
//	"msngr/taxi/set"
	"strings"

	"sort"
)


var reg = regexp.MustCompile("[а-яА-Я]{2,3}\\.")

var corrects = map[string]string{
	"ул.":"улица",
	"пл.":"площадь",
	"пр.":"проспект",
	"Ул.":"Улица",
	"пер.":"переулок",
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

type Coordinates struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func (c Coordinates) String() string {
	return fmt.Sprintf("{%v,%v}", c.Lat, c.Lon)
}

type ElEntity struct {
	Coordinates Coordinates `json:"coordinate"`
	State       OsmName `json:"state"`
	City        OsmName `json:"city"`
	Name        OsmName `json:"name"`
	Street      OsmName `json:"street"`
	OSM_ID      int64 `json:"osm_id"`
	OSM_key     string `json:"osm_key"`
	Importance  float64 `json:"importance"`
}

func (e ElEntity) String() string {
	return fmt.Sprintf("[%v] {%v} Name: %v |%v| %v", e.OSM_ID, e.OSM_key, e.Name, e.Importance, e.City)
}

type AutocompleteEntity struct {
	Name       string `json:"name"`
	OSM_ID     int64 `json:"osm_id"`
	City       string `json:"city"`
	Location   Coordinates `json:"location"`
}

func (e AutocompleteEntity) String() string {
	return fmt.Sprintf("[%v] %v at %v in %v", e.OSM_ID, e.Name, e.City, e.Location)
}


func NameCorrection(in string) string {
	if reg.MatchString(in) {
		if subst, ok := corrects[reg.FindAllString(in, -1)[0]]; ok {
			//correcting
			if !strings.Contains(in, " ") {
				in = reg.ReplaceAllString(in, fmt.Sprintf("%v ", subst))
			} else {
				in = reg.ReplaceAllString(in, subst)
			}
			in = strings.TrimSpace(in)
		}
	}
	return in
}

func EnsureAutocomplete(index_name string) {
	client, err := elastic.NewClient()
	if err != nil {
		log.Printf("elastic err: %v", err)
		return
	}
	termQuery := elastic.NewTermQuery("osm_key", "highway")
	searchPhotonResult, err := client.Search().Index("photon").Query(&termQuery).Size(math.MaxInt32).Pretty(true).Do()
	if err != nil {
		log.Printf("ERROR: %v", err)
		return
	}
	log.Println(searchPhotonResult)
	var eet ElEntity
	var count int
	for _, photon_hit := range searchPhotonResult.Each(reflect.TypeOf(eet)) {
		if entity, ok := photon_hit.(ElEntity); ok {
			name_to_save := entity.Name.GetAny()
			if name_to_save == "" {
				continue
			}
			name_to_save = NameCorrection(name_to_save)
			index_el := AutocompleteEntity{Name:name_to_save, OSM_ID: entity.OSM_ID, City:entity.City.GetAny(), Location:entity.Coordinates}
			_, err = client.Index().Index(index_name).Type("name").Id(fmt.Sprintf("%v", entity.OSM_ID)).BodyJson(index_el).Do()
			if err != nil {
				log.Printf("Error at adding to autocomplete: %v", err)
				continue
			}
			count += 1
		}

	}
	log.Printf("Was processed: %v results", count)
}




func CountsOfCities(index string) map[string]int {
	client, err := elastic.NewClient()
	cities := map[string]int{}
	if err != nil {
		log.Printf("elastic err: %v", err)
		return cities
	}
	var eet AutocompleteEntity
	result, err := client.Search().Index(index).Query(elastic.NewMatchAllQuery()).Size(math.MaxInt32).Do()
	for _, hit := range result.Each(reflect.TypeOf(eet)) {
		if entity, ok := hit.(AutocompleteEntity); ok {
			if val, ok := cities[entity.City]; ok {
				cities[entity.City] = val + 1
			}else {
				cities[entity.City] = 1
			}
		}
	}
	return cities
}


type ByCitySize struct {
	data   []string
	cities map[string]int
}

func ByCitySizeWithCityCounts(data []string, cities map[string]int) ByCitySize {
	return ByCitySize{data:data, cities:cities}
}
func (s ByCitySize) Len() int {
	return len(s.data)
}
func (s ByCitySize) Swap(i, j int) {
	s.data[i], s.data[j] = s.data[j], s.data[i]
}
func (s ByCitySize) Less(i, j int) bool {
	if s_c_i, ok_i := s.cities[s.data[i]]; ok_i {
		if s_c_j, ok_j := s.cities[s.data[j]]; ok_j {
			return s_c_i < s_c_j
		}
	}
	return false
}

func to_list_keys(input map[string]int) []string {
	res := []string{}
	for k, _ := range input {
		res = append(res, k)
	}
	return res
}


func ensure() {
	EnsureAutocomplete("autocomplete")

	cs := CountsOfCities("autocomplete")
	keys := ByCitySizeWithCityCounts(to_list_keys(cs), cs)
	sort.Sort(keys)
	sort.Reverse(keys)

	for _, k := range keys.data {
		log.Printf("%v : %v", k, cs[k])
	}
}

func main() {
	ensure()
}
