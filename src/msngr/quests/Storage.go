package quests

import (
	"msngr/db"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"errors"
	"time"
	"fmt"
)

type Message struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	SID       string
	From      string `bson:"from"`
	To        string `bson:"to"`
	Body      string `bson:"body"`
	Time      time.Time `bson:"time"`
	TimeStamp int64 `bson:"time_stamp"`
	Answered  bool `bson:"is_answered"`
	IsKey     bool `bson:"is_key"`
}
type Key struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	SID         string
	Founded     bool `bson:"is_found"`
	FoundedBy   string `bson:"found_by"`
	StartKey    string `bson:"start_key"`
	NextKey     string `bson:"next_key"`
	Description string `bson:"description"`
}

type TeamMember struct {
	UserId   string `bson:"user_id"`
	Phone    string `bson:"phone"`
	Name     string `bson:"name"`
	TeamName string `bson:"team_name"`
}

type Team struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	SID       string
	Name      string `bson:"name"`
	FoundKeys []string `bson:"found_keys"`
}

type QuestStorage struct {
	db.DbHelper
	Keys        *mgo.Collection
	Messages    *mgo.Collection
	Teams       *mgo.Collection
	TeamMembers *mgo.Collection
}

func (qks *QuestStorage) ensureIndexes() {
	collection := qks.Session.DB(qks.DbName).C("quest_keys")
	collection.EnsureIndex(mgo.Index{
		Key:        []string{"start_key"},
		Unique:    true,
	})
	collection.EnsureIndex(mgo.Index{
		Key:        []string{"is_found"},
	})
	collection.EnsureIndex(mgo.Index{
		Key:        []string{"next_key"},
	})
	qks.Keys = collection

	message_collection := qks.Session.DB(qks.DbName).C("quest_messages")
	message_collection.EnsureIndex(mgo.Index{
		Key:[]string{"from"},
	})
	message_collection.EnsureIndex(mgo.Index{
		Key:[]string{"to"},
	})
	message_collection.EnsureIndex(mgo.Index{
		Key:[]string{"time"},
	})
	message_collection.EnsureIndex(mgo.Index{
		Key:[]string{"time_stamp"},
	})
	message_collection.EnsureIndex(mgo.Index{
		Key:[]string{"is_key"},
	})
	message_collection.EnsureIndex(mgo.Index{
		Key:[]string{"is_answered"},
	})
	qks.Messages = message_collection

	teams_collection := qks.Session.DB(qks.DbName).C("quest_teams")
	teams_collection.EnsureIndex(mgo.Index{
		Key:[]string{"name"},
		Unique:true,
	})
	qks.Teams = teams_collection

	members_collection := qks.Session.DB(qks.DbName).C("quest_team_members")
	members_collection.EnsureIndex(mgo.Index{
		Key:[]string{"name"},
	})
	members_collection.EnsureIndex(mgo.Index{
		Key:[]string{"user_id"},
		Unique:true,
	})
	members_collection.EnsureIndex(mgo.Index{
		Key:[]string{"team_name"},
	})
	qks.TeamMembers = members_collection
}

func NewQuestStorage(conn, dbname string) *QuestStorage {
	helper := db.NewDbHelper(conn, dbname)
	result := QuestStorage{DbHelper:*helper}
	result.ensureIndexes()
	return &result
}

//KEYS
func (qks *QuestStorage) AddKey(start_key, description, next_key string) error {
	kw := Key{}
	err := qks.Keys.Find(bson.M{"start_key":start_key}).One(&kw)
	if err == mgo.ErrNotFound {
		err = qks.Keys.Insert(Key{StartKey:start_key, Description:description, NextKey:next_key})
		return err
	} else if err != nil {
		return err
	} else {
		return errors.New(fmt.Sprintf("Key [%v] already exists", start_key))
	}
}

func (qks *QuestStorage) DeleteKey(key_id string) error {
	err := qks.Keys.RemoveId(bson.ObjectIdHex(key_id))
	return err
}

func (qks *QuestStorage) UpdateKey(key_id, start_key, description, next_key string) error {
	err := qks.Keys.UpdateId(bson.ObjectIdHex(key_id), bson.M{"$set":bson.M{"description":description, "start_key":start_key, "next_key":next_key}})
	return err
}

func (qs *QuestStorage) GetKeys(query bson.M) ([]Key, error) {
	result := []Key{}
	err := qs.Keys.Find(query).All(&result)
	if err != nil && err != mgo.ErrNotFound {
		return result, err
	}
	return result, nil
}

func (qs *QuestStorage) GetKey(key string) (*Key, error) {
	result := Key{}
	err := qs.Keys.Find(bson.M{"start_key":key}).One(&result)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	} else if err == mgo.ErrNotFound {
		return nil, nil
	}
	return &result, nil
}

func (qs *QuestStorage) SetKeyFounded(key, by string) error {
	err := qs.Keys.Update(bson.M{"start_key":key}, bson.M{"$set":bson.M{"found_by":by, "is_found":true}})
	if err != nil {
		return err
	}
	err = qs.Teams.Update(bson.M{"name":by}, bson.M{"$addToSet":bson.M{"found_keys": key}})
	return err
}

