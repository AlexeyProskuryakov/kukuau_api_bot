package notify

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
	from    string
}

func NewNotifier(addr, key string, dbh *db.MainDb) *Notifier {
	return &Notifier{address: addr, key: key, _db:dbh}
}

func (n *Notifier) SetFrom(from string) {
	n.from = from
}

func (n *Notifier) Notify(outPkg s.OutPkg) (*db.MessageWrapper, error) {
	jsoned_out, err := json.Marshal(&outPkg)
	if err != nil {
		log.Printf("NTF error at unmarshal %v", err)
		return nil, err
	}
	body := bytes.NewBuffer(jsoned_out)
	req, err := http.NewRequest("POST", n.address, body)
	if err != nil {
		log.Printf("NTF error at for request %v", err)
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", n.key)

	log.Printf("N\n>> %+v \n>> %+v \n>> %v\n>>%s", n.address, n.key, req.Header, jsoned_out)
	var resultMessage *db.MessageWrapper
	if n.from == "" {
		resultMessage, _ = n._db.Messages.StoreNotificationMessage("me", outPkg.To, outPkg.Message.Body, outPkg.Message.ID)
	} else {
		resultMessage, _ = n._db.Messages.StoreNotificationMessage(n.from, outPkg.To, outPkg.Message.Body, outPkg.Message.ID)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("NTF error at do request %v", err)
		n._db.Messages.UpdateMessageStatus(outPkg.Message.ID, "error", err.Error())
		return resultMessage, err
	}

	if resp != nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("N << ERROR:%+v", err)
			n._db.Messages.UpdateMessageStatus(outPkg.Message.ID, "error", fmt.Sprintf("%v", resp.StatusCode))
			return resultMessage, err
		} else {
			log.Printf("N << %v", string(body))
			n._db.Messages.UpdateMessageStatus(outPkg.Message.ID, "sended", "ok")
		}
		defer resp.Body.Close()
	} else {
		n._db.Messages.UpdateMessageStatus(outPkg.Message.ID, "error", "404")
		return resultMessage, errors.New("404")
	}
	return resultMessage, nil
}

func (n *Notifier) NotifyText(to, text string) (*s.OutPkg, *db.MessageWrapper, error) {
	result := s.OutPkg{To:to, Message:&s.OutMessage{ID:utils.GenStringId(), Type:"chat", Body:text}}
	message, err := n.Notify(result)
	return &result, message, err
}

func (n *Notifier) NotifyTextWithCommands(to, text string, commands *[]s.OutCommand) (*s.OutPkg, *db.MessageWrapper, error) {
	result := s.OutPkg{To:to, Message:&s.OutMessage{ID:utils.GenStringId(), Type:"chat", Body:text, Commands:commands}}
	message, err := n.Notify(result)
	return &result, message, err
}

func (n *Notifier) NotifyTextToMembers(text string) (*s.OutPkg, *db.MessageWrapper, error) {
	result := s.OutPkg{Message:&s.OutMessage{ID:utils.GenStringId(), Type:"chat", Body:text}}
	message, err := n.Notify(result)
	return &result, message, err
}

func (n *Notifier)SendMessageToPeople(people []db.UserWrapper, text string) {
	go func() {
		for _, user := range people {
			n.NotifyText(user.UserId, text)
		}
	}()
}
