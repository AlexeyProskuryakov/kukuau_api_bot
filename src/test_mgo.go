package main

import (
//	"gopkg.in/mgo.v2"
//	"gopkg.in/mgo.v2/bson"
//	"log"
	ms "msngr/db"
	t "msngr/taxi"
	"log"
)

type SubDoc struct {
	SubC  map[string]interface{}
	SubCS map[string]string
}

type Doc struct {
	Content string
	SubDoc  SubDoc
}

func main() {
	dbh := ms.NewDbHandler("localhost:27017", "test")

	order := t.Order{IDCar:100500, ID:100500600, Cost:100400}
	order_id := int64(1)

	order_data := order.ToOrderData()
	log.Println(order_data)

	//	dbh.Orders.AddOrder(order_id, "foo")
	dbh.Orders.SetState(order_id, 1, order_data)

	order_wrpr := dbh.Orders.GetByOrderId(order_id)
	log.Println(order_wrpr.OrderData)
}