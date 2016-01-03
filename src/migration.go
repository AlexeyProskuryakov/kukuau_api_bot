package main

import (
	db "msngr/db"
	c "msngr/configuration"
	s "msngr/structs"
	"gopkg.in/mgo.v2"
	"log"
)

func main() {
	conf := c.ReadConfig()
	handler := db.NewMainDb(conf.Main.Database.ConnString, conf.Main.Database.Name)
	s.StartAfter(func() (string, bool) {
		return "", handler.Check()
	}, func() {
		log.Printf("connection established....")
	})

	log.Println("Starting migration 0...")
	orders_collection := handler.Session.DB(conf.Main.Database.Name).C("orders")
	err := orders_collection.DropIndex("order_id")
	if err != nil {
		panic(err)
	}
	err = orders_collection.EnsureIndex(mgo.Index{
		Key:        []string{"order_id"},
		Background: true,
		Unique:     false,
		DropDups:   true,
	})
	log.Printf("Index for order_id field is recreated...%v", err)
}