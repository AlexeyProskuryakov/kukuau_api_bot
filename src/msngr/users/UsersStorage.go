package users

type Contact struct {
	ID               string `bson:"_id"`
	Name             string `bson:"name"`
	NewMessagesCount int `bson:"not_answered_count"`
	Phone            string
	LastMessageTime  int64 `bson:"time"`
}
