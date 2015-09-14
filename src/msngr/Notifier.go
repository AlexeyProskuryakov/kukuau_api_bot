package msngr

import (
	"bytes"
	"encoding/json"
	// "io/ioutil"
	"log"
	"net/http"
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

//todo notifier must be more flexiability. you can notify different messages for different engines (shop, taxi)
//you must think about it

func NewNotifier(addr, key string) *Notifier {
	return &Notifier{address: addr, key: key}
}

func (n Notifier) Notify(outPkg OutPkg) {
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
		// log.Println("N response Status:", resp.Status)
		// log.Println("N response Headers:", resp.Header)
		// resp_body, _ := ioutil.ReadAll(resp.Body)
		// log.Println("N response Body:", string(resp_body))
	}

}
