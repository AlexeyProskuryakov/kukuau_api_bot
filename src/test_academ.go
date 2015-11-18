package main

import (
	i "msngr/taxi/infinity"
	c "msngr/configuration"
	"log"
	"fmt"
)


func main() {
	config := c.ReadConfig()
	taxi_name := "academ"
	tc, ok := config.Taxis[taxi_name]
	if !ok {
		panic(fmt.Sprintf("not %v taxi in config!!!", taxi_name))
	}
	inf := i.GetInfinityAPI(tc.Api)
	log.Printf("inf: %+v", inf)
	log.Printf("connected? %v", inf.IsConnected())
	log.Printf("orders: %+v", inf.Orders())
	log.Printf("cars: %+v", inf.GetCarsInfo())

	as := i.GetInfinityAddressSupplier(tc.Api)
	log.Printf("is connected? %v", as.IsConnected())
	log.Println(as.AddressesSearch("Никол"))
	log.Println(as.AddressesSearch("Росс"))
	log.Println(as.AddressesSearch("ман"))
	log.Println(as.AddressesSearch("клав"))
}