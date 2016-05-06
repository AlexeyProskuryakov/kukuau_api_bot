package chat

import (
	"msngr/db"
	"time"
	"log"
	n "msngr/notify"
)

func WatchNotAnsweredMessages(mainStore *db.MainDb, configStorage *db.ConfigurationStorage, ntfAddress string) {
	for {
		configs, err := configStorage.GetAllChatsConfig()
		//log.Printf("WNAM found %v configs", len(configs))
		if err != nil {
			log.Printf("WNAM ERROR when get chat config %v", err)
			time.Sleep(10 * time.Second)
			continue
		}
		for _, config := range configs {
			notifier := n.NewNotifier(ntfAddress, config.Key, mainStore)
			for _, autoAnswerCfg := range config.AutoAnswers {
				if autoAnswerCfg.After != 0 {
					messages, err := mainStore.Messages.GetMessagesForAutoAnswer(config.CompanyId, autoAnswerCfg.After)
					if err != nil {
						log.Printf("CW ERROR AT REtrieving messages %v", err)
						continue
					}
					froms := map[string]bool{}
					for _, message := range messages {
						if _, ok := froms[message.From]; !ok {
							notifier.NotifyText(message.From, autoAnswerCfg.Text)
							mainStore.Messages.SetMessagesAutoAnswered(message.From, config.CompanyId, autoAnswerCfg.After)
							froms[message.From] = true
						}
					}
				}
			}
		}

		time.Sleep(10 * time.Second)
	}
}
