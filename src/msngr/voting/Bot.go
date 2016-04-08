package voting

import (
	s "msngr/structs"
	c "msngr/configuration"
	m "msngr"
	cns "msngr/console"
	d  "msngr/db"
	"fmt"
	"log"
	"gopkg.in/mgo.v2"
	"msngr/quests"
	"time"
	"strings"
)

const (
	ROLE_CLIENT = "клиент"
	ALREADY_ADDED_MSG = "Вы уже добавляли такую компанию (услугу), добавьте другую."
	ROLE_CLIENT_MSG = "Ваша заявка на регистрацию принята! В течении дня наш менеджер с Вами свяжется."
	ROLE_CLIENT_INFO = "Мы добавили Вашу Компанию для голосования другим Пользователям, количество проголосовавших: %v. Приглашайте своих сотрудников и друзей к голосованию! Добавляйте Компании для решения своих дел!"
	NEED_NAME_OR_SERVICE = "Нужно ввестии хотябы имя компании и/или название услуги, а то че так-то :("
	DEFAULT_INFO_MESSAGE = "Расскажите нам какую компанию или услугу вы хотели бы видеть в нашем мессенджере и мы ее добавим."
	ERROR_MESSAGE = "Упс. Что-то пошло не так."
)

func FormVoteBotContext(conf c.Configuration, db_handler *d.MainDb) *m.BotContext {
	context := m.BotContext{}
	vh, _ := NewVotingHandler(conf.Main.Database.ConnString, conf.Main.Database.Name)
	context.Check = func() (string, bool) {
		if vh.Check() {
			return "", true
		}
		return "", false
	}
	qs := quests.NewQuestStorage(conf.Main.Database.ConnString, conf.Main.Database.Name)
	context.Request_commands = map[string]s.RequestCommandProcessor{
		"commands": &VoteCommandProcessor{DictUrl: conf.Vote.DictUrl},
	}
	context.Message_commands = map[string]s.MessageCommandProcessor{
		"add_company": &VoteConsiderCompanyProcessor{Storage:vh, DictUrl:conf.Vote.DictUrl, Answers:conf.Vote.Answers, MainStorage:db_handler},
		"information": &VoteInformationProcessor{Storage:vh, DictUrl:conf.Vote.DictUrl},
		"":cns.ConsoleMessageProcessor{MainDb:*db_handler, QuestStorage:qs},
	}
	return &context
}

func getCommands(dictUrlPrefix string) []s.OutCommand {
	nameSearchUrl := fmt.Sprintf("%v/name", dictUrlPrefix)
	serviceSearchUrl := fmt.Sprintf("%v/service", dictUrlPrefix)
	citySearchUrl := fmt.Sprintf("%v/city", dictUrlPrefix)
	roleSearchUrl := fmt.Sprintf("%v/role", dictUrlPrefix)

	commands := []s.OutCommand{
		s.OutCommand{
			Title: "Добавить компанию",
			Action: "add_company",
			Position:0,
			Form: &s.OutForm{
				Title: "Форма добавления компании",
				Type:  "form",
				Name:  "add_company_form",
				Text:  "Название компании: ?(name);\nНазвание услуги: ?(service);\nГород: ?(city)\nВаш статус в компании: ?(user_role);\nОписание и/или комментарий: ?(description).",
				Fields: []s.OutField{
					s.OutField{
						Name: "name",
						Type: "dict",
						Attributes: s.FieldAttribute{
							Label:    "",
							Required: false,
							URL:      &nameSearchUrl,
						},
					},
					s.OutField{
						Name: "service",
						Type: "dict",
						Attributes: s.FieldAttribute{
							Required: false,
							URL:      &serviceSearchUrl,
						},
					},
					s.OutField{
						Name: "city",
						Type: "dict",
						Attributes: s.FieldAttribute{
							Required: false,
							URL:      &citySearchUrl,
						},
					},
					s.OutField{
						Name: "description",
						Type: "text",
						Attributes: s.FieldAttribute{
							Required: false,
						},
					},
					s.OutField{
						Name: "role",
						Type: "dict",
						Attributes: s.FieldAttribute{
							Required: false,
							URL:      &roleSearchUrl,
						},
					},
				},
			},
		},
	}
	return commands
}

type VoteCommandProcessor struct {
	DictUrl string
}

func (vcp *VoteCommandProcessor) ProcessRequest(in *s.InPkg) *s.RequestResult {
	commands := getCommands(vcp.DictUrl)
	return &s.RequestResult{Commands:&commands}
}

type VoteInformationProcessor struct {
	Storage *VotingDataHandler
	DictUrl string
}

