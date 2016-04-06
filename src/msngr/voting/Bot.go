package voting

import (
	s "msngr/structs"
	c "msngr/configuration"
	m "msngr"
	"fmt"
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
		"add_company": &VoteMessageProcessor{Storage:vh},
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
						Name: "user_role",
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
	Storage *VotingDataHandler
}

func (vmp *VoteMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	result := &s.MessageResult{}
	return result
}