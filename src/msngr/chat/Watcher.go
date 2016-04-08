package chat

import (
	"msngr/db"
	"time"
	"gopkg.in/mgo.v2/bson"
	"log"
	"msngr/notify"
	"msngr/configuration"
)

func Watch(messageStore *db.MessageHandler, ntf *msngr.Notifier, config configuration.ChatConfig) {
	for {
		froms := map[string]bool{}
		timeStampLess := time.Now().Add(-(time.Duration(config.AutoAnswer.After) * time.Minute)).Unix()
		//log.Printf("CW will get not answered messages to %v, and with time stamp less than: %v", config.CompanyId, timeStampLess)
		messages, err := messageStore.GetMessages(bson.M{
			"to":config.CompanyId,
			"not_answered":1,
			"unread":1,
			"time_stamp":bson.M{
				"$lte": timeStampLess,
			}})

		if err != nil {
			log.Printf("CW ERROR %v", err)
		}
		for _, message := range messages {
			if _, ok := froms[message.From]; !ok {
				ntf.NotifyText(message.From, config.AutoAnswer.Text)
				messageStore.SetMessagesAnswered(message.From, config.CompanyId, "bot")
				froms[message.From] = true
			}
		}
		//log.Printf("CW was process %v messages", len(messages))
		time.Sleep(10 * time.Second)
	}
}
