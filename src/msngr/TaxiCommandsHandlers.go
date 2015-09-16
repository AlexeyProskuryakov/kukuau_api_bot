package msngr

import (
	"errors"
	"fmt"
	"log"

	taxi "msngr/taxi"

	"strconv"
	"time"
)

const (
	timeFormat = "2006-01-02 15:04:05"
)

func FormTaxiCommands(im *taxi.InfinityMixin, db DbHandlerMixin) *BotContext {
	context := BotContext{}

	context.Check = func() (string, bool) {
		var ok bool
		var detail string
		ok = im.API.IsConnected()
		log.Printf("CHECK api: %+v, ok: %v", im.API, ok)
		if !ok {
			detail = "Ошибка в подключении к сервису"
		}
		return detail, ok
	}

	context.Request_commands = map[string]RequestCommandProcessor{
		"commands": TaxiCommandsProcessor{DbHandlerMixin: db},
	}

	context.Message_commands = map[string]MessageCommandProcessor{
		"information":      TaxiInformationProcessor{DbHandlerMixin: db},
		"new_order":        TaxiNewOrderProcessor{InfinityMixin: *im, DbHandlerMixin: db},
		"cancel_order":     TaxiCancelOrderProcessor{InfinityMixin: *im, DbHandlerMixin: db},
		"calculate_price":  TaxiCalculatePriceProcessor{InfinityMixin: *im},
		"feedback":         TaxiFeedbackProcessor{InfinityMixin: *im, DbHandlerMixin: db},
		"write_dispatcher": SupportMessageProcessor{},
	}

	return &context
}

func FormNotification(order_id int64, state int, ohm DbHandlerMixin, carCache *taxi.CarsCache) *OutPkg {
	order_wrapper := ohm.Orders.GetByOrderId(order_id)
	car_id := order_wrapper.OrderObject.IDCar
	car_info := carCache.CarInfo(car_id)
	var commands *[]OutCommand
	var text string
	switch state := order_wrapper.OrderState; state {
	case 2:
		text = fmt.Sprintf("Вам назначен %v %v c номером %v, время подачи %v.", car_info.Color, car_info.Model, car_info.Number, get_time_after(5*time.Minute, "15:04"))
	case 4:
		text = "Машина на месте. Приятной Вам поездки!"
	case 7:
		text = "Заказ выполнен! Спасибо что воспользовались услугами нашей компании."
		commands = &commands_for_order_feedback
	}

	if text != "" {
		out := OutPkg{To: order_wrapper.Whom, Message: &OutMessage{ID: genId(), Type: "chat", Body: text, Commands: commands}}
		return &out
	}
	return nil
}

