package taxi

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
	"encoding/json"

	"gopkg.in/mgo.v2"

	s "msngr/structs"
	d "msngr/db"
	u "msngr/utils"
	c "msngr/configuration"
	m "msngr"
)

const (
	timeFormat = "2006-01-02 15:04:05"
	car_info_update_time = 5.0
)

type CarInfoProvider struct {
	Cache      *CarsCache
	LastUpdate time.Time
}

func NewCarInfoProvider(cache *CarsCache) *CarInfoProvider {
	return &CarInfoProvider{Cache:cache, LastUpdate:time.Now()}
}

func (cip *CarInfoProvider) GetCarInfo(car_id int64) *CarInfo {
	if time.Now().Sub(cip.LastUpdate).Seconds() > car_info_update_time {
		cip.Cache.Reload()
		cip.LastUpdate = time.Now()
	}
	car_info := cip.Cache.GetCarInfo(car_id)
	return car_info
}


func FormTaxiBotContext(im *ExternalApiMixin, db_handler *d.MainDb, tc c.TaxiConfig, ah AddressHandler, cc *CarsCache) *m.BotContext {
	context := m.BotContext{}
	context.Check = func() (detail string, ok bool) {
		ok = im.API.IsConnected()
		if !ok {
			detail = "Ошибка в подключении к сервису. Попробуйте позже."
		} else {
			return "", db_handler.Check()
		}
		return detail, ok
	}
	context.Commands = GetCommands(tc.DictUrl)
	context.Commands = EnsureAvailableCommands(context.Commands, tc.AvailableCommands)

	context.Name = tc.Name
	context.Request_commands = map[string]s.RequestCommandProcessor{
		"commands": &TaxiCommandsProcessor{MainDb: *db_handler, context: &context},
	}
	context.Message_commands = map[string]s.MessageCommandProcessor{
		"information":      &TaxiInformationProcessor{information:&(tc.Information.Text)},
		"new_order":        &TaxiNewOrderProcessor{ExternalApiMixin: *im, MainDb: *db_handler, context:&context, AddressHandler:ah},
		"cancel_order":     &TaxiCancelOrderProcessor{ExternalApiMixin: *im, MainDb: *db_handler, context:&context, alert_phone:tc.Information.Phone},
		"calculate_price":  &TaxiCalculatePriceProcessor{ExternalApiMixin: *im, context:&context, AddressHandler:ah},
		"feedback":         &TaxiFeedbackProcessor{ExternalApiMixin: *im, MainDb: *db_handler, context:&context},
		"write_dispatcher": &TaxiWriteDispatcherMessageProcessor{ExternalApiMixin: *im},
		"callback_request": &TaxiCallbackRequestMessageProcessor{ExternalApiMixin:*im},
		"where_it":         &TaxiWhereItMessageProcessor{ExternalApiMixin:*im, MainDb:*db_handler, context:&context},
		"car_position":     &TaxiCarPositionMessageProcessor{ExternalApiMixin: *im, MainDb:*db_handler, context:&context, Cars:NewCarInfoProvider(cc)},
	}
	context.Settings = make(map[string]interface{})
	context.Settings["not_send_price"] = tc.Api.NotSendPrice
	if tc.Markups != nil {
		context.Settings["markups"] = *tc.Markups
	}
	return &context
}

const (
	CMDS_CREATED_ORDER = "created_order"
	CMDS_NOT_CREATED_ORDER = "not_created_order"
	CMDS_FEEDBACK = "feedback"
)

var not_point = string("не указан")

var CommandsData = map[string]s.OutCommand{
	"where_it":s.OutCommand{
		Title:"Не вижу машины",
		Action:"where_it",
	},
	"callback_request":s.OutCommand{
		Title:"Заказать обратный звонок",
		Action:"callback_request",
	},
	"car_position":s.OutCommand{
		Title: "Узнать местонахождение машины",
		Action:"car_position",
	},


}

