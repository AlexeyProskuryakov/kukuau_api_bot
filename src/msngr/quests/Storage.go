package quests

import (
	"msngr/db"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"time"
	"sort"
	"errors"
	"log"
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
}

type Step struct {
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
	Steps    *mgo.Collection
	Messages *mgo.Collection
	Teams    *mgo.Collection
	Peoples  *mgo.Collection
}

func (qks *QuestStorage) ensureIndexes() {
	collection := qks.Session.DB(qks.DbName).C("quest_steps")
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
	qks.Steps = collection

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
func (qks *QuestStorage) AddKey(start_key, description, next_key string) (*Step, error) {
	kw := Step{}
	err := qks.Steps.Find(bson.M{"start_key":start_key}).One(&kw)
	if err == mgo.ErrNotFound {
		kw = Step{StartKey:start_key, Description:description, NextKey:next_key}
		if team_name, err := GetTeamNameFromKey(start_key); team_name != "" && err == nil {
			kw.ForTeam = team_name
		}
		err = qks.Steps.Insert(kw)
		return &kw, err
	} else if err != nil {
		return nil, err
	} else {
		return &kw, errors.New("Key already exist!")
	}
}

func (qks *QuestStorage) DeleteKey(key_id string) error {
	err := qks.Steps.RemoveId(bson.ObjectIdHex(key_id))
	return err
}

func (qks *QuestStorage) UpdateKey(key_id, start_key, description, next_key string) error {
	err := qks.Steps.UpdateId(bson.ObjectIdHex(key_id), bson.M{"$set":bson.M{"description":description, "start_key":start_key, "next_key":next_key}})
	return err
}

func (qs *QuestStorage) GetKeys(query bson.M) ([]Step, error) {
	result := []Step{}
	err := qs.Steps.Find(query).All(&result)
	if err != nil && err != mgo.ErrNotFound {
		return result, err
	}
	for i, k := range result {
		result[i].SID = k.ID.Hex()
	}
	return result, nil
}

func (qs *QuestStorage) GetKeyByStartKey(start_key string) (*Step, error) {
	result := Step{}
	err := qs.Steps.Find(bson.M{"start_key":start_key}).One(&result)

	if err != nil && err != mgo.ErrNotFound {
		log.Printf("QS: Error %v", err)
		return nil, err
	} else if err == mgo.ErrNotFound {
		log.Printf("QS: Not found key")
		return nil, nil
	}

	return &result, nil
}

func (qs *QuestStorage) SetKeyFounded(key, by string) error {
	err := qs.Steps.Update(bson.M{"start_key":key}, bson.M{"$set":bson.M{"found_by":by, "is_found":true}})
	if err != nil {
		return err
	}
	err = qs.Teams.Update(bson.M{"name":by}, bson.M{"$addToSet":bson.M{"found_keys": key}})
	return err
}

func (qs *QuestStorage) GetKeyByNextKey(next_key string) (*Step, error) {
	result := Step{}
	err := qs.Steps.Find(bson.M{"next_key":next_key}).One(&result)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	} else if err == mgo.ErrNotFound {
		return nil, nil
	}
	return &result, nil
}

func (qks *QuestStorage) GetAllKeys() ([]Step, error) {
	result := []Step{}
	err := qks.Steps.Find(bson.M{}).All(&result)
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
	return result, err
}

func (qs *QuestStorage) SetMessagesAnswered(from, by string) error {
	_, err := qs.Messages.UpdateAll(
		bson.M{"from":from, "not_answered":bson.M{"$ne":0}},
		bson.M{"$set":bson.M{
			"answered_by":by,
			"not_answered":0}})
	return err
}

func (qs *QuestStorage) GetMessages(query bson.M) ([]Message, error) {
	messages := []Message{}
	query["is_key"] = false
	err := qs.Messages.Find(query).Sort("-time").Limit(25).All(&messages)
	return messages, err
}

func (qs *QuestStorage) GetMessagesKeys(query bson.M) ([]Message, error) {
	messages := []Message{}
	query["is_key"] = true
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
	Time             int64 `bson:"time"`
}

//CONTACTS
type ByContactsTeam []Contact

func (s ByContactsTeam) Len() int {
	return len(s)
}
func (s ByContactsTeam) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByContactsTeam) Less(i, j int) bool {
	if (s[i].IsTeam && s[j].IsTeam) || (!s[i].IsTeam && !s[j].IsTeam) {
		return s[i].Time > s[j].Time
	}
	return false
}

func (qs *QuestStorage) GetContacts(teams []Team) ([]Contact, error) {
	resp := []Contact{}
	err := qs.Messages.Pipe([]bson.M{
		bson.M{"$match":bson.M{"is_key":false}},
		bson.M{"$group": bson.M{"_id":"$from", "not_answered_count":bson.M{"$sum":"$not_answered"}, "name":bson.M{"$first":"$from"}, "time":bson.M{"$max":"$time_stamp"}}}}).All(&resp)
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
		bson.M{"$match":bson.M{"time_stamp":bson.M{"$gt":after}, "from":bson.M{"$ne":ME}, "to":bson.M{"$ne":ALL}, "is_key":false}},
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