func (vip *VoteInformationProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	commands := getCommands(vip.DictUrl)
	userName := in.From
	cm, _ := vip.Storage.GetLastVote(userName)
	if cm != nil {
		if voter := cm.GetVoter(userName); voter != nil && voter.Role == ROLE_CLIENT {
			return &s.MessageResult{Body:fmt.Sprintf(ROLE_CLIENT_INFO, cm.VoteInfo.VoteCount), Commands:&commands}
		}
	}
	cms, err := vip.Storage.GetUserVotes(userName)
	if err == mgo.ErrNotFound || len(cms) == 0 {
		return &s.MessageResult{Body:DEFAULT_INFO_MESSAGE, Commands:&commands}
	} else if err != nil {
		log.Printf("VB Error at get user votes")
		return &s.MessageResult{Body:ERROR_MESSAGE, Commands:&commands}
	}else {
		text := "За ваши компании проголосовало:\n"
		for _, cm := range cms {
			text = fmt.Sprintf("%v%v (%v) в %v: %v человек;\n", text, cm.Name, cm.Service, cm.City, cm.VoteInfo.VoteCount)
		}
		return &s.MessageResult{Body:text, Commands:&commands}
	}
}

type VoteConsiderCompanyProcessor struct {
	DictUrl     string
	Storage     *VotingDataHandler
	MainStorage *d.MainDb
	Answers     []string
}

func prepareMessageText(role, phone string, cmp *CompanyModel) string {
	text := "Я"
	if role != "" {
		text = fmt.Sprintf("%v, являясь %vом, хочу добавить", text, strings.ToLower(role))
	}else {
		text = fmt.Sprintf("%v хочу добавить", text)
	}
	if cmp.Name != "" {
		text = fmt.Sprintf("%v компанию. Мой телефон: %v", text, phone)
	}else {
		text = fmt.Sprintf("%v услугу.", text)
	}
	return text
}

func (vmp *VoteConsiderCompanyProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	result := &s.MessageResult{}
	userName := in.From

	if in.Message.Commands != nil {
		cmdsPtr := in.Message.Commands
		for _, command := range *cmdsPtr {
			if command.Action == "add_company" {
				commands := getCommands(vmp.DictUrl)
				name, nOk := command.Form.GetValue("name")
				service, sOk := command.Form.GetValue("service")
				if !nOk && !sOk {
					return &s.MessageResult{
						Body:NEED_NAME_OR_SERVICE,
						Commands:&commands,
					}
				}
				city, _ := command.Form.GetValue("city")
				description, _ := command.Form.GetValue("description")
				role, _ := command.Form.GetValue("role")
				log.Printf("VB Receive name: %v, service: %v, city: %v, descr: %v, role: %v", name, service, city, description, role)
				cmp, err := vmp.Storage.ConsiderCompany(name, city, service, description, userName, role)
				if err != nil {
					if _, ok := err.(AlreadyConsider); ok {
						return &s.MessageResult{
							Body:ALREADY_ADDED_MSG,
							Commands:&commands,
						}
					}else {
						log.Printf("VB ERROR at conside company! %v", err)
						return &s.MessageResult{
							Body:ERROR_MESSAGE,
							Commands:&commands,
						}
					}
				}
				vmp.MainStorage.Users.StoreUser(userName, in.UserData.Name, in.UserData.Phone, in.UserData.Email)
				vmp.MainStorage.Messages.StoreMessageObject(d.MessageWrapper{
					MessageID:in.Message.ID,
					From:userName,
					To:"me",
					Body:prepareMessageText(role, in.UserData.Phone, cmp),
					Unread:1,
					NotAnswered:1,
					Time:time.Now(),
					TimeStamp:time.Now().Unix(),
					TimeFormatted: time.Now().Format(time.Stamp),
					Attributes:[]string{"vote"},
					AdditionalData:cmp.ToMap(),
				})
				if err != nil {
					log.Printf("VB ERROR when storing message")
				}
				if role == ROLE_CLIENT {
					return &s.MessageResult{
						Body:ROLE_CLIENT_MSG,
						Commands:&commands,
					}
				}
				votes, err := vmp.Storage.GetUserVotes(userName)

				var text string
				if err != nil {
					log.Printf("VB ERROR at getting user votes")
					text = vmp.Answers[0]
				}else {
					if len(votes) >= len(vmp.Answers) {
						text = vmp.Answers[len(vmp.Answers) - 1]
					}else {
						text = vmp.Answers[len(votes) - 1]
					}
				}
				return &s.MessageResult{
					Body:text,
					Commands:&commands,
				}

			}
		}
	}
	return result
}