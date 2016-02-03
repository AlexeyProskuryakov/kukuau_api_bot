package main

import (
	s "msngr/structs"
	c "msngr/configuration"
	q "msngr/quests"
	"msngr/test"
	"msngr/utils"
	"log"

	"fmt"
	"strings"
	"gopkg.in/mgo.v2/bson"
)

const QUEST_NAME = "quest_time"

var config = c.ReadConfig()
var address = fmt.Sprintf("http://localhost:%v/quest/%v", config.Main.Port, QUEST_NAME)
var salt = "УРА!FUCKING112-012$%^&*()-1"

func send_key(key string, userData *s.InUserData) string {
	out := s.InPkg{UserData:userData, Message:&s.InMessage{Type:"chat", ID:utils.GenId(), Thread:utils.GenId(), Body:&key}, From:userData.Name}
	in, err := test.POST(address, &out)
	if err != nil {
		log.Printf("Error: %v", err)
		return ""
	}
	return strings.TrimSpace(in.Message.Body)
}

//todo create prepare
func PREPARE_KEYS(salt string, commands_count int) {
	qs := q.NewQuestStorage(config.Main.Database.ConnString, config.Main.Database.Name)
	qs.Keys.RemoveAll(bson.M{})
	qs.Messages.RemoveAll(bson.M{})
	qs.Peoples.RemoveAll(bson.M{})
	qs.Teams.RemoveAll(bson.M{})

	for i := 1; i <= commands_count; i++ {
		for j := 0; j <= 10; j++ {
			key, err := qs.AddKey(k(j, i), kr(j, i), k(j + 1, i))
			log.Printf("Q T Add key %+v err: %v", key, err)
		}
		key, err := qs.AddKey(k(11, i), kr(11, i), "")
		log.Printf("Q T Add last key %+v err: %v", key, err)
	}

}
func kr(kn, c int) string {
	return fmt.Sprintf("%+v-test-%s-%v", kn, salt, c)
}
func k(kn, c int) string {
	return fmt.Sprintf("#%+v-%v", kn, c)
}
func ok(kn, c int, who *s.InUserData) {
	result := send_key(k(kn, c), who)
	log.Printf("ASSERT KEY IS OK:   |||%v||| real: %v | want: %v | by: %v", result == kr(kn, c), result, kr(kn, c), who.Name)

}
func not_ok(kn, c int, who *s.InUserData) {
	result := send_key(k(kn, c), who)
	log.Printf("ASSERT KEY NOT OK:  |||%v||| real: %v | not want: %v | by: %v", result != kr(kn, c), result, kr(kn, c), who.Name)
}

func prepare_command(n_t, ph string, count int) []*s.InUserData {
	result := []*s.InUserData{}
	for i := int(0); i < count; i++ {
		user := &s.InUserData{Name:fmt.Sprintf("%v_%v", n_t, i), Phone:fmt.Sprintf("%s_%s", ph, i), Email:fmt.Sprintf("a_%s0@qa.ru", i)}
		result = append(result, user)
	}
	return result
}

func test_sequented(team []*s.InUserData) {
	not_ok(1, 9, team[7])
	not_ok(1, 9, team[9])
	not_ok(1, 9, team[8])

	not_ok(1, 8, team[1])
	not_ok(1, 8, team[1])
	not_ok(1, 8, team[2])
	not_ok(1, 8, team[3])

	not_ok(2, 6, team[1])
	not_ok(3, 6, team[3])
	//
	ok(0, 7, team[7])
	ok(1, 7, team[7])
	ok(2, 7, team[7])
	ok(3, 7, team[7])
	//
	ok(0, 7, team[8])
	ok(4, 7, team[8])

	not_ok(6, 7, team[8])
	not_ok(6, 7, team[7])

	ok(5, 7, team[7])
	ok(5, 7, team[8])

	not_ok(8, 7, team[8])
	not_ok(8, 7, team[7])

	ok(5, 7, team[7])
	ok(5, 7, team[8])

	not_ok(8, 7, team[8])
	not_ok(8, 7, team[7])

	ok(5, 7, team[7])
	ok(5, 7, team[8])
}

func test_new_user() {
	PREPARE_KEYS(salt, 3)
	team := prepare_command("alesha", "TEST", 2)

	not_ok(1, 1, team[1])
	not_ok(1, 1, team[1])
	ok(0,1, team[1])
	ok(0,1, team[0])

	ok(1,1, team[0])
	ok(2,1, team[0])
	ok(0,1, team[0])


	ok(2,1, team[1])


}

func main() {
	PREPARE_KEYS(salt, 10)
	//prepare_keys(salt, 10)
	team := prepare_command("alesha", "TEST", 20)
	ok(0,1,team[1])
	ok(0,2,team[1])
	ok(1,2, team[1])
	ok(2,2, team[1])
	ok(3,2, team[1])
	ok(4,2, team[1])

	ok(0,1, team[1])
	ok(1,1, team[1])
	test_sequented(team)
	test_new_user()


	//next_ok(0,2, team[1])
	//next_ok(1,2, team[1])


	//test_register(team, 2)

}
