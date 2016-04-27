package utils

import (
	"testing"
	"log"
)

func TestGenId(t *testing.T){
	first_id := GenStringId()
	second_id := GenStringId()
	if first_id == second_id {
		t.Errorf("First (%+v) == Second (%+v)", first_id, second_id)
	}
	log.Printf("ids: \n%v\n%v", first_id, second_id)
}
