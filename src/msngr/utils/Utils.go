package utils
import (
	"time"
	"fmt"
	"math/rand"
	"reflect"
)


func GetTimeAfter(d time.Duration, format string) string {
	result := time.Now().Add(d)
	return result.Format(format)
}

func GenId() string {
	return fmt.Sprintf("%d", rand.Int63())
}

func CheckErr(e error){
	if e != nil {
		panic(e)
	}
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

func Priority(data ...interface{}) interface{} {
	for _, data_el := range data {
		if data_el != ""{
			return data_el
		}
	}
	return ""
}
