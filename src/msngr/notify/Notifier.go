package msngr

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	s "msngr/structs"
	"io/ioutil"
	"msngr/utils"
	db "msngr/db"
	"fmt"
	"errors"
)

type Notifier struct {
	address string
	key     string
	_db     *db.MainDb
}

func NewNotifier(addr, key string, dbh *db.MainDb) *Notifier {
	return &Notifier{address: addr, key: key, _db:dbh}
}

func (n Notifier) Notify(outPkg s.OutPkg) error {
	jsoned_out, err := json.Marshal(&outPkg)
	if err != nil {
		log.Printf("NTF error at unmarshal %v", err)
		return err
	}
	body := bytes.NewBuffer(jsoned_out)
	req, err := http.NewRequest("POST", n.address, body)
	if err != nil {
		log.Printf("NTF error at for request %v", err)
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", n.key)

	log.Printf("N >> %+v \n>>%+v \n>>%s", n.address, req.Header, jsoned_out)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("NTF error at do request %v", err)
		return err
	}
	n._db.Messages.StoreMessage("ME", outPkg.To, outPkg.Message.Body, outPkg.Message.ID)
	if resp != nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("N << ERROR:%+v", err)
			n._db.Messages.UpdateMessageStatus(outPkg.Message.ID, "error", fmt.Sprintf("%v", resp.StatusCode))
			return err
		} else {
			log.Printf("N << %v", string(body))
			n._db.Messages.UpdateMessageStatus(outPkg.Message.ID, "sended", "ok")
		}
		defer resp.Body.Close()
	}else{
		n._db.Messages.UpdateMessageStatus(outPkg.Message.ID, "error", "404")
		return errors.New("404")
	}
	return nil
}

func (n Notifier) NotifyText(to, text string) (*s.OutPkg, error) {
	result := s.OutPkg{To:to, Message:&s.OutMessage{ID:utils.GenId(), Type:"chat", Body:text}}
	err := n.Notify(result)
	return &result, err
}
