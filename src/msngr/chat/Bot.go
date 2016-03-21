package chat

import (
	"msngr"
	"msngr/db"
	s "msngr/structs"

	"errors"
	"msngr/configuration"
	"log"
)

var (
	DB_ERROR = errors.New("Ошибка на стороне базы данных, попробуйте позже.")
	DB_ERROR_RESULT = &s.MessageResult{Type:"chat", Body:DB_ERROR.Error()}
)

type ChatRequestProcessor struct {
	Commands *[]s.OutCommand
}

func (crp *ChatRequestProcessor)ProcessRequest(in *s.InPkg) *s.RequestResult {
	result := s.RequestResult{Commands:crp.Commands}
	return &result
}

type ChatInformationProcessor struct {
	Information string
}

func (cip ChatInformationProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	result := s.MessageResult{Type:"chat", Body:cip.Information}
	return &result
}

type ChatMessageProcessor struct {
	Storage        *ChatStorage
	MessageStorage *db.MessageHandler
	CompanyId      string
	Answer         string
}

func (cmp ChatMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	body := in.Message.Body
	userData := in.UserData
	if body != nil && userData != nil {
		user, err := cmp.Storage.GlobalUsers.GetUserById(in.From)
		if err != nil {
			log.Printf("CB ERROR after getting user %v", err)
			return DB_ERROR_RESULT
		}
		if user == nil {
			cmp.Storage.GlobalUsers.AddUser(in.From, userData.Name, userData.Phone, userData.Email)
		} else {
			cmp.Storage.GlobalUsers.UpdateUserData(in.From, userData.Name, userData.Phone, userData.Email)
		}
		cmp.Storage.SetUserCompany(in.From, cmp.CompanyId)

		r_body := *body
		res, _ := cmp.MessageStorage.StoreMessage(in.From, cmp.CompanyId, r_body, in.Message.ID)
		log.Printf("CB persist message:\n %+v", res)

		if cmp.Answer != "" {
			return &s.MessageResult{Type:"chat", Body:cmp.Answer, IsDeferred:false}
		}
		return &s.MessageResult{Type:"chat", Body:"", IsDeferred:true}
	}else {
		return &s.MessageResult{Type:"chat", Body:"Нет данных для сообщения или данных пользователя"}
	}
}

func FormChatBotContext(config configuration.ChatConfig, store *db.MainDb) *msngr.BotContext {
	result := msngr.BotContext{}
	result.Request_commands = map[string]s.RequestCommandProcessor{
		"commands":&ChatRequestProcessor{
			Commands:&[]s.OutCommand{
				s.OutCommand{
					Title:    "Информация",
					Action:   "information",
					Position: 0,
				},
			},
		},
	}
	chatStorage := NewChatStorage(store)

	result.Message_commands = map[string]s.MessageCommandProcessor{
		"information":&ChatInformationProcessor{Information:config.Information},
		"":&ChatMessageProcessor{Storage:chatStorage, CompanyId:config.CompanyId, MessageStorage:store.Messages, Answer:config.BotAnswer},
	}

	return &result
}