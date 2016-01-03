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
	cc := cs.Session.DB(cs.DbName).C("commands")
	cc.EnsureIndex(mgo.Index{
		Key:        []string{"name", "provider"},
		Background: true,
		DropDups:   true,
	})

	cs.Commands = cc

	ccfg := cs.Session.DB(cs.DbName).C("configs")
	ccfg.EnsureIndex(mgo.Index{
		Key:        []string{"name", "type"},
		Background: true,
		DropDups:   true,
	})
	cs.Configs = ccfg
}

func NewConfigurationStorage(conn, dbname string) *ConfigStorage {
	helper := db.NewMainDb(conn, dbname)
	res := &ConfigurationStorage{DbHelper:helper}
	res.ensureIndexes()
	return res
}

type ConfigStorage interface {
	SaveCommand(provider, name string, command structs.OutCommand) error
	LoadCommands(provider, name string) ([]structs.OutCommand, error)
	GetCommand(req bson.M) (*structs.OutCommand, error)
}

func (cs *ConfigurationStorage) SaveCommand(provider, name string, command structs.OutCommand) error {
	err := cs.Commands.Insert(bson.M{"name":name, "provider":provider, "command":command})
	if err != nil {
		log.Printf("CS Error at saving command %v", err)
	}
	return err
}

func (cs *ConfigurationStorage) LoadCommands(provider, name string) ([]structs.OutCommand, error) {
	res := []structs.OutCommand{}
	err := cs.Commands.Find(bson.M{"name":name, "provider":provider}).Sort("position").All(&res)
	if err != nil {
		log.Printf("CS Error at find commands by name: %v provider: %v", name, provider)
	}
	return res, err
}

func (cs *ConfigurationStorage) GetCommand(req bson.M) (*structs.OutCommand, error) {
	res := structs.OutCommand{}
	err := cs.Commands.Find(req).One(&res)
	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	} else if err == mgo.ErrNotFound {
		return nil, nil
	}
	return res, nil
}

