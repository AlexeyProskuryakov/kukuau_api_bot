package ruposts


import (
	s "msngr/structs"
	m "msngr"
	"log"
	"text/template"
	"bytes"
	"fmt"
)
var req = "<последние 14 цифр>"
var form_for_tracking = &s.OutForm{
	Title: "Форма запроса информации о посылке",
	Type:  "form",
	Name:  "tracking_form",
	Text:  "Номер отправления: ?(code)",
	Fields: []s.OutField{
		s.OutField{
			Name: "code",
			Type: "number",
			Attributes: s.FieldAttribute{
				Label:    "<последние 14 цифр>",
				Required: true,
				EmptyText: &req,
			},
		},
	},
}

var out_g_commands = &[]s.OutCommand{
	s.OutCommand{
		Title:    "Поиск посылок по почтовому идентификатору",
		Action:   "tracking",
		Position: 0,
		Repeated: true,
		Form:     form_for_tracking,
	},
}


/*
info	Объект	Содержит поля:
code - номер идентификатора
name - наименование отправления
destination - адрес назначения: id, county, index, adress
weight - вес отправления в граммах
category - общие категории отправления: info, rank, mark, type
weight - финансовые категории отправления: payment, value, weight, insurance, air, rate
latest - информация о последней операции
operations	Объект	Содержит поля:
date - дата в формате "день месяц год, часы минуты"
dateiso - дата в формате ISO 8601
timestamp - дата в формате Unix Timestamp
adress - адрес регистрации: index, description
operation - наименование операции
attr - описание операции
 */
type RuPostCommandsProcessor struct {

}
const LETTER = `Посылка № {{.Info.Code}} {{.Info.Name}} в {{.Info.Destinaiton}}\n
весом {{.Info.Weight}} гр. \n
Имела следующие операции: {{.Operations}}
`
var LETTER_TEMPLATE = template.Must(template.New("post_leter").Parse(LETTER))

func (rpcp RuPostCommandsProcessor) ProcessRequest(in *s.InPkg) *s.RequestResult {
	result := s.RequestResult{Commands:out_g_commands}
	return &result
}
type RuPostTrackingProcessor struct {
	Url string
}
func (rptp RuPostTrackingProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	commands_ptr := in.Message.Commands
	if commands_ptr != nil {
		commands := *commands_ptr
		for _, command := range commands {
			if command.Action == "tracking" && command.Form.Name == "tracking_form" {
				for _, field := range command.Form.Fields {
					if field.Name == "code" {
						code := field.Data.Value
						result, err := Load(code, rptp.Url)
						if err != nil {
							return s.ErrorMessageResult(err, out_g_commands)
						}
						var text string
						if result.ResponseId < 0 {
							text = fmt.Sprintf("Ошибка в почте № %v (%v), попробуйте как-нибудь по-другому.", result.ResponseId, result.Message)
						} else if result.ResponseId == 0 {
							text = "Нет результатов у такого почтового номера, попробуйте какой-нибудь другой."
						} else {
							var wrtr bytes.Buffer
							err = LETTER_TEMPLATE.ExecuteTemplate(&wrtr, "post_letter", result)
							if err != nil {
								log.Printf("err in execut templatE:%v", err)
							}
							text = wrtr.String()
						}
						mr := s.MessageResult{Commands:out_g_commands, Body:text, Type:"chat"}
						return &mr
					}
				}
			} else {
				log.Printf("RU POST TP WARNING: i have command with verififcatio fail: %+v \n with form: %+v", command, command.Form)
			}

		}
	}
	return nil
}

func FormRPBotContext(conf m.Configuration) *s.BotContext {
	result := s.BotContext{}
	result.Request_commands = map[string]s.RequestCommandProcessor{
		"commands":RuPostCommandsProcessor{},
	}
	result.Message_commands = map[string]s.MessageCommandProcessor{
		"tracking":RuPostTrackingProcessor{Url:conf.RuPost.ExternalUrl},
	}
	return &result
}