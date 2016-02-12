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
	"msngr/db"
)

const (
	ME = "me"
	BAD_KEY = "Не верный ключ."
	NOT_NEXT_KEY = "Вы не должны были найти этот ключ сейчас. Верните его на место. Ищите ключ согласно описанию."
	BAD_GROUP_INPUT = "Не могу определить группу по введеному ключу :("
	NOT_TEAM_MEMBER = "Вы не являетесь участником квеста."
)

var (
	DB_ERROR = errors.New("Ошибка на стороне базы данных")
	DB_ERROR_RESULT = &s.MessageResult{Type:"chat", Body:DB_ERROR.Error()}
	BAD_KEY_RESULT = &s.MessageResult{Type:"chat", Body:BAD_KEY}
	USER_DATA_ERROR_RESULT = &s.MessageResult{Type:"chat", Body:"Не хватает данных для сохранения сообщения :("}
)

func WRONG_TEAM_MEMBER(bad, good string) string {
	return fmt.Sprintf("Вы не являетесь участником группы %s. Вы учасник группы %s.", bad, good)
}

type QuestCommandRequestProcessor struct {
	c.ConfigStorage
	Storage *QuestStorage
}

func (qcp *QuestCommandRequestProcessor) ProcessRequest(in *s.InPkg) *s.RequestResult {
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

var marker_reg = regexp.MustCompile("^\\#[\\w\\dа-яА-Я]+\\-?(?P<team>[\\w\\da-zа-я]+)?")

func ValidateKeyBySequent(team *Team, key_info *Step, qs *QuestStorage) (string, error, bool) {
	//return description or some text for user or "" if error
	previous_key, _ := qs.GetKeyByNextKey(key_info.StartKey)
	if previous_key != nil {
		log.Printf("Q: i found previous key which have next_key == %v and: " +
		"\nit was in team founded keys? %v," +
		"\nit founded? %v" +
		"\nitfounded by this team? %v (by %v)",
			key_info.StartKey,
			utils.InS(previous_key.StartKey, team.FoundKeys),
			previous_key.Founded,
			previous_key.FoundedBy == team.Name,
			previous_key.FoundedBy,
		)
		if utils.InS(previous_key.StartKey, team.FoundKeys) && previous_key.Founded && previous_key.FoundedBy == team.Name {
			return key_info.Description, nil, true
		} else {
			return NOT_NEXT_KEY, nil, false
		}
	}
	log.Printf("Q: i not found any key which have next_key == %v and i think that it is first key in sequence", key_info)
	return key_info.Description, nil, true
}

func GetTeamNameFromKey(key string) (string, error) {
	found := marker_reg.FindStringSubmatch(key)
	if len(found) == 2 {
		return found[1], nil
	} else {
		return "", errors.New(BAD_GROUP_INPUT)
	}
}

func (qmpp QuestMessagePersistProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	if in.Message.Body != nil {
		pkey := in.Message.Body
		key := *pkey
		log.Printf("Q: Processing message %v from %v [%+v]", key, in.From, in.UserData)
		key = strings.TrimSpace(key)
		if marker_reg.MatchString(key) {
			key = strings.ToLower(key)
			log.Printf("Q: Here is key: %v", key)
			key_info, err := qmpp.Storage.GetKeyByStartKey(key)
			if err != nil {
				log.Printf("QUEST key [%v] is ERR! %v", err)
				return DB_ERROR_RESULT
			}
			if key_info == nil {
				return BAD_KEY_RESULT
			}

			team_name, err := GetTeamNameFromKey(key)
			if err != nil {
				return &s.MessageResult{Type:"chat", Body:BAD_GROUP_INPUT}
			}
			team, err := qmpp.Storage.GetTeamByName(team_name)
			if team == nil {
				team, err = qmpp.Storage.AddTeam(team_name)
				if err != nil {
					log.Printf("Q E : at adding team %v", err)
					return DB_ERROR_RESULT
				}
			}
			log.Printf("Q:Team recognised: %+v", team)
			prev_key, err := qmpp.Storage.GetKeyByNextKey(key);
			log.Printf("Q:prevkey? %+v , Err? %v", prev_key, err)
			if err != nil {
				return DB_ERROR_RESULT
			}
			member, err := qmpp.Storage.GetTeamMemberByUserId(in.From)

			if member == nil {
				if prev_key == nil {
					log.Printf("Q:Recognised register key from %v [%+v], add him to team: %v", in.From, in.UserData, team_name)
					member, err = qmpp.Storage.AddTeamMember(in.From, in.UserData.Name, in.UserData.Phone, team)
					if err != nil {
						log.Printf("Q E : at adding team member%v", err)
						return DB_ERROR_RESULT
					}
				}else {
					log.Printf("Q:Register key [%v] not recognised because we have previous key: %v, " +
					"\nQ:but member for[%v] is nil:( all in:\n%+v", key, prev_key, in.UserData, in)
					return BAD_KEY_RESULT
				}
			} else {
				if prev_key == nil {
					log.Printf("Q:will change team at member [%v]  %v -> %v", member.Name, member.TeamName, team_name)
					qmpp.Storage.SetTeamForTeamMember(team, member)
					member, err = qmpp.Storage.AddTeamMember(in.From, in.UserData.Name, in.UserData.Phone, team)
				} else if !member.Passersby && member.TeamName != "" && member.TeamSID != "" && member.TeamName != team_name {
					return &s.MessageResult{Type:"chat", Body:WRONG_TEAM_MEMBER(team_name, member.TeamName)}
				} else if member.Passersby && prev_key != nil {
					return &s.MessageResult{Type:"chat", Body:NOT_TEAM_MEMBER}
				}
			}

			if err != nil {
				log.Printf("Q E : at getting or persisting user team %v", err)
				return DB_ERROR_RESULT
			}
			descr, err, ok := ValidateKeyBySequent(team, key_info, qmpp.Storage)
			log.Printf("QUESTS want to send key %v i have this answer for key: %v, err: %v, ok? %v", key, descr, err, ok)
			if err != nil {
				log.Printf("Q E : at processing key result %v", err)
				return DB_ERROR_RESULT
			}
			if ok {
				qmpp.Storage.SetKeyFounded(key, team_name)
				_, err = qmpp.Storage.StoreMessage(team.Name, ME, key, true)
				if err != nil {
					log.Printf("Q E : at storing key as message %v", err)
					return DB_ERROR_RESULT
				}

			}
			return &s.MessageResult{Type:"chat", Body:descr, }
		} else {
			var from string
			log.Printf("Q: From %v is message: [%v]", in.From, key)
			member, _ := qmpp.Storage.GetTeamMemberByUserId(in.From)
			if member != nil {
				from = member.TeamName
				log.Printf("Q: message from member %v of group %v ", in.From, from)

			} else if in.UserData != nil {
				user_data := in.UserData
				from = in.From
				qmpp.Storage.AddPasserby(in.From, user_data.Phone, user_data.Name)
				log.Printf("Q: message from passersby %v", in.From)
			} else {
				log.Printf("Q: but %v it is not team member and not have userdata", in.From)
				return USER_DATA_ERROR_RESULT
			}
			log.Printf("Q: will storing msg:[%v] from:v to:%v as not key answer", key, from, ME)
			qmpp.Storage.StoreMessage(from, ME, key, false)
		}
	} else {
		return &s.MessageResult{Type:"chat", Body:"Сообщения нет :( ", }
	}
	return &s.MessageResult{Type:"chat", Body:"Ваше сообщение доставлено. Скоро вам ответят.", }
}

func FormQuestBotContext(conf c.Configuration, qname string, cs c.ConfigStorage, qs *QuestStorage, db *db.MainDb) *m.BotContext {
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
	notifier := msngr.NewNotifier(conf.Main.CallbackAddr, qconf.Key, db)
	go Run(qconf, qs, notifier)

	return &result
}