package console

import (
	"time"
)

var users = map[string]string{
	"alesha":"sederfes100500",
	"leha":"qwerty100500",
}


type IndexInfo struct {
	UserName    string
	CurrentTime time.Time
	UpTime      time.Duration
}

type ConsoleInfo struct {
	TaxiOrdersCount int
	TaxiUserOrders  map[string]string

	ShopUsersCount  int
	ShopLoginUsers  []string

}

