package msngr

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

func FormShopCommands(db DbHandlerMixin) *BotContext {
	var ShopRequestCommands = map[string]RequestCommandProcessor{
		"commands": ShopCommandsProcessor{DbHandlerMixin: db},
	}

	var ShopMessageCommands = map[string]MessageCommandProcessor{
		"information":     ShopInformationProcessor{},
		"authorise":       ShopAuthoriseProcessor{DbHandlerMixin: db},
		"log_out":         ShopLogOutMessageProcessor{DbHandlerMixin: db},
		"orders_state":    ShopOrderStateProcessor{DbHandlerMixin: db},
		"support_message": SupportMessageProcessor{},
		"balance":         ShopBalanceProcessor{},
	}

	context := BotContext{}
	context.Check = func() (string, bool) { return "", true }
	context.Message_commands = ShopMessageCommands
	context.Request_commands = ShopRequestCommands
	return &context
}

var authorised_commands = []OutCommand{
	OutCommand{
		Title:    "Мои заказы",
		Action:   "orders_state",
		Position: 0,
	},
	OutCommand{
		Title:    "Мой баланс",
		Action:   "balance",
		Position: 1,
	},
	OutCommand{
		Title:    "Задать вопрос",
		Action:   "support_message",
		Position: 2,
		Fixed:    true,
		Form: &OutForm{
			Type: "form",
			Text: "?(text)",
			Fields: []OutField{
				OutField{
					Name: "text",
					Type: "text",
					Attributes: FieldAttribute{
						Label:    "Текст вопроса",
						Required: true,
					},
				},
			},
		},
	},
	OutCommand{
		Title:    "Выйти",
		Action:   "log_out",
		Position: 3,
	},
}
var not_authorised_commands = []OutCommand{
	OutCommand{
		Title:    "Авторизоваться",
		Action:   "authorise",
		Position: 0,
		Form: &OutForm{
			Name: "Форма ввода данных пользователя",
			Type: "form",
			Text: "Пользователь: ?(username), пароль: ?(password)",
			Fields: []OutField{
				OutField{
					Name: "username",
					Type: "text",
					Attributes: FieldAttribute{
						Label:    "имя",
						Required: true,
					},
				},
				OutField{
					Name: "password",
					Type: "password",
					Attributes: FieldAttribute{
						Label:    "пароль",
						Required: true,
					},
				},
			},
		},
	},
}

func _get_user_and_password(fields []InField) (string, string) {
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
	DbHandlerMixin
}

func (cp ShopCommandsProcessor) ProcessRequest(in InPkg) RequestResult {
	user_state, err := cp.Users.GetUserState(in.From)
	if err != nil {
		cp.Users.AddUser(in.From, in.UserData.Phone)
	}
	commands := []OutCommand{}
	if user_state == LOGIN {
		commands = authorised_commands
	} else {
		commands = not_authorised_commands
	}
	return RequestResult{Commands:&commands}
}

type ShopAuthoriseProcessor struct {
	DbHandlerMixin
}

func (sap ShopAuthoriseProcessor) ProcessMessage(in InPkg) MessageResult {
	command := *in.Message.Commands
	user, password := _get_user_and_password(command[0].Form.Fields)

	var body string
	var commands []OutCommand

	if sap.Users.CheckUserPassword(user, password) {
		sap.Users.SetUserState(in.From, LOGIN)
		body = "Добро пожаловать в интернет магазин Desprice Markt!"
		commands = authorised_commands
	}else {
		body = "Не правильные логин или пароль :("
		commands = not_authorised_commands
	}
	return MessageResult{Body:body, Commands:&commands}

}

type ShopOrderStateProcessor struct {
	DbHandlerMixin
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

func (osp ShopOrderStateProcessor) ProcessMessage(in InPkg) MessageResult {
	user_state, err := osp.Users.GetUserState(in.From)
	_check(err)
	var result string
	var commands []OutCommand
	if user_state == LOGIN {
		result = fmt.Sprintf("Ваш заказ #%v (%v) %v.", rand.Int31n(10000), __choiceString(order_products[:]), __choiceString(order_states[:]))
		commands = authorised_commands
	} else {
		result = "Авторизуйтесь пожалуйста!"
		commands = not_authorised_commands
	}
	return MessageResult{Body:result, Commands:&commands}
}

type SupportMessageProcessor struct {}

func contains(container string, elements []string) bool {
	container_elements := regexp.MustCompile("[a-zA-Zа-яА-Я]+").FindAllString(container, -1)
	log.Printf("SCH splitted: %v", strings.Join(container_elements, ","))
	// container_elements = strings.Fields(container)
	// log.Printf("SCH splitted: %v", strings.Join(container_elements, ","))
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

func make_one_string(fields []InField) string {
	var buffer bytes.Buffer
	for _, field := range fields {
		buffer.WriteString(field.Data.Value)
		buffer.WriteString(field.Data.Text)
	}
	return buffer.String()
}

func (sm SupportMessageProcessor) ProcessMessage(in InPkg) MessageResult {
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
	return MessageResult{Body:body}
}

type ShopInformationProcessor struct {}

func (ih ShopInformationProcessor) ProcessMessage(in InPkg) MessageResult {
	return MessageResult{Body:"Desprice Markt - интернет-магазин бытовой техники и электроники в Новосибирске и других городах России. Каталог товаров мировых брендов."}
}

type ShopLogOutMessageProcessor struct {
	DbHandlerMixin
}

func (lop ShopLogOutMessageProcessor) ProcessMessage(in InPkg) MessageResult {
	lop.Users.SetUserState(in.From, LOGOUT)
	return MessageResult{Body:"До свидания! ", Commands:&not_authorised_commands}
}

type ShopBalanceProcessor struct {
}

func (sbp ShopBalanceProcessor) ProcessMessage(in InPkg) MessageResult {
	return MessageResult{Body: fmt.Sprintf("Ваш баланс на %v составляет %v бонусных баллов.", time.Now().Format("01.02.2006"), rand.Int31n(1000) + 10), Commands: &authorised_commands}
}
