package main

import (
	tm "msngr/taxi/master"
	"log"
)


func main() {
	api := tm.TaxiMasterAPI{}
	log.Printf("tm api: %+v", api)
}