func TaxiOrderWatch(db DbHandlerMixin, im taxi.InfinityMixin, carsCache *taxi.CarsCache, n *Notifier) {
	for {
		api_orders := im.API.Orders()
		// log.Printf("OW api have %v orders", len(api_orders))
		for _, api_order := range api_orders {
			db_order_state := db.Orders.GetState(api_order.ID)
			// log.Printf("OW state of %+v is: %v\n", order, order_state)
			if db_order_state == -1 {
				log.Printf("OW order %+v is not present in system :(\n", api_order)
				continue
			}
			if api_order.State != db_order_state {
				log.Printf("OW state of %+v \nis updated (api: %v != db: %v)", api_order, api_order.State, db_order_state)
				err := db.Orders.SetState(api_order.ID, api_order.State, &api_order)
				if err != nil {
					log.Printf("for order %+v can not update status %+v", api_order.ID, api_order.State)
					continue
				}
				notification_data := FormNotification(api_order.ID, api_order.State, db, carsCache)
				if notification_data != nil {
					n.Notify(*notification_data)
				}
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}

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
}

var commands_for_order_feedback = []OutCommand{
	OutCommand{
		Title:    "Отзыв о поездке",
		Action:   "feedback",
		Position: 0,
		Form: &OutForm{
			Type: "form",
			Text: "?(text)",
			Fields: []OutField{
				OutField{
					Name: "text",
					Type: "text",
					Attributes: FieldAttribute{
						Label:    "Ваш отзыв",
						Required: true,
					},
				},
			},
		},
	},
	OutCommand{
		Title:    "Вызвать такси",
		Action:   "new_order",
		Position: 1,
		Repeated: true,
		Form:     taxi_call_form,
	},
	// OutCommand{
	// 	Title:    "Рассчитать цену",
	// 	Action:   "calculate_price",
	// 	Position: 2,
	// 	Form:     taxi_call_form,
	// },
}

var commands_at_not_created_order = []OutCommand{
	OutCommand{
		Title:    "Вызвать такси",
		Action:   "new_order",
		Position: 0,
		Repeated: true,
		Form:     taxi_call_form,
	},
	// OutCommand{
	// 	Title:    "Рассчитать цену",
	// 	Action:   "calculate_price",
	// 	Position: 1,
	// 	Form:     taxi_call_form,
	// },
}

// var feedback_form = &OutForm{
// 	Title: "Форма крутоты",
// 	Type:  "form",
// 	Name:  "feedback_form",
// 	Text:  "?(yes) ?(no)",
// 	Fields: []OutField{
// 		OutField{
// 			Name: "yes",
// 			Type: "text",
// 			Attributes: FieldAttribute{
// 				Label:    "ДА :)",
// 				Required: false,
// 			},
// 		},
// 		OutField{
// 			Name: "no",
// 			Type: "text",
// 			Attributes: FieldAttribute{
// 				Label:    "Нет :(",
// 				Required: false,
// 			},
// 		},
// 	},
// }

var taxi_call_form = &OutForm{
	Title: "Форма вызова такси",
	Type:  "form",
	Name:  "call_taxi",
	Text:  "Откуда: ?(street_from), ?(house_from), ?(entrance). Куда: ?(street_to), ?(house_to).",
	Fields: []OutField{
		OutField{
			Name: "street_from",
			Type: "dict",
			Attributes: FieldAttribute{
				Label:    "улица",
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
				Label:    "улицa",
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
		// OutField{
		// 	Name: "time",
		// 	Type: "datetime",
		// 	Attributes: FieldAttribute{
		// 		Label:    "время",
		// 		Required: false,
		// 	},
		// },
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
		if order_wrapper.OrderState == taxi.ORDER_PAYED && time.Now().Sub(order_wrapper.When) < time.Hour && order_wrapper.Feedback == "" {
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
	phone, err := _get_phone(in)
	if err != nil {
		return nil, err
	}
	cp.Users.AddUser(in.From, *phone)
	result := FormCommands(in.From, cp.DbHandlerMixin)
	return *result, nil
}

type TaxiInformationProcessor struct {
	DbHandlerMixin
}

func (ih TaxiInformationProcessor) ProcessMessage(in InPkg) (string, *[]OutCommand, error) {
	return "Срочный заказ такси в Новосибирске. Быстрая подача. Оплата наличными или картой. ", FormCommands(in.From, ih.DbHandlerMixin), nil
}

func _get_time_from_timestamp(tst string) time.Time {
	i, err := strconv.ParseInt(tst, 10, 64)
	_check(err)
	dst := time.Unix(i, 0)
	return dst
}

func _form_order(fields []InField) (new_order taxi.NewOrder) {
	var from_info, to_info, hf, ht string
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
			// case "time": //todo see time! with exceptions
			// 	when = field.Data.Value
			// 	log.Println("!time of order: ", when)
			// 	if when == "0" || when == "" {
			// 		new_order.DeliveryMinutes = 0
			// 	} else {
			// 		new_order.DeliveryTime = _get_time_from_timestamp(when).Format(timeFormat)
			// 	}
		}
	}
	//fucking hardcode //todo refactor
	new_order.IdService = 5001753333
	new_order.Notes = "KUKU-AU"
	new_order.Attributes = [2]int64{1000113000, 1000113002}
	//end fucking hardcode

	new_order.Delivery = taxi.H_get_delivery(from_info, hf)
	new_order.Destinations = []taxi.Destination{taxi.H_get_destination(to_info, ht)}

	return
}

type TaxiNewOrderProcessor struct {
	taxi.InfinityMixin
	DbHandlerMixin
}

func _get_phone(in InPkg) (phone *string, err error) {
	if user_data := in.UserData; user_data != nil {
		if phone := user_data.Phone; phone != "" {
			return &phone, nil
		}
	}
	return nil, errors.New("no row at message.UserData.Phone")
}

func (nop TaxiNewOrderProcessor) ProcessMessage(in InPkg) (string, *[]OutCommand, error) {
	order_wrapper := nop.Orders.GetByOwner(in.From)
	log.Printf("NOP order_wrapper: %+v\n", order_wrapper)
	if order_wrapper == nil || taxi.IsOrderNotAvaliable(order_wrapper.OrderState) {
		commands := *in.Message.Commands
		new_order := _form_order(commands[0].Form.Fields)
		phone, err := _get_phone(in)
		if err != nil {
			uwrpr, err := nop.Users.GetById(in.From)
			if err != nil {
				return "Error of user data", nil, errors.New("You must provide phone of user (at user data in this message or in `commands` message)")
			} else {
				phone = &(uwrpr.Phone)
			}

		} else {
			new_order.Phone = *phone
		}
		ans, ord_error := nop.API.NewOrder(new_order)
		if ord_error != nil {
			panic(ord_error)
		}
		nop.Orders.AddOrder(ans.Content.Id, in.From)
		text := fmt.Sprintf("Ваш заказ создан! Стоймость поездки составит %+v рублей.", ans.Content.Cost)
		return text, &commands_at_created_order, nil
	} else {
		return "Заказ уже создан!", &commands_at_created_order, errors.New("Заказ уже создан!")
	}
}

type TaxiCancelOrderProcessor struct {
	taxi.InfinityMixin
	DbHandlerMixin
}

func (cop TaxiCancelOrderProcessor) ProcessMessage(in InPkg) (string, *[]OutCommand, error) {
	order_wrapper := cop.Orders.GetByOwner(in.From)
	if order_wrapper != nil {
		cop.API.CancelOrder(order_wrapper.OrderId)
		return "Ваш заказ отменен!", &commands_at_not_created_order, nil
		//todo check at infinity orders
	}
	return "У вас нет заказов!", FormCommands(in.From, cop.DbHandlerMixin), errors.New("У вас нет заказов!")
}

type TaxiCalculatePriceProcessor struct {
	taxi.InfinityMixin
}

func (cpp TaxiCalculatePriceProcessor) ProcessMessage(in InPkg) (string, *[]OutCommand, error) {
	commands := *in.Message.Commands
	order := _form_order(commands[0].Form.Fields)
	s, _ := cpp.API.CalcOrderCost(order)
	cost := strconv.Itoa(s)
	return fmt.Sprintf("Стоймость будет всего лишь %v рублей!", cost), nil, nil
}

type TaxiFeedbackProcessor struct {
	taxi.InfinityMixin
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
	f := taxi.Feedback{IdOrder: order_id, Rating: rating, Notes: fdbk}
	fp.API.Feedback(f)

	return "Спасибо! Ваш отзыв очень важен для нас:)", FormCommands(in.From, fp.DbHandlerMixin), nil
}
