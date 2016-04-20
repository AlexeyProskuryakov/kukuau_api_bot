package notify

import (
	"msngr/db"
	"gopkg.in/mgo.v2/bson"
	"time"
	"log"
)

type NotifyConfiguration struct {
	Name         string
	SecondsAfter int64
	Text         string
	Key          string
}

type WatchManager struct {
	Store   *db.MainDb
	Configs []NotifyConfiguration
	Address string
}

func NewWatchManager(store *db.MainDb, address string) *WatchManager {
	log.Printf("WUM initializing Watch Unread Manager")
	result := WatchManager{Store:store, Configs:[]NotifyConfiguration{}, Address:address}
	return &result
}

func (wm *WatchManager)AddConfiguration(profileName, text, key string, secondsAfter int64) {
	log.Printf("WUM Add configuration for %v with text [%v] and seconds after %v", profileName, text, secondsAfter)
	wm.Configs = append(wm.Configs, NotifyConfiguration{Name:profileName, SecondsAfter:secondsAfter, Text:text, Key:key})
}

func (wm *WatchManager)WatchUnreadMessages() {
	log.Printf("WUM start...")
	for {
		for _, config := range wm.Configs {
			messages, err := wm.Store.Messages.GetMessages(bson.M{
				"unread":1,
				"to":config.Name,
				"time_stamp":bson.M{"$lte":time.Now().Unix() - config.SecondsAfter},
				"notification_sent":false,
			})
			if err != nil {
				log.Printf("WUM ERROR at retrieve messages for notify [%v]: %v", config.Name, err)
				continue
			}
			if len(messages) > 0 {
				log.Printf("WUM Will notify by %v messages for [%v]", len(messages), config.Name)
				notifier := NewNotifier(wm.Address, config.Key, wm.Store)
				_, err = notifier.NotifyTextToMembers(config.Name, config.Text)
				if err != nil {
					log.Printf("WUM ERROR at send notification to %v: %v", config.Name, err)
					continue
				}
				for _, message := range messages {
					wm.Store.Messages.SetMessageNotificationSent(message.MessageID)
				}
			}
		}
		time.Sleep(10 * time.Second)
	}
}
