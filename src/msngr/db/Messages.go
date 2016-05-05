package db

import (
	"gopkg.in/mgo.v2/bson"
	"time"
	"gopkg.in/mgo.v2"
	"errors"
	"fmt"
	"log"
	"msngr/utils"
)

type AdditionalDataElement struct {
	Key   string
	Value string
	Name  string
}

type AdditionalFuncElement struct {
	MessageId string `bson:"message_id"`
	Name      string `bson:"name"`
	Action    string `bson:"action"`
	Context   map[string]interface{} `bson:"context"`
	Used      bool `bson:"used"`
}

type NotificationElement struct {
	After int `bson:"after"`
}

type MessageWrapper struct {
	ID                bson.ObjectId `bson:"_id,omitempty"`
	SID               string
	From              string `bson:"from"`
	FromName          string `bson:"from_name,omitempty"`
	Body              string `bson:"body"`
	To                string `bson:"to"`
	ToName            string `bson:"to_name,omitempty"`
	Time              time.Time `bson:"time"`
	TimeStamp         int64 `bson:"time_stamp"`
	TimeFormatted     string `bson:",omitempty" json:"time"`
	NotAnswered       int `bson:"not_answered"`
	AnsweredBy        string `bson:"answered_by"`
	Unread            int `bson:"unread"`
	MessageID         string `bson:"message_id"`
	MessageStatus     string `bson:"message_status"`
	MessageCondition  string `bson:"message_condition"`
	IsDeleted         bool `bson:"is_deleted"`
	Attributes        []string `bson:"attributes,omitempty"`
	AdditionalData    []AdditionalDataElement `bson:"additional_data,omitempty"`
	AdditionalFuncs   []AdditionalFuncElement  `bson:"additional_funcs,omitempty"`
	RelatedOrder      int64 `bson:"related_order,omitempty"`
	RelatedOrderState string `bson:"related_order_state,omitempty"`
	IsNotification    bool `bson:"is_notification,omitempty"`
	AutoAnswers       []NotificationElement `bson:"auto_answers,omitempty"`
	Notifications     []NotificationElement `bson:"notifications,omitempty"`
}

func (mw MessageWrapper) IsAttrPresent(attrName string) bool {
	return utils.InS(attrName, mw.Attributes)
}

func NewMessageForWeb(sid, from, to, body string) *MessageWrapper {
	result := MessageWrapper{From:from, To:to, Body:body, TimeFormatted:time.Now().Format(time.Stamp), SID:sid}
	return &result
}

type MessageHandler struct {
	MessagesCollection  *mgo.Collection
	FunctionsCollection *mgo.Collection
	parent              *MainDb
}

