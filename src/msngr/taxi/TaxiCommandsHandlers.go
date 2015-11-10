package taxi

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
	s "msngr/structs"
	d "msngr/db"
	u "msngr/utils"
	c "msngr/configuration"
	"gopkg.in/mgo.v2"
)

const (
	timeFormat = "2006-01-02 15:04:05"

)

func FormTaxiBotContext(im *ExternalApiMixin, db_handler *d.DbHandlerMixin, tc c.TaxiConfig, ah *GoogleAddressHandler) *s.BotContext {

	context := s.BotContext{}

	context.Check = func() (detail string, ok bool) {
		ok = im.API.IsConnected()
		if !ok {
			detail = "Ошибка в подключении к сервису попробуйте позже"
		} else {
			return db_handler.Check()
		}
		return detail, ok
	}

	context.Commands = GetCommands(tc.DictUrl)
	context.Name = tc.Name

	context.Request_commands = map[string]s.RequestCommandProcessor{
		"commands": &TaxiCommandsProcessor{DbHandlerMixin: *db_handler, context: &context},
	}

	log.Printf("ADDRESS SUPPLIER: %+v", ah.ExternalAddressSupplier)
	context.Message_commands = map[string]s.MessageCommandProcessor{
		"information":      &TaxiInformationProcessor{DbHandlerMixin: *db_handler, context:&context, information:&(tc.Information.Text)},
		"new_order":        &TaxiNewOrderProcessor{ExternalApiMixin: *im, DbHandlerMixin: *db_handler, context:&context, AddressHandler:ah},
		"cancel_order":     &TaxiCancelOrderProcessor{ExternalApiMixin: *im, DbHandlerMixin: *db_handler, context:&context, alert_phone:tc.Information.Phone},
		"calculate_price":  &TaxiCalculatePriceProcessor{ExternalApiMixin: *im, context:&context, AddressHandler:ah},
		"feedback":         &TaxiFeedbackProcessor{ExternalApiMixin: *im, DbHandlerMixin: *db_handler, context:&context},
		"write_dispatcher": &TaxiSupportMessageProcessor{},
	}

	context.Settings = make(map[string]interface{})
	context.Settings["not_send_price"] = tc.Api.NotSendPrice

	log.Printf("For %+v will not send price? %+v\nAll settings is: %+v", tc.Name, tc.Api.NotSendPrice, context.Settings)

	return &context
}

var not_point = string("не указан")

func GetCommands(dictUrl string) map[string]*[]s.OutCommand {
	result := make(map[string]*[]s.OutCommand)

	var taxi_call_form = &s.OutForm{
		Title: "Форма вsызова такси",
		Type:  "form",
		Name:  "call_taxi",
		Text:  "Откуда: ?(street_from), ?(house_from), ?(entrance). Куда: ?(street_to), ?(house_to).",
		Fields: []s.OutField{
			s.OutField{
				Name: "street_from",
				Type: "dict",
				Attributes: s.FieldAttribute{
					Label:    "улица",
					Required: true,
					URL:      &dictUrl,
				},
			},
			s.OutField{
				Name: "house_from",
				Type: "text",
				Attributes: s.FieldAttribute{
					Label:    "дом",
					Required: true,
				},
			},
			s.OutField{
				Name: "entrance",
				Type: "number",
				Attributes: s.FieldAttribute{
					Label:    "подъезд",
					Required: false,
					EmptyText: &not_point,
				},
			},
			s.OutField{
				Name: "street_to",
				Type: "dict",
				Attributes: s.FieldAttribute{
					Label:    "улицa",
					Required: true,
					URL:      &dictUrl,
				},
			},
			s.OutField{
				Name: "house_to",
				Type: "text",
				Attributes: s.FieldAttribute{
					Label:    "дом",
					Required: true,
				},
			},
		},
	}

	result["commands_at_created_order"] = &[]s.OutCommand{
		s.OutCommand{
			Title:    "Отменить заказ",
			Action:   "cancel_order",
			Position: 0,
		},
		s.OutCommand{
			Title:    "Написать диспетчеру",
			Action:   "write_dispatcher",
			Position: 1,
			Fixed:    true,
			Form: &s.OutForm{
				Type: "form",
				Text: "?(text)",
				Fields: []s.OutField{
					s.OutField{
						Name: "text",
						Type: "text",
						Attributes: s.FieldAttribute{
							Label:    "Текст сообщения",
							Required: true,
						},
					},
				},
			},
		},
	}
	result["commands_for_order_feedback"] = &[]s.OutCommand{
		s.OutCommand{
			Title:    "Отзыв о поездке",
			Action:   "feedback",
			Position: 0,
			Form: &s.OutForm{
				Type: "form",
				Text: "?(text)",
				Fields: []s.OutField{
					s.OutField{
						Name: "text",
						Type: "text",
						Attributes: s.FieldAttribute{
							Label:    "Ваш отзыв",
							Required: true,
						},
					},
				},
			},
		},

		s.OutCommand{
			Title:    "Вызвать такси",
			Action:   "new_order",
			Position: 1,
			Repeated: true,
			Form:     taxi_call_form,
		},
	}

	result["commands_at_not_created_order"] = &[]s.OutCommand{
		s.OutCommand{
			Title:    "Вызвать такси",
			Action:   "new_order",
			Position: 0,
			Repeated: true,
			Form:     taxi_call_form,
		},
	}

	return result
}

