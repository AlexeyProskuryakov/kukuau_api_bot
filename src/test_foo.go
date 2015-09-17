package main

import (
	"log"
	u "msngr/utils"
)

func main() {
	i := 5
	if 3 < i && i < 7 {
		log.Println("foo")
	}
	log.Println(u.Priority("", "", 1234))

}
