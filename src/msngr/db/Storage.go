package db

import (
	"log"
	"time"
	"errors"
	"fmt"
	"sync"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"reflect"

	u "msngr/utils"
	s "msngr/structs"
	c "msngr/configuration"
)

const (
	LOGOUT = "LOGOUT"
	LOGIN = "LOGIN"
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
	} else {
		return nil
	}
}

type OrderWrapper struct {
	OrderState int   `bson:"order_state"`
	OrderId    int64 `bson:"order_id"`
	When       time.Time `bson:"when"`
	Whom       string  `bson:"whom"`
	OrderData  OrderData `bson:"data"`
	Feedback   string `bson:"feedback,omitempty"`
	Source     string `bson:"source"`
	Active     bool `bson:"is_active"`
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

func NewDbHelper(conn, dbname string) *DbHelper {
	res := &DbHelper{Conn:conn, DbName:dbname}
	res.reConnect()
	return res
}

type MainDb struct {
	DbHelper
	Orders   *orderHandler
	Users    *UserHandler
	Errors   *errorHandler
	Messages *MessageHandler
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
		Key:[]string{"is_active"},
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

	error_collection := odbh.Session.DB(odbh.DbName).C("errors")
	error_collection.EnsureIndex(mgo.Index{
		Key: []string{"username"},
		Unique:false,
	})
	error_collection.EnsureIndex(mgo.Index{
		Key:[]string{"time"},
		Unique:false,
	})

	odbh.Orders.Collection = orders_collection
	odbh.Errors.Collection = error_collection
	odbh.Users.ensureIndexes()
	odbh.Messages.ensureIndexes()
}

func NewMainDb(conn, dbname string) *MainDb {
	helper := DbHelper{Conn:conn, DbName:dbname}
	odbh := MainDb{DbHelper:helper}

	odbh.Users = &UserHandler{parent:&odbh}
	odbh.Orders = &orderHandler{parent:&odbh}
	odbh.Errors = &errorHandler{parent:&odbh}
	odbh.Messages = &MessageHandler{parent:&odbh}

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
		u.After(oh.parent.Check, func() {
			oh.SetActive(order_id, source, state)
		})
		return nil
	}
	err := oh.Collection.Update(bson.M{"order_id": order_id, "source":source}, bson.M{"$set":bson.M{"is_active":state}})
	if err == mgo.ErrNotFound {
		log.Printf("DB: update not existed %v %v to active %v", order_id, source, state)
	}
	return err
}

func (oh *orderHandler) SetState(order_id int64, source string, new_state int, order_data *OrderData) error {
	if !oh.parent.Check() {
		log.Printf("DB: can not set state for [%v] now... Will do it after.", order_id)
		u.After(oh.parent.Check, func() {
			oh.SetState(order_id, source, new_state, order_data)
		})
		return nil
	}
	to_set := bson.M{"order_state": new_state, "when": time.Now()}
	if order_data != nil {
		to_set["data"] = order_data
	}

	change := bson.M{"$set": to_set}
	log.Println("DB: change:", change["$set"])
	err := oh.Collection.Update(bson.M{"order_id": order_id, "source":source}, change)
	if err != nil && err != mgo.ErrNotFound {
		log.Printf("DB: state [%v] for order [%v] %v is not stated because %v", new_state, order_id, source, err)
		return err
	}
	if err == mgo.ErrNotFound {
		log.Printf("DB: for order %v at %v not found :(( ", order_id, source)
	}
	return err
}

func (oh *orderHandler) SetFeedback(for_whom string, for_state int, feedback string, source string) (*int64, error) {
	if !oh.parent.Check() {
		u.After(oh.parent.Check, func() {
			oh.SetFeedback(for_whom, for_state, feedback, source)
		})
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

func (oh *orderHandler) AddOrderObject(order OrderWrapper) error {
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
	} else if err != nil {
		return nil, err
	}
	return &result, nil
}

func (oh *orderHandler) GetByOwner(whom, source string, active bool) (*OrderWrapper, error) {
	if !oh.parent.Check() {
		return nil, errors.New("БД не доступна")
	}
	result := OrderWrapper{}
	err := oh.Collection.Find(bson.M{"whom": whom, "source":source, "is_active":true}).Sort("-when").One(&result)
	if err == mgo.ErrNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &result, nil
}

func (oh *orderHandler) GetOrders(q bson.M) ([]OrderWrapper, error) {
	if !oh.parent.Check() {
		return nil, errors.New("БД не доступна")
	}
	var result []OrderWrapper
	err := oh.Collection.Find(q).Sort("-when").All(&result)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	}
	return result, nil
}

