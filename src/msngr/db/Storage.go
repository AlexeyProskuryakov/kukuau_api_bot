package db

import (
	"crypto/md5"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"time"
)

func _check(e error) {
	if e != nil {
		panic(e)
	}
}

type orderHandler struct {
	collection *mgo.Collection
}

type OrderData struct {
	Content map[string]interface{}
}
func NewOrderData(content map[string]interface{}) OrderData{
	return OrderData{Content:content}
}

func (odh *OrderData) Get(key string) interface{} {
	val, ok := odh.Content[key]
	if ok {
		return val
	}else {
		return nil
	}
}

type OrderWrapper struct {
	OrderState int   `bson:"order_state"`
	OrderId    int64 `bson:"order_id"`
	When       time.Time
	Whom       string
	OrderData OrderData `bson:"data"`
	Feedback   string
}

type userHandler struct {
	collection *mgo.Collection
}

type UserWrapper struct {
	State      string `bson:"user_state"`
	UserId     string `bson:"user_id"`
	UserName   string `bson:"user_name"`
	Password   string
	Phone      string

	LastUpdate time.Time `bson:"last_update"`
}

type DbHandlerMixin struct {
	session *mgo.Session

	Orders  *orderHandler
	Users   *userHandler
}


func (odbh *DbHandlerMixin) reConnect(conn string, dbname string) {
	session, err := mgo.Dial(conn)
	_check(err)
	session.SetMode(mgo.Monotonic, true)
	odbh.session = session

	orders_collection := session.DB(dbname).C("orders")

	orders_index := mgo.Index{
		Key:        []string{"order_id"},
		Background: true,
		Unique:     true,
		DropDups:   true,
	}
	orders_collection.EnsureIndex(orders_index)

	state_index := mgo.Index{
		Key:        []string{"order_state"},
		Background: true,
		Unique:     false,
	}
	orders_collection.EnsureIndex(state_index)

	owners_index := mgo.Index{
		Key:        []string{"whom"},
		Background: true,
		Unique:     false,
	}
	orders_collection.EnsureIndex(owners_index)

	when_index := mgo.Index{
		Key:        []string{"when"},
		Background: true,
		Unique:     false,
	}
	orders_collection.EnsureIndex(when_index)

	odbh.Orders = &orderHandler{collection: orders_collection}

	users_collection := session.DB(dbname).C("users")
	users_collection.EnsureIndex(mgo.Index{
		Key:        []string{"user_id"},
		Background: true,
		Unique:     true,
		DropDups:   true,
	})
	users_collection.EnsureIndex(mgo.Index{
		Key:        []string{"last_update"},
		Unique:     false,
		Background: true,
	})
	users_collection.EnsureIndex(mgo.Index{
		Key:        []string{"user_state"},
		Unique:     false,
		Background: true,
	})
	users_collection.EnsureIndex(mgo.Index{
		Key:        []string{"user_name"},
		Unique:     false,
		Background: true,
	})
	odbh.Users = &userHandler{collection: users_collection}
}

func NewDbHandler(conn string, dbname string) *DbHandlerMixin {
	odbh := DbHandlerMixin{}
	odbh.reConnect(conn, dbname)
	return &odbh
}

func (odbh *orderHandler) GetState(order_id int64) int {
	result := OrderWrapper{}
	err := odbh.collection.Find(bson.M{"order_id": order_id}).One(&result)
	if err != nil {
		return -1
	}
	return result.OrderState
}

func (odbh *orderHandler) SetState(order_id int64, new_state int, order_data *OrderData) error {
	var to_set bson.M
	if order_data!=nil{
		to_set = bson.M{"order_state": new_state, "when": time.Now(), "data": order_data}
	} else{
		to_set = bson.M{"order_state": new_state, "when": time.Now()}
	}
	change := bson.M{"$set": to_set}
	log.Println("change:",change["$set"])
	err := odbh.collection.Update(bson.M{"order_id": order_id}, change)
	return err
}

