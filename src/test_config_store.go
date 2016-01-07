package main
import (
	"msngr/configuration"
	"msngr/quests"
	s "msngr/structs"
	"log"
)

var QUEST_NOT_SUBSCRIBED_COMMANDS = []s.OutCommand{
	s.OutCommand{
		Title:    "Учавствовать",
		Action:   "subscribe",
		Position: 0,
		Repeated: false,
	},
}

var key_input_form = &s.OutForm{
	Title: "Форма ввода ключа для следующего задания",
	Type:  "form",
	Name:  "key_form",
	Text:  "Код: ?(code)",
	Fields: []s.OutField{
		s.OutField{
			Name: "code",
			Type: "text",
			Attributes: s.FieldAttribute{
				Label:    "Ваш найденный код",
				Required: true,
			},
		},
	},
}

var QUEST_SUBSCRIBED_COMMANDS = []s.OutCommand{
	s.OutCommand{
		Title:    "Ввод найденного кода",
		Action:   "key_input",
		Position: 0,
		Repeated: false,
		Form:     key_input_form,
	},
	s.OutCommand{
		Title:    "Перестать участвовать",
		Action:"unsubscribe",
		Position:1,
		Repeated:false,
	},
}


func main() {
	cs := configuration.NewConfigurationStorage("localhost:27017", "bot")
	SUBSCRIBED, _ := cs.LoadCommands(quests.PROVIDER, quests.SUBSCRIBED)
	log.Printf("SUBSCRIBED COMMANDS: %+v", SUBSCRIBED)
	UNSUBSCRIBED, _ := cs.LoadCommands(quests.PROVIDER, quests.UNSUBSCRIBED)
	log.Printf("UNSUBSCRIBED COMMANDS: %+v", UNSUBSCRIBED)
	for _, here := range QUEST_SUBSCRIBED_COMMANDS {
		for _, db := range SUBSCRIBED {
			log.Printf("subscribed : %v\n db: %v, here: %v", here.String() == db.String(), db, here)
		}
	}
	for _, here := range QUEST_NOT_SUBSCRIBED_COMMANDS {
		for _, db := range UNSUBSCRIBED {
			log.Printf("unsubscribed: %v\n db: %v, here: %v", here.String() == db.String(), db, here)
		}
	}


}
