package db

import (
	"testing"
	"log"

	"msngr/configuration"

	"time"
	"gopkg.in/mgo.v2/bson"
)

func TestAutoAnswersStoring(t *testing.T) {
	log.Printf("start tes....")
	config := configuration.ReadTestConfigInRecursive()
	store := NewMainDb(config.Main.Database.ConnString, config.Main.Database.Name)
	store.Messages.Collection.RemoveAll(bson.M{})

	store.Messages.StoreMessage("from", "to", "test", "testMessageId")
	store.Messages.StoreMessage("from", "to", "test1", "testMessageId1")
	messages, _ := store.Messages.GetMessagesForAutoAnswer("to", 0)
	log.Printf("messages: %+v", messages)
	if len(messages) != 2{
		t.Errorf("I must see 2 messages to auto answer 0 but %v", len(messages))
	}

	messages, _ = store.Messages.GetMessagesForAutoAnswer("to", 1)
	log.Printf("messages1: %+v", messages)
	if len(messages) != 0{
		t.Errorf("I must see 0 messages to auto answer 1 but %v", len(messages))
	}
	time.Sleep(time.Minute)

	messages, _ = store.Messages.GetMessagesForAutoAnswer("to", 0)
	log.Printf("messages: %+v", messages)
	if len(messages) != 2{
		t.Errorf("I must see 2 messages to auto answer 0 after sleep but %v", len(messages))
	}

	messages, _ = store.Messages.GetMessagesForAutoAnswer("to", 1)
	log.Printf("messages1after sleep: %+v", messages)
	if len(messages) != 2{
		t.Errorf("I must see 2 messages to auto answer 1 after sleep but %v", len(messages))
	}

	store.Messages.SetMessagesAutoAnswered("from", "to", 0)
	messages, _ = store.Messages.GetMessagesForAutoAnswer("to", 0)
	log.Printf("messages after auto aswer 0: %+v", messages)
	if len(messages) != 0{
		t.Errorf("I must see 2 messages to auto answer 0 after sleep and set auto answered but %v", len(messages))
	}
	store.Messages.SetMessagesAutoAnswered("from", "to", 1)
	messages, _ = store.Messages.GetMessagesForAutoAnswer("to", 1)
	log.Printf("messages1after sleep and answer 0: %+v", messages)
	if len(messages) != 0{
		t.Errorf("I must see 2 messages to auto answer 1 after sleep and set auto answered but %v", len(messages))
	}


}
