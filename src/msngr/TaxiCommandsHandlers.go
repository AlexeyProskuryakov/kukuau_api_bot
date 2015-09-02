package msngr

import (
	"errors"
	"fmt"
	"log"

	inf "msngr/infinity"

	"strconv"
	"time"
)

const (
	timeFormat = "2006-01-02 15:04:05"
)

//todo what about many inifinity apis
var url = "http://localhost:8080/_streets"
var TaxisInfinity = inf.GetInfinityAPI()

var im = inf.InfinityMixin{API: TaxisInfinity}

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
			Type: "dict",
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
			Type: "datetime",
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
	"new_order":       TaxiNewOrderHandler{InfinityMixin: im},
	"cancel_order":    TaxiCancelOrderHandler{InfinityMixin: im},
	"calculate_price": TaxiCalculatePriceHandler{InfinityMixin: im},
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

type TaxiNewOrderHandler struct {
	inf.InfinityMixin
}

func _get_time_from_timestamp(tst string) time.Time {
	i, err := strconv.ParseInt(tst, 10, 64)
	if err != nil {
		panic(err)
	}
	dst := time.Unix(i, 0)
	return dst
}

func _form_order(fields []InField) (new_order inf.NewOrder) {
	var from_info, to_info, hf, ht, when string
	for _, field := range fields {
		switch fn := field.Name; fn {
		case "street_from":
			from_info = field.Data.Value
		case "street_to":
			to_info = field.Data.Value
		case "house_to":
			ht = field.Data.Value
		case "house_from":
			hf = field.Data.Value

		case "time": //todo see time! with exceptions
			when = field.Data.Value

			if when == "0" {
				new_order.DeliveryMinutes = 0
			} else {
				new_order.DeliveryTime = _get_time_from_timestamp(when).Format(timeFormat)
			}
		}
	}
	//fucking hardcode //todo refactor
	new_order.Phone = "89537631628"
	new_order.IdService = 5001753333
	new_order.Notes = "Хочется комфортную машину"
	new_order.Attributes = [2]int64{1000113000, 1000113002}
	//end fucking hardcode

	new_order.Delivery = inf.H_get_delivery(from_info, hf)
	new_order.Destinations = []inf.Destination{inf.H_get_destination(to_info, ht)}

	return
}

func (noh TaxiNewOrderHandler) ProcessMessage(in InPkg) (string, *[]Command, error) {
	log.Println(noh.API)
	state := _taxi_db.GetUserState(in.From)
	if state != ORDER_CREATE {
		new_order := _form_order(in.Message.Command.Form.Fields)
		ans, ord_error := noh.API.NewOrder(new_order)
		if ord_error != nil {
			panic(ord_error)
		}
		_taxi_db.SetUserOrder(in.From, ans.Content.Id)
		result := fmt.Sprintf("Ваш заказ создан! Вот так: %+v и ответ таков: %+v ", new_order, ans)
		return result, &commands_at_created_order, nil
	} else {
		return "Заказ уже создан!", nil, errors.New("Заказ уже создан!")
	}
}

type TaxiCancelOrderHandler struct {
	inf.InfinityMixin
}

func (coh TaxiCancelOrderHandler) ProcessMessage(in InPkg) (string, *[]Command, error) {
	state := _taxi_db.GetUserState(in.From)
	if state == ORDER_CREATE {
		order_id, err := _taxi_db.GetUserOrderId(in.From)
		if err != nil {
			return "У вас нет заказов!", &commands_at_not_created_order, errors.New("У вас нет заказов!")
		}
		ok, info := coh.API.CancelOrder(order_id)
		if !ok {
			err_str := fmt.Sprintf("Какие-то проблемы с отменой заказа %+v", info)
			return err_str, nil, errors.New(err_str)
		}
		_taxi_db.CancelOrderId(in.From)
		return "Ваш заказ отменен", &commands_at_not_created_order, nil
	}
	return "У вас нет заказов!", &commands_at_not_created_order, errors.New("У вас нет заказов!")
}

type TaxiCalculatePriceHandler struct {
	inf.InfinityMixin
}

func (cph TaxiCalculatePriceHandler) ProcessMessage(in InPkg) (string, *[]Command, error) {
	order := _form_order(in.Message.Command.Form.Fields)
	s, details := cph.API.CalcOrderCost(order)
	log.Println(details)
	cost := strconv.Itoa(s)
	return fmt.Sprintf("Стоймость будет всего лишь %v рублей! \nА детали таковы: %v", cost, details), nil, nil
}
