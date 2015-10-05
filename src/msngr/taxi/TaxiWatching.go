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

func FormNotification(whom string, order_id int64, state int, previous_state int, car_info CarInfo, deliv_time time.Time) *s.OutPkg {
	var text string
	switch state {
	case 2:
		_time, _ := time.Parse(deliv_time.Format("15:04"), "15:04")
		text = fmt.Sprintf("%v %v, время подачи %v.", nominated, car_info, _time)
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
	case 7, 9:
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
			log.Print("get ")
			db_order, err := taxiContext.DataBase.Orders.GetById(api_order.ID, botContext.Name)
			if err != nil {
				log.Printf("WATCH some error in retrieve order [%+v]", api_order)
				continue
			}
			if db_order == nil {
				log.Printf("WATCH order [%+v] is not present in system :(\n", api_order)
				continue
			}
			if db_order.OrderState == ORDER_CANCELED {
//				log.Printf("WATCH order [%+v] is CANCELED", db_order.OrderId)
				continue
			}
			if api_order.State != db_order.OrderState {
				log.Printf("WATCH state of: %+v is updated (api: %v != db: %v)", api_order.ID, api_order.State, db_order.OrderState)
				order_data := api_order.ToOrderData()
				err := taxiContext.DataBase.Orders.SetState(api_order.ID, botContext.Name, api_order.State, &order_data)
				if err != nil {
					log.Printf("WATCH for order %+v can not update status %+v", api_order.ID, api_order.State)
					continue
				}
				car_info := taxiContext.Cars.CarInfo(api_order.IDCar)

				if car_info != nil {
					var notification_data *s.OutPkg
					prev_state, ok := previous_states[api_order.ID]
					delivery_time, err := time.Parse("2006-01-02 15:04:05", api_order.DeliveryTime)
					if err != nil {
						delivery_time = time.Now()
					}

					if ok {
						notification_data = FormNotification(db_order.Whom, api_order.ID, api_order.State, prev_state, *car_info, delivery_time)
					} else {
						notification_data = FormNotification(db_order.Whom, api_order.ID, api_order.State, -1, *car_info, delivery_time)
					}
					if notification_data != nil {
						notification_data.Message.Commands = form_commands_for_current_order(db_order, botContext.Commands)
						taxiContext.Notifier.Notify(*notification_data)
					}
				}
			}
			previous_states[api_order.ID] = api_order.State
		}
		time.Sleep(3 * time.Second)
	}
}

