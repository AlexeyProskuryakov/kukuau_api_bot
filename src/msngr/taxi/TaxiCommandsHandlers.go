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
	"reflect"
	"regexp"
)

const (
	CAR_INFO_UPDATE_TIME = 30.0
	NEW_ORDER_TEXT_INFO = "В течении 5 минут Вам будет назначен автомобиль. Или перезвонит оператор если ожидаемое время подачи составит более 15 минут."
)

var CONNECTION_ERROR = s.ErrorMessageResult(errors.New("Система обработки заказов такси не отвечает, попробуйте позже."), nil)

type CarInfoProvider struct {
	Cache      *CarsCache
	LastUpdate time.Time
}

func NewCarInfoProvider(cache *CarsCache) *CarInfoProvider {
	return &CarInfoProvider{Cache:cache, LastUpdate:time.Now()}
}

func (cip *CarInfoProvider) GetCarInfo(car_id int64) *CarInfo {
	if time.Now().Sub(cip.LastUpdate).Seconds() > CAR_INFO_UPDATE_TIME {
		cip.Cache.Reload()
		cip.LastUpdate = time.Now()
	}
	car_info := cip.Cache.GetCarInfo(car_id)
	return car_info
}

func FormTaxiBotContext(im *ExternalApiMixin, db_handler *d.MainDb, cfgStore *d.ConfigurationStorage, tc c.TaxiConfig, ah AddressHandler, cc *CarsCache) *m.BotContext {
	context := m.BotContext{}
	context.Check = func() (detail string, ok bool) {
		ok = im.API.IsConnected()
		if !ok {
			detail = "Ошибка в подключении к сервису такси. Попробуйте позже."
		} else {
			return "", db_handler.Check()
		}
		return detail, ok
	}
	context.Commands = EnsureAvailableCommands(GetCommands(tc.DictUrl), tc.AvailableCommands)

	context.Name = tc.Name
	context.RequestProcessors = map[string]s.RequestCommandProcessor{
		"commands": &TaxiCommandsProcessor{MainDb: *db_handler, context: &context},
	}

	commandsGenerator := func(in *s.InPkg) (*[]s.OutCommand, error) {
		command, err := form_commands(in.From, *db_handler, &context)
		return command, err
	}

	context.MessageProcessors = map[string]s.MessageCommandProcessor{
		"information":      m.NewUpdatableInformationProcessor(cfgStore, commandsGenerator, tc.Chat.CompanyId),
		"new_order":        &TaxiNewOrderProcessor{ExternalApiMixin: *im, MainDb: *db_handler, context:&context, AddressHandler:ah, Config: tc},
		"cancel_order":     &TaxiCancelOrderProcessor{ExternalApiMixin: *im, MainDb: *db_handler, context:&context, alert_phone:tc.Information.Phone},
		"calculate_price":  &TaxiCalculatePriceProcessor{ExternalApiMixin: *im, context:&context, AddressHandler:ah, Config: tc},
		"feedback":         &TaxiFeedbackProcessor{ExternalApiMixin: *im, MainDb: *db_handler, context:&context},
		"write_dispatcher": &TaxiWriteDispatcherMessageProcessor{ExternalApiMixin: *im, MainDb:*db_handler},
		"callback_request": &TaxiCallbackRequestMessageProcessor{ExternalApiMixin:*im},
		"where_it":         &TaxiWhereItMessageProcessor{ExternalApiMixin:*im, MainDb:*db_handler, context:&context},
		"car_position":     &TaxiCarPositionMessageProcessor{ExternalApiMixin: *im, MainDb:*db_handler, context:&context, Cars:NewCarInfoProvider(cc)},
		"":                 &TaxiWriteDispatcherMessageProcessor{ExternalApiMixin: *im, MainDb:*db_handler},
	}
	context.Settings = make(map[string]interface{})
	context.Settings["not_send_price"] = tc.Api.NotSendPrice
	if tc.Markups != nil {
		context.Settings["markups"] = *tc.Markups
	}
	if tc.Api.Data.RefreshOrdersTimeStep != 0 {
		context.Settings["refresh_orders_time_step"] = time.Duration(tc.Api.Data.RefreshOrdersTimeStep) * time.Second
	} else {
		context.Settings["refresh_orders_time_step"] = 10 * time.Second
	}

	return &context
}

