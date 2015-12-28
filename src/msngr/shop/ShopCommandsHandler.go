package shop

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"

	d "msngr/db"
	s "msngr/structs"
	u "msngr/utils"
	c "msngr/configuration"
	"errors"
	"gopkg.in/mgo.v2"
)

const (
	SHOP_STATE_KEY = "shop"

)

func FormShopCommands(db *d.DbHandlerMixin, config *c.ShopConfig) *s.BotContext {
	var ShopRequestCommands = map[string]s.RequestCommandProcessor{
		"commands": ShopCommandsProcessor{DbHandlerMixin: *db},
	}

	var ShopMessageCommands = map[string]s.MessageCommandProcessor{
		"information":     ShopInformationProcessor{Info:config.Info},
		"authorise":       ShopLogInMessageProcessor{DbHandlerMixin: *db},
		"log_out":         ShopLogOutMessageProcessor{DbHandlerMixin: *db},
		"orders_state":    ShopOrderStateProcessor{DbHandlerMixin: *db},
		"support_message": ShopSupportMessageProcessor{},
		"balance":         ShopBalanceProcessor{},
	}

	context := s.BotContext{}
	context.Message_commands = ShopMessageCommands
	context.Request_commands = ShopRequestCommands
	context.Check = func() (string, bool) {
		if !db.Check() {
			return "Ошибка в подключении к БД попробуйте позже", false
		}
		return "All ok!", true


	}
	return &context
}

var AUTH_COMMANDS = []s.OutCommand{
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
var NOT_AUTH_COMMANDS = []s.OutCommand{
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

func (cp ShopCommandsProcessor) ProcessRequest(in *s.InPkg) *s.RequestResult {
	user_state, err := cp.Users.GetUserMultiplyState(in.From, SHOP_STATE_KEY)
	if err == mgo.ErrNotFound {
		user_data := in.UserData
		if user_data != nil && in.UserData.Phone != "" {
			phone := in.UserData.Phone
			cp.Users.AddUser(in.From, phone)
		}
	}
	commands := []s.OutCommand{}
	if user_state == d.LOGIN {
		commands = AUTH_COMMANDS
	} else {
		commands = NOT_AUTH_COMMANDS
	}
	return &s.RequestResult{Commands:&commands}
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

func (osp ShopOrderStateProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	user_state, err := osp.Users.GetUserMultiplyState(in.From, SHOP_STATE_KEY)
	if err != nil && err != mgo.ErrNotFound {
		return s.ErrorMessageResult(err, &NOT_AUTH_COMMANDS)
	}

	var result string
	var commands []s.OutCommand
	if user_state == d.LOGIN {
		result = fmt.Sprintf("Ваш заказ #%v (%v) %v.", rand.Int31n(10000), __choiceString(order_products[:]), __choiceString(order_states[:]))
		commands = AUTH_COMMANDS
	} else {
		result = "Авторизуйтесь пожалуйста!"
		commands = NOT_AUTH_COMMANDS
	}
	return &s.MessageResult{Body:result, Commands:&commands, Type:"chat"}
}

type ShopSupportMessageProcessor struct{}

func make_one_string(fields []s.InField) string {
	var buffer bytes.Buffer
	for _, field := range fields {
		buffer.WriteString(field.Data.Value)
		buffer.WriteString(" ")
		buffer.WriteString(field.Data.Text)
	}
	return buffer.String()
}

func (sm ShopSupportMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	commands := *in.Message.Commands
	var body string
	input := make_one_string(commands[0].Form.Fields)
	if commands != nil {
		if u.Contains(input, []string{"где", "забрать", "заказ"}) {
			body = "Ваш заказ вы можете забрать по адресу: ул. Николаева д. 11."
		} else {
			body = "Спасибо за вопрос. Мы ответим Вам в ближайшее время."
		}
	} else {
		body = "Спасибо за вопрос. Мы ответим Вам в ближайшее время."
	}
	u.SaveToFile(fmt.Sprintf("\n%v | %v", input, in.From), "shop_revue.txt")
	return &s.MessageResult{Body:body, Type:"chat"}
}

type ShopInformationProcessor struct {
	Info string
}

func (ih ShopInformationProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	info := ih.Info
	if info == "" {
		info = "Desprice Markt - интернет-магазин бытовой техники и электроники в Новосибирске и других городах России. Каталог товаров мировых брендов."
	}
	return &s.MessageResult{Body:info, Type:"chat"}
}

type ShopLogOutMessageProcessor struct {
	d.DbHandlerMixin
}

type ShopLogInMessageProcessor struct {
	d.DbHandlerMixin
}

func (sap ShopLogInMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	command := *in.Message.Commands
	user, password := _get_user_and_password(command[0].Form.Fields)
	if user == "" || password == "" {
		return s.ErrorMessageResult(errors.New("Не могу извлечь логин и (или) пароль"), &NOT_AUTH_COMMANDS)
	}

	check, err := sap.Users.CheckUserPassword(user, password)
	if err != nil && err != mgo.ErrNotFound {
		return s.ErrorMessageResult(err, &NOT_AUTH_COMMANDS)
	}

	var body string
	var commands []s.OutCommand

	if check {
		sap.Users.SetUserMultiplyState(in.From, SHOP_STATE_KEY, d.LOGIN)
		body = "Добро пожаловать в интернет магазин Desprice Markt!"
		commands = AUTH_COMMANDS
	}else {
		body = "Не правильные логин или пароль :("
		commands = NOT_AUTH_COMMANDS
	}
	return &s.MessageResult{Body:body, Commands:&commands, Type:"chat"}

}

func (lop ShopLogOutMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	err := lop.Users.SetUserMultiplyState(in.From, SHOP_STATE_KEY, d.LOGOUT)
	if err != nil && err != mgo.ErrNotFound {
		return s.ErrorMessageResult(err, &NOT_AUTH_COMMANDS)
	}
	return &s.MessageResult{Body:"До свидания! ", Commands:&NOT_AUTH_COMMANDS, Type:"chat"}
}

type ShopBalanceProcessor struct {
}

func (sbp ShopBalanceProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	return &s.MessageResult{Body: fmt.Sprintf("Ваш баланс на %v составляет %v бонусных баллов.", time.Now().Format("01.02.2006"), rand.Int31n(1000) + 10), Commands: &AUTH_COMMANDS, Type:"chat"}
}
