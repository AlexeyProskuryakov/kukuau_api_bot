package msngr

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	inf "msngr/infinity"
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
	OrderState  int   `bson:"order_state"`
	OrderId     int64 `bson:"order_id"`
	When        time.Time
	Whom        string
	OrderObject *inf.Order `bson:"order_object"`
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
	orders_index := mgo.Index{
		Key:        []string{"order_id", "order_state"},
		Unique:     true,
		DropDups:   true,
		Background: true,
	}
	collection.EnsureIndex(orders_index)

	owners_index := mgo.Index{
		Key:        []string{"whom"},
		Background: true,
		Unique:     false,
	}
	collection.EnsureIndex(owners_index)

	when_index := mgo.Index{
		Key:        []string{"when"},
		Background: true,
		Unique:     false,
	}

	collection.EnsureIndex(when_index)

	odbh.collection = collection
}

func NewOrderHandler(conn string, dbname string) *ordersDbHandler {
	odbh := ordersDbHandler{}
	odbh.reConnect(conn, dbname)
	return &odbh
}

func (odbh *ordersDbHandler) GetState(order_id int64) int {
	result := OrderWrapper{}
	err := odbh.collection.Find(bson.M{"order_id": order_id}).One(&result)
	if err != nil {
		return -1
	}
	return result.OrderState
}

func (odbh *ordersDbHandler) SetState(order_id int64, new_state int, order inf.Order) {
	change := bson.M{"$set": bson.M{"order_state": new_state, "when": time.Now(), "order_object": order}}
	err := odbh.collection.Update(bson.M{"order_id": order_id}, change)
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

func (odbh *ordersDbHandler) GetByOwner(whom string) *OrderWrapper {
	result := OrderWrapper{}
	err := odbh.collection.Find(bson.M{"whom": whom}).Sort("when").One(&result)
	if err != nil {
		return nil
	}
	return &result
}

func (odbh *ordersDbHandler) GetByOrderId(order_id int64) *OrderWrapper {
	result := OrderWrapper{}
	err := odbh.collection.Find(bson.M{"order_id": order_id}).One(&result)
	if err != nil {
		return nil
	}
	return &result
}
