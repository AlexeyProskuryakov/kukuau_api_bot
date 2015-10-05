package taxi

import (
	"log"
	"math/rand"
	"time"
	"encoding/json"
)

//////////////////////////////////////////////////////////////////////////
///////THIS IS FAKE API FOR TEST//////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////

var fakeInstance *FakeTaxiAPI

func GetFakeInfinityAPI(params ApiParams) TaxiInterface {
	if fakeInstance == nil {
		log.Print("params:::",params.Fake)
		fakeInstance = &FakeTaxiAPI{SleepTime:params.Fake.SleepTime, SendedStates:params.Fake.SendedStates}
	}
	return fakeInstance
}


type FakeTaxiAPI struct {
	SleepTime    int
	SendedStates []int
	orders       []Order
}

func send_states(order_id int64, inf *FakeTaxiAPI) {
	log.Printf("FA will send fake states for order %v", order_id, inf.SendedStates)
	for _, i := range inf.SendedStates {
		time.Sleep(time.Duration(inf.SleepTime) * time.Second)
		log.Println("FA send state: ", i)
		inf.set_order_state(order_id, i)
	}
}

func (inf *FakeTaxiAPI) set_order_state(order_id int64, new_state int) {
	for i, order := range inf.orders {
		if order.ID == order_id {
			inf.orders[i].State = new_state
		}
	}
}

func (inf *FakeTaxiAPI) NewOrder(order NewOrder) Answer {
	saved_order := Order{
		ID:    rand.Int63(),
		State: 1,
		Cost:  150,
		IDCar: 5033615557,
	}
	result, _ := json.Marshal(order)
	log.Printf("4 END NO RESULT DATA\n%+v\n", string(result))

	inf.orders = append(inf.orders, saved_order)

	ans := Answer{
		IsSuccess: true,
		Message:   "test order was formed",
	}
	ans.Content.Id = saved_order.ID
	ans.Content.Cost = 150
	log.Println("FA now i have orders: ", len(inf.orders))

	go send_states(saved_order.ID, inf)

	return ans
}

func (inf *FakeTaxiAPI) Orders() []Order {
	return inf.orders
}

func (inf *FakeTaxiAPI) CancelOrder(order_id int64) (bool, string) {
	log.Println("FA order was canceled", order_id)
	for i, order := range inf.orders {
		if order.ID == order_id {
			inf.orders[i].State = ORDER_CANCELED
			return true, "test order was cancelled"
		}
	}
	return true, "Test order not found :( "
}

func (p *FakeTaxiAPI) CalcOrderCost(order NewOrder) (int, string) {
	log.Println("FA calulate cost for order: ", order)
	return 100500, "Good cost!"
}

func (p *FakeTaxiAPI) Feedback(f Feedback) (bool, string) {
	return true, "Test feedback was received! Thanks!"
}

func (p *FakeTaxiAPI) IsConnected() bool {
	return true
}

func (p *FakeTaxiAPI) GetCarsInfo() []CarInfo {
	return []CarInfo{
		CarInfo{
			ID:5033615557,
			Number:"В777ОР",
			Color:"ультрамариновый",
			Model:"Боливар",
		},
	}
}