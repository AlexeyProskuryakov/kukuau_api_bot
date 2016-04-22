package chat

import (
	"msngr/db"
	"time"
	"log"
	n "msngr/notify"
	cfg "msngr/configuration"
)

func Watch(messageStore *db.MessageHandler, configStorage *cfg.ConfigurationStorage, notifier *n.Notifier, company_id string) {
	for {
		config, err := configStorage.GetChatConfig(company_id)
		if err != nil {
			log.Printf("AW ERROR when get chat config for %v", company_id)
			time.Sleep(10 * time.Second)
			continue
		}
		for _, autoAnswerCfg := range config.AutoAnswers {
			messages, err := messageStore.GetMessagesForAutoAnswer(company_id, autoAnswerCfg.After)
			if err != nil {
				log.Printf("CW ERROR AT REtrieving messages %v", err)
			}
			froms := map[string]bool{}
			for _, message := range messages {
				if _, ok := froms[message.From]; !ok {
					notifier.NotifyText(message.From, autoAnswerCfg.Text)
					messageStore.SetMessagesAutoAnswer(message.From, config.CompanyId, autoAnswerCfg.After)
					froms[message.From] = true
				}
			}
		}
		time.Sleep(10 * time.Second)
	}
}
