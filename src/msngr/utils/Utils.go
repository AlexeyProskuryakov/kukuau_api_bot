package utils

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"
	"regexp"
	"strings"

	"os"
	"log"
	"net/http"
	"io/ioutil"
	"path"
	"crypto/md5"
	"strconv"
)

func GenId() string {
	t := time.Now().UnixNano()
	s := rand.NewSource(t)
	r := rand.New(s)
	return fmt.Sprintf("%d", r.Int63())
}

func PHash(pwd string) (string) {
	input := []byte(pwd)
	output := md5.Sum(input)
	result := string(output[:])
	return result
}

func get_parent_path(path string) string {
	separator := RuneToAscii(os.PathSeparator)
	path_elements := strings.Split(path, separator)
	return strings.Join(path_elements[:len(path_elements) - 1], separator)
}

func RuneToAscii(r rune) string {
	if r < 128 {
		return string(r)
	} else {
		return "\\u" + strconv.FormatInt(int64(r), 16)
	}
}
func FoundFile(fname string) *string {
	log.Printf("Found file: %v\nPath sep: %s", fname, RuneToAscii(os.PathSeparator))
	dir, err := os.Getwd()

	prev_dir := dir
	if err != nil {
		return nil
	}
	for {
		log.Printf("Search config at: %v", dir)
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			return nil
		}
		log.Printf("files: %+v", files)
		for _, f := range files {
			log.Printf("analyse file %v", f)
			if fname == f.Name() {
				result := path.Join(dir, fname)
				return &result
			}
		}
		dir = get_parent_path(dir)
		if prev_dir == dir {
			log.Printf("Can not found file: %v", fname)
			return nil
		}else {
			prev_dir = dir
		}
		log.Printf("now dir is: %v", dir)

	}
	return nil
}

func ToMap(in interface{}, tag string) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// we only accept structs
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("ToMap only accepts structs; got %T", v)
	}

	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		// gets us a StructField
		fi := typ.Field(i)
		if tagv := fi.Tag.Get(tag); tagv != "" {
			out[tagv] = v.Field(i).Interface()
		}
	}
	return out, nil
}

func FirstOf(data ...interface{}) interface{} {
	for _, data_el := range data {
		if data_el != "" {
			return data_el
		}
	}
	return ""
}

func In(p int, a []int) bool {
	for _, v := range a {
		if p == v {
			return true
		}
	}
	return false
}

func InS(p string, a []string) bool {
	for _, v := range a {
		if p == v {
			return true
		}
	}
	return false
}

func IntersectionS(a1, a2 []string) bool {
	for _, v1 := range a1 {
		for _, v2 := range a2 {
			if v1 == v2 {
				return true
			}
		}
	}
	return false
}

func Contains(container string, elements []string) bool {
	container_elements := regexp.MustCompile("[a-zA-Zа-яА-Я]+").FindAllString(container, -1)
	ce_map := make(map[string]bool)
	for _, ce_element := range container_elements {
		ce_map[strings.ToLower(ce_element)] = true
	}
	result := true
	for _, element := range elements {
		_, ok := ce_map[strings.ToLower(element)]
		result = result && ok
	}
	return result
}

func SaveToFile(what, fn string) {
	f, err := os.OpenFile(fn, os.O_APPEND | os.O_WRONLY | os.O_CREATE, 0600)
	if err != nil {
		log.Printf("ERROR when save to file in open file %v [%v]", fn, err)
	}

	defer f.Close()

	if _, err = f.WriteString(what); err != nil {
		log.Printf("ERROR when save to file in write to file %v [%v]", fn, err)
	}
}

func GET(url string, params *map[string]string) (*[]byte, error) {
	//	log.Printf("GET > [%+v] |%+v|", url, params)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("ERROR IN GET FORM REQUEST! [%v]\n%v", url, err)
		return nil, err
	}

	if params != nil {
		values := req.URL.Query()
		for k, v := range *params {
			values.Add(k, v)
		}
		req.URL.RawQuery = values.Encode()
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if res == nil || err != nil {
		log.Println("ERROR IN GET DO REQUEST!\nRESPONSE: ", res, "\nERROR: ", err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	//	log.Printf("GET < \n%v\n", string(body), )
	return &body, err
}

type Predicate func() (bool)

func After(p Predicate, what func()) {
	go func() {
		for {
			if p() {
				what()
				break
			}
			time.Sleep(2 * time.Second)
		}
	}()
}

func DoDeferred(after time.Duration, what func()) {
	go func() {
		time.Sleep(after)
		what()
	}()
}