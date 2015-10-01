package main

import (
	"gopkg.in/mgo.v2"
	m "msngr"
)

func main() {
	conf := m.ReadConfig()

	session, err := mgo.Dial(conf.Database.ConnString)
	if err != nil {
		panic(err)
	}

	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	err = session.DB(conf.Database.Name).DropDatabase()
	if err != nil {
		panic(err)
	}
}
