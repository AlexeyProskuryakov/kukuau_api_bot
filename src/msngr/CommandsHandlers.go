package msngr

import (
	"errors"
	"fmt"
	"math/rand"
	// "strconv"
)

type requestCommandProcessor interface {
	ProcessRequest(in inPkg) ([]Command, error)
}

type messageCommandProcessor interface {
	ProcessMessage(in inPkg) (string, error)
}

var requestCommands = map[string]requestCommandProcessor{
	"commands": CommandsHandler{},
}

var messageCommands = map[string]messageCommandProcessor{
	"information":     InformationHandler{},
	"new_order":       NewOrderHandler{},
	"cancel_order":    CancelOrderHandler{},
	"calculate_price": CalculatePriceHandler{},
}

type CommandsHandler struct{}

func (s CommandsHandler) ProcessRequest(in inPkg) ([]Command, error) {
	uh := GetUserHandler()
	state := uh.GetUserState(in.From)
	if state == ORDER_CREATE {
		return []Command{
			Command{
				Title:    "Отменить заказ",
				Action:   "cancel_order",
				Position: 0,
			},
		}, nil

	} else {
		taxi_call_form := &OutForm{
			Title: "Форма вызова такси",
			Type:  "form",
			Name:  "call_taxi",
			Text:  "Откуда: ?(street_from), ?(house_from), ?(entrance). Куда: ?(street_to), ?(house_to). Когда: ?(time)",
			Fields: []OutField{
				OutField{
					Name:     "street_from",
					Required: true,
					Type:     "dict",
					Label:    "FromLabel",
					Value:    "FromValue",
					Attributes: FieldAttribute{
						Label:    "улица/район",
						Required: true,
						URL:      "http://foo.bar",
					},
				},
				OutField{
					Name:     "house_from",
					Required: true,
					Type:     "text",
					Label:    "house_from",
					Value:    "house_from",
					Attributes: FieldAttribute{
						Label:    "дом",
						Required: true,
					},
				},
				OutField{
					Name:     "entrance",
					Required: false,
					Type:     "number",
					Label:    "entrance",
					Value:    "entrance",
					Attributes: FieldAttribute{
						Label:    "подъезд",
						Required: false,
					},
				},
				OutField{
					Name:     "street_to",
					Required: true,
					Type:     "text",
					Label:    "time_label",
					Value:    "time_value",
					Attributes: FieldAttribute{
						Label:    "улица/район",
						Required: true,
						URL:      "http://foo.bar",
					},
				},
				OutField{
					Name:     "house_to",
					Required: true,
					Type:     "text",
					Label:    "house_to",
					Value:    "house_to",
					Attributes: FieldAttribute{
						Label:    "дом",
						Required: true,
					},
				},
				OutField{
					Name:     "time",
					Required: false,
					Type:     "text",
					Label:    "time",
					Value:    "time",
					Attributes: FieldAttribute{
						Label:    "время",
						Required: false,
					},
				},
			},
		}
		commands := []Command{
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
		return commands, nil
	}
}

type InformationHandler struct{}

func (ih InformationHandler) ProcessMessage(in inPkg) (string, error) {
	return "!!! This is TAXI !!! ", nil
}

type NewOrderHandler struct{}

func (noh NewOrderHandler) ProcessMessage(in inPkg) (string, error) {
	uh := GetUserHandler()
	state := uh.GetUserState(in.From)
	if state != ORDER_CREATE {
		var from, to, hf, ht, t string
		for _, field := range in.Message.Command.Form.Fields {
			switch fn := field.Name; fn {
			case "street_from":
				from = field.Data.Text
			case "street_to":
				to = field.Data.Text
			case "house_to":
				ht = field.Data.Text
			case "house_from":
				hf = field.Data.Text
			case "time":
				fv := field.Data.Value
				if fv == "0" {
					t = "сейчас"
				} else {
					t = fmt.Sprintf("через %v минут", rand.Int31n(10)+10)
				}
			}

		}

		uh.SetUserState(in.From, ORDER_CREATE)
		result := fmt.Sprintf("Ваш заказ создан! Поедем из ул %v дом %v, на %v к дому %v. Cтоймость %v, машина прибудет %v", from, hf, to, ht, rand.Int31n(500)+50, t)
		return result, nil
	} else {
		return "Заказ уже создан!", errors.New("Заказ уже создан!")
	}

}

type CancelOrderHandler struct{}

func (coh CancelOrderHandler) ProcessMessage(in inPkg) (string, error) {
	uh := GetUserHandler()
	uh.SetUserState(in.From, ORDER_CANCELED)
	return "Ваш заказ отменен", nil
}

type CalculatePriceHandler struct {
}

func (cph CalculatePriceHandler) ProcessMessage(in inPkg) (string, error) {
	return fmt.Sprintf("Стоймость будет всего лишь %v рублей!", rand.Int31n(500)+50), nil
}
