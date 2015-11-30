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
	Coordinates Coordinates `json:"coordinates"`
	State       OsmName `json:"state"`
	City        OsmName `json:"city"`
	Name        OsmName `json:"name"`
	Street      OsmName `json:"street"`
	OSM_ID      int64 `json:"osm_id"`
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
	Sort("user", true).// sort by "user" field, ascending
	Pretty(true).// pretty print request and response JSON
	Do()                // execute
	if err != nil {
		// Handle error
		panic(err)
	}
	var eet ElEntity
	for _, photon_hit := range searchPhotonResult.Each(reflect.TypeOf(eet)) {
		if entity, ok := photon_hit.(ElEntity); ok {
			_, err = client.Index().
			Index("autocomplete").
			Type("osm_hw").
			Id(fmt.Sprintf("%v", entity.OSM_ID)).
			BodyJson(entity).
			Do()
			if err != nil {
				// Handle error
				log.Printf("add el erro: %v", err)
				continue
			}
		}

	}

}