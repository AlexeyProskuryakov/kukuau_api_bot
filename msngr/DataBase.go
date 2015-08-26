package msngr

import (
	"sync"
)

type userHandler struct {
	state_map map[string]int
}

var instance *userHandler
var once sync.Once

func GetUserHandler() *userHandler {
	once.Do(func() {
		instance = &userHandler{state_map: make(map[string]int)}
	})
	return instance
}

const (
	ORDER_CREATE   = 1
	ORDER_CANCELED = 0
	USER_NOT_KNOWN = -1
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
