package configuration

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"msngr/structs"
	"msngr/db"
)

type ConfigurationStorage struct {
	db.DbHelper
	Commands *mgo.Collection
	Configs  *mgo.Collection
}

func (cs *ConfigurationStorage) ensureIndexes() {
	commands_collection := cs.Session.DB(cs.DbName).C("commands")
	commands_collection.EnsureIndex(mgo.Index{
		Key:        []string{"name", "provider"},
		Background: true,
		DropDups:   true,
		Unique:    true,
	})

	cs.Commands = commands_collection

	configs_collection := cs.Session.DB(cs.DbName).C("configs")
	configs_collection.EnsureIndex(mgo.Index{
		Key:        []string{"name", "type"},
		Background: true,
		DropDups:   true,
	})
	cs.Configs = configs_collection
}

func NewConfigurationStorage(conn, dbname string) ConfigStorage {
	helper := db.NewDbHelper(conn, dbname)
	res := &ConfigurationStorage{DbHelper:*helper}
	res.ensureIndexes()
	return res
}

type ConfigStorage interface {
	SaveCommand(provider, name string, command structs.OutCommand) error
	LoadCommands(provider, name string) ([]structs.OutCommand, error)
	GetCommand(req bson.M) (*structs.OutCommand, error)
}

type CommandsWrapper struct {
	Name     string `bson:"name"`
	Provider string `bson:"provider"`
	Commands []structs.OutCommand `bson:"commands"`
}

func (cs *ConfigurationStorage) SaveCommand(provider, name string, command structs.OutCommand) error {
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

func (cs *ConfigurationStorage) LoadCommands(provider, name string) ([]structs.OutCommand, error) {
	res := CommandsWrapper{}
	err := cs.Commands.Find(bson.M{"name":name, "provider":provider}).Sort("commands.position").One(&res)
	if err != nil {
		log.Printf("CS Error at find commands by name: %v provider: %v %v", name, provider, err)
	}
	return res.Commands, err
}

func (cs *ConfigurationStorage) GetCommand(req bson.M) (*structs.OutCommand, error) {
	res := &structs.OutCommand{}
	err := cs.Commands.Find(req).One(res)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	} else if err == mgo.ErrNotFound {
		return nil, nil
	}
	return res, nil
}

