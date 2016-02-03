package quests

import (
	"msngr/db"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"time"
	"log"
	"sort"
)

type Message struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	SID         string
	From        string `bson:"from"`
	To          string `bson:"to"`
	Body        string `bson:"body"`
	Time        time.Time `bson:"time"`
	TimeStamp   int64 `bson:"time_stamp"`
	NotAnswered int `bson:"not_answered"`
	IsKey       bool `bson:"is_key"`
	AnswerOf    string `bson:"answer_of,omitempty"`
	AnsweredBy  string `bson:"answered_by,omitempty"`
}
type Key struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	SID         string
	Founded     bool `bson:"is_found"`
	FoundedBy   string `bson:"found_by"`
	Founder     string `bson:"founder"`
	StartKey    string `bson:"start_key"`
	NextKey     string `bson:"next_key"`
	Description string `bson:"description"`
	ForTeam     string `bson:"for_team"`
}

type TeamMember struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	UserId    string `bson:"user_id"`
	Phone     string `bson:"phone"`
	Name      string `bson:"name"`
	TeamName  string `bson:"team_name,omitempty"`
	TeamSID   string `bson:"team_sid,omitempty"`
	Passersby bool  `bson:"is_passerby,omitempty"`
}

type Passersby struct {
	TeamMember
}

type Team struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	SID       string
	Name      string `bson:"name"`
	FoundKeys []string `bson:"found_keys"`
}

type QuestStorage struct {
	db.DbHelper
	Keys     *mgo.Collection
	Messages *mgo.Collection
	Teams    *mgo.Collection
	Peoples  *mgo.Collection
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
		Key:[]string{"not_answered"},
	})

	qks.Messages = message_collection

	teams_collection := qks.Session.DB(qks.DbName).C("quest_teams")
	teams_collection.EnsureIndex(mgo.Index{
		Key:[]string{"name"},
		Unique:true,
	})
	qks.Teams = teams_collection

	members_collection := qks.Session.DB(qks.DbName).C("quest_peoples")
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
	members_collection.EnsureIndex(mgo.Index{
		Key:[]string{"is_passerby"},
	})

	qks.Peoples = members_collection
}

func NewQuestStorage(conn, dbname string) *QuestStorage {
	helper := db.NewDbHelper(conn, dbname)
	result := QuestStorage{DbHelper:*helper}
	result.ensureIndexes()
	return &result
}

