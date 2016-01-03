package db

import (
	"log"
	"time"
	"errors"
	"fmt"
	"sync"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"msngr/utils"
)

const (
	LOGOUT = "LOGOUT"
	LOGIN = "LOGIN"
	REGISTERED = "REGISTERED"
)

type OrderData struct {
	Content map[string]interface{}
}

func NewOrderData(content map[string]interface{}) OrderData {
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
	OrderData  OrderData `bson:"data"`
	Feedback   string
	Source     string
	Active     bool
}

type UserWrapper struct {
	GlobalState string `bson:"global_state"`
	States      map[string]string `bson:"states"`
	UserId      string `bson:"user_id"`
	UserName    string `bson:"user_name"`
	Password    string
	Phone       string

	LastUpdate  time.Time `bson:"last_update"`
}


func (uw *UserWrapper) GetStateValue(state_key string) (string, bool) {
	res, ok := uw.States[state_key]
	return res, ok
}

type ErrorWrapper struct {
	Username string
	Error    string
	Time     time.Time
}

type orderHandler struct {
	Collection *mgo.Collection
	parent     *MainDb
}

type userHandler struct {
	Collection *mgo.Collection
	parent     *MainDb
}

type errorHandler struct {
	Collection *mgo.Collection
	parent     *MainDb
}

type CheckedMixin interface {
	Check() bool
}
type DbHelper struct {
	sync.Mutex
	CheckedMixin

	Conn           string
	DbName         string
	try_to_connect bool

	Session        *mgo.Session

}

func NewDbHelper(conn, dbname string) *DbHelper{
	res := &DbHelper{Conn:conn, DbName:dbname}
	res.reConnect()
	return res
}

type MainDb struct {
	DbHelper
	Orders         *orderHandler
	Users          *userHandler
	Errors         *errorHandler
}

var DELETE_DB = false

func (odbh *DbHelper) Check() bool {
	if odbh.Session != nil && odbh.Session.Ping() == nil {
		return true
	} else if !odbh.try_to_connect {
		go odbh.reConnect()
		return false
	}
	return false
}


func (odbh *DbHelper) reConnect() {
	odbh.Lock()
	odbh.try_to_connect = true
	defer func() {
		odbh.try_to_connect = false
		odbh.Unlock()
	}()

	count := 2500 * time.Millisecond
	var err error
	var session *mgo.Session

	for {
		session, err = mgo.Dial(odbh.Conn)
		if err == nil {
			log.Printf("Connection to mongodb established!")
			session.SetMode(mgo.Strong, true)
			err = session.Ping()
			if err != nil {
				log.Printf("Connection to mongodb is not verified")
				continue
			}
			odbh.Session = session
			log.Printf("Db session is establised")
			break
		} else {
			count += count
			log.Printf("can not connect to db, will sleep %+v and try", count)
			time.Sleep(count)
		}
	}
}


func (odbh *MainDb) ensureIndexes() {
	orders_collection := odbh.Session.DB(odbh.DbName).C("orders")
	orders_collection.EnsureIndex(mgo.Index{
		Key:        []string{"order_id"},
		Background: true,
		DropDups:   true,
	})
	orders_collection.EnsureIndex(mgo.Index{
		Key:        []string{"order_state"},
		Background: true,
	})
	orders_collection.EnsureIndex(mgo.Index{
		Key:[]string{"active"},
		Background:true,
	})
	orders_collection.EnsureIndex(mgo.Index{
		Key:        []string{"whom"},
		Background: true,

	})
	orders_collection.EnsureIndex(mgo.Index{
		Key:        []string{"when"},
		Background: true,
	})
	orders_collection.EnsureIndex(mgo.Index{
		Key:    []string{"source"},
		Background:true,
		Unique:false,
	})

	users_collection := odbh.Session.DB(odbh.DbName).C("users")
	users_collection.EnsureIndex(mgo.Index{
		Key:        []string{"user_id"},
		Background: true,
		Unique:     true,
		DropDups:   true,
	})
	users_collection.EnsureIndex(mgo.Index{
		Key:        []string{"last_update"},
		Background: true,
	})
	users_collection.EnsureIndex(mgo.Index{
		Key:        []string{"user_state"},
		Background: true,
	})
	users_collection.EnsureIndex(mgo.Index{
		Key:        []string{"user_name"},
		Background: true,
	})

	error_collection := odbh.Session.DB(odbh.DbName).C("errors")

	error_collection.EnsureIndex(mgo.Index{
		Key: []string{"username"},
		Unique:false,
	})
	error_collection.EnsureIndex(mgo.Index{
		Key:[]string{"time"},
		Unique:false,
	})

	odbh.Users.Collection = users_collection
	odbh.Orders.Collection = orders_collection
	odbh.Errors.Collection = error_collection
}

