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

var DictUrl string

var commands_at_created_order = []OutCommand{
	OutCommand{
		Title:    "Отменить заказ",
		Action:   "cancel_order",
		Position: 0,
	},
	OutCommand{
		Title:    "Написать диспетчеру",
		Action:   "write_dispatcher",
		Position: 1,
		Fixed:    true,
	},
}

var commands_for_order_feedback = []OutCommand{
	OutCommand{
		Title:    "Было круто?",
		Action:   "feedback",
		Position: 0,
		Form:     feedback_form,
	},
	OutCommand{
		Title:    "Вызвать такси",
		Action:   "new_order",
		Position: 1,
		Repeated: true,
		Form:     taxi_call_form,
	},
	OutCommand{
		Title:    "Рассчитать цену",
		Action:   "calculate_price",
		Position: 2,
		Form:     taxi_call_form,
	},
}

var commands_at_not_created_order = []OutCommand{
	OutCommand{
		Title:    "Вызвать такси",
		Action:   "new_order",
		Position: 0,
		Repeated: true,
		Form:     taxi_call_form,
	},
	OutCommand{
		Title:    "Рассчитать цену",
		Action:   "calculate_price",
		Position: 1,
		Form:     taxi_call_form,
	},
}

var feedback_form = &OutForm{
	Title: "Форма крутоты",
	Type:  "form",
	Name:  "feedback_form",
	Text:  "?(yes) ?(no)",
	Fields: []OutField{
		OutField{
			Name: "yes",
			Type: "text",
			Attributes: FieldAttribute{
				Label:    "ДА :)",
				Required: false,
			},
		},
		OutField{
			Name: "no",
			Type: "text",
			Attributes: FieldAttribute{
				Label:    "Нет :(",
				Required: false,
			},
		},
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
				URL:      &DictUrl,
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
				URL:      &DictUrl,
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

func in(p int, a []int) bool {
	for _, v := range a {
		if p == v {
			return true
		}
	}
	return false
}

func FormCommands(username string, ohm DbHandlerMixin) *[]OutCommand {
	order_wrapper := ohm.Orders.GetByOwner(username)
	if order_wrapper != nil {
		if order_wrapper.OrderState == inf.ORDER_PAYED && time.Now().Sub(order_wrapper.When) < time.Hour && order_wrapper.Feedback == "" {
			return &commands_for_order_feedback
		} else if in(order_wrapper.OrderState, []int{7, 8, 9, 13, 15}) {
			return &commands_at_not_created_order
		}
		return &commands_at_created_order
	} else {
		log.Printf("F_C not orders for user %+v\n", username)
		return &commands_at_not_created_order
	}
}

type TaxiCommandsProcessor struct {
	DbHandlerMixin
}

func (cp TaxiCommandsProcessor) ProcessRequest(in InPkg) ([]OutCommand, error) {
	result := FormCommands(in.From, cp.DbHandlerMixin)
	return *result, nil
}

type TaxiInformationProcessor struct{}

func (ih TaxiInformationProcessor) ProcessMessage(in InPkg) (string, *[]OutCommand, error) {
	return "Срочный заказ такси в Новосибирске. Быстрая подача. Оплата наличными или картой. ", nil, nil
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
			log.Println("!time of order: ", when)
			if when == "0" || when == "" {
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

type TaxiNewOrderProcessor struct {
	inf.InfinityMixin
	DbHandlerMixin
}

func (nop TaxiNewOrderProcessor) ProcessMessage(in InPkg) (string, *[]OutCommand, error) {
	order_wrapper := nop.Orders.GetByOwner(in.From)
	if order_wrapper == nil || inf.IsOrderNotAvaliable(order_wrapper.OrderState) {
		commands := *in.Message.Commands
		new_order := _form_order(commands[0].Form.Fields)
		ans, ord_error := nop.API.NewOrder(new_order)
		if ord_error != nil {
			panic(ord_error)
		}
		nop.Orders.AddOrder(ans.Content.Id, in.From)
		//todo check at infinity orders
		text := fmt.Sprintf("Ваш заказ создан! Стоймость всего лишь %+v рублей", ans.Content.Cost)
		return text, &commands_at_created_order, nil
	} else {
		return "Заказ уже создан!", &commands_at_created_order, errors.New("Заказ уже создан!")
	}
}

type TaxiCancelOrderProcessor struct {
	inf.InfinityMixin
	DbHandlerMixin
}

func (cop TaxiCancelOrderProcessor) ProcessMessage(in InPkg) (string, *[]OutCommand, error) {
	order_wrapper := cop.Orders.GetByOwner(in.From)
	if order_wrapper != nil {
		ok, info := cop.API.CancelOrder(order_wrapper.OrderId)
		if ok {
			return fmt.Sprintf("Ваш заказ отменен (%+v)", info), FormCommands(in.From, cop.DbHandlerMixin), nil
		} else {
			err_str := fmt.Sprintf("Какие-то проблемы с отменой заказа %+v", info) //todo send alarm command
			return err_str, FormCommands(in.From, cop.DbHandlerMixin), errors.New(err_str)
		}
		//todo check at infinity orders
	}
	return "У вас нет заказов!", FormCommands(in.From, cop.DbHandlerMixin), errors.New("У вас нет заказов!")
}

type TaxiCalculatePriceProcessor struct {
	inf.InfinityMixin
}

func (cpp TaxiCalculatePriceProcessor) ProcessMessage(in InPkg) (string, *[]OutCommand, error) {
	commands := *in.Message.Commands
	order := _form_order(commands[0].Form.Fields)
	s, details := cpp.API.CalcOrderCost(order)
	cost := strconv.Itoa(s)
	return fmt.Sprintf("Стоймость будет всего лишь %v рублей! \nА детали таковы: %v", cost, details), nil, nil
}

type TaxiFeedbackProcessor struct {
	inf.InfinityMixin
	DbHandlerMixin
}

func _get_feedback(fields []InField) (int, string) {
	for _, v := range fields {
		if v.Name == "yes" && v.Data.Value == "" && v.Data.Text == "" {
			return 0, "Not cool"
		}
	}
	return 10, "Cool"
}

func (fp TaxiFeedbackProcessor) ProcessMessage(in InPkg) (string, *[]OutCommand, error) {
	commands := *in.Message.Commands
	rating, fdbk := _get_feedback(commands[0].Form.Fields)
	order_id := fp.Orders.SetFeedback(in.From, fdbk)
	f := inf.Feedback{IdOrder: order_id, Rating: rating, Notes: fdbk}
	fp.API.Feedback(f)

	return "Спасибо! Ваш отзыв очень важен для нас:)", FormCommands(in.From, fp.DbHandlerMixin), nil
}
