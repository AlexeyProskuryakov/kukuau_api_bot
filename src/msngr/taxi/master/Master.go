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

type TaxiMasterAPI struct {
	ConnString  string
	BearerToken string
}

type TMAPIResponse struct{
	Code int `json:"code"`
	Description string `json:"descr"`
	Data map[string]interface{} `json:"data"`
}

func (m *TaxiMasterAPI) _get_request(method string, params map[string]string, security bool) ([]byte, error) {
	req, err := http.NewRequest("GET", m.ConnString + method, nil)
	if err != nil {
		log.Printf("TM gr error in request")
	}
	if security {
		params_string := ""
		for k, v := range params {
			params_string += fmt.Sprintf("%v=%v?", k, v)
		}
		params_string+=m.BearerToken
		h := md5.New()
		io.WriteString(h, params_string)
		signature := fmt.Sprintf("%x", h.Sum(nil))
		req.Header.Add("Signature", signature)
	}

	values := req.URL.Query()
	for k, v := range params {
		values.Add(k, v)
	}

	req.URL.RawQuery = values.Encode()

	client := &http.Client{}
	res, err := client.Do(req)
	if res == nil || err != nil {
		log.Println("TM response is: ", res, "; error is:", err, ". I will reconnect and will retrieve data again after 3s.")
		time.Sleep(3 * time.Second)

	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	return body, err
}

func (m *TaxiMasterAPI) _post_request(method string, params map[string]string, security bool) ([]byte, error){
	req, err := http.NewRequest("POST", m.ConnString + method, nil)
	if err != nil {
		log.Printf("TM pr error in request")
	}
	if security {
		params_string := ""
		for k, v := range params {
			params_string += fmt.Sprintf("%v=%v?", k, v)
		}
		params_string+=m.BearerToken
		h := md5.New()
		io.WriteString(h, params_string)
		signature := fmt.Sprintf("%x", h.Sum(nil))
		req.Header.Add("Signature", signature)
	}

	for k, v := range params {
		req.Form.Add(k, v)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if res == nil || err != nil {
		log.Println("TM response is: ", res, "; error is:", err, ". I will reconnect and will retrieve data again after 3s.")
		time.Sleep(3 * time.Second)

	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	return body, err
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
type Tariff struct{
	Id int
	Name string
	IsActive bool
}

type TariffWrapper struct {
	TMAPIResponse
	Data
}
func (m *TaxiMasterAPI) GetTariffList() {
	result, err := m._get_request("get_tarif_list", map[string]string{}, true)
	if err != nil{
		log.Println("error at getting tarif list")
	}
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
	return []t.Order
}
func (m *TaxiMasterAPI)Feedback(f t.Feedback) (bool, string) {
	return false, ""
}
func (m *TaxiMasterAPI)GetCarsInfo() []t.CarInfo {
	return []t.CarInfo
}

func (m *TaxiMasterAPI)AddressesSearch(query string) t.FastAddress {
	return t.FastAddress{}
}

func (m *TaxiMasterAPI)IsConnected() bool {
	return false
}