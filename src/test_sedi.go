package main

import (
	"msngr/taxi/sedi"
	"log"
)

func main() {
	s := sedi.NewSediAPI("http://test2.sedi.ru")
	order_statuses, err := s.GetOrderStatuses()
	if err != nil {
		log.Printf("order statuses err : %v", err)
		return
	}
	for _, order_status := range order_statuses {
		log.Printf("order status : %+v \n", order_status)
	}
	activation_key_resp, err := s.GetActivationKey("+79811064022")
	if err != nil {
		log.Printf("activation key err: %+v\n", err)
		return
	}
	log.Printf("Activation key response: %+v", activation_key_resp)

	login_info, err := s.GetProfile()
	log.Printf("profile: %v\n",login_info)

	login_info, err = s.AuthoriseCustomer("Михаил Егоренков", "+79612183729")
	log.Printf("customer auth: %v\n",login_info)
}