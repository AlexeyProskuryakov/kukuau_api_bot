package main
import (
	"msngr/configuration"
	"msngr/quests"
	"log"
)

func main() {
	cs := configuration.NewConfigurationStorage("localhost:27017", "test_cs")
	for _, command := range quests.QUEST_SUBSCRIBED_COMMANDS {
		err := cs.SaveCommand("quest", "subscribed", command)
		log.Printf("command saved : %+v, err? : %s", command, err)
	}
	cmds, err := cs.LoadCommands("quest", "subscribed")
	if err != nil {
		log.Printf("Error load commands : %v", err)
	}
	if (cmds == quests.QUEST_NOT_SUBSCRIBED_COMMANDS) {
		log.Printf("commands equals: %+v", cmds)
	} else {
		log.Printf("commands not equals: %+v", cmds)
	}


}