type TaxiSupportMessageProcessor struct {
}

func (smp *TaxiSupportMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	return &s.MessageResult{Body:"Спасибо за ваш отзыв!", Type:"chat"}
}

func form_commands_for_current_order(order_wrapper *d.OrderWrapper, commands map[string]*[]s.OutCommand) *[]s.OutCommand {
	if order_wrapper != nil {
		//if time and not fedb and state is end of driver
		if time.Now().Sub(order_wrapper.When) < time.Hour && order_wrapper.Feedback == "" && order_wrapper.OrderState == ORDER_PAYED {
			return commands["commands_for_order_feedback"]
		}
		//if order state less than order payed in StatusesMap
		if order_wrapper.OrderState < ORDER_PAYED {
			return commands["commands_at_created_order"]
		}
	}
	return commands["commands_at_not_created_order"]
}

func FormCommands(username string, db d.DbHandlerMixin, context *s.BotContext) (*[]s.OutCommand, error) {
	order_wrapper, err := db.Orders.GetByOwner(username, context.Name)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	} else if err == mgo.ErrNotFound {
		return context.Commands["commands_at_not_created_order"], nil
	}
	return form_commands_for_current_order(order_wrapper, context.Commands), nil
}

type TaxiCommandsProcessor struct {
	d.DbHandlerMixin
	context *s.BotContext
}

func (cp *TaxiCommandsProcessor) ProcessRequest(in *s.InPkg) *s.RequestResult {
	phone, err := _get_phone(in)
	if err != nil {
		return s.ExceptionRequestResult(err, cp.context.Commands["commands_at_not_created_order"])
	}
	cp.Users.AddUser(&(in.From), phone)

	result, err := FormCommands(in.From, cp.DbHandlerMixin, cp.context)
	if err != nil {
		return s.ExceptionRequestResult(err, cp.context.Commands["commands_at_not_created_order"])
	}
	return &s.RequestResult{Commands:result}
}

type TaxiInformationProcessor struct {
	d.DbHandlerMixin
	context     *s.BotContext
	information *string
}

func (ih *TaxiInformationProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	var info_text string
	if ih.information == nil {
		info_text = "Срочный заказ такси. Быстрая подача. Оплата наличными или картой. "
	} else {
		info_text = *ih.information
	}
	return &s.MessageResult{
		Body: info_text,
		Type:"chat",
	}
}

type AddressNotHere struct {
	From string
	To   string
}
func (a *AddressNotHere) Error() string {
	return fmt.Sprintf("Адрес \n %+v --> %+v \n не поддерживается этим такси.", a.From, a.To)
}

