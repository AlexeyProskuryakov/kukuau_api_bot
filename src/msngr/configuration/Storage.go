package configuration

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"msngr/structs"
	"msngr/db"
	"reflect"
	u "msngr/utils"
	"github.com/derekparker/delve/config"
)

type CommandsStorage struct {
	db.DbHelper
	Commands *mgo.Collection
}

func (cs *CommandsStorage) ensureIndexes() {
	commands_collection := cs.Session.DB(cs.DbName).C("commands")
	commands_collection.EnsureIndex(mgo.Index{
		Key:        []string{"name", "provider"},
		Background: true,
		DropDups:   true,
		Unique:    true,
	})

	cs.Commands = commands_collection
}

func NewCommandsStorage(conn, dbname string) *CommandsStorage {
	helper := db.NewDbHelper(conn, dbname)
	res := &CommandsStorage{DbHelper:*helper}
	res.ensureIndexes()
	return res
}

type CommandsStore interface {
	SaveCommand(provider, name string, command structs.OutCommand) error
	LoadCommands(provider, name string) ([]structs.OutCommand, error)
	GetCommand(req bson.M) (*structs.OutCommand, error)
}

type CommandsWrapper struct {
	Name     string `bson:"name"`
	Provider string `bson:"provider"`
	Commands []structs.OutCommand `bson:"commands"`
}

func (cs *CommandsStorage) SaveCommand(provider, name string, command structs.OutCommand) error {
	res := &CommandsWrapper{}
	err := cs.Commands.Find(bson.M{"name":name, "provider":provider}).One(res)
	if res.Name != "" && res.Provider != "" && len(res.Commands) > 0 {
		err = cs.Commands.Update(bson.M{"name":name, "provider":provider}, bson.M{"$addToSet":bson.M{"commands":command}})
	} else {
		err = cs.Commands.Insert(CommandsWrapper{Name:name, Provider:provider, Commands:[]structs.OutCommand{command}})
		if err != nil {
			log.Printf("CS Error at saving command %v", err)
		}
	}
	return err
}

func (cs *CommandsStorage) LoadCommands(provider, name string) ([]structs.OutCommand, error) {
	res := CommandsWrapper{}
	err := cs.Commands.Find(bson.M{"name":name, "provider":provider}).Sort("commands.position").One(&res)
	if err != nil {
		log.Printf("CS Error at find commands by name: %v provider: %v %v", name, provider, err)
	}
	return res.Commands, err
}

func (cs *CommandsStorage) GetCommand(req bson.M) (*structs.OutCommand, error) {
	res := &structs.OutCommand{}
	err := cs.Commands.Find(req).One(res)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	} else if err == mgo.ErrNotFound {
		return nil, nil
	}
	return res, nil
}

type ConfigurationStorage struct {
	db.DbHelper
	Collection *mgo.Collection
}

func NewConfigurationStorage(connection MongoDbConfig) *ConfigurationStorage {
	helper := db.NewDbHelper(connection.ConnString, connection.Name)
	result := &ConfigurationStorage{DbHelper:*helper}
	result.ensureIndexes()
	return result
}

func (cs *ConfigurationStorage) ensureIndexes() {
	collection := cs.Session.DB(cs.DbName).C("bots_configs")
	collection.EnsureIndex(mgo.Index{
		Key:[]string{"company_id"},
		Unique:true,
	})
	collection.EnsureIndex(mgo.Index{
		Key:[]string{"notifications"},
	})
	collection.EnsureIndex(mgo.Index{
		Key:[]string{"auto_answers"},
	})
}

func (cs *ConfigurationStorage) SetChatConfig(config ChatConfig, update bool) error {
	found := ChatConfig{}
	err := cs.Collection.Find(bson.M{"company_id":config.CompanyId}).One(&found)
	if err != nil && err != mgo.ErrNotFound {
		log.Printf("CFGS ERROR set chat config at find %v", err)
		return err
	}
	if err == mgo.ErrNotFound {
		cs.Collection.Insert(config)
	} else if update {
		modify := bson.M{}
		if reflect.DeepEqual(found.Notifications, config.Notifications) {
			modify["notifications"] = config.Notifications
		}
		if reflect.DeepEqual(found.AutoAnswers, config.AutoAnswers) {
			modify["auto_answers"] = config.AutoAnswers
		}
		update, _ := u.GetStringUpdates(found, config, "bson", true)
		for k, v := range update {
			modify[k] = v
		}
		cs.Collection.Update(bson.M{"company_id":config.CompanyId}, bson.M{"$set":modify})
	}
	return nil
}

func (cs *ConfigurationStorage) GetChatConfig(companyId string) (*ChatConfig, error) {
	found := ChatConfig{}
	err := cs.Collection.Find(bson.M{"company_id":companyId}).One(&found)
	if err != nil && err != mgo.ErrNotFound {
		log.Printf("CFGS ERROR set chat config at find %v", err)
		return nil, err
	} else if err == mgo.ErrNotFound {
		return nil, nil
	}
	return &found, nil
}