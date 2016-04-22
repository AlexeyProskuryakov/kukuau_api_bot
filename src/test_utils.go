package main

import (
	"msngr/utils"
	cfg "msngr/configuration"
	"log"
)

func main() {
	cc1 := cfg.ChatConfig{CompanyId:"test", Notifications:[]cfg.TimedAnswer{}, AutoAnswers:[]cfg.TimedAnswer{cfg.TimedAnswer{Text:"test", }}, Key:"key"}
	cc2 := cfg.ChatConfig{CompanyId:"test", Notifications:[]cfg.TimedAnswer{}, AutoAnswers:[]cfg.TimedAnswer{cfg.TimedAnswer{Text:"test", }}, Key:"key2", User:"foo"}
	cc3 := cfg.ChatConfig{CompanyId:"test", Notifications:[]cfg.TimedAnswer{}, AutoAnswers:[]cfg.TimedAnswer{cfg.TimedAnswer{Text:"test", }}, User:"foo", Name:"kkkookkoo"}

	result1, _ := utils.GetStringUpdates(cc1, cc2, "bson", true)
	result2, _ := utils.GetStringUpdates(cc2, cc1, "bson", true)
	result3, _ := utils.GetStringUpdates(cc2, cc3, "bson", true)
	log.Printf("1 -> 2: %+v", result1)
	log.Printf("2 -> 1: %+v", result2)
	log.Printf("2 -> 3: %+v", result3)
}
