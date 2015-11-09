package main

import (
	"log"
	d "msngr/db"
	c "msngr/configuration"
	s "msngr/structs"
	sh "msngr/shop"
	"time"
	"reflect"
)
func test_shop_support() {
	conf := c.ReadConfig()

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

func test_shop_login_logout() {
	conf := c.ReadConfig()
	d.DELETE_DB = true
	if d.DELETE_DB {
		log.Println("!start at test mode!")
		conf.Database.Name = conf.Database.Name + "_test"
	}

	user, pwd := "test", "test"

	db := d.NewDbHandler(conf.Database.ConnString, conf.Database.Name)
	err := db.Users.SetUserPassword(&user, &pwd)
	if err != nil {
		go func() {
			for err == nil {
				time.Sleep(1 * time.Second)
				err = db.Users.SetUserPassword(&user, &pwd)
				log.Printf("trying add user for test shops... now we have err:%+v", err)
			}
		}()
	}


	for _, shop_conf := range conf.Shops {
		if shop_conf.Name != "test_shop" {
			continue
		}

		db := d.NewDbHandler(conf.Database.ConnString, conf.Database.Name)
		for !db.IsConnected() {
			log.Printf("wait wile db is connected...")
			time.Sleep(1 * time.Second)
		}
		bot_context := sh.FormShopCommands(db, &shop_conf)
		request_commands := bot_context.Request_commands
		message_commands := bot_context.Message_commands

		in := s.InPkg{From:"TEST", UserData:&s.InUserData{Phone:"TEST123"}, Request:&s.InRequest{ID:"1234", Type:"get"}}
		in.Request.Query.Action = "COMMANDS"

		request_result := request_commands["commands"].ProcessRequest(&in)
		rr_commands := *request_result.Commands
		log.Printf("request commands result: %#v", reflect.DeepEqual(rr_commands, sh.NOT_AUTH_COMANDS))

		in_auth := s.InPkg{
			From:"TEST",
			UserData:&s.InUserData{Phone:"TEST123"},
			Message:&s.InMessage{Commands:&[]s.InCommand{
				s.InCommand{
					Action:"authorise",
					Form:s.InForm{
						Fields:[]s.InField{
							s.InField{
								Name:"username",
								Data:s.InFieldData{
									Value:"test",
								},

							},
							s.InField{
								Name:"password",
								Data:s.InFieldData{
									Value:"test",
								},
							},
						}},
				},
			}}}
		auth_result := message_commands["authorise"].ProcessMessage(&in_auth)
		ar_commnads := *auth_result.Commands
		log.Printf("auth commands result: %v", reflect.DeepEqual(ar_commnads, sh.AUTH_COMMANDS))

		log.Println("request commands next...")
		request_result = request_commands["commands"].ProcessRequest(&in)
		rr_commands = *request_result.Commands
		log.Printf("request commands result: %#v", reflect.DeepEqual(rr_commands, sh.AUTH_COMMANDS))
	}
}
func main() {
	//	test_shop_support()
	test_shop_login_logout()
}