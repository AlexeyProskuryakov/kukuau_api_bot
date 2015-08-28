package msngr

//experiments with finite state machines
import (
	"fmt"
	"github.com/looplab/fsm"
)

type Taxi struct {
	uh    *userHandler
	state *fsm.FSM
}

const (
	_NO_ORDERS     = "no orders"
	_ORDER_CREATED = "order created"

	_create_order  = "create order"
	_cancel_order  = "cancel order"
	_process_order = "process order"
)

func NewTaxi() {
	_uh := GetUserHandler()
	t := Taxi{
		uh: _uh,
	}
	t.state = fsm.NewFSM(
		_NO_ORDERS,
		fsm.Events{
			{Name: _create_order, Src: []string{_NO_ORDERS}, Dst: _ORDER_CREATED},
			{Name: _cancel_order, Src: []string{_ORDER_CREATED}, Dst: _NO_ORDERS},
			{Name: _process_order, Src: []string{_ORDER_CREATED}, Dst: _NO_ORDERS},
		},
		fsm.Callbacks{
			"enter_create order": func(e *fsm.Event) { t.enterState(e) },
		},
	)
}

func (t *Taxi) enterState(e *fsm.Event) {
	fmt.Println(e, e.Src, e.Dst)
}

type UsersTaxi struct {
	content map[string]Taxi
}

var _usersTaxi = UsersTaxi{
	content: make(map[string]Taxi),
}

func (ut *UsersTaxi) processInput(in InPkg) (out OutPkg) {
	out = OutPkg{}
	ut_map := _usersTaxi.content
	if _, err := ut_map[in.From]; !err {
		if in.Request != nil {
			// out.Request = in.Request

		} else if in.Message != nil {
			// out.Message = in.Message

		}
	}
	return
}