func (oh *orderHandler) SetFeedback(for_whom string, for_state int, feedback string) int64 {
	order := OrderWrapper{}
	err := oh.collection.Find(bson.M{"whom": for_whom, "order_state": for_state}).Sort("-when").One(&order)
	if err!=nil{
		return -1
	}
	oh.collection.Update(bson.M{"order_id": order.OrderId}, bson.M{"$set": bson.M{"feedback": feedback}})
	order_id := order.OrderId
	return order_id
}

func (odbh *orderHandler) AddOrder(order_id int64, whom string) {
	wrapper := OrderWrapper{
		When:       time.Now(),
		Whom:       whom,
		OrderId:    order_id,
		OrderState: 0,
	}
	err := odbh.collection.Insert(&wrapper)
	if err != nil{
		log.Println(err)
	}
}

func (odbh *orderHandler) GetByOwner(whom string) *OrderWrapper {
	result := OrderWrapper{}
	err := odbh.collection.Find(bson.M{"whom": whom}).Sort("-when").One(&result)
	if err != nil {
		return nil
	}
	return &result
}

func (odbh *orderHandler) GetByOrderId(order_id int64) *OrderWrapper {
	result := OrderWrapper{}
	err := odbh.collection.Find(bson.M{"order_id": order_id}).One(&result)
	if err != nil {
		return nil
	}
	return &result
}

const (
	LOGOUT = "LOGOUT"
	LOGIN = "LOGIN"
	REGISTERED = "REGISTERED"
)

func phash(pwd string) (res string) {
	input := []byte(res)
	output := md5.Sum(input)
	res = string(output[:])
	return
}

func (uh *userHandler) CheckUser(req bson.M) *UserWrapper {
	tmp := UserWrapper{}
	err := uh.collection.Find(req).One(&tmp)
	log.Printf("Checking User result is: %+v [%+v]", tmp, err)
	if err != nil {
		return nil
	}
	return &tmp
}

func (uh *userHandler) AddUser(user_id, phone string) {
	tmp := uh.CheckUser(bson.M{"user_id": user_id, "phone": phone})
	if tmp == nil {
		err := uh.collection.Insert(&UserWrapper{UserId: user_id, State: REGISTERED, Phone: phone, LastUpdate: time.Now()})
		_check(err)
	}
}

func (uh *userHandler) SetUserState(user_id string, state string) {
	tmp := uh.CheckUser(bson.M{"user_id": user_id})
	if tmp == nil {
		err := uh.collection.Insert(&UserWrapper{UserId: user_id, State: state, LastUpdate: time.Now()})
		_check(err)
	} else {
		err := uh.collection.Update(
			bson.M{"user_id": user_id},
			bson.M{"$set": bson.M{"user_state": state, "last_update": time.Now()}},
		)
		_check(err)
	}
}
func (uh *userHandler) SetUserPassword(username, password string) {
	tmp := uh.CheckUser(bson.M{"user_name": username})
	if tmp == nil {
		err := uh.collection.Insert(&UserWrapper{UserId: username, UserName: username, Password: password, State: REGISTERED, LastUpdate: time.Now()})
		_check(err)
	} else if phash(password) != tmp.Password {
		log.Println("changing password! for user ", username)
		err := uh.collection.Update(
			bson.M{"user_name": username},
			bson.M{"$set": bson.M{"password": phash(password), "last_update": time.Now()}},
		)
		_check(err)
	}
}
func (uh *userHandler) GetUserState(user_id string) (string, error) {
	result := UserWrapper{}
	err := uh.collection.Find(bson.M{"user_id": user_id}).One(&result)
	return result.State, err
}

func (uh *userHandler) CheckUserPassword(username, password string) bool {
	tmp := UserWrapper{}
	err := uh.collection.Find(bson.M{"user_name": username, "password": phash(password)}).One(&tmp)
	log.Println("ST checking user password", tmp, err)
	return err != nil
}

func (uh *userHandler) GetById(user_id string) (*UserWrapper, error) {
	result := UserWrapper{}
	err := uh.collection.Find(bson.M{"user_id": user_id}).One(&result)
	return &result, err
}
