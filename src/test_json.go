package main

import "encoding/json"
import "fmt"


type TMAPIResponce struct {
	Code        int `json:"code"`
	Description string `json:"descr"`
	Data        string `json:"data"`
}

type Tariff struct {
	Id       int `json:"id"`
	Name     string `json:"name"`
	IsActive bool `json:"is_active"`
}

type TariffWrapper struct {
	TMAPIResponce
	Data struct {
			 Tariffs []Tariff `json:"tariffs"`
		 } `json:"data"`
}

var json_str = `{
	"code":1,
	"descr":"OK",
	"data":{"tariffs":
				[
					{"id":1,"name":"TARIFF1","is_active":true},
					{"id":2,"name":"TARIFF2","is_active":true}
				]
			}
	}`


func main() {
	res := TariffWrapper{}
	err:= json.Unmarshal([]byte(json_str), &res)
	fmt.Println(err)
	fmt.Println(res.Code)
	fmt.Println(res.Description)
	fmt.Println(res.Data.Tariffs)
}