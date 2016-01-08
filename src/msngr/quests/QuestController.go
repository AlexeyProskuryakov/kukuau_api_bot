package quests

import (
	"log"
	"time"

	s "msngr/structs"
	c "msngr/configuration"
	m "msngr"
	"msngr/db"

	"fmt"
	"gopkg.in/mgo.v2"
	"msngr/notify"
	"regexp"
)

const (
	QUEST_STATE_KEY = "quest"
	SUBSCRIBED = "subscribed"
	UNSUBSCRIBED = "unsubscribed"

	PROVIDER = "quests"
)

type QuestCommandRequestProcessor struct {
	db.MainDb
	c.ConfigStorage
}

func getCommands(in *s.InPkg, db db.MainDb, cs c.ConfigStorage) []s.OutCommand {
	var result_commands []s.OutCommand
	if state, err := db.Users.GetUserMultiplyState(in.From, QUEST_STATE_KEY); err == nil {
		if state == SUBSCRIBED {
			result_commands, _ = cs.LoadCommands(PROVIDER, SUBSCRIBED)
		} else if state == UNSUBSCRIBED {
			result_commands, _ = cs.LoadCommands(PROVIDER, UNSUBSCRIBED)
		}
	} else {
		result_commands, _ = cs.LoadCommands(PROVIDER, UNSUBSCRIBED)
	}
	return result_commands
}

func (qcp *QuestCommandRequestProcessor) ProcessRequest(in *s.InPkg) *s.RequestResult {
	result_commands := getCommands(in, qcp.MainDb, qcp.ConfigStorage)
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
	db.MainDb
	c.ConfigStorage
}


func (qump *QuestUnsubscribeMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	log.Printf("QUESTS Want unsubscribe: %s", in.From)
	err := qump.Users.SetUserMultiplyState(in.From, QUEST_STATE_KEY, UNSUBSCRIBED)
	if err != nil {
		commands, _ := qump.LoadCommands(PROVIDER, SUBSCRIBED)
		return &s.MessageResult{Commands:&commands, Body:fmt.Sprintf("Что-то пошло не так. Попробуйте снова. Вот с такая ошибешка: %s", err), Type:"chat"}
	}
	commands, _ := qump.LoadCommands(PROVIDER, UNSUBSCRIBED)
	return &s.MessageResult{Commands:&commands, Body:"Теперь вы не учавствуете в квесте. \nПечаль :( ", Type:"chat"}
}

type QuestSubscribeMessageProcessor struct {
	db.MainDb
	c.ConfigStorage
	AcceptPhrase   string
	RejectedPhrase string
	ErrorPhrase    string
}

func (qsmp *QuestSubscribeMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	log.Printf("QUESTS Want subscribe %s", in.From)
	user, err := qsmp.Users.GetUserById(in.From)
	var text string
	if err != nil && err != mgo.ErrNotFound {
		text = fmt.Sprintf("%s: [%v]", qsmp.ErrorPhrase, err)
		commands, _ := qsmp.LoadCommands(PROVIDER, UNSUBSCRIBED)
		return &s.MessageResult{Commands:&commands, Body:text, Type:"chat"}
	}
	if user != nil {
		if state, ok := user.GetStateValue(QUEST_STATE_KEY); ok && state == SUBSCRIBED {
			text = qsmp.RejectedPhrase
			commands, _ := qsmp.LoadCommands(PROVIDER, SUBSCRIBED)
			return &s.MessageResult{Commands:&commands, Body:text, Type:"chat"}
		} else {
			qsmp.Users.SetUserMultiplyState(in.From, QUEST_STATE_KEY, SUBSCRIBED)
			text = qsmp.AcceptPhrase
			commands, _ := qsmp.LoadCommands(PROVIDER, SUBSCRIBED)
			return &s.MessageResult{Commands:&commands, Body:text, Type:"chat"}
		}
	} else {
		qsmp.Users.SetUserMultiplyState(in.From, QUEST_STATE_KEY, SUBSCRIBED)
		text = qsmp.AcceptPhrase
		commands, _ := qsmp.LoadCommands(PROVIDER, SUBSCRIBED)
		return &s.MessageResult{Commands:&commands, Body:text, Type:"chat"}
	}
}

