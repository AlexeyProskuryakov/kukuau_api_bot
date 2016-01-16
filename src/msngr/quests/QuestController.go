package quests

import (
	"log"
	"time"

	s "msngr/structs"
	c "msngr/configuration"
	m "msngr"

	"fmt"
	"gopkg.in/mgo.v2"
	"msngr/notify"
	"regexp"
	"strings"
)

const (
	SUBSCRIBED = "subscribed"
	UNSUBSCRIBED = "unsubscribed"
	PROVIDER = "quests"
)

type QuestCommandRequestProcessor struct {
	c.ConfigStorage
	Storage *QuestStorage
}


func (qcp *QuestCommandRequestProcessor) ProcessRequest(in *s.InPkg) *s.RequestResult {
	result := s.RequestResult{Commands:&[]s.OutCommand{}}
	return &result
}

type QuestInfoMessageProcessor struct {
	Information string
}

func (qimp QuestInfoMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	return &s.MessageResult{Body:qimp.Information, Type:"chat"}
}

func ProcessKeyUserResult(user_id, key string, qs *QuestStorage) (string, error, bool) {
	//return description or some text for user or "" if error
	key_info, err := qs.GetKeyInfo(key)
	if err != nil {
		log.Printf("QUEST key [%v] is ERR! %v", err)
		return "", err, false
	}
	user_info, err := qs.GetUserInfo(user_id, PROVIDER)
	if err != nil {
		log.Printf("QUESTS key [%v] user [%v] is NOT found because: %+v", user_id, err)
		return "", err, false
	}
	log.Printf("QUESTS proces key result: \nuser [%v] is founded: %+v \nand key [%v] is founded: %+v", user_id, user_info, key, key_info)
	return key_info.Description, nil, true
}

type QuestMessagePersistProcessor struct {
	c.ConfigStorage
	Storage *QuestStorage
}

var key_reg = regexp.MustCompile("^\\#[\\w\\dа-яА-Я]+")

func (qmpp QuestMessagePersistProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	commands := []s.OutCommand{}//getCommands(in, qmpp.Storage, qmpp.ConfigStorage)
	log.Printf("QUESTS want to send simple message")
	if in.Message.Body != nil {
		pkey := in.Message.Body
		key := *pkey
		key = strings.TrimSpace(key)

		if _, err := qmpp.Storage.GetUserInfo(in.From, PROVIDER); err == mgo.ErrNotFound {
			var name, email, phone string
			if in.UserData != nil {
				name, email, phone = in.UserData.Name, in.UserData.Email, in.UserData.Phone
			}
			log.Printf("Adding user [%v] [%v] [%v]", name, email, phone)
			qmpp.Storage.AddUser(in.From, name, email, phone, UNSUBSCRIBED, PROVIDER)
		}

		if key_reg.MatchString(key) {
			key = strings.TrimSpace(strings.ToLower(key))
			descr, err, ok := ProcessKeyUserResult(in.From, key, qmpp.Storage)
			log.Printf("QUESTS want to send key %v i have this answer for key: %v, err: %v, ok? %v", key, descr, err, ok)
			if err == nil {
				if ok {
					qmpp.Storage.SetUserState(in.From, SUBSCRIBED, PROVIDER)
					qmpp.Storage.SetUserLastKey(in.From, key, PROVIDER)
					qmpp.Storage.StoreMessage(in.From, key, time.Now(), true)
				}
				return &s.MessageResult{Type:"chat", Body:descr, Commands:&commands}
			} else if err == mgo.ErrNotFound {
				return &s.MessageResult{Type:"chat", Body:"Неверный ключ.", Commands:&commands}
			} else {
				return &s.MessageResult{Type:"chat", Body:fmt.Sprintf("Упс ошибочка. %v.", err.Error()), Commands:&commands}
			}
		}
		//else storing this message
		err := qmpp.Storage.StoreMessage(in.From, *in.Message.Body, time.Now(), false)
		if err != nil {
			return &s.MessageResult{Type:"chat", Body:err.Error(), Commands:&commands}
		}
	} else {
		return &s.MessageResult{Type:"chat", Body:"Сообщения нет :( ", Commands:&commands}
	}
	return &s.MessageResult{Type:"chat", Body:"Ваше сообщение доставлено. Скоро вам ответят.", Commands:&commands}
}

func FormQuestBotContext(conf c.Configuration, qname string, cs c.ConfigStorage, qs *QuestStorage) *m.BotContext {
	result := m.BotContext{}
	qconf, ok := conf.Quests[qname]
	if !ok {
		panic(fmt.Sprintf("Quest configuration with name %v is not exist :(", qname))
	}

	result.Request_commands = map[string]s.RequestCommandProcessor{
		"commands":&QuestCommandRequestProcessor{Storage:qs, ConfigStorage:cs},
	}

	result.Message_commands = map[string]s.MessageCommandProcessor{
		"information":&QuestInfoMessageProcessor{Information:qconf.Info},
		"":QuestMessagePersistProcessor{Storage:qs, ConfigStorage:cs},
	}

	result.CommandsStorage = cs
	notifier := msngr.NewNotifier(conf.Main.CallbackAddr, qconf.Key)
	go Run(qconf, qs, notifier)

	return &result

}