func (qs *QuestStorage) GetKeyByNextKey(next_key string) (*Key, error) {
	result := Key{}
	err := qs.Keys.Find(bson.M{"next_key":next_key}).One(&result)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	} else if err == mgo.ErrNotFound {
		return nil, nil
	}
	return &result, nil
}

func (qks *QuestStorage) GetAllKeys() ([]Key, error) {
	result := []Key{}
	err := qks.Keys.Find(bson.M{}).All(&result)
	for i, key := range result {
		result[i].SID = key.ID.Hex()
	}
	return result, err
}


//TEAMS
func (qs *QuestStorage) AddTeam(name string) error {
	err := qs.Teams.Insert(Team{Name:name})
	return err
}

func (qs *QuestStorage) GetAllTeams() ([]Team, error) {
	result := []Team{}
	err := qs.Teams.Find(bson.M{}).All(&result)
	for i, team := range result {
		result[i].SID = team.ID.Hex()
	}
	return result, err
}

func (qs *QuestStorage) GetTeamByName(name string) (*Team, error) {
	result := Team{}
	err := qs.Teams.Find(bson.M{"name":name}).One(&result)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	} else if err == mgo.ErrNotFound {
		return nil, nil
	}
	result.SID = result.ID.Hex()
	return &result, nil
}


//TEAM MEMBERS
func (qs *QuestStorage) AddTeamMember(user_id, m_name, phone, t_name string) error {
	team := Team{}
	err := qs.Teams.Find(bson.M{"name":t_name}).One(&team)
	if err != nil {
		return err
	}
	err = qs.TeamMembers.Insert(TeamMember{Name:m_name, Phone:phone, TeamName:t_name, UserId:user_id})
	return err
}

func (qs *QuestStorage) GetMembersOfTeam(team_name string) ([]TeamMember, error) {
	res := []TeamMember{}
	err := qs.TeamMembers.Find(bson.M{"team_name":team_name}).All(&res)
	return res, err
}

func (qs *QuestStorage) GetAllTeamMembers() ([]TeamMember, error) {
	res := []TeamMember{}
	err := qs.TeamMembers.Find(bson.M{}).All(&res)
	return res, err
}

func (qs *QuestStorage) GetTeamMembers(query bson.M) ([]TeamMember, error) {
	res := []TeamMember{}
	err := qs.TeamMembers.Find(query).All(&res)
	if err != nil && err != mgo.ErrNotFound {
		return res, err
	}
	return res, nil
}

func (qs *QuestStorage)GetTeamMemberByUserId(user_id string) (*TeamMember, error) {
	res := TeamMember{}
	err := qs.TeamMembers.Find(bson.M{"user_id":user_id}).One(&res)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	} else if err == mgo.ErrNotFound {
		return nil, nil
	}
	return &res, nil
}


//MESSAGES
func (qs *QuestStorage) StoreMessage(from, to, body string, is_key bool) error {
	result := Message{
		From: from,
		To:to,
		Body: body,
		Time: time.Now(),
		TimeStamp: time.Now().Unix(),
		Answered: false,
		IsKey: is_key,
	}
	err := qs.Messages.Insert(&result)
	return err
}


//type KeyWrapper struct {
//	ID          bson.ObjectId `bson:"_id,omitempty"`
//	SID         string
//	Key         string `bson:"key"`
//	Description string `bson:"description"`
//
//	IsFirst     bool   `bson:"is_first"`
//	NextKey     *string `bson:"next_key"`
//}
//
//func (kw KeyWrapper) String() string {
//	return fmt.Sprintf("KW [%s] [%s] %v -> \n%s\n -> %s", kw.ID, kw.IsFirst, kw.Key, kw.Description, kw.NextKey)
//}
//
//type QuestMessageWrapper struct {
//	ID        bson.ObjectId `bson:"_id,omitempty"`
//	SID       string
//	From      string `bson:"from"`
//	Body      string `bson:"body"`
//	TimeStamp int64 `bson:"time"`
//	Time      time.Time `bson:"time_obj"`
//	Answered  bool `bson:"answered"`
//	IsKey     bool `bson:"is_key"`
//}
//

//

