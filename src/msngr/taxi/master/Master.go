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
var OrderCreatedErrorCodes = map[int]string{
	100:    "Заказ с такими параметрами уже создан",
	101:    "Тариф не найден",
	102:    "Группа экипажа не найдена",
	103:    "Служба ЕДС не найдена",
}


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
		log.Printf("TM GET request error in request")
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
		log.Printf("TM POST request error in request")
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
		log.Println("error at unmarshal ping result")
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
	r := TariffWrapper{}
	result, err := m._get_request("get_tarif_list", map[string]string{}, true)
	if err != nil {
		log.Printf("error at getting tarif list %v", err)
		return r.Data
	}
	err = json.Unmarshal(result, &r)
	if err != nil {
		log.Println("error at unarshaling tarif list data %v, [%s]", err, result)
	}
	return r.Data
}

type CreateOrderAnswer struct {
	OrderId int64 `json:"order_id"`
}
type CreateOrderAnswerWrapper struct {
	TMAPIResponse
	Data CreateOrderAnswer `json:"data"`
}

func (m *TaxiMasterAPI)NewOrder(order t.NewOrder) t.Answer {
	phone := order.Phone
	source := fmt.Sprintf("%v дом: %v", order.Delivery.Street, order.Delivery.House)
	source_time := time.Now().Add(5 * time.Minute).Format("20060102150405")
	dest := fmt.Sprintf("%v дом: %v", order.Destinations[0].Street, order.Destinations[0].House)

	result := t.Answer{}

	res, err := m._post_request(
		"create_order",
		map[string]string{
			"phone":phone,
			"source":source,
			"source_time":source_time,
			"dest":dest,
		},
		true,
	)
	if err != nil {
		log.Printf("Error at creating TM order, %v" % err)
		return result
	}
	coaw := CreateOrderAnswerWrapper{}
	err = json.Unmarshal(res, &coaw)
	if err != nil {
		log.Printf("Error at unmarshalling TM order data %v [%s]", err, res)
	}
	if message, ok := OrderCreatedErrorCodes[coaw.Code]; ok {
		result.IsSuccess = false
		result.Message = message
		return result
	}

	result.IsSuccess = true
	result.Content.Id = coaw.Data.OrderId
	return result
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
	return []t.CarInfo{}
}

func (m *TaxiMasterAPI)AddressesSearch(query string) t.AddressPackage {
	return t.AddressPackage{}
}

func (m *TaxiMasterAPI)IsConnected() bool {
	return false
}