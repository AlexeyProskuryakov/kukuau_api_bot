package msngr

import (
	"fmt"
	"math/rand"
	"time"
)

// var shop_db = GetUserHandler()

var authorised_commands = []OutCommand{
	OutCommand{
		Title:    "Мои заказы",
		Action:   "orders_state",
		Position: 0,
	},
	OutCommand{
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
	OutCommand{
		Title:    "Выйти",
		Action:   "log_out",
		Position: 2,
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

func (cp ShopCommandsProcessor) ProcessRequest(in InPkg) ([]OutCommand, error) {
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
	return commands, nil
}

type ShopAuthoriseProcessor struct {
	DbHandlerMixin
}

func (sap ShopAuthoriseProcessor) ProcessMessage(in InPkg) (string, *[]OutCommand, error) {
	command := *in.Message.Commands
	user, password := _get_user_and_password(command[0].Form.Fields)
	if sap.Users.CheckUserPassword(user, password) {
		sap.Users.SetUserState(in.From, LOGIN)
		return "Вы авторизовались. Ура!", &authorised_commands, nil
	}
	return "Не правильные логин или пароль :(", nil, nil

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

var order_states = [3]string{"обработан", "создан", "отправлен"}

func (osp ShopOrderStateProcessor) ProcessMessage(in InPkg) (string, *[]OutCommand, error) {
	user_state, err := osp.Users.GetUserState(in.From)
	_check(err)
	if user_state == LOGIN {
		result := fmt.Sprintf("Ваш заказ с номером %v %v", rand.Int31n(10000), __choiceString(order_states[:]))
		return result, &authorised_commands, nil
	}
	return "Авторизуйтесь пожалуйста!", nil, nil
}

type SupportMessageProcessor struct{}

func (sm SupportMessageProcessor) ProcessMessage(in InPkg) (string, *[]OutCommand, error) {
	return "Ваш отзыв важен для нас, спасибо.", nil, nil
}

type ShopInformationProcessor struct{}

func (ih ShopInformationProcessor) ProcessMessage(in InPkg) (string, *[]OutCommand, error) {
	return "Покупки в тысячах проверенных магазинов! (тестовый логин: test, пароль: 123)", nil, nil
}

type ShopLogOutMessageProcessor struct {
	DbHandlerMixin
}

func (lop ShopLogOutMessageProcessor) ProcessMessage(in InPkg) (string, *[]OutCommand, error) {
	lop.Users.SetUserState(in.From, LOGOUT)
	return "Вы вышли. Ура!", &not_authorised_commands, nil //todo
}
