package main

import (
	tm "msngr/taxi/master"
)


func main() {
	api := tm.TaxiMasterAPI{}
	api.GetTariffList()
}