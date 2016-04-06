package geo
import (
	"testing"
	"msngr/configuration"
	"msngr/taxi/infinity"
	"fmt"
	"log"
)

func GetFakeInfinityOwnAddressHandler() *OwnAddressHandler {
	conf := configuration.ReadConfigInRecursive()
	ext := infinity.GetInfinityAddressSupplier(conf.Taxis["fake"].Api, "fake")
	own_ah := NewOwnAddressHandler(conf.Main.ElasticConn, conf.Taxis["fake"].Api.GeoOrbit, ext)
	return own_ah
}

func GetAllExtInfoByString(s_n string, ah *OwnAddressHandler) ([]string, bool) {
	addrs_pkg := ah.AddressesAutocomplete(s_n)
	result := []string{}
	if addrs_pkg.Rows != nil {
		rows := *addrs_pkg.Rows
		for i, _ := range rows {
			adrs := rows[i]
			ext_address, err := ah.GetExternalInfo(fmt.Sprintf("%v", adrs.OSM_ID), adrs.Name)
			if ext_address != nil && err == nil {
				result = append(result, fmt.Sprintf("%v %v", ext_address.Name, ext_address.City))
				//				return fmt.Sprintf("for %v > %v[%v] %+v", s_n, adrs.Name, adrs.City, ext_address), true
			}
		}
		return result, true
	}
	return []string{fmt.Sprintf("for %v not found autocomplete", s_n)}, false
}


func GetFirstExtInfoByString(s_n string, ah *OwnAddressHandler) (string, bool) {
	addrs_pkg := ah.AddressesAutocomplete(s_n)
	if addrs_pkg.Rows != nil {
		rows := *addrs_pkg.Rows
		if len(rows) > 0 {
			adrs := rows[0]
			ext_address, err := ah.GetExternalInfo(fmt.Sprintf("%v", adrs.OSM_ID), adrs.Name)
			if ext_address != nil && err == nil {
				return fmt.Sprintf("for %v > %v[%v] %+v", s_n, adrs.Name, adrs.City, ext_address), true
			}
		} else {
			return fmt.Sprintf("for %v rows is 0", s_n), false
		}
	}
	return fmt.Sprintf("for %v not found", s_n), false
}

func TestAddressRecognise(t *testing.T) {
	ah := GetFakeInfinityOwnAddressHandler()
	s_n := "российс"
	if res, ok := GetFirstExtInfoByString(s_n, ah); !ok {
		t.Errorf("Res %+v for [%v] is bad", res, s_n)
	}
	s_n = "демако"
	if res, ok := GetFirstExtInfoByString(s_n, ah); !ok {
		t.Errorf("Res %+v for [%v] is bad", res, s_n)
	}else {
		log.Printf("OK! %v for %v", res, s_n)
	}
	s_n = "карла маркс"
	if res, ok := GetFirstExtInfoByString(s_n, ah); !ok {
		t.Errorf("Res %+v for [%v] is bad", res, s_n)
	}else {
		log.Printf("OK! %v for %v", res, s_n)
	}
	s_n = "ленин"
	if res, ok := GetFirstExtInfoByString(s_n, ah); !ok {
		t.Errorf("Res %+v for [%v] is bad", res, s_n)
	}else {
		log.Printf("OK! %v for %v", res, s_n)
	}
	s_n = "лесосе"
	if res, ok := GetFirstExtInfoByString(s_n, ah); !ok {
		t.Errorf("Res %+v for [%v] is bad", res, s_n)
	}else {
		log.Printf("OK! %v for %v", res, s_n)
	}
	s_n = "бульвар молод"
	if res, ok := GetFirstExtInfoByString(s_n, ah); !ok {
		t.Errorf("Res %+v for [%v] is bad", res, s_n)
	}else {
		log.Printf("OK! %v for %v", res, s_n)
	}
	s_n = "весенний"
	if res, ok := GetFirstExtInfoByString(s_n, ah); !ok {
		t.Errorf("Res %+v for [%v] is bad", res, s_n)
	}else {
		log.Printf("OK! %v for %v", res, s_n)
	}
	s_n = "детский"
	if res, ok := GetFirstExtInfoByString(s_n, ah); !ok {
		t.Errorf("Res %+v for [%v] is bad", res, s_n)
	}else {
		log.Printf("OK! %v for %v", res, s_n)
	}
	s_n = "пер"
	if res, ok := GetFirstExtInfoByString(s_n, ah); !ok {
		t.Errorf("Res %+v for [%v] is bad", res, s_n)
	}else {
		log.Printf("OK! %v for %v", res, s_n)
	}
	s_n = "улиц"
	if res, ok := GetFirstExtInfoByString(s_n, ah); !ok {
		t.Errorf("Res %+v for [%v] is bad", res, s_n)
	}else {
		log.Printf("OK! %v for %v", res, s_n)
	}
	s_n = "советска"
	if res, ok := GetFirstExtInfoByString(s_n, ah); !ok {
		t.Errorf("Res %+v for [%v] is bad", res, s_n)
	}else {
		log.Printf("OK! %v for %v", res, s_n)
	}
	s_n = "ключ"
	if res, ok := GetFirstExtInfoByString(s_n, ah); !ok {
		t.Errorf("Res %+v for [%v] is bad", res, s_n)
	}else {
		log.Printf("OK! %v for %v", res, s_n)
	}
	s_n = "горская"
	if res, ok := GetFirstExtInfoByString(s_n, ah); !ok {
		t.Errorf("Res %+v for [%v] is bad", res, s_n)
	}else {
		log.Printf("OK! %v for %v", res, s_n)
	}

	s_n = "хуй"
	if res, ok := GetFirstExtInfoByString(s_n, ah); ok {
		t.Errorf("Res  %+v for [%v] must be bad", res, s_n)
	}else {
		log.Printf("OK! %v for %v", res, s_n)
	}

	s_n = "курва"
	if res, ok := GetFirstExtInfoByString(s_n, ah); ok {
		t.Errorf("Res  %+v for [%v] must be bad", res, s_n)
	}else {
		log.Printf("OK! %v for %v", res, s_n)
	}


}
