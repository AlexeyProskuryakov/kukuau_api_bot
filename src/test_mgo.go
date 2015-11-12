package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	d "msngr/db"
	t "msngr/taxi"
	"log"
	"time"
)

type SubDoc struct {
	SubC  map[string]interface{}
	SubCS map[string]string
}

type Doc struct {
	Content string
	SubDoc  SubDoc
}

type Person struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	Name      string
	Phone     string
	Timestamp time.Time
}

var (
	IsDrop = true
)

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
	
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic(err)
	}

	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	// Drop Database
	if IsDrop {
		err = session.DB("test").DropDatabase()
		if err != nil {
			panic(err)
		}
	}

	// Collection People
	c := session.DB("test").C("people")

	// Index
	index := mgo.Index{
		Key:        []string{"name", "phone"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}

	// Insert Datas
	err = c.Insert(&Person{Name: "Ale", Phone: "+55 53 1234 4321", Timestamp: time.Now()},
		&Person{Name: "Cla", Phone: "+66 33 1234 5678", Timestamp: time.Now()})

	if err != nil {
		panic(err)
	}

	// Query One
	result := Person{}
	err = c.Find(bson.M{"name": "Ale"}).Select(bson.M{"phone": 0}).One(&result)
	if err != nil {
		panic(err)
	}
	log.Println("Phone", result)

	// Query All
	var results []Person
	err = c.Find(bson.M{"name": "Ale"}).Sort("-timestamp").All(&results)

	if err != nil {
		panic(err)
	}
	log.Println("Results All: ", results)

	// Update
	colQuerier := bson.M{"name": "Ale"}
	change := bson.M{"$set": bson.M{"phone": "+86 99 8888 7777", "timestamp": time.Now()}}
	err = c.Update(colQuerier, change)
	if err != nil {
		panic(err)
	}

	// Query All
	err = c.Find(bson.M{"name": "Ale"}).Sort("-timestamp").All(&results)

	if err != nil {
		panic(err)
	}
	log.Println("Results All: ", results)

}

func main() {
	test_fundamental()
}