func EnsureAvailableCommands(default_cmds map[string]*[]s.OutCommand, available_cmds_info map[string][]string) map[string]*[]s.OutCommand {
	//ensuring commands
	for cmd_type, cmd_names := range available_cmds_info {
		if cmds_p, ok := default_cmds[cmd_type]; ok {
			cmds := *cmds_p
			position := cmds[len(cmds) - 1].Position
			for _, cmd_name := range cmd_names {
				position += 1
				cmd := CommandsData[cmd_name]
				cmd.Position = position
				cmds = append(cmds, cmd)
			}
			default_cmds[cmd_type] = &cmds
		}
	}
	return default_cmds
}

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

	result[CMDS_CREATED_ORDER] = &[]s.OutCommand{
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

	result[CMDS_FEEDBACK] = &[]s.OutCommand{
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

	result[CMDS_NOT_CREATED_ORDER] = &[]s.OutCommand{
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

type TaxiCarPositionMessageProcessor struct {
	ExternalApiMixin
	d.MainDb
	Cars    *CarInfoProvider
	context *m.BotContext

}

func (cp *TaxiCarPositionMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	order_wrapper, err := cp.Orders.GetByOwner(in.From, cp.context.Name, true)
	if err != nil {
		return s.ErrorMessageResult(err, cp.context.Commands[CMDS_NOT_CREATED_ORDER])
	}

	if order_wrapper != nil && !IsOrderNotActual(order_wrapper.OrderState) {
		car_id_ := order_wrapper.OrderData.Get("IDCar")
		if car_id_ == nil {
			return &s.MessageResult{Body: "Не найден идентификатор автомобиля у вашего заказа :("}
		}
		car_id, ok := car_id_.(int64)
		if !ok {
			return &s.MessageResult{Body:fmt.Sprintf("Не понятен идентификатор автомобиля у вашего заказа :( %#v, %T", car_id_, car_id_)}
		}
		car_info := cp.Cars.GetCarInfo(car_id)
		return &s.MessageResult{Body:fmt.Sprintf("Lat:%v;Lon:%v", car_info.Lat, car_info.Lon)}

	}
	commands, err := FormCommands(in.From, cp.MainDb, cp.context)
	if err != nil {
		return s.ErrorMessageResult(err, cp.context.Commands[CMDS_NOT_CREATED_ORDER])
	}
	return &s.MessageResult{Body: "У вас нет активных заказов!", Commands:commands}
}

type TaxiWriteDispatcherMessageProcessor struct {
	ExternalApiMixin
}

func get_text(in s.InCommand) (s string, err error) {
	if len(in.Form.Fields) > 0 {
		s = in.Form.Fields[0].Data.Text
		return s, nil
	} else {
		err = errors.New("No fields in input command")
		return s, err
	}
}

func (smp *TaxiWriteDispatcherMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	cmds := in.Message.Commands
	if cmds == nil {
		return &s.MessageResult{Body:"Нет данных"}
	}
	commands := *cmds
	message, err := get_text(commands[0])
	if err != nil {
		return &s.MessageResult{Body:fmt.Sprintf("Ошибка! %v", err)}
	}
	ok, result := smp.API.WriteDispatcher(message)
	var text string
	if ok {
		text = fmt.Sprintf("Спасибо за ваш отзыв!\n%s", result)
	} else {
		text = fmt.Sprintf("Спасибо за ваш отзыв! Но сообщение доставленно с ошибкой\n%s\nопробуйте снова", result)
	}
	return &s.MessageResult{Body:text, Type:"chat"}
}

type TaxiCallbackRequestMessageProcessor struct {
	ExternalApiMixin
}

func (crmp *TaxiCallbackRequestMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	phone, err := _get_phone(in)
	if err != nil {
		return &s.MessageResult{Body:"Ошибка! Не предоставлен номер телефона", Type:"chat"}
	}
	ok, result := crmp.API.CallbackRequest(*phone)
	var text string
	if ok {
		text = fmt.Sprintf("Ожидайте звонка оператора\n%s", result)
	}else {
		text = fmt.Sprintf("Ошибка при отправке запроса на обратный звонок\n%s", result)
	}
	return &s.MessageResult{Body:text, Type:"chat"}
}


type TaxiWhereItMessageProcessor struct {
	ExternalApiMixin
	d.MainDb
	context *m.BotContext
}

func (twmp *TaxiWhereItMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	order_wrapper, err := twmp.Orders.GetByOwner(in.From, twmp.context.Name, true)
	if err != nil {
		return s.ErrorMessageResult(err, twmp.context.Commands[CMDS_NOT_CREATED_ORDER])
	}

	if order_wrapper != nil && !IsOrderNotActual(order_wrapper.OrderState) {
		ok, result := twmp.API.WhereIt(order_wrapper.OrderId)
		var text string
		if ok {
			text = fmt.Sprintf("О том что вы не видите машину диспетчер уведомлен\n%s", result)
		}else {
			text = fmt.Sprintf("Ошибка!\n%s", result)
		}
		return &s.MessageResult{Body:text, Type:"chat"}
	}
	return s.ErrorMessageResult(errors.New("Не найден активный заказ"), twmp.context.Commands[CMDS_NOT_CREATED_ORDER])

}

func form_commands_for_current_order(order_wrapper *d.OrderWrapper, commands map[string]*[]s.OutCommand) *[]s.OutCommand {
	if order_wrapper != nil {
		//if time and not fedb and state is end of driver
		if time.Now().Sub(order_wrapper.When) < time.Hour && order_wrapper.Feedback == "" && order_wrapper.OrderState == ORDER_PAYED {
			return commands[CMDS_FEEDBACK]
		}
		//if order state less than order payed in StatusesMap
		if order_wrapper.Active == true {
			return commands[CMDS_CREATED_ORDER]
		}
	}
	return commands[CMDS_NOT_CREATED_ORDER]
}

func FormCommands(username string, db d.MainDb, context *m.BotContext) (*[]s.OutCommand, error) {
	order_wrapper, err := db.Orders.GetByOwnerLast(username, context.Name)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	} else if err == mgo.ErrNotFound {
		return context.Commands[CMDS_NOT_CREATED_ORDER], nil
	}
	return form_commands_for_current_order(order_wrapper, context.Commands), nil
}

type TaxiCommandsProcessor struct {
	d.MainDb
	context *m.BotContext
}

func (cp *TaxiCommandsProcessor) ProcessRequest(in *s.InPkg) *s.RequestResult {
	phone, _ := _get_phone(in)
	if phone != nil {
		cp.Users.AddUser(in.From, *phone)
	}

	result, err := FormCommands(in.From, cp.MainDb, cp.context)
	if err != nil {
		return s.ExceptionRequestResult(err, cp.context.Commands[CMDS_NOT_CREATED_ORDER])
	}
	return &s.RequestResult{Commands:result}
}

type TaxiInformationProcessor struct {
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

func _form_order(fields []s.InField, ah AddressHandler) (*NewOrderInfo, error) {
	var from_key, from_name, to_key, to_name, hf, ht, entrance string
	log.Printf("NEW ORDER fields: %+v", fields)
	for _, field := range fields {
		switch fn := field.Name; fn {
		case "street_from":
			from_key = field.Data.Value
			from_name = field.Data.Text
		case "street_to":
			to_key = field.Data.Value
			to_name = field.Data.Text
		case "house_to":
			ht = u.FirstOf(field.Data.Value, field.Data.Text).(string)
		case "house_from":
			hf = u.FirstOf(field.Data.Value, field.Data.Text).(string)
		case "entrance":
			entrance = u.FirstOf(field.Data.Value, field.Data.Text).(string)

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

	new_order := NewOrderInfo{}
	note_info := "Тестирование."
	new_order.Notes = note_info

	var dest, deliv AddressF
	if ah != nil {
		if !ah.IsHere(from_key) && !ah.IsHere(to_key) {
			return nil, &AddressNotHere{From:from_key, To:to_key}
		}
		del_id_street_, err := ah.GetExternalInfo(from_key, from_name)
		dest_id_street_, err := ah.GetExternalInfo(to_key, to_name)
		if err != nil {
			return nil, err
		}

		deliv = *del_id_street_
		dest = *dest_id_street_
	} else {
		log.Printf("Address handler is nil. Using street id and name values... [%v] %v --> [%v] %v", from_key, from_name, to_key, to_name)
		err := json.Unmarshal([]byte(from_key), &deliv)
		if err != nil {
			log.Printf("FORM ORDER Error at unmarshal address from; %v\n%v", err, from_key)
		}
		err = json.Unmarshal([]byte(to_key), &dest)
		if err != nil {
			log.Printf("FORM ORDER Error at unmarshal address to; %v\n%v", err, to_key)
		}
		log.Printf("\nDelivery: %v\nDestination: %v", deliv, dest)
	}
	deliv_id := strconv.FormatInt(deliv.ID, 10)
	dest_id := strconv.FormatInt(dest.ID, 10)

	delivery := Delivery{
		Street:u.FirstOf(from_name, deliv.Name).(string),
		House:hf,
		City:deliv.City,
		Entrance:entrance,
		IdRegion:deliv.IDRegion,
		IdAddress:deliv_id}
	destination := Destination{
		Street:u.FirstOf(to_name, dest.Name).(string),
		House:ht,
		IdRegion:dest.IDRegion,
		IdAddress:dest_id,
		City:dest.City}

	new_order.Delivery = &delivery
	new_order.Destinations = []*Destination{&destination}
	return &new_order, nil
}

type TaxiNewOrderProcessor struct {
	ExternalApiMixin
	d.MainDb
	AddressHandler AddressHandler
	context        *m.BotContext
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
	order_wrapper, err := nop.Orders.GetByOwnerLast(in.From, nop.context.Name)
	if err != nil {
		return s.ErrorMessageResult(err, nop.context.Commands[CMDS_NOT_CREATED_ORDER])
	}

	if order_wrapper == nil || order_wrapper.Active == false {
		commands := *in.Message.Commands
		phone, err := _get_phone(in)
		if err != nil {
			uwrpr, err := nop.Users.GetUserById(in.From)
			if err != nil {
				return s.ErrorMessageResult(errors.New("Не предоставлен номер телефона"), nop.context.Commands[CMDS_NOT_CREATED_ORDER])
			} else {
				phone = &(uwrpr.Phone)
			}
		}

		new_order, err := _form_order(commands[0].Form.Fields, nop.AddressHandler)
		if err != nil {
			log.Printf("Error at forming order: %+v", err)
			if err_val, ok := err.(*AddressNotHere); ok {
				log.Printf("Addrss not here! %+v", err_val)
				return &s.MessageResult{
					Body: "Адрес не поддерживается этим такси.",
					Commands: nop.context.Commands[CMDS_NOT_CREATED_ORDER],
					Type: "chat",
				}
			}else {
				return s.ErrorMessageResult(errors.New("Не могу определить адрес"), nop.context.Commands[CMDS_NOT_CREATED_ORDER])
			}
		}

		new_order.Phone = *phone
		if mrkps, ok := nop.context.Settings["markups"]; ok {
			markups, ok := mrkps.([]string)
			if ok {
				new_order.Markups = markups
			}else {
				log.Printf("markups setting present but it is not []string %+v, %T", mrkps, mrkps)
			}
		}
		//get info about cost
		cost, _ := nop.API.CalcOrderCost(*new_order)
		//send command to create order to external api
		ans := nop.API.NewOrder(*new_order)
		//check is answer of new order in external api has error
		if !ans.IsSuccess {
			nop.Errors.StoreError(in.From, ans.Message)
			return s.ErrorMessageResult(errors.New(ans.Message), nop.context.Commands[CMDS_NOT_CREATED_ORDER])
		}
		log.Printf("Order was created! %+v \n with content: %+v", ans, ans.Content)

		//not send price settings
		not_send_price := false
		nsp_, ok := nop.context.Settings["not_send_price"]
		if ok {
			if _nsp, ok := nsp_.(bool); ok {
				not_send_price = _nsp
			}
		}
		//persisting order
		err = nop.Orders.AddOrderObject(&d.OrderWrapper{OrderState:ORDER_CREATED, Whom:in.From, OrderId:ans.Content.Id, Source:nop.context.Name})
		err = nop.Orders.SetActive(ans.Content.Id, nop.context.Name, true)
		if err != nil {
			//if error we must cancel order at external api
			ok, message := nop.API.CancelOrder(ans.Content.Id)
			log.Printf("Error at persist order. Cancelling order at external api with result: %v, %v", ok, message)
			return s.ErrorMessageResult(err, nop.context.Commands[CMDS_NOT_CREATED_ORDER])
		}
		//forming text
		text := ""
		if not_send_price {
			text = "Ваш заказ создан!"
		} else {
			text = fmt.Sprintf("Ваш заказ создан! Стоймость поездки составит %+v рублей.", cost)
		}

		return &s.MessageResult{Body:text, Commands:nop.context.Commands[CMDS_CREATED_ORDER], Type:"chat"}
	}
	order_state, _ := InfinityStatusesName[order_wrapper.OrderState]
	return &s.MessageResult{Body: fmt.Sprintf("Заказ уже создан! Потому что в состоянии %v", order_state), Commands: nop.context.Commands[CMDS_NOT_CREATED_ORDER], Type:"chat"}
}

type TaxiCancelOrderProcessor struct {
	ExternalApiMixin
	d.MainDb
	context     *m.BotContext
	alert_phone string
}

func (cop *TaxiCancelOrderProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	order_wrapper, err := cop.Orders.GetByOwnerLast(in.From, cop.context.Name)
	if err != nil {
		return s.ErrorMessageResult(err, cop.context.Commands[CMDS_NOT_CREATED_ORDER])
	}
	if order_wrapper == nil || order_wrapper.Active == false {
		if order_wrapper != nil {
			log.Printf("Order is %v active? (%v)", order_wrapper, order_wrapper.Active)
		} else {
			log.Printf("Order not found...")
		}
		return s.ErrorMessageResult(errors.New("Order for it operation is unsuitable :("), cop.context.Commands[CMDS_NOT_CREATED_ORDER])
	}
	is_success, message := cop.API.CancelOrder(order_wrapper.OrderId)
	if is_success {
		cop.Orders.SetState(order_wrapper.OrderId, cop.context.Name, ORDER_CANCELED, nil)
		cop.Orders.SetActive(order_wrapper.OrderId, order_wrapper.Source, false)
		return &s.MessageResult{Body:"Ваш заказ отменен!", Commands: cop.context.Commands[CMDS_NOT_CREATED_ORDER], Type:"chat"}
	} else {
		return &s.MessageResult{Body:fmt.Sprintf("Проблемы с отменой заказа %v\nЗвони скорее: %+v ", message, cop.alert_phone), Commands: cop.context.Commands[CMDS_NOT_CREATED_ORDER], Type:"chat"}
	}

	commands, err := FormCommands(in.From, cop.MainDb, cop.context)
	if err != nil {
		return s.ErrorMessageResult(err, cop.context.Commands[CMDS_NOT_CREATED_ORDER])
	}
	return &s.MessageResult{Body: "У вас нет активных заказов!", Commands:commands}
}

type TaxiCalculatePriceProcessor struct {
	ExternalApiMixin
	context        *m.BotContext
	AddressHandler AddressHandler
}

func (cpp *TaxiCalculatePriceProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	commands := *in.Message.Commands
	order, err := _form_order(commands[0].Form.Fields, cpp.AddressHandler)
	if err != nil {
		return s.ErrorMessageResult(err, cpp.context.Commands[CMDS_NOT_CREATED_ORDER])
	}
	cost_s, _ := cpp.API.CalcOrderCost(*order)
	cost := strconv.Itoa(cost_s)
	return &s.MessageResult{Body: fmt.Sprintf("Стоймость будет всего лишь %v рублей!", cost), Type:"chat"}
}

type TaxiFeedbackProcessor struct {
	ExternalApiMixin
	d.MainDb
	context *m.BotContext
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
	phone, err := _get_phone(in)
	if phone == nil {
		user, uerr := fp.Users.GetUserById(in.From)
		if uerr != nil {
			log.Printf("Error at implying user by id %v", in.From)
			return s.ErrorMessageResult(err, fp.context.Commands[CMDS_FEEDBACK])
		}else {
			phone = &(user.Phone)
		}
	}

	order_id, err := fp.Orders.SetFeedback(in.From, ORDER_PAYED, fdbk, fp.context.Name)
	if err != nil {
		return s.ErrorMessageResult(err, fp.context.Commands[CMDS_NOT_CREATED_ORDER])
	}
	if order_id != nil {
		f := Feedback{IdOrder: *order_id, Rating: rate, FeedBackText: fdbk, Phone:*phone}
		fp.API.Feedback(f)
		result_commands, err := FormCommands(in.From, fp.MainDb, fp.context)
		if err != nil {
			return s.ErrorMessageResult(err, fp.context.Commands[CMDS_NOT_CREATED_ORDER])
		}
		return &s.MessageResult{Body:"Спасибо! Ваш отзыв очень важен для нас:)", Commands: result_commands, Type:"chat"}
	} else {
		return &s.MessageResult{Body:"Оплаченный заказ не найден :( Отзывы могут быть только для оплаченных заказов", Commands:fp.context.Commands[CMDS_NOT_CREATED_ORDER], Type:"chat"}
	}
}
