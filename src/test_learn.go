package main

import (
	l "learn"
	"log"
)

func main() {
	channel := l.ReadToChan("config.json")
	var line string
	ok := true
	for {
		log.Println(line, ok)
		if !ok {
			break
		} else {
			line, ok = <-channel
		}
	}
}
