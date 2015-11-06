package main

import (
	u "msngr/utils"
	"log"
)


func main() {
	result := u.FoundFile("test_res")
	log.Println(*result)
}