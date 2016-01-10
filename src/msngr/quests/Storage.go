package quests

import (
	"msngr/db"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"errors"
	"time"
	"fmt"
)

type QuestStorage struct {
	db.DbHelper
	Keys     *mgo.Collection
	Messages *mgo.Collection
	Users    *mgo.Collection
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

	users_collection := qks.Session.DB(qks.DbName).C("quest_users")
	users_collection.EnsureIndex(mgo.Index{
		Key:[]string{"time"},
		Unique:false,
	})
	qks.Users = users_collection
}

func NewQuestStorage(conn, dbname string) *QuestStorage {
	helper := db.NewDbHelper(conn, dbname)
	result := QuestStorage{DbHelper:*helper}
	result.ensureIndexes()
	return &result
}

type KeyWrapper struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	SID         string
	Key         string `bson:"key"`
	Description string `bson:"description"`
	Position    int64  `bson:"position"`
}

func (qks *QuestStorage) AddKey(key, description string, position int64) error {
	kw := KeyWrapper{}
	err := qks.Keys.Find(bson.M{"key":key}).One(&kw)
	if err != nil && err != mgo.ErrNotFound {
		return err
	} else if err == mgo.ErrNotFound {
		err = qks.Keys.Insert(KeyWrapper{Key:key, Description:description, Position:position})
		return err
	}
	return errors.New("This key already exists")
}

func (qks *QuestStorage) GetAllKeys() ([]KeyWrapper, error) {
	result := []KeyWrapper{}
	err := qks.Keys.Find(bson.M{}).All(&result)
	for i, key := range result{
		result[i].SID = key.ID.Hex()
	}
	return result, err
}

func (qks *QuestStorage) GetKeyInfo(key string) (*KeyWrapper, error) {
	kw := KeyWrapper{}
	err := qks.Keys.Find(bson.M{"key":key}).One(&kw)
	if err != nil {
		return nil, err
	}
	return &kw, nil
}

func (qks *QuestStorage) DeleteKey(key_id string) error {
	err := qks.Keys.RemoveId(bson.ObjectIdHex(key_id))
	return err
}

func (qks *QuestStorage) GetDescription(key string) (string, error) {
	result := KeyWrapper{}
	err := qks.Keys.Find(bson.M{"key":key}).One(&result)
	return result.Description, err
}

func (qks *QuestStorage) StoreMessage(from, body string, time time.Time) error {
	result := db.MessageWrapper{From:from, Body:body, Time:time, Answered:false}
	err := qks.Messages.Insert(&result)
	return err
}

func (qks *QuestStorage) SetMessageAnswer(message_id bson.ObjectId) error {
	err := qks.Messages.UpdateId(message_id, bson.M{"$set":bson.M{"answered":true}})
	return err
}

func (qs *QuestStorage) GetMessage(message_id string) (*db.MessageWrapper, error) {
	result := db.MessageWrapper{}
	err := qs.Messages.FindId(bson.ObjectIdHex(message_id)).One(&result)
	return &result, err
}

func (qks *QuestStorage) GetMessages(query bson.M) ([]db.MessageWrapper, error) {
	result := []db.MessageWrapper{}
	err := qks.Messages.Find(query).Sort("-time").All(&result)
	for i, message := range result {
		result[i].SID = message.ID.Hex()
	}
	return result, err
}

type QuestUserWrapper struct {
	UserId           string `bson:"user_id"`
	State            map[string]string `bson:"state"`
	Keys             map[string][]string `bson:"found_keys"`
	LastKeyPositions map[string]*int64 `bson:"last_key_positions"`
}

func (qks *QuestStorage) SetUserState(user_id, state, provider string) error {
	find := bson.M{"user_id":user_id}
	user := QuestUserWrapper{}
	err := qks.Users.Find(find).One(&user)
	if err == mgo.ErrNotFound {
		qks.Users.Insert(QuestUserWrapper{UserId:user_id, State:map[string]string{provider:state}})
	} else if err != nil {
		return err
	} else {
		qks.Users.Update(find, bson.M{"$set":bson.M{fmt.Sprintf("state.%s", provider):state}})
	}
	return nil
}

func (qks *QuestStorage) GetUserState(user_id, provider string) (string, error) {
	find := bson.M{"user_id":user_id}
	user := QuestUserWrapper{}
	err := qks.Users.Find(find).One(&user)
	if err != nil {
		return "", err
	}
	if state, ok := user.State[provider]; ok {
		return state, nil
	} else {
		return "", nil
	}
}

func (qks *QuestStorage) SetUserLastKey(user_id, key, provider string) error {
	find := bson.M{"user_id":user_id}
	user := QuestUserWrapper{}
	err := qks.Users.Find(find).One(&user)
	if err != nil {
		return err
	}
	key_info, err := qks.GetKeyInfo(key)
	if err != nil {
		return errors.New(fmt.Sprintf("Key added is errored: %v", err.Error()))
	}
	err = qks.Users.Update(find, bson.M{
		"$addToSet":bson.M{fmt.Sprintf("found_keys.%s", provider):key},
		"$set":bson.M{fmt.Sprintf("last_key_positions.%s", provider):key_info.Position},
	})
	return err
}

type CurrentProviderUserInfo struct {
	UserId          string
	State           string
	FoundKeys       []string
	LastKeyPosition *int64
}

func (qks *QuestStorage) GetUserInfo(user_id, provider string) (*CurrentProviderUserInfo, error) {
	find := bson.M{"user_id":user_id}
	user := QuestUserWrapper{}
	err := qks.Users.Find(find).One(&user)
	if err != nil {
		return nil, err
	}
	state, _ := user.State[provider]
	keys, _ := user.Keys[provider]
	position, _ := user.LastKeyPositions[provider]


	return &CurrentProviderUserInfo{
		UserId:user.UserId,
		State:state,
		FoundKeys:keys,
		LastKeyPosition:position,
	}, nil

}

func (qks *QuestStorage) GetUserKeys(user_id, key, provider string) ([]string, error) {
	find := bson.M{"user_id":user_id}
	user := QuestUserWrapper{}
	err := qks.Users.Find(find).One(&user)
	if err != nil {
		return []string{}, err
	} else {
		if keys, ok := user.Keys[provider]; ok {
			return keys, nil
		}else {
			return []string{}, nil
		}
	}
}


func (qks *QuestStorage) GetSubscribedUsers() ([]QuestUserWrapper, error) {
	users := []QuestUserWrapper{}
	err := qks.Users.Find(bson.M{fmt.Sprintf("state.%s", PROVIDER):SUBSCRIBED}).All(&users)
	return users, err
}
