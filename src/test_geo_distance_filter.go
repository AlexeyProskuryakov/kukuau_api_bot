package main

import (
	"gopkg.in/olivere/elastic.v2"
	"log"
)

func main() {
	log.Println(elastic.NewGeoDistanceFilter("Foo").Lat(21.221).Lon(22.33).Distance("12km").DistanceType("km").Source())
}
