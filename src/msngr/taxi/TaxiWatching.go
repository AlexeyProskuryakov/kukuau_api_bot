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

func FormNotification(whom string, order_id int64, state int, previous_state int, car_info InfinityCarInfo) *s.OutPkg {
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

func TaxiOrderWatch(db d.DbHandlerMixin, im InfinityMixin, carsCache *CarsCache, notifier *n.Notifier) {
	previous_states := map[int64]int{}
	for {
		api_orders := im.API.Orders()
		for _, api_order := range api_orders {
			db_order_state := db.Orders.GetState(api_order.ID)
			if db_order_state == -1 {
				log.Printf("OW order %+v is not present in system :(\n", api_order)
				continue
			}
			if api_order.State != db_order_state {
				log.Printf("OW state of:\n%+v \nis updated (api: %v != db: %v)", api_order, api_order.State, db_order_state)
				order_data := api_order.ToOrderData()
				err := db.Orders.SetState(api_order.ID, api_order.State, &order_data)
				if err != nil {
					log.Printf("OW for order %+v can not update status %+v", api_order.ID, api_order.State)
					continue
				}
				order_wrapper := db.Orders.GetByOrderId(api_order.ID)
				log.Printf("OW updated order: %+v", order_wrapper)
				car_info := carsCache.CarInfo(api_order.IDCar)

				if car_info != nil {
					var notification_data *s.OutPkg
					prev_state, ok := previous_states[api_order.ID]
					if ok {
						notification_data = FormNotification(order_wrapper.Whom, api_order.ID, api_order.State, prev_state, *car_info)
					} else {
						notification_data = FormNotification(order_wrapper.Whom, api_order.ID, api_order.State, -1, *car_info)
					}
					if notification_data != nil {
						notification_data.Message.Commands = form_commands_for_current_order(order_wrapper)
						notifier.Notify(*notification_data)
					}
				}
			}

			previous_states[api_order.ID] = api_order.State

		}
		time.Sleep(3 * time.Second)
	}
}

