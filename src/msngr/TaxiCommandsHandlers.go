package msngr

import (
	"errors"
	"fmt"

	"math/rand"
)

var url = "http://foo.bar.baz"
var commands_at_created_order = []Command{
	Command{
		Title:    "Отменить заказ",
		Action:   "cancel_order",
		Position: 0,
	},
}
var commands_at_not_created_order = []Command{
	Command{
		Title:    "Вызвать такси",
		Action:   "new_order",
		Position: 0,
		Form:     taxi_call_form,
	},

	Command{
		Title:    "Рассчитать цену",
		Action:   "calculate_price",
		Position: 1,
		Form:     taxi_call_form,
	},
}
var taxi_call_form = &OutForm{
	Title: "Форма вызова такси",
	Type:  "form",
	Name:  "call_taxi",
	Text:  "Откуда: ?(street_from), ?(house_from), ?(entrance). Куда: ?(street_to), ?(house_to). Когда: ?(time)",
	Fields: []OutField{
		OutField{
			Name: "street_from",
			Type: "dict",
			Attributes: FieldAttribute{
				Label:    "улица/район",
				Required: true,
				URL:      &url,
			},
		},
		OutField{
			Name: "house_from",
			Type: "text",
			Attributes: FieldAttribute{
				Label:    "дом",
				Required: true,
			},
		},
		OutField{
			Name: "entrance",
			Type: "number",
			Attributes: FieldAttribute{
				Label:    "подъезд",
				Required: false,
			},
		},
		OutField{
			Name: "street_to",
			Type: "text",
			Attributes: FieldAttribute{
				Label:    "улица/район",
				Required: true,
				URL:      &url,
			},
		},
		OutField{
			Name: "house_to",
			Type: "text",
			Attributes: FieldAttribute{
				Label:    "дом",
				Required: true,
			},
		},
		OutField{
			Name: "time",
			Type: "text",
			Attributes: FieldAttribute{
				Label:    "время",
				Required: false,
			},
		},
	},
}

var TaxiRequestCommands = map[string]RequestCommandProcessor{
	"commands": TaxiCommandsHandler{},
}

var TaxiMessageCommands = map[string]MessageCommandProcessor{
	"information":     TaxiInformationHandler{},
	"new_order":       TaxiNewOrderHandler{},
	"cancel_order":    TaxiCancelOrderHandler{},
	"calculate_price": TaxiCalculatePriceHandler{},
}

var _taxi_db = GetUserHandler()

type TaxiCommandsHandler struct{}

func (s TaxiCommandsHandler) ProcessRequest(in InPkg) ([]Command, error) {
	state := _taxi_db.GetUserState(in.From)
	if state == ORDER_CREATE {
		return commands_at_created_order, nil
	} else {
		return commands_at_not_created_order, nil
	}
}

type TaxiInformationHandler struct{}

func (ih TaxiInformationHandler) ProcessMessage(in InPkg) (string, *[]Command, error) {
	return "Срочный заказ такси в Новосибирске. Быстрая подача. Оплата наличными или картой. ", nil, nil
}

type TaxiNewOrderHandler struct{}

func (noh TaxiNewOrderHandler) ProcessMessage(in InPkg) (string, *[]Command, error) {
	state := _taxi_db.GetUserState(in.From)
	if state != ORDER_CREATE {
		var from, to, hf, ht, t string
		for _, field := range in.Message.Command.Form.Fields {
			switch fn := field.Name; fn {
			case "street_from":
				from = field.Data.Text
			case "street_to":
				to = field.Data.Text
			case "house_to":
				ht = field.Data.Value
			case "house_from":
				hf = field.Data.Value
			case "time":
				fv := field.Data.Value
				if fv == "0" {
					t = "сейчас"
				} else {
					t = fmt.Sprintf("через %v минут", rand.Int31n(10)+10)
				}
			}

		}
		_taxi_db.SetUserState(in.From, ORDER_CREATE)
		result := fmt.Sprintf("Ваш заказ создан! Поедем из %v дом %v, на %v к дому %v. Cтоймость %v рублей, машина прибудет %v", from, hf, to, ht, rand.Int31n(500)+50, t)
		return result, &commands_at_created_order, nil
	} else {
		return "Заказ уже создан!", nil, errors.New("Заказ уже создан!")
	}

}

type TaxiCancelOrderHandler struct{}

func (coh TaxiCancelOrderHandler) ProcessMessage(in InPkg) (string, *[]Command, error) {
	_taxi_db.SetUserState(in.From, ORDER_CANCELED)
	return "Ваш заказ отменен", nil, nil
}

type TaxiCalculatePriceHandler struct{}

func (cph TaxiCalculatePriceHandler) ProcessMessage(in InPkg) (string, *[]Command, error) {
	return fmt.Sprintf("Стоймость будет всего лишь %v рублей!", rand.Int31n(500)+50), nil, nil
}
