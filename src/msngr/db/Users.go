package db

import (
	"gopkg.in/mgo.v2/bson"
	"time"
	"gopkg.in/mgo.v2"
	"msngr/utils"
	"fmt"
	"errors"
	"log"
)

type UserData struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	States      map[string]string `bson:"states"`
	UserId      string `bson:"user_id"`
	UserName    string `bson:"user_name"`
	ShowedName  string `bson:"showed_name"`
	Password    string `bson:"password"`
	Phone       string `bson:"phone"`
	Email       string `bson:"email"`
	LastUpdate  time.Time `bson:"last_update"`

	LastLogged  time.Time `bson:"last_logged"`
	Role        string `bson:"role"`
	ReadRights  []string `bson:"read_rights"`
	WriteRights []string `bson:"write_rights"`
	Auth        bool `bson:"auth"`
}

func (uw *UserData) GetStateValue(state_key string) (string, bool) {
	res, ok := uw.States[state_key]
	return res, ok
}

func (uw *UserData) GetName() string {
	if uw.ShowedName != "" {
		return uw.ShowedName
	}
	return uw.UserName
}

type UserHandler struct {
	UsersCollection *mgo.Collection
	parent          *MainDb
}

func (uh *UserHandler) ensureIndexes() {
	usersCollection := uh.parent.Session.DB(uh.parent.DbName).C("users")
	usersCollection.EnsureIndex(mgo.Index{
		Key:        []string{"user_id"},
		Background: true,
		Unique:     true,
		DropDups:   true,
	})
	usersCollection.EnsureIndex(mgo.Index{
		Key:        []string{"last_update"},
	})
	usersCollection.EnsureIndex(mgo.Index{
		Key:        []string{"user_state"},
	})
	usersCollection.EnsureIndex(mgo.Index{
		Key:        []string{"user_name"},
	})
	usersCollection.EnsureIndex(mgo.Index{
		Key:        []string{"role"},
	})
	usersCollection.EnsureIndex(mgo.Index{
		Key:        []string{"last_marker"},
	})
	usersCollection.EnsureIndex(mgo.Index{
		Key:        []string{"auth"},
	})
	uh.UsersCollection = usersCollection
}

func (uh *UserHandler) UserLogin(userId string) (*UserData, error) {
	found := &UserData{}
	err := uh.UsersCollection.Find(bson.M{"user_id":userId}).One(found)
	if err != nil {
		return nil, err
	}
	err = uh.UsersCollection.Update(bson.M{"user_id":userId}, bson.M{"$set":bson.M{"auth":true, "last_logged":time.Now()}})
	return found, err
}

func (uh *UserHandler) UserLogout(userId string) (*UserData, error) {
	found := &UserData{}
	err := uh.UsersCollection.Find(bson.M{"user_id":userId}).One(found)
	if err != nil {
		return nil, err
	}
	err = uh.UsersCollection.Update(bson.M{"user_id":userId}, bson.M{"$set":bson.M{"auth":false}})
	return found, err
}

func (uh *UserHandler) GetUserDataForWeb(userName, password string) (*UserData, error) {
	tmp := UserData{}
	err := uh.UsersCollection.Find(bson.M{"$or":[]bson.M{bson.M{"user_name": userName}, bson.M{"email":userName}}, "password": utils.PHash(password)}).One(&tmp)
	return &tmp, err
}

func (uh *UserHandler) AddRightToUser(userId string, rightType string, rights ...string) error {
	if !utils.InS(rightType, []string{"read", "write"}) {
		errText := fmt.Sprintf("Unsupported type of right, except read or wright, but %v", rightType)
		log.Printf("DB USERS ERROR %v", errText)
		return errors.New(errText)
	}
	err := uh.UsersCollection.Update(bson.M{"user_id":userId}, bson.M{"$addToSet":bson.M{fmt.Sprintf("%v_rights", rightType):rights}})
	return err
}