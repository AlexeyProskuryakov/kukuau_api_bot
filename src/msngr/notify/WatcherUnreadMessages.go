package notify

import (
	"msngr/db"
	"gopkg.in/mgo.v2/bson"
	"time"
	"log"
	cfg "msngr/configuration"
)

type WatchManager struct {
	store       *db.MainDb
	address     string
	configStore *cfg.ConfigurationStorage
	companies   []string
}

func NewWatchManager(store *db.MainDb, address string) *WatchManager {
	log.Printf("WUM initializing Watch Unread Manager")
	result := WatchManager{store:store, address:address}
	return &result
}

func (wm *WatchManager)AddConfiguration(companyId string) {
	log.Printf("WUM Will watch for %v", companyId)
	wm.companies = append(wm.companies, companyId)
}

func (wm *WatchManager)WatchUnreadMessages() {
	log.Printf("WUM start...")
	for {
		for _, company := range wm.companies {
			config, err := wm.configStore.GetChatConfig(company)
			if err != nil {
				log.Printf("WUM ERROR at retrieve chat config for %v", company)
				continue
			}
			for _, cfg_notification := range config.Notifications {
				messages, err := wm.store.Messages.GetMessages(bson.M{
					"unread":1,
					"to":config.CompanyId,
					"time_stamp":bson.M{"$lte":time.Now().Unix() - int64(cfg_notification.After) * 60},
					"notification_sent":false,
				})
				if err != nil {
					log.Printf("WUM ERROR at retrieve messages for notify [%v]: %v", config.CompanyId, err)
					continue
				}
				if len(messages) > 0 {
					log.Printf("WUM Will notify by %v messages for [%v]", len(messages), config.CompanyId)
					notifier := NewNotifier(wm.address, config.Key, wm.store)
					_, _, err = notifier.NotifyTextToMembers(cfg_notification.Text)
					if err != nil {
						log.Printf("WUM ERROR at send notification to %v: %v", config.CompanyId, err)
						continue
					}
					for _, message := range messages {
						wm.store.Messages.SetMessageNotificationSent(message.From)
					}
				}
			}
		}
		time.Sleep(10 * time.Second)
	}
}
