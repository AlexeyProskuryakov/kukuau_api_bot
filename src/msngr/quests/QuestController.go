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

func getCommands(in *s.InPkg, qs *QuestStorage, cs c.ConfigStorage) []s.OutCommand {
	var result_commands []s.OutCommand
	if state, err := qs.GetUserState(in.From, PROVIDER); err == nil && state == SUBSCRIBED {
		result_commands, _ = cs.LoadCommands(PROVIDER, SUBSCRIBED)
	} else {
		result_commands, _ = cs.LoadCommands(PROVIDER, UNSUBSCRIBED)
	}
	return result_commands
}

func (qcp *QuestCommandRequestProcessor) ProcessRequest(in *s.InPkg) *s.RequestResult {
	result_commands := getCommands(in, qcp.Storage, qcp.ConfigStorage)
	result := s.RequestResult{Commands:&result_commands}
	return &result
}

type QuestInfoMessageProcessor struct {
	Information string
}

func (qimp QuestInfoMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	return &s.MessageResult{Body:qimp.Information, Type:"chat"}
}

type QuestUnsubscribeMessageProcessor struct {
	Storage *QuestStorage
	c.ConfigStorage
}


func (qump *QuestUnsubscribeMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	log.Printf("QUESTS Want unsubscribe: %s", in.From)
	err := qump.Storage.SetUserState(in.From, UNSUBSCRIBED, PROVIDER)
	if err != nil {
		commands, _ := qump.LoadCommands(PROVIDER, SUBSCRIBED)
		return &s.MessageResult{Commands:&commands, Body:fmt.Sprintf("Что-то пошло не так. Попробуйте снова. Вот с такая ошибешка: %s", err), Type:"chat"}
	}
	commands, _ := qump.LoadCommands(PROVIDER, UNSUBSCRIBED)
	return &s.MessageResult{Commands:&commands, Body:"Теперь вы не учавствуете в квесте. \nПечаль :( ", Type:"chat"}
}

type QuestSubscribeMessageProcessor struct {
	Storage        *QuestStorage
	c.ConfigStorage
	AcceptPhrase   string
	RejectedPhrase string
	ErrorPhrase    string
}

func (qsmp *QuestSubscribeMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	log.Printf("QUESTS Want subscribe %s", in.From)
	user_state, err := qsmp.Storage.GetUserState(in.From, PROVIDER)
	var text string

	if err != nil && err != mgo.ErrNotFound {
		text = fmt.Sprintf("%s: [%v]", qsmp.ErrorPhrase, err)
		commands, _ := qsmp.LoadCommands(PROVIDER, UNSUBSCRIBED)
		return &s.MessageResult{Commands:&commands, Body:text, Type:"chat"}
	} else if err == mgo.ErrNotFound {
		qsmp.Storage.SetUserState(in.From, SUBSCRIBED, PROVIDER)
		commands, _ := qsmp.LoadCommands(PROVIDER, SUBSCRIBED)
		text = qsmp.AcceptPhrase
		return &s.MessageResult{Commands:&commands, Body:text, Type:"chat"}
	}

	commands, _ := qsmp.LoadCommands(PROVIDER, SUBSCRIBED)
	if user_state == SUBSCRIBED {
		text = qsmp.RejectedPhrase
		return &s.MessageResult{Commands:&commands, Body:text, Type:"chat"}
	} else {
		qsmp.Storage.SetUserState(in.From, SUBSCRIBED, PROVIDER)
		text = qsmp.AcceptPhrase
		return &s.MessageResult{Commands:&commands, Body:text, Type:"chat"}
	}
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
		log.Printf("QUEST user [%v] is not found because: %+v", user_id, err)
		return "", err, false
	}
	log.Printf("QUEST user [%v] is found: %+v", user_id, user_info)

	if user_info.LastKeyPosition == nil && key_info.Position == 0{
		return key_info.Description, nil, true
	} else if user_info.LastKeyPosition != nil{
		user_last_key_position := user_info.LastKeyPosition
		if (*user_last_key_position + 1) == key_info.Position{
			return key_info.Description, nil, true
		} else if utils.InS(key_info.Key, user_info.FoundKeys) {
			return "Вы уже вводили этот ключ", nil, false
		} else {
			return "Вы не можете использовать этот ключ сейчас.", nil, false
		}
	}
	return "", nil, false
}

type QuestKeyInputMessageProcessor struct {
	Storage *QuestStorage
	c.ConfigStorage

}