func _form_order(fields []s.InField, ah *GoogleAddressHandler) (*NewOrder, error) {
	var from_info, to_info, hf, ht string
	var entrance *string
	log.Printf("NEW ORDER fields: %+v", fields)
	for _, field := range fields {
		switch fn := field.Name; fn {
		case "street_from":
			from_info = u.FirstOf(field.Data.Value, field.Data.Text).(string)
		case "street_to":
			to_info = u.FirstOf(field.Data.Value, field.Data.Text).(string)
		case "house_to":
			ht = u.FirstOf(field.Data.Value, field.Data.Text).(string)
		case "house_from":
			hf = u.FirstOf(field.Data.Value, field.Data.Text).(string)
		case "entrance":
			entrance_ := u.FirstOf(field.Data.Value, field.Data.Text).(string)
			entrance = &entrance_

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

	new_order := NewOrder{}
	note_info := "Тестирование."
	new_order.Notes = &note_info

	//	new_order.Attributes = [2]int64{1000113000, 1000113002}
	//	end fucking hardcode

	if !ah.IsHere(from_info) && !ah.IsHere(to_info) {
		return nil, &AddressNotHere{From:from_info, To:to_info}
	}
	delivery_street_info, err := ah.GetStreetInfo(from_info)
	if err != nil {
		return nil, err
	}
	destination_street_info, err := ah.GetStreetInfo(to_info)
	if err != nil {
		return nil, err
	}
	delivery := Delivery{IdStreet:delivery_street_info.ID, Street:delivery_street_info.Name, House:hf, Entrance:entrance, IdRegion:delivery_street_info.IDRegion}
	destination := Destination{IdStreet:destination_street_info.ID, Street:delivery_street_info.Name, House:ht, IdRegion:destination_street_info.IDRegion}
	new_order.Delivery = delivery
	new_order.Destinations = []Destination{destination}
	log.Printf("NEW ORDER: \ndelivery:%+v\ndestination:%+v", delivery, destination)
	return &new_order, nil
}

type TaxiNewOrderProcessor struct {
	ExternalApiMixin
	d.DbHandlerMixin
	AddressHandler *GoogleAddressHandler
	context        *s.BotContext
}

func _get_phone(in *s.InPkg) (phone *string, err error) {
	if user_data := in.UserData; user_data != nil {
		if phone := user_data.Phone; phone != "" {
			return &phone, nil
		}
	}
	return nil, errors.New("Нет записи UserData.Phone")
}

func (nop *TaxiNewOrderProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	order_wrapper, err := nop.Orders.GetByOwner(in.From, nop.context.Name)
	log.Printf("NOP saved_order info: %+v\n", order_wrapper)
	if err != nil {
		return s.ErrorMessageResult(err, nop.context.Commands["commands_at_not_created_order"])
	}

	if order_wrapper == nil || IsOrderNotAvailable(order_wrapper.OrderState) {
		commands := *in.Message.Commands
		phone, err := _get_phone(in)
		if err != nil {
			uwrpr, err := nop.Users.GetUserById(in.From)
			if err != nil {
				return s.ErrorMessageResult(errors.New("Не предоставлен номер телефона"), nop.context.Commands["commands_at_not_created_order"])
			} else {
				phone = uwrpr.Phone
			}
		}

		new_order, err := _form_order(commands[0].Form.Fields, nop.AddressHandler)
		if err_val, ok := err.(*AddressNotHere); ok {
			log.Printf("Addrss not here! %+v", err_val)
			return &s.MessageResult{
				Body: "Адрес не поддерживается этим такси.",
				Commands: nop.context.Commands["commands_at_not_created_order"],
				Type: "chat",
			}
		}
		if err != nil {
			return s.ErrorMessageResult(errors.New("Не могу определить адрес"), nop.context.Commands["commands_at_not_created_order"])
		}

		new_order.Phone = *phone

		ans := nop.API.NewOrder(*new_order)
		log.Printf("Order was created! %+v \n with content: %+v", ans, ans.Content)
		cost := ans.Content.Cost
		if cost == 0 {
			cost, _ = nop.API.CalcOrderCost(*new_order)
			if cost == 0 {
				log.Printf("ALERT! Создан заказ [%+v] без денег!", ans.Content.Id)
			}
		}
		//not send price settings
		not_send_price := false
		nsp_, ok := nop.context.Settings["not_send_price"]
		if ok {
			if _nsp, ok := nsp_.(bool); ok {
				not_send_price = _nsp
			}
		}
		text := ""
		if not_send_price {
			text = "Ваш заказ создан!"
		} else {
			text = fmt.Sprintf("Ваш заказ создан! Стоймость поездки составит %+v рублей.", cost)
		}
		//check is answer of new order in external api has error
		if !ans.IsSuccess {
			nop.Errors.StoreError(in.From, ans.Message)
			return s.ErrorMessageResult(errors.New(ans.Message), nop.context.Commands["commands_at_not_created_order"])
		}
		//persisting order
		err = nop.Orders.AddOrderObject(&d.OrderWrapper{OrderState:ORDER_CREATED, Whom:in.From, OrderId:ans.Content.Id, Source:nop.context.Name})
		if err != nil {
			return s.ErrorMessageResult(err, nop.context.Commands["commands_at_not_created_order"])
		}

		return &s.MessageResult{Body:text, Commands:nop.context.Commands["commands_at_created_order"], Type:"chat"}
	}
	return &s.MessageResult{Body: "Заказ уже создан!", Commands: nop.context.Commands["commands_at_created_order"], Type:"chat"}
}

type TaxiCancelOrderProcessor struct {
	ExternalApiMixin
	d.DbHandlerMixin
	context     *s.BotContext
	alert_phone string
}

func (cop *TaxiCancelOrderProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	order_wrapper, err := cop.Orders.GetByOwner(in.From, cop.context.Name)
	if err != nil {
		return s.ErrorMessageResult(err, cop.context.Commands["commands_at_not_created_order"])
	}

	if order_wrapper != nil && !IsOrderNotAvailable(order_wrapper.OrderState) {
		is_success, message := cop.API.CancelOrder(order_wrapper.OrderId)
		if is_success {
			cop.Orders.SetState(order_wrapper.OrderId, cop.context.Name, ORDER_CANCELED, nil)
			return &s.MessageResult{Body:"Ваш заказ отменен!", Commands: cop.context.Commands["commands_at_not_created_order"], Type:"chat"}
		} else {
			return &s.MessageResult{Body:fmt.Sprintf("Проблемы с отменой заказа %v\nЗвони скорее: %+v ", message, cop.alert_phone), Commands: cop.context.Commands["commands_at_not_created_order"], Type:"chat"}
		}
	}
	commands, err := FormCommands(in.From, cop.DbHandlerMixin, cop.context)
	if err != nil {
		return s.ErrorMessageResult(err, cop.context.Commands["commands_at_not_created_order"])
	}
	return &s.MessageResult{Body: "У вас нет активных заказов!", Commands:commands}
}

type TaxiCalculatePriceProcessor struct {
	ExternalApiMixin
	context        *s.BotContext
	AddressHandler *GoogleAddressHandler
}

func (cpp *TaxiCalculatePriceProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	commands := *in.Message.Commands
	order, err := _form_order(commands[0].Form.Fields, cpp.AddressHandler)
	if err != nil {
		return s.ErrorMessageResult(err, cpp.context.Commands["commands_at_not_created_order"])
	}
	cost_s, _ := cpp.API.CalcOrderCost(*order)
	cost := strconv.Itoa(cost_s)
	return &s.MessageResult{Body: fmt.Sprintf("Стоймость будет всего лишь %v рублей!", cost), Type:"chat"}
}

type TaxiFeedbackProcessor struct {
	ExternalApiMixin
	d.DbHandlerMixin
	context *s.BotContext
}

func _get_feedback(fields []s.InField) (fdb string, rate int) {
	//todo return not only string represent also int rating
	for _, v := range fields {
		if v.Name == "text" {
			fdb = u.FirstOf(v.Data.Value, v.Data.Text).(string)
		}
	}
	return fdb, rate
}

func (fp *TaxiFeedbackProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	commands := *in.Message.Commands
	fdbk, rate := _get_feedback(commands[0].Form.Fields)

	order_id, err := fp.Orders.SetFeedback(in.From, ORDER_PAYED, fdbk, fp.context.Name)
	if err != nil {
		return s.ErrorMessageResult(err, fp.context.Commands["commands_at_not_created_order"])
	}

	user, err := fp.Users.GetUserById(in.From)
	if err != nil {
		log.Printf("Error at implying user by id %v", in.From)
	}

	if order_id != nil && user != nil {
		f := Feedback{IdOrder: *order_id, Rating: rate, FeedBackText: fdbk, Phone:user.Phone}
		fp.API.Feedback(f)
		result_commands, err := FormCommands(in.From, fp.DbHandlerMixin, fp.context)
		if err != nil {
			return s.ErrorMessageResult(err, fp.context.Commands["commands_at_not_created_order"])
		}
		return &s.MessageResult{Body:"Спасибо! Ваш отзыв очень важен для нас:)", Commands: result_commands, Type:"chat"}
	} else if order_id == nil {
		return &s.MessageResult{Body:"Оплаченный заказ не найден :( Отзывы могут быть только для оплаченных заказов", Commands:fp.context.Commands["commands_at_not_created_order"], Type:"chat"}
	} else if user == nil {
		return &s.MessageResult{Body:"Не найдена информация о пользователе...", Commands:fp.context.Commands["commands_at_not_created_order"], Type:"chat"}
	} else {
		return &s.MessageResult{Body:"Нет ни пользователя ни заказа с нужным состоянием :(", Commands:fp.context.Commands["commands_at_not_created_order"], Type:"chat"}
	}
}
