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
	"msngr/utils"
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

func IsSubscribedKey(key string, qs *QuestStorage) (bool, error) {
	key_info, err := qs.GetKeyInfo(key)
	if err != nil {
		return false, err
	}
	log.Printf("QUESTS checking is key was subscribed... key: %+v, is first? %+v", key_info, key_info.IsFirst)
	return key_info.IsFirst, nil
}

func ProcessKeyUserResult(user_id, key string, qs *QuestStorage) (string, error, bool) {
	//return description or some text for user or "" if error
	key_info, err := qs.GetKeyInfo(key)
	if err != nil {
		return "", err, false
	}
	log.Printf("QUEST key [%v] is found: %+v", key, key_info)

	user_info, err := qs.GetUserInfo(user_id, PROVIDER)
	if err != nil {
		log.Printf("QUESTS user [%v] is not found because: %+v", user_id, err)
		return "", err, false
	}
	log.Printf("QUESTS user [%v] is found: %+v", user_id, user_info)

	if user_info.LastKey == nil && key_info.IsFirst {
		return key_info.Description, nil, true
	} else if user_info.LastKey != nil {
		user_last_key_p := user_info.LastKey
		user_last_key := *user_last_key_p
		previous_key, err := qs.GetKeyInfo(user_last_key)
		if err != nil {
			return "", err, false
		}
		if utils.InS(key_info.Key, user_info.FoundKeys) {
			return "Вы уже вводили этот ключ", nil, false
		}
		if previous_key.NextKey == nil || *previous_key.NextKey == key {
			return key_info.Description, nil, true
		}

		return "Вы не можете использовать этот ключ сейчас.", nil, false

	}
	return "", nil, false
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
		//try recognise code at simple message
		pkey := in.Message.Body
		key := *pkey
		key = strings.TrimSpace(key)
		if key_reg.MatchString(key) {
			key = strings.ToLower(key)
			if is_first, err := IsSubscribedKey(key, qmpp.Storage); is_first && err == nil {
				qmpp.Storage.SetUserState(in.From, SUBSCRIBED, PROVIDER)
			}
			descr, err, ok := ProcessKeyUserResult(in.From, key, qmpp.Storage)
			log.Printf("QUESTS want to send key %v i have this answer for key: %v, err: %v, ok? %v", key, descr, err, ok)
			if err == nil {
				if ok {
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
		//		"subscribe":&QuestSubscribeMessageProcessor{Storage:qs, AcceptPhrase:qconf.AcceptPhrase, RejectedPhrase:qconf.RejectPhrase, ErrorPhrase:qconf.ErrorPhrase, ConfigStorage:cs},
		//		"unsubscribe":&QuestUnsubscribeMessageProcessor{Storage:qs, ConfigStorage:cs},
		//		"key_input":&QuestKeyInputMessageProcessor{Storage:qs, ConfigStorage:cs},
		"information":&QuestInfoMessageProcessor{Information:qconf.Info},
		"":QuestMessagePersistProcessor{Storage:qs, ConfigStorage:cs},
	}

	result.CommandsStorage = cs
	notifier := msngr.NewNotifier(conf.Main.CallbackAddr, qconf.Key)
	go Run(qconf, qs, notifier)

	return &result

}