func (mh *MessageHandler) ensureIndexes() {
	messageCollection := mh.parent.Session.DB(mh.parent.DbName).C("user_messages")
	messageCollection.EnsureIndex(mgo.Index{
		Key:[]string{"from"},
		Unique:false,
	})
	messageCollection.EnsureIndex(mgo.Index{
		Key:[]string{"to"},
		Unique:false,
	})
	messageCollection.EnsureIndex(mgo.Index{
		Key:[]string{"not_answered"},
		Unique:false,
	})
	messageCollection.EnsureIndex(mgo.Index{
		Key:[]string{"unread"},
		Unique:false,
	})
	messageCollection.EnsureIndex(mgo.Index{
		Key:[]string{"time_stamp"},
		Unique:false,
	})
	messageCollection.EnsureIndex(mgo.Index{
		Key:[]string{"message_id"},
		Unique:true,
	})
	messageCollection.EnsureIndex(mgo.Index{
		Key:[]string{"related_order"},
	})
	messageCollection.EnsureIndex(mgo.Index{
		Key:[]string{"is_deleted"},
	})
	mh.MessagesCollection = messageCollection

	functions := mh.parent.Session.DB(mh.parent.DbName).C("user_messages_functions")
	functions.EnsureIndex(mgo.Index{
		Key:[]string{"message_id", "action"},
		Unique:true,
	})
	functions.EnsureIndex(mgo.Index{
		Key:[]string{"message_id"},
	})
	mh.FunctionsCollection = functions
}
func (mh *MessageHandler) StoreNotificationMessage(from, to, body, message_id string) (*MessageWrapper, error) {
	if !mh.parent.Check() {
		return nil, errors.New("БД не доступна")
	}
	found, err := mh.GetMessageByMessageId(message_id)
	if found == nil && err == nil {
		result := MessageWrapper{
			From:from,
			To:to,
			Body:body,
			TimeStamp:time.Now().Unix(),
			Time:time.Now(),
			NotAnswered:1,
			Unread:1,
			MessageID:message_id,
			IsDeleted:false,
			TimeFormatted: time.Now().Format(time.Stamp),
			IsNotification:true,
		}
		err := mh.MessagesCollection.Insert(&result)
		return &result, err
	}
	return nil, errors.New(fmt.Sprintf("I have duplicate!%+v", found))
}
func (mh *MessageHandler) StoreMessage(from, to, body, message_id string) (*MessageWrapper, error) {
	if !mh.parent.Check() {
		return nil, errors.New("БД не доступна")
	}
	found, err := mh.GetMessageByMessageId(message_id)
	if found == nil && err == nil {
		result := MessageWrapper{
			From:from,
			To:to,
			Body:body,
			TimeStamp:time.Now().Unix(),
			Time:time.Now(),
			NotAnswered:1,
			Unread:1,
			MessageID:message_id,
			IsDeleted:false,
			TimeFormatted: time.Now().Format(time.Stamp),
		}
		err := mh.MessagesCollection.Insert(&result)
		return &result, err
	}
	return nil, errors.New(fmt.Sprintf("I have duplicate!%+v", found))
}
func (mh *MessageHandler) StoreMessageObject(message MessageWrapper) (error) {
	if !mh.parent.Check() {
		return errors.New("БД не доступна")
	}
	found, err := mh.GetMessageByMessageId(message.MessageID)
	if found == nil && err == nil {
		if len(message.AdditionalFuncs) > 0 {
			for _, function := range message.AdditionalFuncs {
				err = mh.InsertMessageFunction(function)
				if err != nil {
					log.Printf("DB ERROR at storing message additional function %v", err)
				}
			}
			message.AdditionalFuncs = []AdditionalFuncElement{}
		}
		err := mh.MessagesCollection.Insert(&message)
		if err != nil {
			log.Printf("DB ERROR at storing message %+v", err)
		}
		return err
	}

	return errors.New("Already exists")
}

func (mh *MessageHandler) InsertMessageFunction(function AdditionalFuncElement) error {
	err := mh.FunctionsCollection.Insert(function)
	return err
}

func (mh *MessageHandler) SetMessageFunctionUsed(messageId, action string) error {
	err := mh.FunctionsCollection.Update(bson.M{"message_id":messageId, "action":action}, bson.M{"$set":bson.M{"used":true}})
	return err
}

func (mh *MessageHandler) GetMessageFunctions(messageId string) ([]AdditionalFuncElement, error) {
	result := []AdditionalFuncElement{}
	err := mh.FunctionsCollection.Find(bson.M{"message_id":messageId}).All(&result)
	if err != mgo.ErrNotFound && err != nil {
		return result, err
	}
	return result, nil
}

func (mh *MessageHandler) SetMessagesAnswered(from, to, by string) error {
	if !mh.parent.Check() {
		return errors.New("БД не доступна")
	}
	_, err := mh.MessagesCollection.UpdateAll(
		bson.M{"from":from, "to":to, "not_answered":1},
		bson.M{"$set":bson.M{"not_answered":0, "answered_by":by}},
	)
	return err
}

func (mh *MessageHandler) GetMessagesForAutoAnswer(to string, after int) ([]MessageWrapper, error) {
	result := []MessageWrapper{}
	if !mh.parent.Check() {
		return result, errors.New("БД не доступна")
	}
	timeStampLess := time.Now().Add(-(time.Duration(after) * time.Minute)).Unix()
	err := mh.MessagesCollection.Find(bson.M{
		"to":to,
		"not_answered":1,
		"time_stamp":bson.M{"$lte": timeStampLess},
		"auto_answers":bson.M{"$not":bson.M{"$elemMatch":bson.M{"after":after}}},
	}).All(&result)
	return result, err
}
func (mh *MessageHandler) SetMessagesAutoAnswered(from, to string, after int) error {
	if !mh.parent.Check() {
		return errors.New("БД не доступна")
	}
	_, err := mh.MessagesCollection.UpdateAll(
		bson.M{"from":from, "to":to, "not_answered":1},
		bson.M{"$push":bson.M{"auto_answers":NotificationElement{After:after}}},
	)
	return err
}

