package msngr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	inf "msngr/infinity"
	"net/http"
	"time"
)

func warn(err error) {
	if err != nil {
		log.Println("notifier: ", err)
	}
}
func warnp(err error) {
	if err != nil {
		log.Println("notifier: ", err)
		panic(err)
	}
}

type Notifier struct {
	address string
	key     string
}

func NewNotifier(addr, key string) *Notifier {
	return &Notifier{address: addr, key: key}
}

func (n Notifier) Notify(outPkg OutPkg) {
	jsoned_out, err := json.Marshal(&outPkg)
	warn(err)

	body := bytes.NewBuffer(jsoned_out)
	req, err := http.NewRequest("POST", n.address, body)
	warnp(err)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", n.key)

	log.Printf("N >> %+v", req)

	client := &http.Client{}
	resp, err := client.Do(req)
	warn(err)

	if resp != nil {
		defer resp.Body.Close()
		fmt.Println("N response Status:", resp.Status)
		fmt.Println("N response Headers:", resp.Header)
		resp_body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("N response Body:", string(resp_body))
	}

}

//todo notifier must be more flexiability. you can notify different messages for different engines (shop, taxi)
//you must think about it
func FormNotification(order_id int64, state int, ohm DbHandlerMixin, carCache *inf.CarsCache) *OutPkg {
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
		text = "Заказ выполнен. Спасибо что воспользовались услугами нашей компании."
		commands = commands_for_order_feedback
	}

	if text != "" {
		out := OutPkg{To: order_wrapper.Whom, Message: &OutMessage{ID: genId(), Type: "chat", Body: text, Commands: commands}}
		return &out
	}
	return nil
}
