package taxi

import (
	"log"
	"fmt"
	"time"

	s "msngr/structs"
	d "msngr/db"
	u "msngr/utils"
	n "msngr/notify"
	m "msngr"
)

const (
	car_arrived = "Машина на месте."
	car_set_out = "Машина выехала"
	good_passage = "Приятной Вам поездки!"
	nominated = "Вам назначен: "
	order_canceled = "Ваш заказ отменен!"
)

func FormNotification(context *TaxiContext, ow *d.OrderWrapper, previous_state int, car_info CarInfo, deliv_time time.Time) *s.OutPkg {
	var text string
	switch ow.OrderState {
	case ORDER_ASSIGNED:
		if previous_state != ORDER_ASSIGNED {
			text = fmt.Sprintf("%v %v, время подачи %v.", nominated, car_info, deliv_time.Format("15:04"))
		}
	case ORDER_CAR_SET_OUT:
		if previous_state != ORDER_CAR_SET_OUT {
			text = fmt.Sprintf("%v, время подачи %v", car_set_out, deliv_time.Format("15:04"))
		}
	case ORDER_CLIENT_WAIT:
		if previous_state == ORDER_CREATED {
			text = fmt.Sprintf("%v %v %v %v.", car_arrived, good_passage, nominated, car_info)
		} else {
			text = fmt.Sprintf("%v %v", car_arrived, good_passage)
		}
	case ORDER_IN_PROCESS:
		if u.In(previous_state, []int{ORDER_CLIENT_WAIT, ORDER_DOWNTIME}) {
			return nil
		} else if previous_state == ORDER_CREATED {
			text = fmt.Sprintf("%v %v %v %v.", car_arrived, good_passage, nominated, car_info)
		} else {
			text = fmt.Sprintf("%v %v", car_arrived, good_passage)
		}

	case ORDER_PAYED:
		text = "Заказ выполнен! Спасибо что воспользовались услугами нашей компании."
		context.DataBase.Orders.SetActive(ow.OrderId, ow.Source, false)

	case ORDER_CANCELED:
		if !u.In(previous_state, []int{ORDER_PAYED, ORDER_NOT_PAYED}) {
			text = "Заказ выполнен! Спасибо что воспользовались услугами нашей компании."
		} else {
			text = order_canceled
		}
		context.DataBase.Orders.SetActive(ow.OrderId, ow.Source, false)

	//	default:
	//		status, _ := StatusesMap[state]
	//		text = fmt.Sprintf("Машина %v %v c номером %v перешла в состояние [%v]", car_info.Color, car_info.Model, car_info.Number, status)
	}

	if text != "" {
		out := s.OutPkg{To: ow.Whom, Message: &s.OutMessage{ID: u.GenId(), Type: "chat", Body: text}}
		return &out
	}
	return nil
}

type CarsCache struct {
	cars map[int64]CarInfo
	api  TaxiInterface
}

func _create_cars_map(i TaxiInterface) map[int64]CarInfo {
	cars_map := make(map[int64]CarInfo)
	for !i.IsConnected() {
		log.Printf("Can not create cars cache because taxi api is not response")
		time.Sleep(3 * time.Second)
	}
	cars_info := i.GetCarsInfo()
	if len(cars_info) == 0 {
		log.Printf("Cars cache will be empty :( Because api is responsed empty cars list")
	}
	for _, info := range cars_info {
		cars_map[info.ID] = info
	}
	return cars_map
}

func NewCarsCache(i TaxiInterface) *CarsCache {
	cars_map := _create_cars_map(i)
	handler := CarsCache{cars: cars_map, api: i}
	return &handler
}
func (ch *CarsCache) Reload() {
	ch.cars = _create_cars_map(ch.api)
}

func (ch *CarsCache) GetCarInfo(car_id int64) *CarInfo {
	key, ok := ch.cars[car_id]
	if !ok {
		ch.cars = _create_cars_map(ch.api)
		key, ok = ch.cars[car_id]
		if !ok {
			return nil
		}
	}
	return &key
}

type TaxiContext struct {
	API      TaxiInterface
	DataBase *d.MainDb
	Cars     *CarsCache
	Notifier *n.Notifier
}

func get_arrival_time(api_order Order) time.Time {
	arrival_time := api_order.TimeArrival
	if arrival_time == nil {
		arrival_time = api_order.TimeDelivery
	}
	if arrival_time == nil || arrival_time.Before(time.Now()) {
		arrival_time_ := time.Now().Add(5 * time.Minute)
		arrival_time = &arrival_time_
	}
	return *arrival_time
}

