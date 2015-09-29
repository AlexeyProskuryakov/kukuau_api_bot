package taxi

import (
	"log"
	"fmt"
	s "msngr/structs"
	d "msngr/db"
	u "msngr/utils"
	n "msngr/notify"
	"time"

)

const (
	car_arrived = "Машина на месте."
	car_set_out = "Машина выехала."
	good_passage = "Приятной Вам поездки!"
	nominated = "Вам назначен: "

)

func FormNotification(whom string, order_id int64, state int, previous_state int, car_info CarInfo) *s.OutPkg {
	var text string
	switch state {
	case 2:
		text = fmt.Sprintf("%v %v, время подачи %v.", nominated, car_info, u.GetTimeAfter(5 * time.Minute, "15:04"))
	case 3:
		text = fmt.Sprintf("%v", car_set_out)
	case 4:
		if previous_state == 1 {
			text = fmt.Sprintf("%v %v %v %v.", car_arrived, good_passage, nominated, car_info)
		} else {
			text = fmt.Sprintf("%v %v", car_arrived, good_passage)
		}
	case 5:
		if previous_state == 4 {
			return nil
		} else if previous_state == 1 {
			text = fmt.Sprintf("%v %v %v %v.", car_arrived, good_passage, nominated, car_info)
		} else {
			text = fmt.Sprintf("%v %v", car_arrived, good_passage)
		}
	case 7:
		text = "Заказ выполнен! Спасибо что воспользовались услугами нашей компании."
	//	default:
	//		status, _ := StatusesMap[state]
	//		text = fmt.Sprintf("Машина %v %v c номером %v перешла в состояние [%v]", car_info.Color, car_info.Model, car_info.Number, status)
	}
	if text != "" {
		out := s.OutPkg{To: whom, Message: &s.OutMessage{ID: u.GenId(), Type: "chat", Body: text}}
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
	cars_info := i.GetCarsInfo()
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

func (ch *CarsCache) CarInfo(car_id int64) *CarInfo {
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
	DataBase *d.DbHandlerMixin
	Cars     *CarsCache
	Notifier *n.Notifier
}

func TaxiOrderWatch(taxiContext *TaxiContext, botContext *s.BotContext) {
	previous_states := map[int64]int{}
	for {
		api_orders := taxiContext.API.Orders()
		for _, api_order := range api_orders {
			db_order_state, err := taxiContext.DataBase.Orders.GetState(api_order.ID, botContext.Name)
			if err != nil {
				log.Printf("OW order %+v is not present in system :(\n", api_order)
				continue
			}
			if api_order.State != db_order_state.OrderState {
				log.Printf("OW state of: %+v is updated (api: %v != db: %v)", api_order.ID, api_order.State, db_order_state.OrderState)
				order_data := api_order.ToOrderData()
				err := taxiContext.DataBase.Orders.SetState(api_order.ID, botContext.Name, api_order.State, &order_data)
				if err != nil {
					log.Printf("!!! OW for order %+v can not update status %+v", api_order.ID, api_order.State)
					continue
				}
				order_wrapper := taxiContext.DataBase.Orders.GetOrderById(api_order.ID, botContext.Name)
				log.Printf("OW updated order: %+v", order_wrapper)
				car_info := taxiContext.Cars.CarInfo(api_order.IDCar)

				if car_info != nil {
					var notification_data *s.OutPkg
					prev_state, ok := previous_states[api_order.ID]
					if ok {
						notification_data = FormNotification(order_wrapper.Whom, api_order.ID, api_order.State, prev_state, *car_info)
					} else {
						notification_data = FormNotification(order_wrapper.Whom, api_order.ID, api_order.State, -1, *car_info)
					}
					if notification_data != nil {
						notification_data.Message.Commands = form_commands_for_current_order(order_wrapper, botContext.Commands)
						taxiContext.Notifier.Notify(*notification_data)
					}
				}
			}
			previous_states[api_order.ID] = api_order.State
		}
		time.Sleep(3 * time.Second)
	}
}

