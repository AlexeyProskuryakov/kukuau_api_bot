package msngr

import (
	"sync"
)

type userHandler struct {
	state_map     map[string]int
	passwords_map map[string]string
}

var instance *userHandler
var once sync.Once

func GetUserHandler() *userHandler {
	once.Do(func() {
		instance = &userHandler{
			state_map:     make(map[string]int),
			passwords_map: make(map[string]string),
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

func (uh userHandler) DeleteUserState(username string) {
	delete(uh.state_map, username)
}

func (uh userHandler) SetUserPassword(username string, password string) {
	uh.passwords_map[username] = password
}

func (uh userHandler) CheckUserPassword(username string, password string) bool {
	p, ok := uh.passwords_map[username]
	if ok && p == password {
		return true
	}
	return false
}
