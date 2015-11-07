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


func test_fundamental() {
	session, _ := mgo.Dial("localhost:27017")

	collection := session.DB("test").C("test")

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
	test_fundamental()
}