func NewMainDb(conn, dbname string) *MainDb {
	helper := DbHelper{Conn:conn, DbName:dbname}
	odbh := MainDb{DbHelper:helper}

	odbh.Users = &userHandler{parent:&odbh}
	odbh.Orders = &orderHandler{parent:&odbh}
	odbh.Errors = &errorHandler{parent:&odbh}

	log.Printf("start reconnecting")
	odbh.reConnect()
	odbh.ensureIndexes()
	return &odbh
}

func (oh *orderHandler) GetById(order_id int64, source string) (*OrderWrapper, error) {
	if !oh.parent.Check() {
		return nil, errors.New("БД не доступна")
	}
	result := OrderWrapper{}
	err := oh.Collection.Find(bson.M{"order_id": order_id, "source": source}).One(&result)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	} else if err == mgo.ErrNotFound {
		return nil, nil
	} else {
		return &result, nil
	}
}

func (oh *orderHandler) SetActive(order_id int64, source string, state bool) error {
	if !oh.parent.Check() {
		utils.After(oh.parent.Check, func() { oh.SetActive(order_id, source, state) })
		return nil
	}
	err := oh.Collection.Update(bson.M{"order_id": order_id, "source":source}, bson.M{"$set":bson.M{"active":state}})
	if err == mgo.ErrNotFound {
		log.Printf("update not existed %v %v to active %v", order_id, source, state)
	}
	return err
}

func (oh *orderHandler) SetState(order_id int64, source string, new_state int, order_data *OrderData) error {
	if !oh.parent.Check() {
		utils.After(oh.parent.Check, func() {oh.SetState(order_id, source, new_state, order_data)})
		return nil
	}
	var to_set bson.M
	if order_data != nil {
		to_set = bson.M{"order_state": new_state, "when": time.Now(), "data": order_data, "active":true}
	} else {
		to_set = bson.M{"order_state": new_state, "when": time.Now()}
	}

	change := bson.M{"$set": to_set}
	log.Println("DB: change:", change["$set"])
	err := oh.Collection.Update(bson.M{"order_id": order_id, "source":source}, change)
	if err != nil && err != mgo.ErrNotFound {
		log.Printf("State [%v] for order [%v] %v is not stated because order is not found", new_state, order_id, source)
		return err
	}
	if err == mgo.ErrNotFound {
		log.Printf("DB: for order %v at %v not found :(( ", order_id, source)
	}
	return err
}

func (oh *orderHandler) SetFeedback(for_whom string, for_state int, feedback string, source string) (*int64, error) {
	if !oh.parent.Check() {
		utils.After(oh.parent.Check, func() { oh.SetFeedback(for_whom, for_state, feedback, source) })
		return nil, nil
	}
	order := OrderWrapper{}
	err := oh.Collection.Find(bson.M{"whom": for_whom, "order_state": for_state, "source":source}).Sort("-when").One(&order)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	}
	if err == mgo.ErrNotFound {
		return nil, errors.New("Заказ не найден!")
	}
	err = oh.Collection.Update(bson.M{"order_id": order.OrderId, "source":source}, bson.M{"$set": bson.M{"feedback": feedback}})
	order_id := order.OrderId
	return &order_id, err
}

