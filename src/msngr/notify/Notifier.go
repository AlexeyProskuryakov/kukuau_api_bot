package msngr

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	s "msngr/structs"
)

func warn(err error) {
	if err != nil {
		log.Println("notifier: ", err)
	}
}
func warnp(err error) {
	if err != nil {
		log.Println("notifier: ", err)
		panic(err)
	}
}

type Notifier struct {
	address string
	key     string
}


func NewNotifier(addr, key string) *Notifier {
	return &Notifier{address: addr, key: key}
}

func (n Notifier) Notify(outPkg s.OutPkg) {
	jsoned_out, err := json.Marshal(&outPkg)
	warn(err)

	body := bytes.NewBuffer(jsoned_out)
	req, err := http.NewRequest("POST", n.address, body)
	warnp(err)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", n.key)

	log.Printf("N >> %+v", req)

	client := &http.Client{}
	resp, err := client.Do(req)
	warn(err)

	if resp != nil {
		defer resp.Body.Close()
	}

}
