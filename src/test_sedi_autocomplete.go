package main

import (
	"msngr/configuration"

//	"msngr/taxi"
	"msngr/taxi/sedi"
	"msngr"
	"log"
//	"msngr/taxi"
//	"fmt"
//	"time"
)

func main() {
	msngr.DEBUG = true
//	my_street := "Российская"
	my_address_from := "Российская 3"
	my_street_to := "Российская 28"

	conf := configuration.ReadConfig()
	api_data := conf.Taxis["fake"].Api.GetAPIData()
	s := sedi.NewSediAPI(&api_data)

	//	order_statuses, err := s.GetOrderStatuses()
	//	if err != nil {
	//		log.Printf("order statuses err : %v", err)
	//		return
	//	}
	//	for _, order_status := range order_statuses {
	//		log.Printf("order status : %+v \n", order_status)
	//	}

	log.Println(s)

	address_from := s.AddressesSearch(my_address_from)
	adress_to := s.AddressesSearch(my_street_to)
	log.Printf("\nFrom %+v\nTo %+v\n", address_from, adress_to)


//	order := taxi.NewOrderInfo{
//		Delivery:taxi.Delivery{Street:my_street, House:"3"},
//		Destinations:[]taxi.Destination{taxi.Destination{Street:my_street, House:"28"}},
//		Phone:"+79811064022",
//	}
//	res, mes := s.CalcOrderCost(order)
//	log.Printf("calc res: %v, mes: %s", res, mes)
//
//	ans := s.NewOrder(order)
//	log.Printf("ans of new order response: %+v", ans)
//
//	for {
//		orders := s.Orders()
//		for i, order := range orders {
//			log.Printf("%v\t%v\n", i, order)
//		}
//		time.Sleep(10 * time.Second)
//	}
//
//	if ans.IsSuccess {
//		fmt.Println(s.CancelOrder(ans.Content.Id))
//	}
}