//KEYS
func (qks *QuestStorage) AddKey(start_key, description, next_key string) (*Key, error) {
	kw := Key{}
	err := qks.Keys.Find(bson.M{"start_key":start_key}).One(&kw)
	if err == mgo.ErrNotFound {
		kw = Key{StartKey:start_key, Description:description, NextKey:next_key}
		if team_name, err := GetTeamNameFromKey(start_key); team_name != "" && err == nil {
			kw.ForTeam = team_name
		}
		err = qks.Keys.Insert(kw)
		return &kw, err
	} else if err != nil {
		return nil, err
	} else {
		return &kw, nil
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
	for i, k := range result {
		result[i].SID = k.ID.Hex()
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
func (qs *QuestStorage) AddTeam(name string) (*Team, error) {
	team := Team{Name:name}
	err := qs.Teams.Insert(team)
	team.SID = team.ID.Hex()
	return &team, err
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
func (qs *QuestStorage) AddTeamMember(user_id, m_name, phone string, team *Team) (*TeamMember, error) {
	tm := TeamMember{}
	err := qs.Peoples.Find(bson.M{"user_id":user_id}).One(&tm)
	if err == mgo.ErrNotFound {
		tm = TeamMember{Name:m_name, Phone:phone, TeamName:team.Name, TeamSID:team.ID.Hex(), UserId:user_id, Passersby:false}
		err = qs.Peoples.Insert(tm)
	} else {
		err = qs.Peoples.UpdateId(tm.ID, bson.M{"$set":bson.M{"is_passerby":false, "team_name":team.Name, "team_sid":team.SID}})
	}
	return &tm, err
}
func (qs *QuestStorage) SetTeamForTeamMember(new_tn *Team, user_id *TeamMember) error {
	return qs.Peoples.Update(bson.M{"user_id":user_id.UserId}, bson.M{"$set":bson.M{"team_name":new_tn.Name, "team_sid":new_tn.SID}})
}
func (qs *QuestStorage) GetMembersOfTeam(team_name string) ([]TeamMember, error) {
	res := []TeamMember{}
	err := qs.Peoples.Find(bson.M{"team_name":team_name}).All(&res)
	return res, err
}

func (qs *QuestStorage) GetAllTeamMembers() ([]TeamMember, error) {
	res := []TeamMember{}
	err := qs.Peoples.Find(bson.M{"team_name":bson.M{"$exists":true}}).All(&res)
	return res, err
}

func (qs *QuestStorage) GetPeoples(query bson.M) ([]TeamMember, error) {
	res := []TeamMember{}
	err := qs.Peoples.Find(query).All(&res)
	if err != nil && err != mgo.ErrNotFound {
		return res, err
	}
	return res, nil
}

func (qs *QuestStorage) GetManByUserId(user_id string) (*TeamMember, error) {
	res := TeamMember{}
	err := qs.Peoples.Find(bson.M{"user_id":user_id}).One(&res)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	}else if err == mgo.ErrNotFound {
		return nil, nil
	}
	return &res, nil
}

func (qs *QuestStorage)GetTeamMemberByUserId(user_id string) (*TeamMember, error) {
	res := TeamMember{}
	err := qs.Peoples.Find(bson.M{"user_id":user_id, "team_name":bson.M{"$exists":true}}).One(&res)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	} else if err == mgo.ErrNotFound {
		return nil, nil
	}
	return &res, nil
}

func (qs *QuestStorage)AddPasserby(user_id, phone, name string) (*TeamMember, error) {
	res := TeamMember{}
	err := qs.Peoples.Find(bson.M{"user_id":user_id, "phone":phone, "name":name}).One(&res)
	if err == mgo.ErrNotFound {
		tm := TeamMember{Name:name, UserId:user_id, Phone:phone, Passersby:true}
		err = qs.Peoples.Insert(tm)
	} else {
		err = qs.Peoples.UpdateId(res.ID, bson.M{"$set":bson.M{"is_passerby":true}})
	}
	return &res, err
}

func (qs *QuestStorage)GetPassersby(query bson.M) (*TeamMember, error) {
	res := TeamMember{}
	query["is_passerby"] = true
	err := qs.Peoples.Find(query).One(&res)
	if err == mgo.ErrNotFound {
		return nil, nil
	}else if err != nil {
		return nil, err
	}
	return &res, err
}

//MESSAGES
func (qs *QuestStorage) StoreMessage(from, to, body string, is_key bool) (Message, error) {
	result := Message{
		From: from,
		To:to,
		Body: body,
		Time: time.Now(),
		TimeStamp: time.Now().Unix(),
		NotAnswered: 1,
		IsKey: is_key,
	}
	err := qs.Messages.Insert(result)
	log.Printf("QS: Message stored: id is: %v", result.ID)
	return result, err
}

func (qs *QuestStorage) SetMessagesAnswered(from, by string) error {
	_, err := qs.Messages.UpdateAll(
		bson.M{"from":from},
		bson.M{"$set":bson.M{
			"answered_by":by,
			"not_answered":0}})
	return err
}

func (qs *QuestStorage) GetMessages(query bson.M) ([]Message, error) {
	messages := []Message{}
	err := qs.Messages.Find(query).Sort("-time").Limit(25).All(&messages)
	return messages, err
}

type Contact struct {
	ID               string `bson:"_id"`
	Name             string `bson:"name"`
	NewMessagesCount int `bson:"not_answered_count"`
	Team             *Team
	Phone            string
	IsPassersby      bool
	IsTeam           bool
}

//CONTACTS
type ByContactsTeam []Contact

// We implement `sort.Interface` - `Len`, `Less`, and
// `Swap` - on our type so we can use the `sort` package's
// generic `Sort` function. `Len` and `Swap`
// will usually be similar across types and `Less` will
// hold the actual custom sorting logic. In our case we
// want to sort in order of increasing string length, so
// we use `len(s[i])` and `len(s[j])` here.
func (s ByContactsTeam) Len() int {
    return len(s)
}
func (s ByContactsTeam) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}
func (s ByContactsTeam) Less(i, j int) bool {
	if s[i].IsTeam && !s[j].IsTeam{
		return true
	}
	return false
}

func (qs *QuestStorage) GetContacts(teams []Team) ([]Contact, error) {
	resp := []Contact{}
	err := qs.Messages.Pipe([]bson.M{
		bson.M{"$group": bson.M{"_id":"$from", "not_answered_count":bson.M{"$sum":"$not_answered"}, "name":bson.M{"$first":"$from"}}}}).All(&resp)
	if err != nil {
		return resp, err
	}
	teams_map := map[string]*Team{}
	for _, team := range teams {
		teams_map[team.Name] = &team
	}
	result := []Contact{}

	for i, contact := range resp {
		if team, ok := teams_map[contact.Name]; ok {
			resp[i].Team = team
			resp[i].IsTeam = true
			resp[i].IsPassersby = false
			result = append(result, resp[i])

		} else {
			pb, _ := qs.GetPassersby(bson.M{"user_id":resp[i].ID})
			if pb != nil {
				resp[i].Phone = pb.Phone
				resp[i].Name = pb.Name
				resp[i].IsPassersby = true
				resp[i].IsTeam = false
				result = append(result, resp[i])
			}
		}
	}
	sort.Sort(ByContactsTeam(result))
	return result, err
}
func (qs QuestStorage) GetContactsAfter(after int64) ([]Contact, error) {
	resp := []Contact{}
	err := qs.Messages.Pipe([]bson.M{
		bson.M{"$match":bson.M{"time_stamp":bson.M{"$gt":after}, "from":bson.M{"$ne":ME}, "to":bson.M{"$ne":ALL}}},
		bson.M{"$group": bson.M{"_id":"$from", "not_answered_count":bson.M{"$sum":"$not_answered"}, "name":bson.M{"$first":"$from"}}}}).All(&resp)
	if err != nil {
		return resp, err
	}
	result := []Contact{}

	for i, contact := range resp {
		t, _ := qs.GetTeamByName(contact.ID)
		if t != nil {
			resp[i].Name = t.Name
			resp[i].IsTeam = true
			resp[i].IsPassersby = false
			result = append(result, resp[i])
		}else {
			m, _ := qs.GetPassersby(bson.M{"user_id":resp[i].ID})
			if m != nil {
				resp[i].Name = m.Name
				resp[i].Phone = m.Phone
				resp[i].IsPassersby = true
				result = append(result, resp[i])
			}
		}
	}
	return result, nil
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