package main

import (
	"log"
	d "msngr/db"
	m "msngr"
	s "msngr/structs"
	sh "msngr/shop"
)
func test_shop_support() {
	conf := m.ReadConfig()

	d.DELETE_DB = true
	if d.DELETE_DB {
		log.Println("!start at test mode!")
		conf.Database.Name = conf.Database.Name + "_test"
	}

	for _, shop_conf := range conf.Shops {
		db := d.NewDbHandler(conf.Database.ConnString, conf.Database.Name)
		bot_context := sh.FormShopCommands(db, &shop_conf)
		request_commands := bot_context.Request_commands
		message_commands := bot_context.Message_commands

		in := s.InPkg{From:"TEST", UserData:&s.InUserData{Phone:"TEST123"}, Request:&s.InRequest{ID:"1234", Type:"get"}}
		in.Request.Query.Action = "COMMANDS"

		request_result := request_commands["commands"].ProcessRequest(&in)
		log.Println("commands: ERROR?: ", request_result.Error)


		in.Message = &s.InMessage{
			ID:"test_id",
			Type:"chat",
			Thread:"test_thread",

			Commands: &[]s.InCommand{
				s.InCommand{
					Title: "Support message command",
					Action: "support_message",
					Form: s.InForm{
						Title:"Support message form",
						Fields:[]s.InField{
							s.InField{
								Name:"support message",
								Data: struct {
									Value string `json:"value"`
									Text  string `json:"text"`
								}{Value:"Value", Text:"Text", },
							},
						},
					},
				},
			},
		}
		message_result := message_commands["support_message"].ProcessMessage(&in)
		log.Println("support message: ERROR?: ", message_result.Error)
	}
}
func main() {
	test_shop_support()
}