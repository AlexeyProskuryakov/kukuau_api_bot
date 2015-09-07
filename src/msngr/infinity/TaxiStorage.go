package infinity

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type ordersDbHandler struct {
	collection *mgo.Collection
	session    *mgo.Session
}

type OrderHandlerMixin struct {
	Orders *ordersDbHandler
}

type OrderWrapper struct {
	OrderState int
	OrderId    int64
	When       time.Time
	Whom       string
}

func except(e error) {
	if e != nil {
		panic(e)
	}
}

func (odbh *ordersDbHandler) reConnect(conn string, dbname string) {
	session, err := mgo.Dial(conn)
	except(err)
	session.SetMode(mgo.Monotonic, true)
	odbh.session = session

	collection := session.DB(dbname).C("orders")
	index := mgo.Index{
		Key:        []string{"order_id", "State"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	collection.EnsureIndex(index)
	odbh.collection = collection
}

func NewOrderHandler(conn string, dbname string) *ordersDbHandler {
	odbh := ordersDbHandler{}
	odbh.reConnect(conn, dbname)
	return &odbh
}

func (odbh *ordersDbHandler) GetState(order_id int64) int {
	result := OrderWrapper{}
	err := odbh.collection.Find(bson.M{"orderid": order_id}).One(&result)
	if err != nil {
		return -1
	}
	return result.OrderState
}

func (odbh *ordersDbHandler) SetState(order_id int64, new_state int) {
	change := bson.M{"$set": bson.M{"orderstate": new_state, "when": time.Now()}}
	err := odbh.collection.Update(bson.M{"orderid": order_id}, change)
	except(err)
}

func (odbh *ordersDbHandler) AddOrder(order_id int64, whom string) {
	wrapper := OrderWrapper{
		When:       time.Now(),
		Whom:       whom,
		OrderId:    order_id,
		OrderState: 0,
	}
	err := odbh.collection.Insert(&wrapper)
	except(err)
}

func (odbh *ordersDbHandler) GetOrderIdByOwner(whom string) int64 {
	result := OrderWrapper{}
	err := odbh.collection.Find(bson.M{"whom": whom}).Sort("when").One(&result)
	if err != nil {
		return -1
	}
	return result.OrderId
}