func (oh *orderHandler) GetOrdersSort(q bson.M, sort string) ([]OrderWrapper, error) {
	if !oh.parent.Check() {
		return nil, errors.New("БД не доступна")
	}
	var result []OrderWrapper
	err := oh.Collection.Find(q).Sort(sort).All(&result)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	}
	return result, nil
}

func (uh *UserHandler) GetUser(req bson.M) (*UserData, error) {
	if !uh.parent.Check() {
		return nil, errors.New("БД не доступна")
	}
	tmp := UserData{}
	err := uh.UsersCollection.Find(req).One(&tmp)
	if err == mgo.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, errors.New(fmt.Sprintf("user for %+v is not found", req))
	}
	return &tmp, nil
}

func (uh *UserHandler) AddUser(user_id, name, phone, email string) error {
	if !uh.parent.Check() {
		return errors.New("БД не доступна")
	}
	tmp, err := uh.GetUser(bson.M{"user_id": user_id, "phone": phone})
	if tmp == nil {
		err = uh.UsersCollection.Insert(&UserData{UserId: user_id, UserName:name, Email:email, Phone: phone, LastUpdate: time.Now()})
		return err
	}
	return errors.New(fmt.Sprintf("Duplicate user! [%v] %v {%v}", user_id, name, phone))
}

func (uh *UserHandler) StoreUser(user_id, name, phone, email string) error {
	if !uh.parent.Check() {
		return errors.New("БД не доступна")
	}
	if u, _ := uh.GetUserById(user_id); u == nil {
		return uh.UsersCollection.Insert(&UserData{UserId: user_id, UserName:name, Email:email, Phone: phone, LastUpdate: time.Now()})
	} else {
		return uh.UpdateUserData(user_id, name, phone, email)
	}
}

func (uh *UserHandler) StoreUserData(user_id string, user_data *s.InUserData) error {
	if !uh.parent.Check() {
		return errors.New("БД не доступна")
	}
	if u, _ := uh.GetUserById(user_id); u == nil {
		return uh.UsersCollection.Insert(&UserData{UserId: user_id, UserName:user_data.Name, Email:user_data.Email, Phone: user_data.Phone, LastUpdate: time.Now()})
	} else {
		return uh.UpdateUserData(user_id, user_data.Name, user_data.Phone, user_data.Email)
	}
}

func (uh UserHandler) AddOrUpdateUserObject(uw UserData) error {
	if !uh.parent.Check() {
		return errors.New("БД не доступна")
	}
	obj := UserData{}
	err := uh.UsersCollection.Find(bson.M{"user_id":uw.UserId}).One(&obj)
	if err == mgo.ErrNotFound {
		err = uh.UsersCollection.Insert(uw)
		return err
	} else {
		err = uh.UsersCollection.UpdateId(obj.ID, uw)
		return err
	}
	return err
}

func (uh *UserHandler) SetUserState(user_id, state_key, state_value string) error {
	/**
	Выставление сосотяние по определенному аспекту. к примеру для квестов. Или для еще какой хуйни, посему требуется ключ да значение.
	Отличается от просто SetUserState тем что там выставляется состояние глобальное
	 */
	if !uh.parent.Check() {
		return errors.New("БД не доступна")
	}
	tmp, _ := uh.GetUser(bson.M{"user_id": user_id})
	if tmp == nil {
		err := uh.UsersCollection.Insert(&UserData{UserId: user_id, States: map[string]string{state_key:state_value}, LastUpdate: time.Now()})
		return err
	} else {
		err := uh.UsersCollection.Update(
			bson.M{"user_id": user_id},
			bson.M{"$set": bson.M{fmt.Sprintf("states.%v", state_key): state_value, "last_update": time.Now()}},
		)
		return err
	}
}

func (uh *UserHandler) GetUserMultiplyState(user_id, state_key string) (string, error) {
	if !uh.parent.Check() {
		return "", errors.New("БД не доступна")
	}
	tmp, _ := uh.GetUser(bson.M{"user_id": user_id})
	if tmp == nil {
		return "", errors.New("Пользователь не найден")
	} else {
		if state, ok := tmp.States[state_key]; ok {
			return state, nil
		}
		return "", errors.New("This user have not this key of state")
	}
}

func (uh *UserHandler) SetUserPassword(username, password string) error {
	if !uh.parent.Check() {
		return errors.New("БД не доступна")
	}
	tmp, _ := uh.GetUser(bson.M{"user_name": username})
	if tmp == nil {
		err := uh.UsersCollection.Insert(&UserData{UserId: username, UserName: username, Password: u.PHash(password), LastUpdate: time.Now()})
		return err
	} else if u.PHash(password) != tmp.Password {
		log.Println("changing password! for user ", username)
		err := uh.UsersCollection.Update(
			bson.M{"user_name": username},
			bson.M{"$set": bson.M{"password": u.PHash(password), "last_update": time.Now()}},
		)
		return err
	}
	return nil
}

