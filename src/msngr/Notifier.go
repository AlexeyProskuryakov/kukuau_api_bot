package msngr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	inf "msngr/infinity"
	"net/http"
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

	req.Header.Add("ContentType", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", n.key)

	log.Printf("N >> %+v", req)

	client := &http.Client{}
	resp, err := client.Do(req)
	warn(err)

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	log.Printf("N << %+v", string(resp))

	fmt.Println("N response Status:", resp.Status)
	fmt.Println("N response Headers:", resp.Header)
	fmt.Println("N response Body:", string(body))
}

func FormNotification(order_id int64, state int, ohm OrderHandlerMixin, carCache *inf.CarsCache) OutPkg {
	order_wrapper := ohm.Orders.GetByOrderId(order_id)
	state_text, ok := inf.StatusesMap[order_wrapper.OrderState]
	if !ok {
		state_text = "не опознанно"
	}
	car_id := order_wrapper.OrderObject.IDCar
	car_info := carCache.CarInfo(car_id)

	text := fmt.Sprintf("%v %v с номером %v %v", car_info.Color, car_info.Model, car_info.Number, state_text)
	out := OutPkg{To: order_wrapper.Whom, Message: &OutMessage{ID: genId(), Type: "chat", Body: text}}
	return out
}