func (mh *MessageHandler) GetMessagesForNotification(to string, after int) ([]MessageWrapper, error) {
	result := []MessageWrapper{}
	if !mh.parent.Check() {
		return result, errors.New("БД не доступна")
	}
	timeStampLess := time.Now().Add(-(time.Duration(after) * time.Minute)).Unix()
	err := mh.MessagesCollection.Find(bson.M{
		"to":to,
		"unread":1,
		"time_stamp":bson.M{"$lte": timeStampLess},
		"notifications":bson.M{"$not":bson.M{"$elemMatch":bson.M{"after":after}}},
	}).All(&result)
	return result, err
}
func (mh *MessageHandler) SetMessagesNotified(from, to string, after int) error {
	if !mh.parent.Check() {
		return errors.New("БД не доступна")
	}
	_, err := mh.MessagesCollection.UpdateAll(
		bson.M{"from":from, "to":to, "unread":1},
		bson.M{"$push":bson.M{"notifications":NotificationElement{After:after}}},
	)
	return err
}

func (mh *MessageHandler) SetMessagesRead(from string) error {
	if !mh.parent.Check() {
		return errors.New("БД не доступна")
	}
	info, err := mh.MessagesCollection.UpdateAll(
		bson.M{"from":from, "unread":1},
		bson.M{"$set":bson.M{"unread":0}},
	)
	log.Printf("Result of messages read for messages from %v is: %+v", from, info)
	return err
}

func (mh *MessageHandler) SetAllMessagesRead(from, to string) error {
	if !mh.parent.Check() {
		return errors.New("БД не доступна")
	}
	info, err := mh.MessagesCollection.UpdateAll(
		bson.M{"$or":[]bson.M{bson.M{"from":from, "to":to}, bson.M{"to":to, "from":from}}, "unread":1},
		bson.M{"$set":bson.M{"unread":0}},
	)
	log.Printf("Result of messages read for messages from %v is: %+v", from, info)
	return err
}

func (mh *MessageHandler) FillMessage(message *MessageWrapper) *MessageWrapper {
	message.TimeFormatted = message.Time.Format(time.Stamp)
	message.SID = message.ID.Hex()
	functions, err := mh.GetMessageFunctions(message.MessageID)
	if err != nil {
		log.Printf("DB ERROR can not get message functions for %+v", message)
	} else {
		message.AdditionalFuncs = functions
	}
	return message
}

func (mh *MessageHandler) GetMessages(query bson.M) ([]MessageWrapper, error) {
	result := []MessageWrapper{}
	if !mh.parent.Check() {
		return result, errors.New("БД не доступна")
	}
	query["is_deleted"] = false
	err := mh.MessagesCollection.Find(query).Sort("time_stamp").All(&result)
	for i, message := range result {
		filledMessage := mh.FillMessage(&message)
		log.Printf("filled message: %+v", *filledMessage)
		result[i] = *filledMessage
	}
	return result, err
}

func (mh *MessageHandler) GetMessageByMessageId(message_id string) (*MessageWrapper, error) {
	result := MessageWrapper{}
	err := mh.MessagesCollection.Find(bson.M{"message_id":message_id, "is_deleted":false}).One(&result)
	if err == mgo.ErrNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return mh.FillMessage(&result), nil
}

func (mh *MessageHandler) GetMessageByRelatedOrder(relatedOrderId int64) (*MessageWrapper, error) {
	result := MessageWrapper{}
	err := mh.MessagesCollection.Find(bson.M{"related_order":relatedOrderId, "is_deleted":false}).One(&result)
	if err == mgo.ErrNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return mh.FillMessage(&result), nil
}

func (mh *MessageHandler) DeleteMessages(from, to string) (int, error) {
	info, err := mh.MessagesCollection.UpdateAll(
		bson.M{"$or":[]bson.M{
			bson.M{"from":from, "to":to},
			bson.M{"to":from, "from":to}},
			"is_deleted":false},
		bson.M{"$set":bson.M{"is_deleted":true}})
	return info.Updated, err
}
func (mh *MessageHandler) UpdateMessageStatus(message_id, status, condition string) error {
	return mh.MessagesCollection.Update(bson.M{"message_id":message_id}, bson.M{"$set":bson.M{"message_status":status, "message_condition":condition}})
}

func (mh *MessageHandler) UpdateMessageRelatedOrderState(messageId, newState string) error {
	return mh.MessagesCollection.Update(
		bson.M{"message_id":messageId},
		bson.M{"$set":bson.M{
			"related_order_state":newState,
		},
		})
}
