package chat

import (
	"msngr/db"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type userCompanyMappingElement struct {
	UserId    string `bson:"user_id"`
	CompanyId string `bson:"company_id"`
}


type ChatStorage struct {
	db.DbHelper

	GlobalUsers *db.UserHandler
	userCompanyMappings     *mgo.Collection
}

func (s *ChatStorage) ensureIndexes() {
	users_collection := s.Session.DB(s.DbName).C("chat_users")
	users_collection.EnsureIndex(mgo.Index{
		Key:        []string{"user_id", "company_id"},
		Background: true,
		DropDups:   true,
		Unique: true,
	})
	s.userCompanyMappings = users_collection
}

func NewChatStorage(main_db *db.MainDb) *ChatStorage {
	result := ChatStorage{DbHelper:main_db.DbHelper, GlobalUsers:main_db.Users}
	result.ensureIndexes()
	return &result
}
func (s *ChatStorage) SetUserCompany(userId, companyId string) error {
	cu := userCompanyMappingElement{}
	err := s.userCompanyMappings.Find(bson.M{"user_id":userId, "company_id":companyId}).One(&cu)
	if err == mgo.ErrNotFound {
		err = s.userCompanyMappings.Insert(userCompanyMappingElement{UserId:userId, CompanyId:companyId})
		return err
	}
	return err
}

func (s *ChatStorage) GetUsersOfCompany(companyId string) ([]db.UserWrapper, error) {
	result := []db.UserWrapper{}
	users_mapping := []userCompanyMappingElement{}
	err := s.userCompanyMappings.Find(bson.M{"company_id":companyId}).All(&users_mapping)
	if err != nil {
		return result, err
	}
	for _, cu := range users_mapping {
		user, _ := s.GlobalUsers.GetUserById(cu.UserId)
		if user != nil {
			result = append(result, *user)
		}
	}
	return result, nil
}
