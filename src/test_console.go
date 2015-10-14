package main
import (
	"msngr/console"
	m "msngr"
	d "msngr/db"
)


func main() {
	conf := m.ReadConfig()
	db := d.NewDbHandler(conf.Database.ConnString, conf.Database.Name)
	console.Run(conf, db)
}