func (qkimp QuestKeyInputMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	var text string
	if state, err := qkimp.Storage.GetUserState(in.From, PROVIDER); err != nil {
		commands, _ := qkimp.LoadCommands(PROVIDER, SUBSCRIBED)
		return &s.MessageResult{Commands:&commands, Body:fmt.Sprintf("Упс. Ошибка: %v", err.Error()), Type:"chat"}
	} else if state != SUBSCRIBED {
		commands, _ := qkimp.LoadCommands(PROVIDER, UNSUBSCRIBED)
		return &s.MessageResult{Commands:&commands, Body:"Вы здесь быть не должны и делать это не можете.", Type:"chat"}
	}

	commands_ptr := in.Message.Commands
	if commands_ptr != nil {
		commands := *commands_ptr
		for _, command := range commands {
			if command.Action == "key_input" && command.Form.Name == "key_form" {
				for _, field := range command.Form.Fields {
					if field.Name == "code" {
						key := field.Data.Value
						log.Printf("QUESTS We have key from %v is: [%v]", in.From, key)
						descr, err, ok := ProcessKeyUserResult(in.From, key, qkimp.Storage)
						if err != nil && err != mgo.ErrNotFound {
							text = fmt.Sprintf("Внутренняя ошибка: %s.", err)
						}else if err == mgo.ErrNotFound {
							text = "Код не верный, попробуйте другой."
						} else {
							text = descr
							if ok {
								qkimp.Storage.SetUserLastKey(in.From, key, PROVIDER)
							}
						}
					}
				}
			}
		}
	}
	commands, _ := qkimp.LoadCommands(PROVIDER, SUBSCRIBED)
	mr := s.MessageResult{Commands:&commands, Body:text, Type:"chat"}
	return &mr
}

type QuestMessagePersistProcessor struct {
	c.ConfigStorage
	Storage *QuestStorage
}

var key_reg = regexp.MustCompile("^\\#[\\w\\dа-яА-Я]+")

func (qmpp QuestMessagePersistProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	commands := getCommands(in, qmpp.Storage, qmpp.ConfigStorage)
	log.Printf("QUESTS want to send simple message")
	if in.Message.Body != nil {
		//try recognise code at simple message
		pkey := in.Message.Body
		key := *pkey
		if key_reg.MatchString(key) {
			descr, err, ok := ProcessKeyUserResult(in.From, key, qmpp.Storage)
			log.Printf("QUESTS want to send key %v i have this answer for key: %v", key, descr)
			if err == nil {
				if ok {
					qmpp.Storage.SetUserLastKey(in.From, key, PROVIDER)
				}
				return &s.MessageResult{Type:"chat", Body:descr, Commands:&commands}
			} else if err == mgo.ErrNotFound {
				return &s.MessageResult{Type:"chat", Body:"Не верный ключ.", Commands:&commands}
			} else {
				return &s.MessageResult{Type:"chat", Body:fmt.Sprintf("Упс ошибочка. %v.", err.Error()), Commands:&commands}
			}
		}
		//else storing this message
		err := qmpp.Storage.StoreMessage(in.From, *in.Message.Body, time.Now())
		if err != nil {
			return &s.MessageResult{Type:"chat", Body:err.Error(), Commands:&commands}
		}
	} else {
		return &s.MessageResult{Type:"chat", Body:"Сообщения нет :( ", Commands:&commands}
	}
	return &s.MessageResult{Type:"chat", Body:"Ваше сообщение доставленно. Скоро вам ответят.", Commands:&commands}
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
		"subscribe":&QuestSubscribeMessageProcessor{Storage:qs, AcceptPhrase:qconf.AcceptPhrase, RejectedPhrase:qconf.RejectPhrase, ErrorPhrase:qconf.ErrorPhrase, ConfigStorage:cs},
		"unsubscribe":&QuestUnsubscribeMessageProcessor{Storage:qs, ConfigStorage:cs},
		"key_input":&QuestKeyInputMessageProcessor{Storage:qs, ConfigStorage:cs},
		"information":&QuestInfoMessageProcessor{Information:qconf.Info},
		"":QuestMessagePersistProcessor{Storage:qs, ConfigStorage:cs},
	}

	result.CommandsStorage = cs
	notifier := msngr.NewNotifier(conf.Main.CallbackAddr, qconf.Key)
	go Run(qconf, qs, notifier)

	return &result

}