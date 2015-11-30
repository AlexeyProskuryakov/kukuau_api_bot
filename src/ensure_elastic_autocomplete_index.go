package main
import (
	"gopkg.in/olivere/elastic.v2"
	"log"
	"reflect"
	"fmt"
)

type OsmName struct {
	Default string `json:"default"`
	Ru      string `json:"ru"`
}

type Coordinates struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type ElEntity struct {
	Coordinates Coordinates `json:"coordinate"`
	State       OsmName `json:"state"`
	City        OsmName `json:"city"`
	Name        OsmName `json:"name"`
	Street      OsmName `json:"street"`
	OSM_ID      int64 `json:"osm_id"`
}

type AutocompleteEntity struct {
	Name string `json:"name"`
	OSM_ID int64 `json:"osm_id"`
	City string `json:"city"`
}

func main() {
	client, err := elastic.NewClient()
	if err != nil {
		log.Printf("elastic err: %v", err)
		return
	}
	termQuery := elastic.NewTermQuery("osm_key", "highway")
	searchPhotonResult, err := client.Search().
	Index("photon").// search in index "twitter"
	Query(&termQuery).// specify the query
	Size(100000000).
	Pretty(true).// pretty print request and response JSON
	Do()                // execute
	if err != nil {
		// Handle error
		panic(err)
	}
	log.Println(searchPhotonResult)
	var eet ElEntity
	var prev_state string
	var count int
	for _, photon_hit := range searchPhotonResult.Each(reflect.TypeOf(eet)) {
		if entity, ok := photon_hit.(ElEntity); ok {
			index_el := AutocompleteEntity{}
			if entity.Name.Ru != ""{
				index_el.Name = entity.Name.Ru
			} else if entity.Name.Default != ""{
				index_el.Name = entity.Name.Default
			} else{
				continue
			}
			index_el.OSM_ID = entity.OSM_ID
			index_el.City = entity.City.Default

			_, err = client.Index().
			Index("autocomplete").
			Type("name").
			Id(fmt.Sprintf("%v", entity.OSM_ID)).
			BodyJson(index_el).
			Do()
			if err != nil {
				// Handle error
				log.Printf("add el erro: %v", err)
				continue
			}
			if prev_state != entity.State.Default {
				log.Printf("load for: %v, loaded: %v", entity.State.Default, count)
				prev_state = entity.State.Default
			}
			count += 1
		}

	}

}