package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	d "msngr/db"
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

func test_taxi(){
	dbh := d.NewDbHandler("localhost:27017", "test")
	order_id := int64(1)

//	o:=dbh.Orders.GetByOrderId(order_id)
//	if o == nil {
//		dbh.Orders.AddOrder(order_id, "foo")
//	}
//
	dbh.Orders.AddOrder(order_id, "foo", "fake")

	order := t.Order{IDCar:100500, ID:100500600, Cost:100400}
	order_data := order.ToOrderData()
	log.Printf("insert: %+v",order_data)

	dbh.Orders.SetState(order_id, 1, order_data)

	order_wrpr := dbh.Orders.GetOrderById(order_id, "fake")
	log.Printf("wrpr: %+v", order_wrpr)
	log.Printf("result: %+v", order_wrpr.OrderData)
	idcar := order_wrpr.OrderData.Get("IDCar")
	log.Printf("result field: %+v %T", idcar, idcar)

	idfoo := order_wrpr.OrderData.Get("IDFoo")
	log.Printf("result field: %+v %T", idfoo, idfoo)
}


func test_fundamental(){
	session, _ := mgo.Dial("localhost:27017")

	collection:=session.DB("test").C("test")

	collection.RemoveAll(bson.M{})
	doc := Doc{Content:"foo"}
	_ = collection.Insert(doc)

	subdoc := SubDoc{SubC:map[string]interface{}{"f1":"123", "f2":123}, SubCS:map[string]string{"foo":"bar", "baz":"baz"}}

	find_key := bson.M{"content":"foo"}
	collection.Update(find_key, bson.M{"$set":bson.M{"subdoc":subdoc}})

	_ = collection.Find(find_key).One(&doc)

	log.Println(doc.SubDoc.SubC, "\n", doc.SubDoc.SubCS)
}

func main() {


	test_taxi()
//	test_fundamental()


}