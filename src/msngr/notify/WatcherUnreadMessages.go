package notify

import (
	"msngr/db"
	"time"
	"log"
)

func WatchUnreadMessages(mainStore *db.MainDb, confStore *db.ConfigurationStorage, address string) {
	log.Printf("WUM start...")
	for {
		chatsConfig, err := confStore.GetAllChatsConfig()
		if err != nil {
			log.Printf("WUM ERROR at gettgin chat config")
			continue
		}
		//log.Printf("WUM find %v configs", len(chatsConfig))
		for _, config := range chatsConfig {
			for _, cfg_notification := range config.Notifications {
				messages, err := mainStore.Messages.GetMessagesForNotification(config.CompanyId, cfg_notification.After)
				if err != nil {
					log.Printf("WUM ERROR at retrieve messages for notify [%v]: %v", config.CompanyId, err)
					continue
				}
				if len(messages) > 0 {
					log.Printf("WUM Will notify by %v messages for [%v]", len(messages), config.CompanyId)
					notifier := NewNotifier(address, config.Key, mainStore)
					_, _, err = notifier.NotifyTextToMembers(cfg_notification.Text)
					if err != nil {
						log.Printf("WUM ERROR at send notification to %v: %v", config.CompanyId, err)
						continue
					}
					for _, message := range messages {
						mainStore.Messages.SetMessagesNotified(message.From, config.CompanyId, cfg_notification.After)
					}
				}
			}
		}
		time.Sleep(10 * time.Second)
	}
}
