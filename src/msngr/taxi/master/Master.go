package master


import (
	t "msngr/taxi"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"log"

	"time"
	"io/ioutil"
	"encoding/json"
	"errors"
)

var ErrorsMap = map[int]string{
	1:"Неизвестная ошибка",
	2:"Неизвестный тип API",
	3:"API отключено в настройках модуля TM API в ТМ2",
	4:"Не совпадает секретный ключ",
	5:"Неподдерживаемая версия API",
	6:"Неизвестное название запроса",
	7:"Неверный тип запроса GET/POST",
	8:"Не хватает входного параметра (в доп. информации ответа будет название отсутствующего параметра)",
	9:"Некорректный входной параметр (в доп. информации ответа будет название некорректного параметра)",
	10:"Внутренняя ошибка обработки запроса",
}

const MAX_TRY_COUNT = 5

func doReq(req *http.Request) ([]byte, error) {
	count := 0
	for {
		client := &http.Client{}
		res, err := client.Do(req)
		if res == nil || err != nil {
			log.Println("TM response is: ", res, "; error is:", err, ". I will reconnect and will retrieve data again after 3s.")
			time.Sleep(3 * time.Second)
			count += 1
			if count > MAX_TRY_COUNT {
				return []byte{}, errors.New(fmt.Sprintf("Max tryings count ended and request not proceed... %#v", req))
			}
			continue
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		return body, err
	}
	return []byte{}, errors.New(fmt.Sprintf("Error at do request... %#v", req))
}




type TaxiMasterAPI struct {
	ConnString  string
	BearerToken string
}

type TMAPIResponse struct {
	Code        int `json:"code"`
	Description string `json:"descr"`
	Data        map[string]interface{} `json:"data"`
}

func (m *TaxiMasterAPI) _createSignature(params map[string]string) string {
	params_string := ""
	for k, v := range params {
		params_string += fmt.Sprintf("%v=%v?", k, v)
	}
	params_string += m.BearerToken
	h := md5.New()
	io.WriteString(h, params_string)
	signature := fmt.Sprintf("%x", h.Sum(nil))
	return signature
}

func (m *TaxiMasterAPI) _get_request(method string, params map[string]string, security bool) ([]byte, error) {
	req, err := http.NewRequest("GET", m.ConnString + method, nil)
	if err != nil {
		log.Printf("TM gr error in request")
	}
	if security {
		signature := m._createSignature(params)
		req.Header.Add("Signature", signature)
	}
	values := req.URL.Query()
	for k, v := range params {
		values.Add(k, v)
	}
	req.URL.RawQuery = values.Encode()
	return doReq(req)
}

func (m *TaxiMasterAPI) _post_request(method string, params map[string]string, security bool) ([]byte, error) {
	req, err := http.NewRequest("POST", m.ConnString + method, nil)
	if err != nil {
		log.Printf("TM pr error in request")
	}
	if security {
		signature := m._createSignature(params)
		req.Header.Add("Signature", signature)
	}
	for k, v := range params {
		req.Form.Add(k, v)
	}
	return doReq(req)
}

func (m *TaxiMasterAPI) Ping() {
	result, err := m._get_request("ping", map[string]string{}, false)
	if err != nil {
		log.Println("error at ping...")
	}
	r := TMAPIResponse{}
	err = json.Unmarshal(result, &r)
	if err != nil {
		log.Println("error at unmarshal ing result")
	}
}

type Tariff struct {
	Id       int `json:"id"`
	Name     string `json:"name"`
	IsActive bool `json:"is_active"`
}

type TariffWrapper struct {
	TMAPIResponse
	Data struct {
			 Tariffs []Tariff `json:"tariffs"`
		 } `json:"data"`
}

func (m *TaxiMasterAPI) GetTariffList() []Tariff {
	request_result, err := m._get_request("get_tarif_list", map[string]string{}, true)
	result_wrapper := TariffWrapper{}
	if err != nil {
		log.Println("error at getting tarif list")
		return result_wrapper.Data.Tariffs
	}
	err = json.Unmarshal(request_result, &result_wrapper)
	if err != nil {
		log.Println("error at unmarshalling json from tariff list")
		return result_wrapper.Data.Tariffs
	}
	return result_wrapper.Data.Tariffs
}

func (m *TaxiMasterAPI)NewOrder(order t.NewOrder) t.Answer {
	return t.Answer{}
}
func (m *TaxiMasterAPI)CancelOrder(order_id int64) (bool, string) {
	return false, ""
}
func (m *TaxiMasterAPI)CalcOrderCost(order t.NewOrder) (int, string) {
	return 0, ""
}
func (m *TaxiMasterAPI)Orders() []t.Order {

}
func (m *TaxiMasterAPI)Feedback(f t.Feedback) (bool, string) {
	return false, ""
}
func (m *TaxiMasterAPI)GetCarsInfo() []t.CarInfo {
	return []t.CarInfo{}
}

func (m *TaxiMasterAPI)AddressesSearch(query string) t.FastAddress {
	return t.FastAddress{}
}

func (m *TaxiMasterAPI)IsConnected() bool {
	return false
}