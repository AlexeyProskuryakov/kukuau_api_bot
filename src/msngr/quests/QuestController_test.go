package quests

import (
	"testing"
	s "msngr/structs"
	"msngr/configuration"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"strings"
)

var keys = []string{"1", "2", "three", "four"}
var teams = []string{"one", "two", "3"}
var salt = "TEST"

func prep_pack(from, message string) *s.InPkg {
	return &s.InPkg{Message:&s.InMessage{Body:&message}, UserData:&s.InUserData{Name:from, Phone:from, Email:from}, From:from, }
}

func f_k(k_id, t_id int) string {
	return fmt.Sprintf("#%v-%v", keys[k_id], teams[t_id])
}
func f_d(k_id, t_id int) string {
	return fmt.Sprintf("%v%v%v", keys[k_id], teams[t_id], salt)
}

func kd(k_id, t_id int) (string, string) {
	return f_k(k_id, t_id), f_d(k_id, t_id)
}

func prep_keys(qs *QuestStorage) {
	qs.Steps.RemoveAll(bson.M{})
	for i, _ := range keys {
		for j, _ := range teams {
			if i + 1 < len(keys) {
				qs.AddKey(f_k(i, j), f_d(i, j), f_k(i + 1, j))
			} else {
				qs.AddKey(f_k(i, j), f_d(i, j), "")
			}
		}
	}
}

//func assert_out_have(out *s.MessageResult, what string){
//	return strings.Contains(out.Body, what)
//}

func TestMessageProcessor(t *testing.T) {
	conf := configuration.ReadConfigInRecursive()
	conf.Main.Database.Name = conf.Main.Database.Name + "_autotest"
	qs := NewQuestStorage(conf.Main.Database.ConnString, conf.Main.Database.Name)
	cs := configuration.NewConfigurationStorage(conf.Main.Database.ConnString, conf.Main.Database.Name)
	qp := QuestMessageProcessor{Storage:qs, ConfigStorage:cs}
	prep_keys(qs)
	k, d := kd(0, 0)
	out := qp.ProcessMessage(prep_pack("U1", k))
	if !strings.Contains(out.Body, d) {
		t.Error("First key not found %v", out.Body)
	}
	k, d = kd(1, 0)
	out = qp.ProcessMessage(prep_pack("U1", k))
	if !strings.Contains(out.Body, d) {
		t.Error("Next key not found %v", out.Body)
	}

	k, d = kd(0, 0)
	out = qp.ProcessMessage(prep_pack("U1", k))
	if !strings.Contains(out.Body, d) {
		t.Error("Next key not found %v", out.Body)
	}
	k, d = kd(1, 0)
	out = qp.ProcessMessage(prep_pack("U1", k))
	if !strings.Contains(out.Body, d) {
		t.Error("Next key not found %v", out.Body)
	}
	k, d = kd(1, 0)
	out = qp.ProcessMessage(prep_pack("U1", k))
	if !strings.Contains(out.Body, d) {
		t.Error("Next key not found %v", out.Body)
	}

	k, d = kd(0, 1)
	out = qp.ProcessMessage(prep_pack("U1", k))
	if !strings.Contains(out.Body, d) {
		t.Error("Go to another command err %v", out.Body)
	}

	k, d = kd(2, 0)
	out = qp.ProcessMessage(prep_pack("U1", k))
	if strings.Contains(out.Body, d) {
		t.Error("From ahother command you can not send keys 2 %v", out.Body)
	}

	k, d = kd(1, 0)
	out = qp.ProcessMessage(prep_pack("U1", k))
	if strings.Contains(out.Body, d) {
		t.Error("From ahother command you can not send keys 1 %v", out.Body)
	}

	k, d = kd(0, 0)
	out = qp.ProcessMessage(prep_pack("U1", k))
	if !strings.Contains(out.Body, d) {
		t.Error("Go to first command err %v", out.Body)
	}

	k, d = kd(1, 1)
	out = qp.ProcessMessage(prep_pack("U1", k))
	if strings.Contains(out.Body, d) {
		t.Error("Can not send code of another team %v", out.Body)
	}

	k, d = kd(1, 2)
	out = qp.ProcessMessage(prep_pack("U1", k))
	if strings.Contains(out.Body, d) {
		t.Error("Can not send code of another team %v", out.Body)
	}

	k, d = kd(1, 2)
	out = qp.ProcessMessage(prep_pack("U2", k))
	if strings.Contains(out.Body, d) {
		t.Error("you must register at first 2 %v", out.Body)
	}
	k, d = kd(1, 1)
	out = qp.ProcessMessage(prep_pack("U2", k))
	if strings.Contains(out.Body, d) {
		t.Error("you must register at first 1 %v", out.Body)
	}
	k, d = kd(1, 0)
	out = qp.ProcessMessage(prep_pack("U2", k))
	if strings.Contains(out.Body, d) {
		t.Error("you must register at first 0 %v", out.Body)
	}

	k, d = kd(0, 0)
	out = qp.ProcessMessage(prep_pack("U2", k))
	if !strings.Contains(out.Body, d) {
		t.Error("Next user not adding to team 0 %v", out.Body)
	}

	k, d = kd(1, 0)
	out = qp.ProcessMessage(prep_pack("U2", k))
	if !strings.Contains(out.Body, d) {
		t.Error("Next user can not send next key to team 0 %v", out.Body)
	}

	k, d = kd(0, 1)
	out = qp.ProcessMessage(prep_pack("U2", k))
	if !strings.Contains(out.Body, d) {
		t.Error("Next user go to another team to team 1 %v", out.Body)
	}

}