func (uh *UserHandler) CheckUserPassword(username, password string) (bool, error) {
	if !uh.parent.Check() {
		return false, errors.New("БД не доступна")
	}
	tmp := UserData{}
	err := uh.UsersCollection.Find(bson.M{"user_name": username, "password": u.PHash(password)}).One(&tmp)
	return err != nil, err
}

func (uh *UserHandler) GetUserById(user_id string) (*UserData, error) {
	if !uh.parent.Check() {
		return nil, errors.New("БД не доступна")
	}
	result := UserData{}
	err := uh.UsersCollection.Find(bson.M{"user_id": user_id}).One(&result)
	if err != nil && err != mgo.ErrNotFound {
		log.Printf("Ощибка определения пользователя %v", err)
		return nil, err
	} else if err == mgo.ErrNotFound {
		return nil, nil
	}
	return &result, err
}

func (uh *UserHandler) SetUserShowedName(user_id, new_name string) error {
	if !uh.parent.Check() {
		return errors.New("БД не доступна")
	}
	user, _ := uh.GetUserById(user_id)
	if user != nil {
		err := uh.UsersCollection.Update(bson.M{"user_id":user_id}, bson.M{"$set":bson.M{"showed_name":new_name}})
		return err
	}
	return errors.New("User not found :(")
}

func (uh *UserHandler) UpdateUserData(user_id, name, phone, email string) error {
	if !uh.parent.Check() {
		return errors.New("БД не доступна")
	}
	to_upd := bson.M{}
	if name != "" {
		to_upd["user_name"] = name
	}
	if phone != "" {
		to_upd["phone"] = phone
	}
	if email != "" {
		to_upd["email"] = email
	}

	err := uh.UsersCollection.Update(bson.M{"user_id":user_id}, bson.M{"$set":to_upd})
	return err
}
func (uh *UserHandler) Count() int {
	r, _ := uh.UsersCollection.Count()
	return r
}

func (uh *UserHandler) GetBy(req bson.M) ([]UserData, error) {
	if !uh.parent.Check() {
		return nil, errors.New("БД не доступна")
	}
	result := []UserData{}
	err := uh.UsersCollection.Find(req).Sort("last_update").All(&result)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	} else if err == mgo.ErrNotFound {
		return result, nil
	}
	return result, nil
}

//ERRORS
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

type CommandsStorage struct {
	DbHelper
	Commands *mgo.Collection
}

func (cs *CommandsStorage) ensureIndexes() {
	commands_collection := cs.Session.DB(cs.DbName).C("commands")
	commands_collection.EnsureIndex(mgo.Index{
		Key:        []string{"name", "provider"},
		Background: true,
		DropDups:   true,
		Unique:    true,
	})

	cs.Commands = commands_collection
}

func NewCommandsStorage(conn, dbname string) *CommandsStorage {
	helper := NewDbHelper(conn, dbname)
	res := &CommandsStorage{DbHelper:*helper}
	res.ensureIndexes()
	return res
}

type CommandsStore interface {
	SaveCommand(provider, name string, command s.OutCommand) error
	LoadCommands(provider, name string) ([]s.OutCommand, error)
	GetCommand(req bson.M) (*s.OutCommand, error)
}

type CommandsWrapper struct {
	Name     string `bson:"name"`
	Provider string `bson:"provider"`
	Commands []s.OutCommand `bson:"commands"`
}

func (cs *CommandsStorage) SaveCommand(provider, name string, command s.OutCommand) error {
	res := &CommandsWrapper{}
	err := cs.Commands.Find(bson.M{"name":name, "provider":provider}).One(res)
	if res.Name != "" && res.Provider != "" && len(res.Commands) > 0 {
		err = cs.Commands.Update(bson.M{"name":name, "provider":provider}, bson.M{"$addToSet":bson.M{"commands":command}})
	} else {
		err = cs.Commands.Insert(CommandsWrapper{Name:name, Provider:provider, Commands:[]s.OutCommand{command}})
		if err != nil {
			log.Printf("CS Error at saving command %v", err)
		}
	}
	return err
}

func (cs *CommandsStorage) LoadCommands(provider, name string) ([]s.OutCommand, error) {
	res := CommandsWrapper{}
	err := cs.Commands.Find(bson.M{"name":name, "provider":provider}).Sort("commands.position").One(&res)
	if err != nil {
		log.Printf("CS Error at find commands by name: %v provider: %v %v", name, provider, err)
	}
	return res.Commands, err
}

