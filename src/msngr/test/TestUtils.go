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
)

const (
	TEST_RESOURCE_PATH = "test_res"
)

func GetTestFileName(fn string) string {
	dir, err := os.Getwd()
	if err != nil{
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
	jsoned_out, err := json.Marshal(out)
	if err != nil {
		log.Printf("TEST POST error at unmarshal %v", err)
		return nil, err
	}

	body := bytes.NewBuffer(jsoned_out)
	req, err := http.NewRequest("POST", address, body)
	if err != nil {
		log.Printf("TEST POST error at for request %v", err)
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	print, _ := json.MarshalIndent(out, "", "	")
	log.Printf("TEST POST >> %+v \n%+v \n %s", address, req.Header, print)

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
			log.Printf("TEST POST << %v", string(body))
			in := s.OutPkg{}
			err := json.Unmarshal(body, &in)
			if err != nil {
				log.Printf("TEST POST err in unmarshal %v", err)
				return nil, err
			}
			return &in, nil
		}
		defer resp.Body.Close()
	}
	return nil, errors.New("response is nil :(")
}

func FakeAddressSupplier() {

}