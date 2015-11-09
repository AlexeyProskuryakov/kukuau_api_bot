package test


import (
	"path"
	u "msngr/utils"
	s "msngr/structs"
	"io/ioutil"
	"log"
	"encoding/json"
)

const (
	TEST_RESOURCE_PATH = "test_res"
)

func GetTestFileName(fn string) *string {
	test_dir := u.FoundFile(TEST_RESOURCE_PATH)
	if test_dir != nil {
		result := path.Join(*test_dir, fn)
		return &result
	}
	return nil
}

func ReadTestFile(fn string) *s.InPkg {
	fileName := GetTestFileName(fn)
	if fileName == nil {
		return nil
	}
	data, err := ioutil.ReadFile(*fileName)
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


func FakeAddressSupplier() {

}