func (cs *CommandsStorage) GetCommand(req bson.M) (*s.OutCommand, error) {
	res := &s.OutCommand{}
	err := cs.Commands.Find(req).One(res)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	} else if err == mgo.ErrNotFound {
		return nil, nil
	}
	return res, nil
}

type ConfigurationStorage struct {
	DbHelper
	Collection *mgo.Collection
}

func NewConfigurationStorage(connection c.MongoDbConfig) *ConfigurationStorage {
	helper := NewDbHelper(connection.ConnString, connection.Name)
	result := &ConfigurationStorage{DbHelper:*helper}
	result.ensureIndexes()
	return result
}

func (cs *ConfigurationStorage) ensureIndexes() {
	collection := cs.Session.DB(cs.DbName).C("bots_configs")
	collection.EnsureIndex(mgo.Index{
		Key:[]string{"company_id"},
		Unique:true,
	})
	collection.EnsureIndex(mgo.Index{
		Key:[]string{"notifications"},
	})
	collection.EnsureIndex(mgo.Index{
		Key:[]string{"auto_answers"},
	})

	cs.Collection = collection
}

func (cs *ConfigurationStorage) SetChatConfig(config c.ChatConfig, update bool) error {
	found := c.ChatConfig{}
	err := cs.Collection.Find(bson.M{"company_id":config.CompanyId}).One(&found)
	if err != nil && err != mgo.ErrNotFound {
		log.Printf("CFGS ERROR set chat config at find %v", err)
		return err
	}
	if err == mgo.ErrNotFound {
		log.Printf("SETTING CHAT CONFIG %+v", config)
		cs.Collection.Insert(config)
		return nil
	}
	if update && err == nil {
		modify := bson.M{}
		if reflect.DeepEqual(found.Notifications, config.Notifications) {
			modify["notifications"] = config.Notifications
		}
		if reflect.DeepEqual(found.AutoAnswers, config.AutoAnswers) {
			modify["auto_answers"] = config.AutoAnswers
		}
		update, _ := u.GetStringUpdates(found, config, "bson", true)
		for k, v := range update {
			modify[k] = v
		}
		cs.Collection.Update(bson.M{"company_id":config.CompanyId}, bson.M{"$set":modify})
	}
	return nil
}

func (cs *ConfigurationStorage) UpdateNotifications(companyId string, notifications []c.TimedAnswer) error {
	ci, err := cs.Collection.Upsert(bson.M{"company_id":companyId}, bson.M{"$set":bson.M{"notifications":notifications}})
	log.Printf("Update notifications %+v", ci)
	return err
}

func (cs *ConfigurationStorage) UpdateAutoAnswers(companyId string, autoAnswers []c.TimedAnswer) error {
	ci, err := cs.Collection.Upsert(bson.M{"company_id":companyId}, bson.M{"$set":bson.M{"auto_answers":autoAnswers}})
	log.Printf("Update auto answers %+v", ci)
	return err
}

func (cs *ConfigurationStorage) UpdateInformation(companyId, information string) error {
	ci, err := cs.Collection.Upsert(bson.M{"company_id":companyId}, bson.M{"$set":bson.M{"information":information}})
	log.Printf("Update information for: %v, new: %v, result: %+v", companyId, information, ci)
	return err
}

func (cs *ConfigurationStorage) GetChatConfig(companyId string) (*c.ChatConfig, error) {
	found := c.ChatConfig{}
	err := cs.Collection.Find(bson.M{"company_id":companyId}).One(&found)
	if err != nil && err != mgo.ErrNotFound {
		log.Printf("CFGS ERROR set chat config at find %v", err)
		return nil, err
	} else if err == mgo.ErrNotFound {
		return nil, nil
	}
	return &found, nil
}
func (cs *ConfigurationStorage) GetAllChatsConfig() ([]c.ChatConfig, error) {
	found := []c.ChatConfig{}
	err := cs.Collection.Find(bson.M{}).All(&found)
	if err != nil && err != mgo.ErrNotFound {
		log.Printf("CFGS ERROR set chat config at find %v", err)
		return nil, err
	} else if err == mgo.ErrNotFound {
		return nil, nil
	}
	return found, nil
}
func (cs *ConfigurationStorage) GetInformation(companyId string) (*string, error) {
	found := c.ChatConfig{}
	err := cs.Collection.Find(bson.M{"company_id":companyId}).One(&found)
	if err != nil && err != mgo.ErrNotFound {
		log.Printf("CFGS ERROR set chat config at find %v", err)
		return nil, err
	} else if err == mgo.ErrNotFound {
		return nil, nil
	}
	information := found.Information
	return &information, nil
}