package voting

import (
	s "msngr/structs"
	c "msngr/configuration"
	m "msngr"
	"fmt"
	"log"
)

func FormVoteBotContext(conf c.Configuration) *m.BotContext {
	context := m.BotContext{}
	vh, _ := NewVotingHandler(conf.Main.Database.ConnString, conf.Main.Database.Name)
	context.Check = func() (string, bool) {
		if vh.Check() {
			return "", true
		}
		return "", false
	}

	context.Request_commands = map[string]s.RequestCommandProcessor{
		"commands": &VoteCommandProcessor{DictUrl: conf.Vote.DictUrl},
	}
	context.Message_commands = map[string]s.MessageCommandProcessor{
		"add_company": &VoteMessageProcessor{Storage:vh, DictUrl:conf.Vote.DictUrl, Answers:conf.Vote.Answers},
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

type VoteMessageProcessor struct {
	DictUrl string
	Storage *VotingDataHandler
	Answers []string
}

func (vmp *VoteMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
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
						Body:"Нужно ввестии хотябы имя компании и/или название услуги, а то че так-то :(",
						Commands:&commands,
					}
				}
				city, _ := command.Form.GetValue("city")
				description, _ := command.Form.GetValue("description")
				role, _ := command.Form.GetValue("role")
				log.Printf("VB Receive name: %v, service: %v, city: %v, descr: %v, role: %v", name, service, city, description, role)
				err := vmp.Storage.ConsiderCompany(name, service, city, description, userName, role)
				if err != nil {
					log.Printf("VB ERROR at conside company! %v", err)
					return &s.MessageResult{
						Body:"Упс. Что-то пошло не так.",
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