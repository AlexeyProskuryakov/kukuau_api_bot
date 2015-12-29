package main

import (
	tm "msngr/text_messages"
	"log"
)

func main() {
	fga := tm.NewFGA()
	for i:=0; i<1000;i++{
		result := fga.GenerateMessage()
		log.Printf(">>>> %s", result)
	}
}