func (oh *orderHandler) AddOrder(order_id int64, whom string, source string) error {
	if !oh.parent.Check() {
		return errors.New("БД не доступна")
	}
	wrapper := OrderWrapper{
		When:       time.Now(),
		Whom:       whom,
		OrderId:    order_id,
		OrderState: 1,
		Source: source,
	}
	err := oh.Collection.Insert(&wrapper)
	return err
}

func (oh *orderHandler) AddOrderObject(order *OrderWrapper) error {
	if !oh.parent.Check() {
		return errors.New("БД не доступна")
	}
	order.When = time.Now()
	err := oh.Collection.Insert(order)
	return err
}


func (oh *orderHandler) Count() int {
	result, _ := oh.Collection.Count()
	return result
}

func (oh *orderHandler) GetBy(req bson.M) ([]OrderWrapper, error) {
	if !oh.parent.Check() {
		return nil, errors.New("БД не доступна")
	}

	result := []OrderWrapper{}
	err := oh.Collection.Find(req).Sort("-when").All(&result)
	return result, err
}

func (oh *orderHandler) GetByOwnerLast(whom, source string) (*OrderWrapper, error) {
	if !oh.parent.Check() {
		return nil, errors.New("БД не доступна")
	}
	result := OrderWrapper{}
	err := oh.Collection.Find(bson.M{"whom": whom, "source":source}).Sort("-when").One(&result)
	if err == mgo.ErrNotFound {
		return nil, nil
	}else if err != nil {
		return nil, err
	}
	return &result, nil
}

func (oh *orderHandler) GetByOwner(whom, source string, active bool) (*OrderWrapper, error) {
	if !oh.parent.Check() {
		return nil, errors.New("БД не доступна")
	}
	result := OrderWrapper{}
	err := oh.Collection.Find(bson.M{"whom": whom, "source":source, "active":true}).Sort("-when").One(&result)
	if err == mgo.ErrNotFound {
		return nil, nil
	}else if err != nil {
		return nil, err
	}
	return &result, nil
}



func (oh *orderHandler) GetOrders(q bson.M) ([]OrderWrapper, error) {
	if !oh.parent.Check() {
		return nil, errors.New("БД не доступна")
	}
	var result []OrderWrapper
	err := oh.Collection.Find(q).Sort("-when").One(&result)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	}
	return result, nil
}

func (uh *userHandler) CheckUser(req bson.M) (*UserWrapper, error) {
	if !uh.parent.Check() {
		return nil, errors.New("БД не доступна")
	}
	tmp := UserWrapper{}
	err := uh.Collection.Find(req).One(&tmp)
	if err == mgo.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, errors.New(fmt.Sprintf("user for %+v is not found", req))
	}
	return &tmp, nil
}

func (uh *userHandler) AddUser(user_id, phone string) error {
	if !uh.parent.Check() {
		return errors.New("БД не доступна")
	}
	tmp, err := uh.CheckUser(bson.M{"user_id": user_id, "phone": phone})
	if tmp == nil {
		err = uh.Collection.Insert(&UserWrapper{UserId: user_id, GlobalState: REGISTERED, Phone: phone, LastUpdate: time.Now()})
		return err
	}
	return nil
}

func (uh *userHandler) SetUserGlobalState(user_id string, state string) error {
	if !uh.parent.Check() {
		return errors.New("БД не доступна")
	}
	tmp, _ := uh.CheckUser(bson.M{"user_id": user_id})
	if tmp == nil {
		err := uh.Collection.Insert(&UserWrapper{UserId: user_id, GlobalState: state, LastUpdate: time.Now()})
		return err
	} else {
		err := uh.Collection.Update(
			bson.M{"user_id": user_id},
			bson.M{"$set": bson.M{"user_state": state, "last_update": time.Now()}},
		)
		return err
	}
}

