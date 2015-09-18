package main

import (
	"log"
	m "msngr"
	taxi "msngr/taxi"
)

func main() {
	conf := m.ReadConfig()

	realInfApi := taxi.GetRealInfinityAPI(conf.Infinity)

	services := realInfApi.GetServices()
	log.Printf("services: %+v", services)

	orders := realInfApi.Orders()
	log.Printf("orders: %+v", orders)



}