//
//func (qks *QuestStorage) GetKeyInfo(key string) (*KeyWrapper, error) {
//	kw := KeyWrapper{}
//	err := qks.Keys.Find(bson.M{"key":key}).One(&kw)
//	if err != nil {
//		return nil, err
//	}
//	return &kw, nil
//}
//
//func (qks *QuestStorage) DeleteKey(key_id string) error {
//	err := qks.Keys.RemoveId(bson.ObjectIdHex(key_id))
//	return err
//}
//
//func (qks *QuestStorage) GetDescription(key string) (string, error) {
//	result := KeyWrapper{}
//	err := qks.Keys.Find(bson.M{"key":key}).One(&result)
//	return result.Description, err
//}
//
//func (qks *QuestStorage) StoreMessage(from, body string, time time.Time, is_key bool) error {
//	result := QuestMessageWrapper{
//		From: from,
//		Body: body,
//		TimeStamp: time.Unix(),
//		Answered: false,
//		IsKey: is_key,
//	}
//	err := qks.Messages.Insert(&result)
//	return err
//}
//
//func (qks *QuestStorage) SetMessageAnswer(message_id bson.ObjectId) error {
//	err := qks.Messages.UpdateId(message_id, bson.M{"$set":bson.M{"answered":true}})
//	return err
//}
//
//func (qs *QuestStorage) GetMessage(message_id string) (*QuestMessageWrapper, error) {
//	result := QuestMessageWrapper{}
//	err := qs.Messages.FindId(bson.ObjectIdHex(message_id)).One(&result)
//	result.SID = result.ID.Hex()
//	return &result, err
//}
//
//func (qks *QuestStorage) GetMessages(query bson.M) ([]QuestMessageWrapper, error) {
//	result := []QuestMessageWrapper{}
//	err := qks.Messages.Find(query).Sort("-time").All(&result)
//	for i, message := range result {
//		result[i].SID = message.ID.Hex()
//		result[i].Time = time.Unix(message.TimeStamp, 0)
//	}
//	return result, err
//}
//
//type QuestUserWrapper struct {
//	ID      bson.ObjectId `bson:"_id,omitempty"`
//	UserId  string    `bson:"user_id"`
//	Name    string    `bson:"name"`
//	Phone   string    `bson:"phone"`
//	EMail   string    `bson:"email"`
//	State   map[string]string `bson:"state"`
//	Keys    map[string][]string `bson:"found_keys"`
//	LastKey map[string]*string `bson:"last_key"`
//}
//
//func (qks *QuestStorage)AddUser(user_id, name, email, phone, state, provider string) error {
//	find := bson.M{"user_id":user_id}
//	user := QuestUserWrapper{}
//	err := qks.Users.Find(find).One(&user)
//	if err == mgo.ErrNotFound {
//		qks.Users.Insert(QuestUserWrapper{UserId:user_id, Name:name, Phone:phone, EMail:email, State:map[string]string{provider:state}})
//	} else if err != nil {
//		return err
//	} else {
//		qks.Users.Update(find, bson.M{"$set":bson.M{fmt.Sprintf("state.%s", provider):state}})
//	}
//	return nil
//}
//func (qs *QuestStorage) SetUserState(user_id, state, provider string) error {
//	find := bson.M{"user_id":user_id}
//	err := qs.Users.Update(find, bson.M{"$set":bson.M{fmt.Sprintf("state.%s", provider):state}})
//	return err
//}
//
//func (qks *QuestStorage) GetUserState(user_id, provider string) (string, error) {
//	find := bson.M{"user_id":user_id}
//	user := QuestUserWrapper{}
//	err := qks.Users.Find(find).One(&user)
//	if err != nil {
//		return "", err
//	}
//	if state, ok := user.State[provider]; ok {
//		return state, nil
//	} else {
//		return "", nil
//	}
//}
//
//func (qks *QuestStorage) SetUserLastKey(user_id, key, provider string) error {
//	find := bson.M{"user_id":user_id}
//	err := qks.Users.Update(find, bson.M{
//		"$addToSet":bson.M{fmt.Sprintf("found_keys.%s", provider):key},
//		"$set":bson.M{fmt.Sprintf("last_key.%s", provider):key},
//	})
//	return err
//}
//
//type CurrentProviderUserInfo struct {
//	UserId    string
//	State     string
//	FoundKeys []string
//	LastKey   *string
//	User      QuestUserWrapper
//}
//
//func (qks *QuestStorage) GetUserInfo(user_id, provider string) (*CurrentProviderUserInfo, error) {
//	find := bson.M{"user_id":user_id}
//	user := QuestUserWrapper{}
//	err := qks.Users.Find(find).One(&user)
//	if err != nil {
//		return nil, err
//	}
//	state, _ := user.State[provider]
//	keys, _ := user.Keys[provider]
//	last_key, _ := user.LastKey[provider]
//
//	return &CurrentProviderUserInfo{
//		UserId:user.UserId,
//		State:state,
//		FoundKeys:keys,
//		LastKey:last_key,
//		User:user,
//	}, nil
//
//}
//
//func (qks *QuestStorage) GetUserKeys(user_id, key, provider string) ([]string, error) {
//	find := bson.M{"user_id":user_id}
//	user := QuestUserWrapper{}
//	err := qks.Users.Find(find).One(&user)
//	if err != nil {
//		return []string{}, err
//	} else {
//		if keys, ok := user.Keys[provider]; ok {
//			return keys, nil
//		}else {
//			return []string{}, nil
//		}
//	}
//}
//
//func (qks *QuestStorage) GetSubscribedUsers() ([]QuestUserWrapper, error) {
//	users := []QuestUserWrapper{}
//	err := qks.Users.Find(bson.M{fmt.Sprintf("state.%s", PROVIDER):SUBSCRIBED}).All(&users)
//	return users, err
//}
//
//func (qs *QuestStorage) GetAllUsers() ([]QuestUserWrapper, error) {
//	users := []QuestUserWrapper{}
//	err := qs.Users.Find(bson.M{}).All(&users)
//	return users, err
//}