package main

import (
	c "msngr/configuration"
	g "msngr/taxi/geo"
	s "msngr/structs"
	"msngr/taxi/set"
	"fmt"
	"msngr/utils"
	"log"
	"encoding/json"
)

const (
	taxi_name = "fake"
)
var config = c.ReadConfig()
var address = fmt.Sprintf("http://localhost:%v/taxi/%v", config.Main.Port, taxi_name)
var address_str = fmt.Sprintf("http://localhost:%v/taxi/%v/streets", config.Main.Port, taxi_name)
var userData = s.InUserData{Phone:"TESTPHONE", Name:"TESTNAME"}


func get_streets(name string) []g.DictItem {
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
//	log.Printf("Result is %+v", s_res)
	return s_res
}


func assert_not_duplicates(input []g.DictItem) bool {
	s_el := func (i g.DictItem) string {
		return fmt.Sprintf("%v%v%v",i.Key, i.Title, i.SubTitle)
	}
	input_set := set.NewSet()
	for _, el := range input{
		if input_set.Contains(s_el(el)){
			log.Printf("%v have duplicate",el)
			return false
		} else{
			input_set.Add(s_el(el))
		}
	}
	log.Printf("OK not duplicates")
	return true
}

func assert_only_street_names(input []g.DictItem)bool {
	for _, el := range input{
		if !g.CC_REGEXP.MatchString(el.Title){
			log.Printf("FAIL Not street name: %v",el.Title)
			return false
		}else{
			log.Printf("OK %v is street", el.Title)
		}
	}
	log.Printf("OK only streets")
	return true
}

func assert_not_empty(input []g.DictItem) bool{
	if len(input)>0{
		log.Printf("OK result have data")
		return true
	}
	log.Printf("FAIL result is empty")
	return false
}

func main() {
	s_res := []g.DictItem{}

//	s_res = get_streets("лес")
//	assert_not_duplicates(s_res)
//	assert_only_street_names(s_res)
//
//	s_res = get_streets("росс")
//	assert_not_duplicates(s_res)
//	assert_only_street_names(s_res)
//
//	s_res = get_streets("карла м")
//	assert_not_duplicates(s_res)
//	assert_only_street_names(s_res)
//
//	s_res = get_streets("ленина")
//	assert_not_duplicates(s_res)
//	assert_only_street_names(s_res)
//
//	s_res = get_streets("дивны")
//	assert_not_duplicates(s_res)
//	assert_only_street_names(s_res)
//	assert_not_empty(s_res)

//	s_res = get_streets("бульвар")
//	assert_not_duplicates(s_res)
//	assert_only_street_names(s_res)
//	assert_not_empty(s_res)

//	s_res = get_streets("энергет")
//	assert_not_duplicates(s_res)
//	assert_only_street_names(s_res)
//
	s_res = get_streets("весенний")
	assert_not_duplicates(s_res)
	assert_only_street_names(s_res)
	assert_not_empty(s_res)
//
//	s_res = get_streets("морск")
//	assert_not_duplicates(s_res)
//	assert_only_street_names(s_res)
//
//	s_res = get_streets("магистраль")
//	assert_not_duplicates(s_res)
//	assert_only_street_names(s_res)
//	assert_not_empty(s_res)

}