type QuestKeyInputMessageProcessor struct {
	db.MainDb
	c.ConfigStorage
	DataStorage *QuestStorage
}

func (qkimp QuestKeyInputMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	var text string
	if state, err := qkimp.Users.GetUserMultiplyState(in.From, QUEST_STATE_KEY); err != nil || state != SUBSCRIBED {
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
						descr, err := qkimp.DataStorage.GetDescription(key)
						if err != nil && err != mgo.ErrNotFound {
							text = fmt.Sprintf("Внутренняя ошибка: %s.", err)
						}else if err == mgo.ErrNotFound {
							text = "Код не верный, попробуйте другой."
						} else {
							text = descr
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
	db.MainDb
	c.ConfigStorage
	DataStorage *QuestStorage
}

var key_reg = regexp.MustCompile("^\\#[\\w\\dа-яА-Я]+")

func (qmpp QuestMessagePersistProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	commands := getCommands(in, qmpp.MainDb, qmpp.ConfigStorage)
	log.Printf("QUESTS want to send simple message")
	if in.Message.Body != nil {
		//try recognise code at simple message
		pkey := in.Message.Body
		key := *pkey
		if key_reg.MatchString(key){
			descr, err := qmpp.DataStorage.GetDescription(key)
			log.Printf("QUESTS want to send key %v i have this answer for key: %v", key, descr)
			if err == nil{
				return &s.MessageResult{Type:"chat", Body:descr, Commands:&commands}
			} else if err == mgo.ErrNotFound {
				return &s.MessageResult{Type:"chat", Body:"Не верный ключ.", Commands:&commands}
			} else {
				return &s.MessageResult{Type:"chat", Body:fmt.Sprintf("Упс ошибочка. %v.", err.Error()), Commands:&commands}
			}
		}
		//else storing this message
		err := qmpp.DataStorage.StoreMessage(in.From, *in.Message.Body, time.Now())
		if err != nil {
			return &s.MessageResult{Type:"chat", Body:err.Error(), Commands:&commands}
		}
	} else {
		return &s.MessageResult{Type:"chat", Body:"Сообщения нет :( ", Commands:&commands}
	}
	return &s.MessageResult{Type:"chat", Body:"Ваше сообщение доставленно. Скоро вам ответят.", Commands:&commands}
}

func FormQuestBotContext(conf c.Configuration, qname string, db_handler *db.MainDb, cs c.ConfigStorage, qs *QuestStorage) *m.BotContext {
	result := m.BotContext{}
	qconf, ok := conf.Quests[qname]
	if !ok {
		panic(fmt.Sprintf("Quest configuration with name %v is not exist :(", qname))
	}

	result.Request_commands = map[string]s.RequestCommandProcessor{
		"commands":&QuestCommandRequestProcessor{MainDb:*db_handler, ConfigStorage:cs},
	}

	result.Message_commands = map[string]s.MessageCommandProcessor{
		"subscribe":&QuestSubscribeMessageProcessor{MainDb:*db_handler, AcceptPhrase:qconf.AcceptPhrase, RejectedPhrase:qconf.RejectPhrase, ErrorPhrase:qconf.ErrorPhrase, ConfigStorage:cs},
		"unsubscribe":&QuestUnsubscribeMessageProcessor{MainDb:*db_handler, ConfigStorage:cs},
		"key_input":&QuestKeyInputMessageProcessor{MainDb:*db_handler, ConfigStorage:cs, DataStorage:qs},
		"information":&QuestInfoMessageProcessor{Information:qconf.Info},
		"":QuestMessagePersistProcessor{MainDb:*db_handler, ConfigStorage:cs, DataStorage:qs},
	}

	result.CommandsStorage = cs
	notifier := msngr.NewNotifier(conf.Main.CallbackAddr, qconf.Key)
	go Run(qconf, qs, notifier)

	return &result

}