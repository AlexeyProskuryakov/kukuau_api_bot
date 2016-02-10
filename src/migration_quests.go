package main

import (
	q "msngr/quests"
	c "msngr/configuration"
	"log"
)

func main() {
	conf := c.ReadConfig()
	qs := q.NewQuestStorage(conf.Main.Database.ConnString, conf.Main.Database.Name)
	users, err := qs.GetAllUsers()
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	for _, user := range users {
		if user.Name == "" && user.Phone == "" && user.EMail == "" {
			err = qs.Users.RemoveId(user.ID)
			if err != nil {
				log.Printf("Error: %v", err)
			}
		}
	}

}