func (uh *userHandler) SetUserMultiplyState(user_id, state_key, state_value string) error {
	/**
	Выставление сосотяние по определенному аспекту. к примеру для квестов. Или для еще какой хуйни, посему требуется ключ да значение.
	Отличается от просто SetUserState тем что там выставляется состояние глобальное
	 */
	if !uh.parent.Check() {
		return errors.New("БД не доступна")
	}
	tmp, _ := uh.CheckUser(bson.M{"user_id": user_id})
	if tmp == nil {
		err := uh.Collection.Insert(&UserWrapper{UserId: user_id, States: map[string]string{state_key:state_value}, LastUpdate: time.Now()})
		return err
	} else {
		err := uh.Collection.Update(
			bson.M{"user_id": user_id},
			bson.M{"$set": bson.M{fmt.Sprintf("states.%v", state_key): state_value, "last_update": time.Now()}},
		)
		return err
	}
}

func (uh *userHandler) GetUserMultiplyState(user_id, state_key string) (string, error) {
	if !uh.parent.Check() {
		return "", errors.New("БД не доступна")
	}
	tmp, _ := uh.CheckUser(bson.M{"user_id": user_id})
	if tmp == nil {
		return "", errors.New("Пользователь не найден")
	} else {
		if state, ok := tmp.States[state_key]; ok {
			return state, nil
		}
		return "", errors.New("This user have not this key of state")
	}
}

func (uh *userHandler) SetUserPassword(username, password string) error {
	if !uh.parent.Check() {
		return errors.New("БД не доступна")
	}
	tmp, _ := uh.CheckUser(bson.M{"user_name": username})
	if tmp == nil {
		err := uh.Collection.Insert(&UserWrapper{UserId: username, UserName: username, Password: password, GlobalState: REGISTERED, LastUpdate: time.Now()})
		return err
	} else if utils.PHash(password) != tmp.Password {
		log.Println("changing password! for user ", username)
		err := uh.Collection.Update(
			bson.M{"user_name": username},
			bson.M{"$set": bson.M{"password": utils.PHash(password), "last_update": time.Now()}},
		)
		return err
	}
	return nil
}

func (uh *userHandler) GetUserGlobalState(user_id string) (*string, error) {
	if !uh.parent.Check() {
		return nil, errors.New("БД не доступна")
	}
	result := UserWrapper{}
	err := uh.Collection.Find(bson.M{"user_id": user_id}).One(&result)
	return &(result.GlobalState), err
}

func (uh *userHandler) CheckUserPassword(username, password string) (bool, error) {
	if !uh.parent.Check() {
		return false, errors.New("БД не доступна")
	}
	tmp := UserWrapper{}
	err := uh.Collection.Find(bson.M{"user_name": username, "password": utils.PHash(password)}).One(&tmp)
	return err != nil, err
}

func (uh *userHandler) GetUserById(user_id string) (*UserWrapper, error) {
	if !uh.parent.Check() {
		return nil, errors.New("БД не доступна")
	}
	result := UserWrapper{}
	err := uh.Collection.Find(bson.M{"user_id": user_id}).One(&result)
	if err != nil && err != mgo.ErrNotFound {
		log.Printf("Ощибка определения пользователя %v", err)
		return nil, err
	}
	return &result, err
}

func (uh *userHandler) Count() int {
	r, _ := uh.Collection.Count()
	return r
}

func (uh *userHandler) GetBy(req bson.M) (*[]UserWrapper, error) {
	if !uh.parent.Check() {
		return nil, errors.New("БД не доступна")
	}
	result := []UserWrapper{}
	err := uh.Collection.Find(req).Sort("last_update").All(&result)
	return &result, err
}


func (eh *errorHandler) StoreError(username, error string) error {
	if !eh.parent.Check() {
		return errors.New("БД не доступна")
	}
	result := ErrorWrapper{Username:username, Error:error, Time:time.Now()}
	err := eh.Collection.Insert(&result)
	return err
}

func (eh *errorHandler) GetBy(req bson.M) (*[]ErrorWrapper, error) {
	if !eh.parent.Check() {
		return nil, errors.New("БД не доступна")
	}

	result := []ErrorWrapper{}
	err := eh.Collection.Find(req).Sort("time").All(&result)
	return &result, err
}