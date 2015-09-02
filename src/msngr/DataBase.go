package msngr

import (
	"errors"
	"sync"
)

type userHandler struct {
	state_map     map[string]int
	passwords_map map[string]string
	orders_map    map[string]int64
}

var instance *userHandler
var once sync.Once

func GetUserHandler() *userHandler {
	once.Do(func() {
		instance = &userHandler{
			state_map:     make(map[string]int),
			passwords_map: make(map[string]string),
			orders_map:    make(map[string]int64),
		}

		instance.SetUserPassword("test", "123")

	})
	return instance
}

const (
	USER_NOT_KNOWN = -1

	ORDER_CREATE   = 1
	ORDER_CANCELED = 0

	USER_AUTHORISED = 2
)

///STATE
func (uh userHandler) GetUserState(username string) int {
	val, ok := uh.state_map[username]
	if ok {
		return val
	}
	return USER_NOT_KNOWN
}
func (uh userHandler) SetUserState(username string, state int) {
	uh.state_map[username] = state
}
func (uh userHandler) RemoveUserState(username string) {
	delete(uh.state_map, username)
}

///ORDER ID
func (uh userHandler) SetUserOrderId(username string, order_id int64) {
	uh.orders_map[username] = order_id
	uh.state_map[username] = ORDER_CREATE
}
func (uh userHandler) GetUserOrderId(username string) (oid int64, e error) {
	oid, ok := uh.orders_map[username]
	if ok {
		return oid, nil
	}
	return -1, errors.New("order not exists!")
}
func (uh userHandler) CancelOrderId(username string) {
	delete(uh.orders_map, username)
	uh.state_map[username] = ORDER_CANCELED
}

///PASSWORDS
func (uh userHandler) SetUserPassword(username string, password string) {
	uh.passwords_map[username] = password
	uh.state_map[username] = USER_AUTHORISED
}

func (uh userHandler) CheckUserPassword(username string, password string) bool {
	p, ok := uh.passwords_map[username]
	if ok && p == password {
		return true
	}
	return false
}