const (
	CMDS_CREATED_ORDER = "created_order"
	CMDS_NOT_CREATED_ORDER = "not_created_order"
	CMDS_FEEDBACK = "feedback"
)

var (
	not_point = string("не указан")
	NOW = string("сейчас")
	EMPTY = string("")
)

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
	//	log.Printf("Avaliable commands: %+v\nResult commands before%+v", available_cmds_info, default_cmds)
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
	//	log.Printf("Result commands after: \n%+v", default_cmds)
	return default_cmds
}

func GetCommands(dictUrl string) map[string]*[]s.OutCommand {
	result := make(map[string]*[]s.OutCommand)
	hours := []string{"00", "01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"}
	minutes := []string{"00", "01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25",
		"26", "27", "28", "29", "30", "31", "32", "33", "34", "35", "36", "37", "38", "39", "39", "40", "41", "42", "43", "44", "45", "46", "48", "48", "49", "50", "51", "52", "53", "54", "54", "55", "56", "57", "58", "59" }
	for i := 0; i < 60; i++ {
		if i < 24 {
			hours = append(hours, fmt.Sprintf("%02d", i))

		}
		minutes = append(minutes, fmt.Sprintf("%02d", i))
	}
	var taxi_call_form = &s.OutForm{
		Title: "Форма вsызова такси",
		Type:  "form",
		Name:  "call_taxi",
		Text:  "Откуда: ?(street_from), ?(house_from), ?(entrance). Куда: ?(street_to), ?(house_to). Когда: ?(thour):?(tmin).",
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
				Type: "text",
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
			s.OutField{
				Name:"thour",
				Type:"list-single",
				Attributes:s.FieldAttribute{
					Label:"ЧЧ",
					EmptyText:&NOW,
				},
				Items:s.FormItems(hours),
			},
			s.OutField{
				Name:"tmin",
				Type:"list-single",
				Attributes:s.FieldAttribute{
					Label:"MM",
					EmptyText:&EMPTY,
				},
				Items:s.FormItems(minutes),
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
	if !cp.API.IsConnected() {
		cp.API.Connect()
		return CONNECTION_ERROR
	}
	order_wrapper, err := cp.Orders.GetByOwner(in.From, cp.context.Name, true)
	if err != nil {
		return s.ErrorMessageResult(err, cp.context.Commands[CMDS_NOT_CREATED_ORDER])
	}

	if order_wrapper != nil && !IsOrderNotActual(order_wrapper.OrderState) {
		car_id_ := order_wrapper.OrderData.Get("IDCar")
		if car_id_ == nil {
			return &s.MessageResult{Body: "Не найден идентификатор автомобиля у вашего заказа (видимо, автомобиль вам еще не назначили) :("}
		}
		car_id, ok := car_id_.(int64)
		if !ok {
			return &s.MessageResult{Body:fmt.Sprintf("Не понятен идентификатор автомобиля у вашего заказа :( %#v, %T", car_id_, car_id_)}
		}
		car_info := cp.Cars.GetCarInfo(car_id)
		if car_info != nil {
			return &s.MessageResult{Body:fmt.Sprintf("Lat:%v;Lon:%v", car_info.Lat, car_info.Lon)}
		} else {
			return s.ErrorMessageResult(errors.New("Неизвестный автомобиль."), cp.context.Commands[CMDS_CREATED_ORDER])
		}

	}
	commands, err := form_commands(in.From, cp.MainDb, cp.context)
	if err != nil {
		return s.ErrorMessageResult(err, cp.context.Commands[CMDS_NOT_CREATED_ORDER])
	}
	return &s.MessageResult{Body: "У вас нет активных заказов!", Commands:commands}
}

type TaxiWriteDispatcherMessageProcessor struct {
	ExternalApiMixin
	d.MainDb
}

func get_text(in s.InCommand) (s string, err error) {
	if len(in.Form.Fields) == 1 {
		//		log.Printf("TAXI Getting text from: %+v\n at: %+v", in.Form.Fields[0])
		if s, ok := u.FirstOf(in.Form.Fields[0].Data.Text, in.Form.Fields[0].Data.Value).(string); ok {
			return s, nil
		}
		return "", errors.New("Can not find text at this form :(")
	} else {
		err = errors.New("No fields in input command :(")
		return s, err
	}
}

func (smp *TaxiWriteDispatcherMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	if !smp.API.IsConnected() {
		smp.API.Connect()
		return CONNECTION_ERROR
	}
	var message string
	var err error
	cmds := in.Message.Commands
	if cmds != nil {
		commands := *cmds
		message, err = get_text(commands[0])
		if err != nil {
			return &s.MessageResult{Body:fmt.Sprintf("Ошибка! %v", err)}
		}
	} else if in.Message.Body != nil && *in.Message.Body != "" {
		message = *in.Message.Body
	} else {
		return &s.MessageResult{Body:"Ошибка, совсем нет букв. Мне нечего отправить диспетчеру :("}
	}
	//	log.Printf("TAXI Write dispatcher message: %s", message)
	user, err := smp.Users.GetUserById(in.From)
	if err != nil {
		return &s.MessageResult{Body:"Ошибка нет телефона у пользователя или самого пользователя :(", Type:"chat"}
	}
	if user != nil {
		message = fmt.Sprintf("%s от %v", message, user.Phone)
	} else {
		if in.UserData != nil {
			message = fmt.Sprintf("%s от %v", message, in.UserData.Phone)
			smp.Users.AddUser(in.From, in.UserData.Name, in.UserData.Phone, in.UserData.Email)
		}
	}
	ok, result := smp.API.WriteDispatcher(message)
	var text string
	if ok {
		text = result
	} else {
		text = fmt.Sprintf("Спасибо за ваш отзыв! Но сообщение доставленно с ошибкой.\n%s", result)
	}
	return &s.MessageResult{Body:text, Type:"chat"}
}

type TaxiCallbackRequestMessageProcessor struct {
	ExternalApiMixin
}

func (crmp *TaxiCallbackRequestMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	if !crmp.API.IsConnected() {
		crmp.API.Connect()
		return CONNECTION_ERROR
	}
	phone, err := _get_phone(in)
	if err != nil {
		return &s.MessageResult{Body:"Ошибка! Не предоставлен номер телефона.", Type:"chat"}
	}
	ok, result := crmp.API.CallbackRequest(*phone)
	var text string
	if ok {
		text = fmt.Sprintf("Ожидайте звонка оператора\n%s", result)
	} else {
		text = fmt.Sprintf("Ошибка при отправке запроса на обратный звонок.\n%s", result)
	}
	return &s.MessageResult{Body:text, Type:"chat"}
}

type TaxiWhereItMessageProcessor struct {
	ExternalApiMixin
	d.MainDb
	context *m.BotContext
}

func (twmp *TaxiWhereItMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	if !twmp.API.IsConnected() {
		twmp.API.Connect()
		return CONNECTION_ERROR
	}
	order_wrapper, err := twmp.Orders.GetByOwner(in.From, twmp.context.Name, true)
	if err != nil {
		return s.ErrorMessageResult(err, twmp.context.Commands[CMDS_NOT_CREATED_ORDER])
	}

	if order_wrapper != nil && !IsOrderNotActual(order_wrapper.OrderState) {
		ok, result := twmp.API.WhereIt(order_wrapper.OrderId)
		var text string
		if ok {
			text = fmt.Sprintf("О том что вы не видите машину диспетчер уведомлен\n%s", result)
		} else {
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

func form_commands(username string, db d.MainDb, context *m.BotContext) (*[]s.OutCommand, error) {
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
	if in.UserData != nil {
		cp.Users.AddUser(in.From, in.UserData.Name, in.UserData.Phone, in.UserData.Email)
	}

	result, err := form_commands(in.From, cp.MainDb, cp.context)
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
		info_text = "Срочный заказ такси. Быстрая подача. Оплата наличными или картой. Для оформления заказа нажмите кнопку меню расположенную в нижнем левом углу."
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
	var dHours, dMin int
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
		case "thour":
			dHours, _ = strconv.Atoi(field.Data.Text)
		case "tmin":
			dMin, _ = strconv.Atoi(field.Data.Text)
		}
	}
	var dest, deliv AddressF
	if ah != nil {
		if from_key == "" || to_key == "" {
			return nil, &AddressNotHere{From:from_key, To:to_key}
		}
		if !ah.IsHere(from_key) && !ah.IsHere(to_key) {
			return nil, &AddressNotHere{From:from_key, To:to_key}
		}
		del_id_street_, err_from := ah.GetExternalInfo(from_key, from_name)
		dest_id_street_, err_to := ah.GetExternalInfo(to_key, to_name)
		if err_from != nil {
			return nil, err_from
		}
		if err_to != nil {
			return nil, err_to
		}
		deliv = *del_id_street_
		dest = *dest_id_street_
	} else {
		//		log.Printf("Address handler is nil. Using street id and name values... [%v] %v --> [%v] %v", from_key, from_name, to_key, to_name)
		err := json.Unmarshal([]byte(from_key), &deliv)
		if err != nil {
			log.Printf("FORM ORDER Error at unmarshal address from; %v\n%v", err, from_key)
		}
		err = json.Unmarshal([]byte(to_key), &dest)
		if err != nil {
			log.Printf("FORM ORDER Error at unmarshal address to; %v\n%v", err, to_key)
		}
		//		log.Printf("\nDelivery: %v\nDestination: %v", deliv, dest)
	}

	deliv_id := strconv.FormatInt(deliv.ID, 10)
	dest_id := strconv.FormatInt(dest.ID, 10)

	delivery := Delivery{
		Street:u.FirstOf(from_name, deliv.Name).(string),
		IdStreet:deliv.ID,
		House:hf,
		City:deliv.City,
		Entrance:entrance,
		IdRegion:deliv.IDRegion,
		IdAddress:deliv_id,
	}

	destination := Destination{
		Street:u.FirstOf(to_name, dest.Name).(string),
		IdStreet:dest.ID,
		House:ht,
		IdRegion:dest.IDRegion,
		IdAddress:dest_id,
		City:dest.City,
	}

	new_order := NewOrderInfo{Notes:"Заказ создан через мессенджер Klichat"}
	if dHours > 0  && dMin > 0 {
		dTime := time.Now().Add(time.Duration(dHours) * time.Hour)
		dTime = time.Now().Add(time.Duration(dMin) * time.Minute)
		log.Printf("Order after: %+v", dTime)
		dTime.Format("2006.01.02 15:04:05")
		new_order.DeliveryTime = dTime.Format("2006.01.02 15:04:05")
	}

	new_order.Delivery = &delivery
	new_order.Destinations = []*Destination{&destination}
	return &new_order, nil
}

type TaxiNewOrderProcessor struct {
	ExternalApiMixin
	d.MainDb
	AddressHandler AddressHandler
	context        *m.BotContext
	Config         c.TaxiConfig
}

func _get_phone(in *s.InPkg) (phone *string, err error) {
	if user_data := in.UserData; user_data != nil {
		if phone := user_data.Phone; phone != "" {
			return &phone, nil
		}
	}
	return nil, errors.New("Нет записи UserData.Phone")
}
func ApplyTransforms(order *NewOrderInfo, transofrmations []c.Transformation) *NewOrderInfo {
	val := reflect.Indirect(reflect.ValueOf(order))
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		json_tag := val.Type().Field(i).Tag.Get("json")
		field_name := val.Type().Field(i).Name
		if field.IsValid() && field.CanSet() && field.Kind() == reflect.String {
			old := field.String()
			for _, transform := range transofrmations {
				if transform.Field == json_tag || transform.Field == field_name {
					reg := regexp.MustCompile(transform.RegexCode)
					if reg.MatchString(old) {
						new := reg.ReplaceAllString(old, transform.To)
						log.Printf("TCH AH: Matched! Was transform to: %v, result: %v", transform.To, new)
						field.SetString(new)
					}
				}
			}
		}
	}
	return order
}
func (nop *TaxiNewOrderProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	log.Printf("check connect")
	if !nop.API.IsConnected() {
		nop.API.Connect()
		return CONNECTION_ERROR
	}
	order_wrapper, err := nop.Orders.GetByOwnerLast(in.From, nop.context.Name)
	if err != nil {
		return s.ErrorMessageResult(err, nop.context.Commands[CMDS_NOT_CREATED_ORDER])
	}

	if order_wrapper == nil || order_wrapper.Active == false {
		commands := *in.Message.Commands
		phone, err := _get_phone(in)
		if err != nil {
			uwrpr, _ := nop.Users.GetUserById(in.From)
			if uwrpr == nil {
				return s.ErrorMessageResult(errors.New("Не предоставлен номер телефона"), nop.context.Commands[CMDS_NOT_CREATED_ORDER])
			} else {
				phone = &(uwrpr.Phone)
			}
		}
		log.Printf("forming order")
		new_order, err := _form_order(commands[0].Form.Fields, nop.AddressHandler)
		if err != nil {
			if _, ok := err.(*AddressNotHere); ok {
				return &s.MessageResult{
					Body: "Адрес не поддерживается этим такси.",
					Commands: nop.context.Commands[CMDS_NOT_CREATED_ORDER],
					Type: "chat",
				}
			} else {
				return s.ErrorMessageResult(
					errors.New(fmt.Sprintf("Не могу определить адрес, потому что %v", err.Error())),
					nop.context.Commands[CMDS_NOT_CREATED_ORDER])
			}
		}
		new_order.Phone = *phone
		if mrkps, ok := nop.context.Settings["markups"]; ok {
			markups, ok := mrkps.([]string)
			if ok {
				new_order.Markups = markups
			}
		}

		new_order = ApplyTransforms(new_order, nop.Config.Api.Transformations)
		log.Printf("sending order")
		ans := nop.API.NewOrder(*new_order)
		if !ans.IsSuccess {
			nop.Errors.StoreError(in.From, ans.Message)
			return s.ErrorMessageResult(errors.New(ans.Message), nop.context.Commands[CMDS_NOT_CREATED_ORDER])
		}
		log.Printf("Order was created! %+v \n with content: %+v", ans, ans.Content)

		err = nop.Orders.AddOrderObject(d.OrderWrapper{OrderState:ORDER_CREATED, Whom:in.From, OrderId:ans.Content.Id, Source:nop.context.Name})
		err = nop.Orders.SetActive(ans.Content.Id, nop.context.Name, true)
		if err != nil {
			ok, message, _ := nop.API.CancelOrder(ans.Content.Id)
			log.Printf("Error at persist order. Cancelling order at external api with result: %v, %v", ok, message)
			return s.ErrorMessageResult(err, nop.context.Commands[CMDS_NOT_CREATED_ORDER])
		}
		text := ""

		not_send_price := false

		if nsp_, ok := nop.context.Settings["not_send_price"]; ok {
			if _nsp, ok := nsp_.(bool); ok {
				not_send_price = _nsp
			}
		}
		if not_send_price {
			text = fmt.Sprintf("Ваш заказ создан! %v", NEW_ORDER_TEXT_INFO)
		} else {
			log.Printf("calculate price")
			cost, _ := nop.API.CalcOrderCost(*new_order)
			if cost == 0 {
				log.Printf("Order %v, %v with ZERO cost", nop.context.Name, new_order)
			}
			//retrieving markup information
			var markup_text string
			if len(new_order.Markups) == 1 {
				markups := nop.API.Markups()
				for _, mkrp := range markups {
					markup_id, _ := strconv.ParseInt(new_order.Markups[0], 10, 64)
					if mkrp.ID == markup_id {
						markup_text = mkrp.Name
						break
					}
				}
				text = fmt.Sprintf("Ваш заказ создан! Стоимость поездки составит %v рублей. %s. %v", cost, markup_text, NEW_ORDER_TEXT_INFO)

			} else {
				text = fmt.Sprintf("Ваш заказ создан! Стоимость поездки составит %v рублей. %v", cost, NEW_ORDER_TEXT_INFO)
			}

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
	if !cop.API.IsConnected() {
		cop.API.Connect()
		return CONNECTION_ERROR
	}
	order_wrapper, err := cop.Orders.GetByOwnerLast(in.From, cop.context.Name)
	if err != nil {
		return s.ErrorMessageResult(err, cop.context.Commands[CMDS_NOT_CREATED_ORDER])
	}
	if order_wrapper == nil || order_wrapper.Active == false {
		return s.ErrorMessageResult(errors.New("У вас нет активных заказов! :("), cop.context.Commands[CMDS_NOT_CREATED_ORDER])
	}

	is_success, message, err := cop.API.CancelOrder(order_wrapper.OrderId)
	if err == nil {
		cop.Orders.SetState(order_wrapper.OrderId, cop.context.Name, ORDER_CANCELED, nil)
		cop.Orders.SetActive(order_wrapper.OrderId, order_wrapper.Source, false)
	}

	if is_success {
		return &s.MessageResult{Body:"Ваш заказ отменен!", Commands: cop.context.Commands[CMDS_NOT_CREATED_ORDER], Type:"chat"}
	} else {
		return &s.MessageResult{Body:fmt.Sprintf("Проблемы с отменой заказа. %v\nЗвони скорее: %+v ", message, cop.alert_phone), Commands: cop.context.Commands[CMDS_NOT_CREATED_ORDER], Type:"chat"}
	}

	commands, err := form_commands(in.From, cop.MainDb, cop.context)
	if err != nil {
		return s.ErrorMessageResult(err, cop.context.Commands[CMDS_NOT_CREATED_ORDER])
	}
	return &s.MessageResult{Body: "У вас нет активных заказов!", Commands:commands}
}

type TaxiCalculatePriceProcessor struct {
	ExternalApiMixin
	context        *m.BotContext
	AddressHandler AddressHandler
	Config         c.TaxiConfig
}

func (cpp *TaxiCalculatePriceProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	if !cpp.API.IsConnected() {
		cpp.API.Connect()
		return CONNECTION_ERROR
	}

	commands := *in.Message.Commands
	order, err := _form_order(commands[0].Form.Fields, cpp.AddressHandler)
	if err != nil {
		return s.ErrorMessageResult(err, cpp.context.Commands[CMDS_NOT_CREATED_ORDER])
	}
	cost_s, _ := cpp.API.CalcOrderCost(*order)
	cost := strconv.Itoa(cost_s)
	return &s.MessageResult{Body: fmt.Sprintf("Стоимость будет всего лишь %v рублей!", cost), Type:"chat"}
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
	rate = 5
	return fdb, rate
}

func (fp *TaxiFeedbackProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	if !fp.API.IsConnected() {
		fp.API.Connect()
		return CONNECTION_ERROR
	}

	commands := *in.Message.Commands
	fdbk, rate := _get_feedback(commands[0].Form.Fields)
	phone, err := _get_phone(in)
	if phone == nil {
		user, _ := fp.Users.GetUserById(in.From)
		if user == nil {
			log.Printf("Error at implying user by id %v", in.From)
			return s.ErrorMessageResult(errors.New("Не могу определить пользователя"), fp.context.Commands[CMDS_FEEDBACK])
		} else {
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
		result_commands, err := form_commands(in.From, fp.MainDb, fp.context)
		if err != nil {
			return s.ErrorMessageResult(err, fp.context.Commands[CMDS_NOT_CREATED_ORDER])
		}
		return &s.MessageResult{Body:"Спасибо! Ваш отзыв очень важен для нас:)", Commands: result_commands, Type:"chat"}
	} else {
		return &s.MessageResult{Body:"Оплаченный заказ не найден :( Отзывы могут быть только для оплаченных заказов", Commands:fp.context.Commands[CMDS_NOT_CREATED_ORDER], Type:"chat"}
	}
}
