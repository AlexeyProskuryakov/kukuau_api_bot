package msngr

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	s "msngr/structs"
	"io/ioutil"
	"msngr/utils"
)


type Notifier struct {
	address string
	key     string
}


func NewNotifier(addr, key string) *Notifier {
	return &Notifier{address: addr, key: key}
}

func (n Notifier) Notify(outPkg s.OutPkg) error{
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
	if resp != nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil{
			log.Printf("N << ERROR:%+v", err)
			return err
		} else{
			log.Printf("N << %v", string(body))
		}
		defer resp.Body.Close()
	}
	return nil
}

func (n Notifier) NotifyText(to, text string) error{
	return n.Notify(s.OutPkg{To:to, Message:&s.OutMessage{ID:utils.GenId(), Type:"chat", Body:text}})
}
