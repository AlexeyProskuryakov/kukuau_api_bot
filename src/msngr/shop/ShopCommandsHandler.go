package shop

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"

	d "msngr/db"
	s "msngr/structs"
)

func FormShopCommands(db *d.DbHandlerMixin) *s.BotContext {
	var ShopRequestCommands = map[string]s.RequestCommandProcessor{
		"commands": ShopCommandsProcessor{DbHandlerMixin: *db},
	}

	var ShopMessageCommands = map[string]s.MessageCommandProcessor{
		"information":     ShopInformationProcessor{},
		"authorise":       ShopAuthoriseProcessor{DbHandlerMixin: *db},
		"log_out":         ShopLogOutMessageProcessor{DbHandlerMixin: *db},
		"orders_state":    ShopOrderStateProcessor{DbHandlerMixin: *db},
		"support_message": ShopSupportMessageProcessor{},
		"balance":         ShopBalanceProcessor{},
	}

	context := s.BotContext{}
	context.Check = func() (string, bool) { return "", true }
	context.Message_commands = ShopMessageCommands
	context.Request_commands = ShopRequestCommands
	return &context
}

var authorised_commands = []s.OutCommand{
	s.OutCommand{
		Title:    "Мои заказы",
		Action:   "orders_state",
		Position: 0,
	},
	s.OutCommand{
		Title:    "Мой баланс",
		Action:   "balance",
		Position: 1,
	},
	s.OutCommand{
		Title:    "Задать вопрос",
		Action:   "support_message",
		Position: 2,
		Fixed:    true,
		Form: &s.OutForm{
			Type: "form",
			Text: "?(text)",
			Fields: []s.OutField{
				s.OutField{
					Name: "text",
					Type: "text",
					Attributes: s.FieldAttribute{
						Label:    "Текст вопроса",
						Required: true,
					},
				},
			},
		},
	},
	s.OutCommand{
		Title:    "Выйти",
		Action:   "log_out",
		Position: 3,
	},
}
var not_authorised_commands = []s.OutCommand{
	s.OutCommand{
		Title:    "Авторизоваться",
		Action:   "authorise",
		Position: 0,
		Form: &s.OutForm{
			Name: "Форма ввода данных пользователя",
			Type: "form",
			Text: "Пользователь: ?(username), пароль: ?(password)",
			Fields: []s.OutField{
				s.OutField{
					Name: "username",
					Type: "text",
					Attributes: s.FieldAttribute{
						Label:    "имя",
						Required: true,
					},
				},
				s.OutField{
					Name: "password",
					Type: "password",
					Attributes: s.FieldAttribute{
						Label:    "пароль",
						Required: true,
					},
				},
			},
		},
	},
}

func _get_user_and_password(fields []s.InField) (string, string) {
	var user, password string
	for _, field := range fields {
		if field.Name == "username" {
			user = field.Data.Value
		} else if field.Name == "password" {
			password = field.Data.Value
		}
	}
	return user, password
}

type ShopCommandsProcessor struct {
	d.DbHandlerMixin
}

func (cp ShopCommandsProcessor) ProcessRequest(in s.InPkg) s.RequestResult {
	user_state, err := cp.Users.GetUserState(in.From)
	if err != nil {
		cp.Users.AddUser(in.From, in.UserData.Phone)
	}
	commands := []s.OutCommand{}
	if user_state == d.LOGIN {
		commands = authorised_commands
	} else {
		commands = not_authorised_commands
	}
	return s.RequestResult{Commands:&commands}
}

type ShopAuthoriseProcessor struct {
	d.DbHandlerMixin
}

func (sap ShopAuthoriseProcessor) ProcessMessage(in s.InPkg) s.MessageResult {
	command := *in.Message.Commands
	user, password := _get_user_and_password(command[0].Form.Fields)

	var body string
	var commands []s.OutCommand

	if sap.Users.CheckUserPassword(user, password) {
		sap.Users.SetUserState(in.From, d.LOGIN)
		body = "Добро пожаловать в интернет магазин Desprice Markt!"
		commands = authorised_commands
	}else {
		body = "Не правильные логин или пароль :("
		commands = not_authorised_commands
	}
	return s.MessageResult{Body:body, Commands:&commands}

}

