package msngr

import (
	"bytes"
	"encoding/json"
	"fmt"
	inf "msngr/infinity"
	"net/http"
)

type Notifier struct {
	address string
}

func NewNotifier(addr string) *Notifier {
	return &Notifier{address: addr}
}

func (n Notifier) Notify(outPkg OutPkg) {
	jsoned_out, err := json.Marshal(&outPkg)
	if err != nil {
		panic(err)
	}

	http.Post(n.address, "application/json", bytes.NewBufferString(string(jsoned_out)))
}

func FormNotification(order_id int64, state int, ohm OrderHandlerMixin, carCache *inf.CarsCache) OutPkg {
	orderd_wrapper := ohm.Orders.GetByOrderId(order_id)
	car_id := orderd_wrapper.OrderObject.IDCar
	out := OutPkg{To: orderd_wrapper.Whom, Message: &OutMessage{ID: genId(), Type: "chat", Body: fmt.Sprintf("car id: %s", car_id)}}
	return out
}