func process_state_change(taxiContext *TaxiContext, botContext *m.BotContext, api_order Order, db_order *d.OrderWrapper, previous_states map[int64]int) {
	log.Printf("WATCH [%v] state of: %+v is updated (api: %v != db: %v)", botContext.Name, api_order.ID, api_order.State, db_order.OrderState)
	order_data := api_order.ToOrderData()
	err := taxiContext.DataBase.Orders.SetState(api_order.ID, botContext.Name, api_order.State, &order_data)
	if err != nil {
		log.Printf("WATCH [%v] for order %+v can not update status %+v", botContext.Name, api_order.ID, api_order.State)
		return
	}
	db_order.OrderState = api_order.State
	db_order.OrderId = api_order.ID

	//если заказ отменил не пользователь
	if api_order.State == ORDER_CANCELED {
		log.Printf("WATCH [%v] NOTIFYING THAT ORDER IS CANCELED", botContext.Name)
		taxiContext.Notifier.Notify(s.OutPkg{
			To:db_order.Whom,
			Message: &s.OutMessage{
				ID: u.GenId(),
				Type: "chat",
				Body: "Ваш заказ отменен!",
				Commands: botContext.Commands["commands_at_not_created_order"],
			},
		})
		previous_states[api_order.ID] = api_order.State
		taxiContext.DataBase.Orders.SetActive(api_order.ID, botContext.Name, false)
		return
	}
	//
	if car_info := taxiContext.Cars.GetCarInfo(api_order.IDCar); car_info != nil {
		var notification_data *s.OutPkg
		arrival_time := get_arrival_time(api_order)
		prev_state, ok := previous_states[api_order.ID]
		if ok {
			notification_data = FormNotification(taxiContext, db_order, prev_state, *car_info, arrival_time)
		} else {
			notification_data = FormNotification(taxiContext, db_order, -1, *car_info, arrival_time)
		}
		if notification_data != nil {
			notification_data.Message.Commands = form_commands_for_current_order(db_order, botContext.Commands)
			taxiContext.Notifier.Notify(*notification_data)
			log.Printf("WATCH [%v] sended for order [%+v]:\n %#v \n and notify that: \n %#v", botContext.Name, db_order.OrderId, notification_data.Message.Commands, *notification_data)
		}
	}
}

func process_car_state(taxiContext *TaxiContext, botContext *m.BotContext, api_order Order, db_order *d.OrderWrapper) {
	var result s.OutPkg

	result.To = db_order.Whom
	car_info := taxiContext.Cars.GetCarInfo(api_order.IDCar)
	if car_info == nil {
		log.Printf("ALERT! CAR CHANGED TO NOT RECOGNIZED ID")
		return
	}
	result.Message = &s.OutMessage{Body:fmt.Sprintf("Ваша машина изменилась на %v", car_info)}
	taxiContext.Notifier.Notify(result)
}

const (
	CHANGE_STATE = "state"
	CHANGE_CAR = "car"
)

func get_changes(api_order Order, db_order *d.OrderWrapper) []string {
	var buff []string
	if result, ok := db_order.OrderData.Get("IDCar").(int64); ok && result != api_order.IDCar {
		buff = append(buff, CHANGE_CAR)
	}
	if api_order.State != db_order.OrderState {
		buff = append(buff, CHANGE_STATE)
	}
	return buff

}

var PreviousStates = make(map[int64]int)

func TaxiOrderWatch(taxiContext *TaxiContext, botContext *m.BotContext) {
	for {
		api_orders := taxiContext.API.Orders()
		//		log.Printf("result of orders request: %v", len(api_orders))
		for _, api_order := range api_orders {
			db_order, err := taxiContext.DataBase.Orders.GetById(api_order.ID, botContext.Name)
			if err != nil {
				log.Printf("WATCH [%v] some error in retrieve order: %v\nOrder:\n[%+v]", botContext.Name, err, api_order)
				continue
			}
			if db_order == nil {
				continue
			}

			for _, el := range get_changes(api_order, db_order) {
				switch el {
				case CHANGE_STATE:
					process_state_change(taxiContext, botContext, api_order, db_order, PreviousStates)
				case CHANGE_CAR:
					process_car_state(taxiContext, botContext, api_order, db_order)
				}
			}
			PreviousStates[api_order.ID] = api_order.State
		}
		time.Sleep(1 * time.Second)
	}
}

