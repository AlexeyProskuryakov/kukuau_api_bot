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
	PhotonName string `json:"photon_name"`
	OSM_ID     int64 `json:"osm_id"`
	City       string `json:"city"`
	Location   Coordinates `json:"location"`
}

func (e AutocompleteEntity) String() string {
	return fmt.Sprintf("[%v] %v at %v in %v", e.OSM_ID, e.Name, e.City, e.Location)
}


func CorrectAddressesAtAutocomplete() {
	client, err := elastic.NewClient()
	searchPhotonResult, err := client.Search().
	Index("photon").
	Size(math.MaxInt32).
	Pretty(true).
	Do()
	if err != nil {
		log.Printf("ERROR: %v", err)
		return
	}
	var eet ElEntity
	count := 0
	log.Printf("Start searching... At %v results", searchPhotonResult.TotalHits())

	for _, photon_hit := range searchPhotonResult.Each(reflect.TypeOf(eet)) {
		if entity, ok := photon_hit.(ElEntity); ok {
			name := fmt.Sprintf("%v %v", entity.Name.Ru, entity.Name.Default)
			if entity.OSM_key == "highway" && reg.MatchString(name) {
				if subst, ok := corrects[reg.FindAllString(name, -1)[0]]; ok {
					var name_to_save string
					if entity.Name.Ru != "" {
						name_to_save = entity.Name.Ru
					} else {
						name_to_save = entity.Name.Default
					}
					//correcting
					if !strings.Contains(name_to_save, " ") {
						name_to_save = reg.ReplaceAllString(name_to_save, fmt.Sprintf("%v ", subst))
					} else {
						name_to_save = reg.ReplaceAllString(name_to_save, subst)
					}
					//delete from autocomplete osm_id
					deleteResult, err := client.Delete().Index("autocomplete").Type("name").Id(fmt.Sprintf("%v", entity.OSM_ID)).Do()
					if err != nil {
						log.Printf("Error at deleting in autocomplete")
						continue
					}
					if deleteResult == nil {
						log.Printf("Delete result: %v", deleteResult)
					}

					index_el := AutocompleteEntity{Name:name_to_save, OSM_ID:entity.OSM_ID, City:entity.City.GetAny(), Location:entity.Coordinates}
					//paste
					log.Printf("Will paste this: %+v", index_el)
					_, err = client.Index().Index("autocomplete").Type("name").Id(fmt.Sprintf("%v", entity.OSM_ID)).BodyJson(index_el).Do()
					if err != nil {
						log.Printf("Error at adding to autocomplete: %v", err)
						continue
					}
					count += 1

				}
			}
		}
	}
	log.Printf("ALL: %v", count)

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

			if reg.MatchString(name_to_save) {
				if subst, ok := corrects[reg.FindAllString(name_to_save, -1)[0]]; ok {
					//correcting
					if !strings.Contains(name_to_save, " ") {
						name_to_save = reg.ReplaceAllString(name_to_save, fmt.Sprintf("%v ", subst))
					} else {
						name_to_save = reg.ReplaceAllString(name_to_save, subst)
					}
					name_to_save = strings.TrimSpace(name_to_save)
				}
			}

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
	//	agg := elastic.NewSumAggregation().Field("city")
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
// We implement `sort.Interface` - `Len`, `Less`, and
// `Swap` - on our type so we can use the `sort` package's
// generic `Sort` function. `Len` and `Swap`
// will usually be similar across types and `Less` will
// hold the actual custom sorting logic. In our case we
// want to sort in order of increasing string length, so
// we use `len(s[i])` and `len(s[j])` here.
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
	CorrectAddressesAtAutocomplete()
	cs := CountsOfCities("autocomplete")
	keys := ByCitySizeWithCityCounts(to_list_keys(cs), cs)
	sort.Sort(keys)
	sort.Reverse(keys)

	for _, k := range keys.data {
		log.Printf("%v : %v", k, cs[k])
	}
}

func main() {
//	query := "россий"
//	q := elastic.NewBoolQuery().Must(elastic.NewTermQuery("osm_key", "highway"), elastic.NewBoolQuery().Should(
//		elastic.NewMatchQuery("collector.default", query).Fuzziness("1").PrefixLength(2).MinimumShouldMatch("90%"),
//		elastic.NewMatchQuery("collector.ru.ngrams", query).Fuzziness("1").PrefixLength(2).MinimumShouldMatch("90%"),
//	).MinimumShouldMatch("1")).Should(
//		elastic.NewMatchQuery("name.ru.raw", query).Type("boolean").Analyzer("search_raw").Boost(200),
//		//		elastic.NewMatchQuery("collector.ru.raw", query).Type("boolean").Analyzer("search_raw").Boost(100),
//	)
//
//	script_funct := elastic.NewScriptFunction("").Lang("groovy").Script("general-score")
//	fsq := elastic.NewFunctionScoreQuery().AddScoreFunc(script_funct).ScoreMode("multiply").BoostMode("multiply").Query(q)
//
//	main_filtered := elastic.NewFilteredQuery(fsq)
//	main_filtered.Filter(elastic.NewOrFilter(elastic.NewTermFilter("osm_key", "highway")))
//	client, _ := elastic.NewClient(elastic.SetURL("http://localhost:9200"))
//	search := client.Search().Index("photon").Query(main_filtered).Size(100).Pretty(true)
//	result, _ := search.Do()
//	if result != nil {
//		log.Printf("%+v", result.TotalHits())
//	}
//	var oae geo.OsmAutocompleteEntity
//	if result.Hits != nil {
//		for _, hit := range result.Hits.Hits {
//			hit_in, e := hit.Source.MarshalJSON()
//			log.Printf("in: %s \n%v", hit_in, e)
//		}
//	}
//	for _, osm_hit := range result.Each(reflect.TypeOf(oae)) {
//
//		//		log.Printf("OWN RAw hit: %+v", osm_hit)
//		if entity, ok := osm_hit.(geo.OsmAutocompleteEntity); ok {
//			log.Printf("found: %+v", entity)
//		}
//
//	}

	ensure()
}
