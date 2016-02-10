package console

import (
	c "msngr/configuration"
	d "msngr/db"
	m "msngr"
	s "msngr/structs"
	n "msngr/notify"
)

const (
	ME = "me"
)

type ConsoleRequestProcessor struct {

}

func (crp *ConsoleRequestProcessor)ProcessRequest(in *s.InPkg) *s.RequestResult {
	result := s.RequestResult{Commands:&[]s.OutCommand{
		s.OutCommand{
			Title:    "Информация",
			Action:   "information",
			Position: 0,
		},
	},
	}
	return &result
}

type ConsoleInformationProcessor struct {
	Information string
}

func (cip ConsoleInformationProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	result := s.MessageResult{Type:"chat", Body:cip.Information}
	return &result
}

type ConsoleMessageProcessor struct {
	d.MainDb
}

func (cmp ConsoleMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	body := in.Message.Body
	userData := in.UserData
	if body != nil && userData != nil {
		u, _ := cmp.Users.GetUserById(in.From)
		if u == nil {
			cmp.Users.AddUser(in.From, userData.Name, userData.Phone, userData.Email)
		}
		cmp.Messages.StoreMessage(in.From, ME, *body, in.Message.ID)
		return &s.MessageResult{Type:"chat", Body:"", IsDeferred:true}
	}else {
		return &s.MessageResult{Type:"chat", Body:"Нет данных для сообщения или данных пользователя"}
	}
}

func FormConsoleBotContext(conf c.Configuration, db_handler *d.MainDb, cs c.ConfigStorage) *m.BotContext {
	result := m.BotContext{}
	result.Request_commands = map[string]s.RequestCommandProcessor{
		"commands":&ConsoleRequestProcessor{},
	}

	result.Message_commands = map[string]s.MessageCommandProcessor{
		"information":&ConsoleInformationProcessor{Information:conf.Console.Information},
		"":ConsoleMessageProcessor{MainDb:*db_handler},
	}

	notifier :=n.NewNotifier(conf.Main.CallbackAddr,conf.Console.Key, db_handler )
	go Run(conf.Console.WebPort, notifier, db_handler, cs)

	return &result
}

