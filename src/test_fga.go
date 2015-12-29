package main

import (
	tm "msngr/text_messages"
	"log"
)

func main() {
	fga := tm.NewTextMessageSupplier()
	for i:=0; i<200;i++{
		result := fga.GenerateMessage()
		log.Printf(">>>> %s", result)
	}
}
