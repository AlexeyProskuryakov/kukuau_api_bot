package users

import (
	"fmt"
	"time"
)

const (
	MANAGER = "manager"

	DEFAULT_USER = "alesha"
	DEFAULT_PWD = "sederfes100500"
)

type ByContactsLastMessageTime []Contact

func (s ByContactsLastMessageTime) Len() int {
	return len(s)
}
func (s ByContactsLastMessageTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByContactsLastMessageTime) Less(i, j int) bool {
	return s[i].LastMessageTime > s[j].LastMessageTime
}

type Contact struct {
	ID               string `bson:"_id"`
	Name             string `bson:"name"`
	NewMessagesCount int `bson:"unread_count"`
	Phone            string
	LastMessageTime  int64 `bson:"time"`
}

func (c Contact) String() string {
	return fmt.Sprintf("[%v] name: %v, new messages: %v, phone: %s, last message: %v", c.ID, c.Name, c.NewMessagesCount, c.Phone, time.Unix(c.LastMessageTime, 0).Format(time.Stamp))
}