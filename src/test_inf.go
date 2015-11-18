package main

import (
	i "msngr/taxi/infinity"
	c "msngr/configuration"
	"log"
)

func main() {
	config := c.ReadConfig()
	tc, ok := config.Taxis["fake"]
	if !ok{
		panic("not fake taxi in config!!!")
	}
	inf := i.GetInfinityAPI(tc.Api)
	log.Printf("inf: %+v", inf)
	log.Printf("connected? %v", inf.IsConnected())
	log.Printf("orders: %+v", inf.Orders())
	log.Printf("cars: %+v", inf.GetCarsInfo())

	as := i.GetInfinityAddressSupplier(tc.Api)
	log.Printf("is connected? %v", as.IsConnected())
	log.Printf("adress result: %+v", as.AddressesSearch("Никол").Rows)
}