package msngr

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

var shop_db = GetUserHandler()

var authorised_commands = []Command{
	Command{
		Title:    "Мои заказы",
		Action:   "orders_state",
		Position: 0,
	},
	Command{
		Title:    "Написать в тех. поддержку",
		Action:   "support_message",
		Position: 1,
		Fixed:    true,
		Form: &OutForm{
			Type: "form",
			Text: "?(text)",
			Fields: []OutField{
				OutField{
					Name: "text",
					Type: "text",
					Attributes: FieldAttribute{
						Label:    "Текст сообщения",
						Required: true,
					},
				},
			},
		},
	},
	Command{
		Title:    "Выйти",
		Action:   "log_out",
		Position: 2,
	},
}
var not_authorised_commands = []Command{
	Command{
		Title:    "Авторизоваться",
		Action:   "authorise",
		Position: 0,
		Form: &OutForm{
			Name: "Форма ввода данных пользователя",
			Type: "form",
			Text: "Пользователь: ?(username), пароль ?(password)",
			Fields: []OutField{
				OutField{
					Name: "username",
					Type: "text",
					Attributes: FieldAttribute{
						Label:    "имя пользователя",
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

var ShopRequestCommands = map[string]RequestCommandProcessor{
	"commands": ShopCommandsHandler{},
}

var ShopMessageCommands = map[string]MessageCommandProcessor{
	"information":     ShopInformationHandler{},
	"authorise":       ShopAuthoriseHandler{},
	"orders_state":    ShopOrderStateHandler{},
	"support_message": ShopSupportMessageHandler{},
	"log_out":         ShopLogOutMessageHandler{},
}

type ShopCommandsHandler struct{}

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

func (ch ShopCommandsHandler) ProcessRequest(in InPkg) ([]Command, error) {
	user_state := shop_db.GetUserState(in.From)
	commands := []Command{}
	if user_state == USER_AUTHORISED {
		commands = authorised_commands
	} else {
		commands = not_authorised_commands
	}
	return commands, nil
}

type ShopAuthoriseHandler struct{}

func (s ShopAuthoriseHandler) ProcessMessage(in InPkg) (string, *[]Command, error) {
	user, password := _get_user_and_password(in.Message.Command.Form.Fields)
	log.Println("user, pass: ", user, password)
	if shop_db.CheckUserPassword(user, password) {
		log.Println("was auth...")
		shop_db.SetUserState(in.From, USER_AUTHORISED)
		return "Вы авторизовались. Ура!", &authorised_commands, nil
	}
	return "Не правильные логин или пароль :(", nil, nil

}

type ShopOrderStateHandler struct{}

func __choiceString(choices []string) string {
	var winner string
	length := len(choices)
	rand.Seed(time.Now().Unix())
	i := rand.Intn(length)
	winner = choices[i]
	return winner
}

var order_states = [3]string{"обработан", "создан", "отправлен"}

func (os ShopOrderStateHandler) ProcessMessage(in InPkg) (string, *[]Command, error) {
	if shop_db.GetUserState(in.From) == USER_AUTHORISED {
		result := fmt.Sprintf("Ваш заказ с номером %v %v", rand.Int31n(10000), __choiceString(order_states[:]))
		return result, &authorised_commands, nil
	}
	return "Авторизуйтесь пожалуйста!", nil, nil
}

type ShopSupportMessageHandler struct{}

func (sm ShopSupportMessageHandler) ProcessMessage(in InPkg) (string, *[]Command, error) {
	return "Ваш отзыв важен для нас, спасибо.", nil, nil
}

type ShopInformationHandler struct{}

func (ih ShopInformationHandler) ProcessMessage(in InPkg) (string, *[]Command, error) {
	return "Покупки в тысячах проверенных магазинов! (тестовый логин: test, пароль: 123)", nil, nil
}

type ShopLogOutMessageHandler struct{}

func (lo ShopLogOutMessageHandler) ProcessMessage(in InPkg) (string, *[]Command, error) {
	shop_db.RemoveUserState(in.From)
	return "Вы вышли. Ура!", &not_authorised_commands, nil //todo
}
