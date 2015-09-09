package main

import (
	"log"
	m "msngr"
	inf "msngr/infinity"
)

func main() {
	conf := m.ReadConfig()

	realInfApi := inf.GetRealInfinityAPI(conf.Infinity)

	services := realInfApi.GetServices()

	log.Println(services)

}
