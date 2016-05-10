package main

import (
	"msngr/configuration"
	"msngr/db"
	"msngr/web"
)

func main() {
	config := configuration.ReadTestConfigInRecursive()
	mainDb := db.NewMainDb(config.Main.Database.ConnString, config.Main.Database.Name)
	web.TestRun(mainDb)
}