type ShopOrderStateProcessor struct {
	d.DbHandlerMixin
}

func __choiceString(choices []string) string {
	var winner string
	length := len(choices)
	rand.Seed(time.Now().Unix())
	i := rand.Intn(length)
	winner = choices[i]
	return winner
}

var order_states = [5]string{"обработан", "доставляется", "отправлен", "поступил в пункт выдачи", "в обработке"}
var order_products = [4]string{"Ноутбук Apple MacBook Air", "Электрочайник BORK K 515", "Аудиосистема Westlake Tower SM-1", "Микроволновая печь Bosch HMT85ML23"}

func (osp ShopOrderStateProcessor) ProcessMessage(in s.InPkg) s.MessageResult {
	user_state, _ := osp.Users.GetUserState(in.From)
	var result string
	var commands []s.OutCommand
	if user_state == d.LOGIN {
		result = fmt.Sprintf("Ваш заказ #%v (%v) %v.", rand.Int31n(10000), __choiceString(order_products[:]), __choiceString(order_states[:]))
		commands = authorised_commands
	} else {
		result = "Авторизуйтесь пожалуйста!"
		commands = not_authorised_commands
	}
	return s.MessageResult{Body:result, Commands:&commands}
}

type ShopSupportMessageProcessor struct {}

func contains(container string, elements []string) bool {
	container_elements := regexp.MustCompile("[a-zA-Zа-яА-Я]+").FindAllString(container, -1)
	log.Printf("SCH splitted: %v", strings.Join(container_elements, ","))
	ce_map := make(map[string]bool)
	for _, ce_element := range container_elements {
		ce_map[strings.ToLower(ce_element)] = true
	}
	result := true
	for _, element := range elements {
		_, ok := ce_map[element]
		result = result && ok
		log.Printf("SCH element: %+v, contains? : %+v => %+v", element, ok, result)
	}
	return result
}

func make_one_string(fields []s.InField) string {
	var buffer bytes.Buffer
	for _, field := range fields {
		buffer.WriteString(field.Data.Value)
		buffer.WriteString(field.Data.Text)
	}
	return buffer.String()
}

func (sm ShopSupportMessageProcessor) ProcessMessage(in s.InPkg) s.MessageResult {
	commands := *in.Message.Commands
	var body string

	if commands != nil {
		log.Printf("SCH: commands: %+v, fields: %+v", commands, commands[0].Form.Fields)
		if contains(make_one_string(commands[0].Form.Fields), []string{"где", "забрать", "заказ"}) {
			body = "Ваш заказ вы можете забрать по адресу: ул. Николаева д. 11."
		} else {
			body = "Спасибо за вопрос. Мы ответим Вам в ближайшее время."
		}
	} else {
		body = "Спасибо за вопрос. Мы ответим Вам в ближайшее время."
	}
	return s.MessageResult{Body:body}
}

type ShopInformationProcessor struct {}

func (ih ShopInformationProcessor) ProcessMessage(in s.InPkg) s.MessageResult {
	return s.MessageResult{Body:"Desprice Markt - интернет-магазин бытовой техники и электроники в Новосибирске и других городах России. Каталог товаров мировых брендов."}
}

type ShopLogOutMessageProcessor struct {
	d.DbHandlerMixin
}

func (lop ShopLogOutMessageProcessor) ProcessMessage(in s.InPkg) s.MessageResult {
	lop.Users.SetUserState(in.From, d.LOGOUT)
	return s.MessageResult{Body:"До свидания! ", Commands:&not_authorised_commands}
}

type ShopBalanceProcessor struct {
}

func (sbp ShopBalanceProcessor) ProcessMessage(in s.InPkg) s.MessageResult {
	return s.MessageResult{Body: fmt.Sprintf("Ваш баланс на %v составляет %v бонусных баллов.", time.Now().Format("01.02.2006"), rand.Int31n(1000) + 10), Commands: &authorised_commands}
}
