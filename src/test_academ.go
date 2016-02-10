package main

import (
	c "msngr/configuration"
	s "msngr/structs"
	g "msngr/taxi/geo"
	"msngr/utils"
	"msngr/test"
	"log"
	"fmt"

	"encoding/json"
)

var config = c.ReadConfig()
var address = fmt.Sprintf("http://localhost:%v/taxi/academ", config.Main.Port)
var address_str = fmt.Sprintf("http://localhost:%v/taxi/academ/streets", config.Main.Port)
var userData = s.InUserData{Phone:"TESTPHONE", Name:"TESTNAME"}


func get_street(name, out_name string) s.InField {
	res, err := utils.GET(address_str, &map[string]string{"q":name})
	if err != nil {
		log.Printf("Error at get streets", err)
		panic(err)
	}
	s_res := []g.DictItem{}
	err = json.Unmarshal(*res, &s_res)
	if err != nil {
		log.Printf(err.Error())
	}
	log.Printf("Result is %+v", s_res)
	result := s.InField{Name:out_name, Type:"dict", Data:s.InFieldData{Value:s_res[0].Key, Text:s_res[0].Title}}
	return result
}

func sendNewOrder(from, to string) {
	street_from := get_street(from, "street_from")
	house_from := s.InField{Data:s.InFieldData{Value:"1"}, Type:"text", Name:"house_from"}

	street_to := get_street(to, "street_to")
	house_to := s.InField{Data:s.InFieldData{Value:"1"}, Type:"text", Name:"house_from"}
	entrance := s.InField{Type:"number", Name:"entrance"}

	form := s.InForm{Fields:[]s.InField{street_from, street_to, house_from, house_to, entrance}}
	command := s.InCommand{Action:"new_order", Form:form}

	out := s.InPkg{
		UserData:&userData,
		Message:&s.InMessage{Type:"chat", ID:utils.GenId(), Thread:utils.GenId(), Commands:&[]s.InCommand{command}},
		From:userData.Name,
	}

	in, err := test.POST(address, &out)
	log.Printf("RESULT: %s\nerr?:%v", in.Message.Body, err)
}

func sendCancelOrder(){
	in := test.ReadTestFile("cancel_order.json")
	in.From = userData.Name
	in.UserData = &userData

	res, err := test.POST(address, in)
	log.Printf("RESULT: %s\nerr?:%v", res.Message.Body, err)
}

func main() {
	sendNewOrder("весенний","детский")
	sendCancelOrder()
}