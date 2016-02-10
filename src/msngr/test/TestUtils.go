package test

import (
	"path"
	s "msngr/structs"
	"io/ioutil"
	"log"
	"encoding/json"
	"net/http"
	"bytes"
	"errors"
	"os"
	"msngr"
)

const (
	TEST_RESOURCE_PATH = "test_res"
)

func GetTestFileName(fn string) string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	test_dir := path.Join(dir, TEST_RESOURCE_PATH, fn)
	return test_dir
}

func ReadTestFile(fn string) *s.InPkg {
	fileName := GetTestFileName(fn)
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Printf("error at read: %q \n", err)
	}
	in := s.InPkg{}
	log.Printf("READ IN FROM FILE:\n%+v", string(data))
	err = json.Unmarshal(data, &in)
	if err != nil {
		log.Printf("error at unmarshal: %+v \n", err)
	}
	return &in
}

func POST(address string, out *s.InPkg) (*s.OutPkg, error) {
	log.Printf("will send: %v to %v", address, out)
	jsoned_out, err := json.Marshal(out)
	if err != nil {
		log.Printf("TEST POST error at unmarshal %v out: %v", err, out)
		return nil, err
	}

	body := bytes.NewBuffer(jsoned_out)
	req, err := http.NewRequest("POST", address, body)
	if err != nil {
		log.Printf("TEST POST error at for request %v", err)
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	//print, _ := json.MarshalIndent(out, "", "	")
	if out.Message != nil && out.Message.Body != nil {
		log.Printf("TP >> %+v [%v]", *out.Message.Body, out.UserData.Name)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("TEST POST error at do request %v", err)
		return nil, err
	}
	if resp != nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("TEST POST << ERROR:%+v", err)
			return nil, err
		} else {
			in := s.OutPkg{}
			err := json.Unmarshal(body, &in)
			if in.Message != nil {
				log.Printf("TP << %+v [%v]", in.Message.Body, out.UserData.Name)
			}
			if err != nil {
				log.Printf("!TEST POST err in unmarshal [%v] body: %s", err, body)
				return nil, err
			}
			return &in, nil
		}
		defer resp.Body.Close()
	}
	return nil, errors.New("response is nil :(")
}

func HandleAddress(address, port string, errors chan s.MessageError) chan string {
	result := make(chan string, 1000)
	go func() {
		log.Printf("TSTU will work at localhost%v/%v", port, address)

		http.HandleFunc(address, func(w http.ResponseWriter, r *http.Request) {
			in, err := msngr.FormInPackage(r)
			log.Printf("TSTU: handling request by %+v", in)
			if err != nil {
				log.Printf("TSTU: at %v can not recognise request %+v because %v", address, r, err)
			}
			out_err := <-errors

			if in.Message.Body != nil {
				body := in.Message.Body
				out := &s.OutPkg{To:in.From, Message:&s.OutMessage{Body:*body, Error:&out_err}}
				log.Printf("TSTU: sending not defered, error %+v", out)
				msngr.PutOutPackage(w, out, true, false)
				result <- out_err.Condition
			}

		})
		server := &http.Server{Addr: port}
		result <- "listen"
		log.Fatal(server.ListenAndServe())

	}()
	return result
}

func FakeAddressSupplier() {

}