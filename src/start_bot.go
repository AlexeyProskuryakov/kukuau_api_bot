package main

import (
	"log"

	m "msngr"
	i "msngr/init"
	d "msngr/db"
	"flag"
	"time"
	"msngr/configuration"
)

func InsertTestUser(db *d.MainDb, user, pwd string) {
	err := db.Users.SetUserPassword(user, pwd)
	if err != nil {
		go func() {
			for err == nil {
				time.Sleep(1 * time.Second)
				err = db.Users.SetUserPassword(user, pwd)
				log.Printf("trying add user for test shops... now we have err:%+v", err)
			}
		}()
	}
}


func main() {

	var test = flag.Bool("test", false, "go in test use?")
	flag.Parse()

	d.DELETE_DB = *test
	m.DEBUG = *test
	m.TEST = *test
	conf := configuration.ReadConfig()

	db := d.NewMainDb(conf.Main.Database.ConnString, conf.Main.Database.Name)
	log.Printf("Is delete DB? [%+v] Is debug? [%v]", d.DELETE_DB, m.DEBUG)
	if d.DELETE_DB {
		log.Println("!!!!!!!!!!start at test mode!!!!!!!!!!!!!")
		conf.Main.Database.Name = conf.Main.Database.Name + "_test"
		db.Session.DB(conf.Main.Database.Name).DropDatabase()
		db = d.NewMainDb(conf.Main.Database.ConnString, conf.Main.Database.Name)
	}

	InsertTestUser(db, "test", "test")
	InsertTestUser(db, "test1", "test1")
	InsertTestUser(db, "test2", "test2")
	result := make(chan string, 1000)
	i.StartBot(db, result)
}
