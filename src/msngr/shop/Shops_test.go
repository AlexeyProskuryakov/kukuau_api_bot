package shop

import (
	"log"
	d "msngr/db"
	c "msngr/configuration"
	s "msngr/structs"

	"testing"
	"time"
	"reflect"
)


func TestLogInOut(t *testing.T) {
	conf := c.ReadConfig()
	d.DELETE_DB = true
	if d.DELETE_DB {
		log.Println("!start at test mode!")
		conf.Main.Database.Name = conf.Main.Database.Name + "_test"
	}

	user, pwd := "test", "test"

	db := d.NewDbHandler(conf.Main.Database.ConnString, conf.Main.Database.Name)
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

		db := d.NewDbHandler(conf.Main.Database.ConnString, conf.Main.Database.Name)
		for !db.IsConnected() {
			log.Printf("wait wile db is connected...")
			time.Sleep(1 * time.Second)
		}
		bot_context := FormShopCommands(db, &shop_conf)
		request_commands := bot_context.Request_commands
		message_commands := bot_context.Message_commands

		in := s.InPkg{From:"TEST", UserData:&s.InUserData{Phone:"TEST123"}, Request:&s.InRequest{ID:"1234", Type:"get"}}
		in.Request.Query.Action = "COMMANDS"

		request_result := request_commands["commands"].ProcessRequest(&in)
		rr_commands := *request_result.Commands
		log.Printf("request commands result: %#v", reflect.DeepEqual(rr_commands, NOT_AUTH_COMANDS))

		if !reflect.DeepEqual(rr_commands, NOT_AUTH_COMANDS){
			t.Errorf("before auth commands is auth")
		}

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
		log.Printf("auth commands result: %v", reflect.DeepEqual(ar_commnads, AUTH_COMMANDS))
		if !reflect.DeepEqual(ar_commnads, AUTH_COMMANDS){
			t.Errorf("after auth result commands is not aut")
		}
		log.Println("request commands next...")
		request_result = request_commands["commands"].ProcessRequest(&in)
		rr_commands = *request_result.Commands
		log.Printf("request commands result: %#v", reflect.DeepEqual(rr_commands, AUTH_COMMANDS))
		if !reflect.DeepEqual(rr_commands, AUTH_COMMANDS){
			t.Errorf("after auth commands not auth")
		}

	}
}
