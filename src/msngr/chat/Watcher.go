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
		timeStampLess := time.Now().Add(-(time.Duration(config.AutoAnswer.After) * time.Minute)).Unix()
		//log.Printf("CW will get not answered messages to %v, and with time stamp less than: %v", config.CompanyId, timeStampLess)
		messages, err := messageStore.GetMessages(bson.M{
			"to":config.CompanyId,
			"not_answered":1,
			"time_stamp":bson.M{
				"$lte": timeStampLess,
			}})
		if err != nil {
			log.Printf("CW ERROR %v", err)
		}
		for _, message := range messages {
			ntf.NotifyText(message.From, config.AutoAnswer.Text)
			messageStore.SetMessageAnswered(message.SID, "bot")
		}
		//log.Printf("CW was process %v messages", len(messages))
		time.Sleep(10 * time.Second)
	}
}
