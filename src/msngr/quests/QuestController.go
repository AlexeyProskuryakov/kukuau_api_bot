package quests

import (
	"log"

	s "msngr/structs"
	c "msngr/configuration"
	m "msngr"

	"fmt"

	"msngr/notify"
	"regexp"
	"strings"
	"errors"
	"msngr/utils"
)

const (
	ME = "me"
	BAD_KEY = "Не верный ключ"
	NOT_SEQUENED_KEY = "Вы не должны были найти этот ключ сейчас. Верните его на место. Ищите ключ согласно описанию."
)

var (
	DB_ERROR = errors.New("Ошибка на стороне базы данных")
	DB_ERROR_RESULT = &s.MessageResult{Type:"chat", Body:DB_ERROR.Error()}
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

type QuestMessagePersistProcessor struct {
	c.ConfigStorage
	Storage *QuestStorage
}

var key_reg = regexp.MustCompile("^\\#[\\w\\dа-яА-Я]+\\-?(?P<team>[\\w\\da-zа-я]+)?")

func ProcessKeyUserResult(team *Team, key string, qs *QuestStorage) (string, error, bool) {
	//return description or some text for user or "" if error
	key_info, err := qs.GetKey(key)
	if err != nil {
		log.Printf("QUEST key [%v] is ERR! %v", err)
		return "", DB_ERROR, false
	}
	if key_info == nil {
		return BAD_KEY, nil, false
	}

	previous_key, err := qs.GetKeyByNextKey(key)
	if previous_key != nil {
		if utils.InS(previous_key.StartKey, team.FoundKeys) && previous_key.Founded && previous_key.FoundedBy == team.Name {
			return key_info.Description, nil, true
		} else {
			return NOT_SEQUENED_KEY, nil, false
		}
	}
	return key_info.Description, nil, true
}

func GetTeamNameFromKey(key string) (string, error) {
	found := key_reg.FindStringSubmatch(key)
	if len(found) == 2 {
		return found[1], nil
	} else {
		return "", errors.New("Не могу определить группу по введеному ключу :(")
	}
}

func GetOrPersistUserTeam(in *s.InPkg, team_name string, qs *QuestStorage) (*Team, error) {
	team, err := qs.GetTeamByName(team_name)
	member, err := qs.GetTeamMemberByUserId(in.From)
	if err != nil {
		return nil, err
	}
	if team == nil {
		err = qs.AddTeam(team_name)
	}
	if member == nil {
		err = qs.AddTeamMember(in.From, in.UserData.Name, in.UserData.Phone, team_name)
	}
	return team, err
}

func (qmpp QuestMessagePersistProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	if in.Message.Body != nil {
		pkey := in.Message.Body
		key := *pkey
		key = strings.TrimSpace(key)
		if key_reg.MatchString(key) {
			key = strings.ToLower(key)
			err := qmpp.Storage.StoreMessage(in.From, ME, key, true)
			if err != nil {
				log.Printf("Q E : at storing key as message %v", err)
				return DB_ERROR_RESULT
			}
			team_name, err := GetTeamNameFromKey(key)
			if err != nil {
				return &s.MessageResult{Type:"chat", Body:"Не могу определить группу по введеному ключу :(", }
			}
			team, err := GetOrPersistUserTeam(in, team_name, qmpp.Storage)
			if err != nil {
				log.Printf("Q E : at getting or persisting user team %v", err)
				return DB_ERROR_RESULT
			}

			descr, err, ok := ProcessKeyUserResult(team, key, qmpp.Storage)
			log.Printf("QUESTS want to send key %v i have this answer for key: %v, err: %v, ok? %v", key, descr, err, ok)
			if err != nil {
				log.Printf("Q E : at processing key result %v", err)
				return DB_ERROR_RESULT
			}
			if ok {
				qmpp.Storage.SetKeyFounded(key, team_name)

			}
			return &s.MessageResult{Type:"chat", Body:descr, }
		} else {
			err := qmpp.Storage.StoreMessage(in.From, ME, key, false)
			if err != nil {
				log.Printf("Q E : at storing message %v", err)
				return DB_ERROR_RESULT
			}
		}
	} else {
		return &s.MessageResult{Type:"chat", Body:"Сообщения нет :( ", }
	}
	return &s.MessageResult{Type:"chat", Body:"Ваше сообщение доставлено. Скоро вам ответят.", }
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