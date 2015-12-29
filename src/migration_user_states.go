package main

import (
	db "msngr/db"
	c "msngr/configuration"
	s "msngr/structs"

	"log"
	"gopkg.in/mgo.v2/bson"
)

func main() {
	conf := c.ReadConfig()
	handler := db.NewDbHandler(conf.Main.Database.ConnString, conf.Main.Database.Name)
	s.StartAfter(func() (string, bool) {
		return "", handler.Check()
	}, func() {
		log.Printf("connection established....")
	})

	log.Println("Starting migration 1...")
	users_collection := handler.Session.DB(conf.Main.Database.Name).C("users")
	users_collection.UpdateAll(bson.M{}, bson.M{"$set":bson.M{"states":bson.M{}}})

}