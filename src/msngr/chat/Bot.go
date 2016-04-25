package chat

import (
	"msngr"
	"msngr/db"
	s "msngr/structs"

	m "msngr"
	"log"
)

type ChatRequestProcessor struct {
	Commands *[]s.OutCommand
}

func (crp *ChatRequestProcessor)ProcessRequest(in *s.InPkg) *s.RequestResult {
	result := s.RequestResult{Commands:crp.Commands}
	return &result
}

type ChatInformationProcessor struct {
	CompanyId     string
	ConfigStorage *db.ConfigurationStorage
}

func (cip ChatInformationProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	chatConfig, err := cip.ConfigStorage.GetChatConfig(cip.CompanyId)
	if err != nil {
		log.Printf("CB ERROR at getting configuration %v", err)
		return &s.MessageResult{Type:"chat", Body:err.Error()}
	}
	return &s.MessageResult{Type:"chat", Body:chatConfig.Information}
}

type ChatMessageProcessor struct {
	Storage        *ChatStorage
	MessageStorage *db.MessageHandler
	CompanyId      string
	ConfigStorage  *db.ConfigurationStorage
}

func (cmp ChatMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	body := in.Message.Body
	userData := in.UserData
	if body != nil && userData != nil {
		user, err := cmp.Storage.GlobalUsers.GetUserById(in.From)
		if err != nil {
			log.Printf("CB ERROR after getting user %v", err)
			return m.DB_ERROR_RESULT
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

		conf, err := cmp.ConfigStorage.GetChatConfig(cmp.CompanyId)
		if err != nil {
			log.Printf("CB ERROR retrieve configuration %v", err)
			return &s.MessageResult{Type:"chat", Body:"", IsDeferred:true}
		}
		for _, aa := range conf.AutoAnswers {
			if aa.After == 0 {
				return &s.MessageResult{Type:"chat", Body:aa.Text, IsDeferred:false}
			}
		}
		return &s.MessageResult{Type:"chat", Body:"", IsDeferred:true}
	} else {
		return &s.MessageResult{Type:"chat", Body:"Нет данных для сообщения или данных пользователя"}
	}
}

func FormChatBotContext(store *db.MainDb, confStore *db.ConfigurationStorage, companyId string) *msngr.BotContext {
	result := msngr.BotContext{}
	result.RequestProcessors = map[string]s.RequestCommandProcessor{
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

	result.MessageProcessors = map[string]s.MessageCommandProcessor{
		"information":&ChatInformationProcessor{CompanyId:companyId, ConfigStorage: confStore},
		"":&ChatMessageProcessor{Storage:chatStorage, CompanyId:companyId, MessageStorage:store.Messages, ConfigStorage:confStore},
	}

	return &result
}