package main

import (
	ntf "msngr/notify"
	tstu "msngr/test"
	//i "msngr/init"
	"msngr"
	"msngr/db"
//"msngr/structs"
//"fmt"
	"log"
	"msngr/configuration"
	"msngr/structs"
)

var address = "http://localhost:9191/quest/quest_time"

func main() {
	msngr.TEST = true
	msngr.DEBUG = true

	cnf := configuration.ReadConfig()
	cnf.SetLogFile("test.log")
	error_chan := make(chan structs.MessageError, 10)
	ha := tstu.HandleAddress("/notify", ":9876", error_chan)
	for {
		res := <-ha
		log.Printf("HANDLE ADDR %v", res)
		if res == "listen" {
			break
		}
	}

	main_db := db.NewMainDb(cnf.Main.Database.ConnString, cnf.Main.Database.Name)
	//start_bot_result := make(chan string, 100)
	//go i.StartBot(main_db, start_bot_result)
	//for {
	//	res := <-start_bot_result
	//	log.Printf("BOT ENGINE: %v", res)
	//	if res == "listen" {
	//		break
	//	}
	//
	//}
	merror := structs.MessageError{
		Code:500,
		Type:"wait",
		Condition:"resource-constraint",
	}
	error_chan <- merror
	ntfr := ntf.NewNotifier(cnf.Main.CallbackAddr, "test_key", main_db)
	out_pkg, err := ntfr.NotifyText("TEST", "TEST MESSAGE")
	log.Printf("> NTF OUT IS: %v,\n err: %v", out_pkg, err)
	result, err := tstu.POST(address, &structs.InPkg{
		Message:&structs.InMessage{
			ID: out_pkg.Message.ID,
			Error: &merror,
			Body:&out_pkg.Message.Body,
		},
		From:out_pkg.To,
		UserData:&structs.InUserData{Name:"TEST"},
	})
	log.Printf("ANSWER must be empty %v, %v", result, err)
	mess := "hui"
	merror.Code = 404
	result, err = tstu.POST(address, &structs.InPkg{
		Message:&structs.InMessage{
			ID: mess,
			Error: &merror,
			Body:&mess,
		},
		From:out_pkg.To,
		UserData:&structs.InUserData{Name:"TEST"},
	})
	log.Printf("ANSWER must be empty %v %v", result, err)
	
	//result_outs := []*structs.OutPkg{}
	//for _, i := range []int{1, 2, 3, 4, 5, 6, 7} {
	//	out, err := ntfr.NotifyText("TEST", fmt.Sprintf("NOTIFY TST MESSSAGE %v", i))
	//	if out != nil {
	//		result_outs = append(result_outs, out)
	//		mw, err := main_db.Messages.GetMessageByMessageId(out.Message.ID)
	//		log.Printf("ASSERT sended status notify message? %v Err? %v", mw.MessageStatus, err)
	//	}else {
	//		log.Printf("WARN sended status notify message? %+v Err? %v", out, err)
	//	}
	//
	//}
	//
	//for _, out := range result_outs {
	//	if out.Message != nil {
	//		in := &structs.InPkg{
	//			Message:&structs.InMessage{
	//				Body:&out.Message.Body,
	//				Type:"error",
	//				Error:&out.Message.Error,
	//			},
	//			From:out.To,
	//			UserData:&structs.InUserData{Name:"TEST"},
	//		}
	//		log.Printf("TST_ERR_MESS send: \n\t%+v", in)
	//		out_echo, err := tstu.POST(address, in)
	//		log.Printf("ASSERT out_echo nil? %v and err:%v", out_echo == nil, err)
	//	}
	//	mw, err := main_db.Messages.GetMessageByMessageId(out.Message.ID)
	//	log.Printf("ASSERT err? %v we have this message? %v They have similar errors?: %v\n...message:\n%+v", err == nil, mw != nil, mw.MessageStatus == "error", mw.MessageCondition == out.Message.Error.Condition, )
	//
	//}
}
