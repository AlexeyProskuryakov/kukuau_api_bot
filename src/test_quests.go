package main

import (
	"encoding/json"
	s "msngr/structs"
	c "msngr/configuration"
	q "msngr/quests"

	"msngr/utils"
	"log"
	"bytes"
	"net/http"

	"fmt"
	"io/ioutil"

)
const QUEST_NAME="quest_time"

var config = c.ReadConfig()
var address = fmt.Sprintf("http://localhost:%v/quest/quest_time", config.Main.Port)

func send_key(key string, userData *s.InUserData) string {
	out := s.InPkg{UserData:userData, Message:&s.InMessage{Type:"chat", ID:utils.GenId(), Thread:utils.GenId(), Body:&key}, From:userData.Name}
	jsoned_out, err := json.Marshal(&out)
	if err != nil {
		log.Printf("TQ error at unmarshal %v", err)
		return err.Error()
	}

	body := bytes.NewBuffer(jsoned_out)
	req, err := http.NewRequest("POST", address, body)
	if err != nil {
		log.Printf("TQ error at for request %v", err)
		return err.Error()
	}

	req.Header.Add("Content-Type", "application/json")

	print, _ := json.MarshalIndent(out, "", "	")
	log.Printf("TQ >> %+v \n%+v \n %s", address, req.Header, print)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("TQ error at do request %v", err)
		return err.Error()
	}
	if resp != nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("TQ << ERROR:%+v", err)
			return err.Error()
		} else {
			log.Printf("TQ << %v", string(body))
			in := s.OutPkg{}
			err := json.Unmarshal(body, &in)
			if err != nil {
				log.Printf("TQ err in unmarshal %v", err)
				return ""
			}
			return in.Message.Body
		}
		defer resp.Body.Close()
	}
	return ""
}

//todo create prepare
func prepare_keys(){
	qs := q.NewQuestStorage(config.Main.Database.ConnString,config.Main.Database.Name)
	qs.AddKey("#1-1", fmt.Sprintf("test_%s", "salt"), "#2-1")
}

func main() {
	user1 := &s.InUserData{Name:"alesha1", Phone:"+79138973664", Email:"a0@qa.ru"}
	user2 := &s.InUserData{Name:"alesha2", Phone:"+79138973665", Email:"a1@qa.ru"}
	user3 := &s.InUserData{Name:"alesha3", Phone:"+79138973666", Email:"a2@qa.ru"}
	user4 := &s.InUserData{Name:"alesha4", Phone:"+79138973667", Email:"a3@qa.ru"}
	log.Printf("1 %s", send_key("#1-1", user1))
	log.Printf("2 %s", send_key("#1-1", user2))
	log.Printf("3 %s", send_key("#1-1", user3))
	log.Printf("4 %s", send_key("#1-1", user4))
	//log.Printf("1", send_key("#1-1", user1))
}
