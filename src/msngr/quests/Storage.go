package quests

import (
	"msngr/db"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"errors"
	"time"
)

type QuestStorage struct {
	db.DbHelper
	Keys *mgo.Collection
	Messages *mgo.Collection
}

func (qks *QuestStorage) ensureIndexes() {
	collection := qks.Session.DB(qks.DbName).C("quest_keys")
	collection.EnsureIndex(mgo.Index{
		Key:        []string{"key"},
		Background: true,
		DropDups:   true,
		Unique:    true,
	})
	qks.Keys = collection

	message_collection := qks.Session.DB(qks.DbName).C("quest_messages")
	message_collection.EnsureIndex(mgo.Index{
		Key:[]string{"from"},
		Unique:false,
	})
	message_collection.EnsureIndex(mgo.Index{
		Key:[]string{"answered"},
		Unique:false,
	})
	message_collection.EnsureIndex(mgo.Index{
		Key:[]string{"time"},
		Unique:false,
	})
	qks.Messages = message_collection
}

func NewQuestKeyStorage(conn, dbname string) *QuestStorage {
	helper := db.NewDbHelper(conn, dbname)
	result := QuestStorage{DbHelper:*helper}
	result.ensureIndexes()
	return &result
}

type KeyWrapper struct {
	Key         string `bson:"key"`
	Description string `bson:"description"`
}

func (qks *QuestStorage) AddKey(key, description string) error {
	kw := KeyWrapper{}
	err := qks.Keys.Find(bson.M{"key":key}).One(&kw)
	if err != nil && err != mgo.ErrNotFound{
		return err
	} else if err == mgo.ErrNotFound{
		err = qks.Keys.Insert(KeyWrapper{Key:key, Description:description})
		return err
	}
	return errors.New("This key already exists")
}

func (qks *QuestStorage) GetAllKeys() ([]KeyWrapper, error){
	result := []KeyWrapper{}
	err := qks.Keys.Find(bson.M{}).All(&result)
	return result, err
}

func (qks *QuestStorage) GetDescription(key string) (string, error){
	result := KeyWrapper{}
	err := qks.Keys.Find(bson.M{"key":key}).One(&result)
	return result.Description, err
}

func (qks *QuestStorage) StoreMessage(from, body string, time time.Time) error{
	result := db.MessageWrapper{From:from, Body:body, Time:time, Answered:false}
	err := qks.Messages.Insert(&result)
	return err
}

func (qks *QuestStorage) GetMessages(query bson.M) ([]db.MessageWrapper, error){
	result := []db.MessageWrapper{}
	err := qks.Messages.Find(query).Sort("time").All(&result)
	return result, err
}