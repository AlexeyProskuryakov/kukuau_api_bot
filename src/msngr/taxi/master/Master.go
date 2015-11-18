package master


import (
	t "msngr/taxi"
	d "msngr/db"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"log"

	"time"
	"io/ioutil"
	"encoding/json"
	"errors"
	"gopkg.in/mgo.v2/bson"
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

var OrderStateMap = map[string]int{
	"new_order":t.ORDER_CREATED,
	"driver_assigned":t.ORDER_ASSIGNED,
	"car_at_place":t.ORDER_CLIENT_WAIT,
	"client_inside":t.ORDER_IN_PROCESS,
	"finished":t.ORDER_PAYED,
	"aborted":t.ORDER_CANCELED,
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
	Source      string
	Storage     *d.DbHandlerMixin
}

func NewTaxiMasterAPI(connString, bearerToken, source string, storage *d.DbHandlerMixin) (*TaxiMasterAPI) {
	result := TaxiMasterAPI{ConnString:connString, BearerToken:bearerToken, Source:source, Storage:storage}
	return &result
}

type TMAPIResponse struct {
	Code        int `json:"code"`
	Description string `json:"descr"`
	Data        map[string]interface{} `json:"data"`
}

func (resp TMAPIResponse) Check() (bool, string) {
	if message, ok := ErrorsMap[resp.Code]; ok {
		return false, message
	}
	return true, ""
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

func (m *TaxiMasterAPI)NewOrder(order t.NewOrderInfo) t.Answer {
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
		log.Printf("Error at creating TM order, %v", err)
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

	if ok, message := coaw.Check(); ok {
		result.IsSuccess = true
		result.Content.Id = coaw.Data.OrderId
	} else {
		result.IsSuccess = false
		result.Message = message
	}
	return result
}

var ChangeOrderResultErrorCodes = map[int]string{
	100:    "Не найден заказ.",
	101:    "Не найдено состояние заказа.",
	102:    "Изменение состояния не соответствует необходимым условиям.",
}

type TMAPIChangeStateResponse struct {
	TMAPIResponse
	Data struct {
			 OrderId  string `json:"order_id"`
			 NewState string `json:"new_state"`
		 }`json:"data"`
}

//todo You must know real CANCEL ORDER state
const TM_API_CANCEL_ORDER = 9

func (m *TaxiMasterAPI)CancelOrder(order_id int64) (bool, string) {

	res, err := m._post_request("change_order_state", map[string]string{"ORDER_ID":order_id, "NEW_STATE":TM_API_CANCEL_ORDER}, true)
	if err != nil {
		log.Printf("Error at request to change order state to cancel: %v (order_id = %v)", err, order_id)
		return false, ""
	}
	response := TMAPIChangeStateResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		log.Printf("Error at unmarshal response to cancel order (%s) %v ", res, err)
	}
	if ok, message := response.Check(); !ok {
		return false, message
	}
	if message, ok := ChangeOrderResultErrorCodes[response.Code]; ok {
		return false, message
	}
	if response.Data.OrderId == order_id && response.Data.NewState == TM_API_CANCEL_ORDER {
		return true, ""
	}
	return false, ""
}

func (m *TaxiMasterAPI)CalcOrderCost(order t.NewOrderInfo) (int, string) {
	//todo You must know tariff_id for calculate order cost
	return 0, ""
}

type TMAPIOrder struct {
	OrderId         int64 `json:"order_id"`
	StateId         int `json:"state_id"`
	StateKind       string `json:"state_kind"`
	CrewId          int64 `json:"crew_id"`
	DriverId        int64 `json:"driver_id"`
	CarId           int64 `json:"car_id"`
	StartTime       string `json:"start_time"`
	SourceTime      string `json:"source_time"`
	FinishTime      string `json:"finish_time"`

	CarMark         string `json:"car_mark"`
	CarModel        string `json:"car_model"`
	CarColor        string `json:"car_color"`
	CarNumber       string `json:"car_number"`

	CrewCoordinates struct {
						Lat float64 `json:"lat"`
						Lon float64 `json:"lon"`
					}`json:"crew_coordinates"`
}

type TMAPIOrderResponseWrapper struct {
	TMAPIResponse
	Data TMAPIOrder `json:"data"`
}

func (m *TaxiMasterAPI)Orders() []t.Order {
	result := []t.Order{}
	//here we get persisted orders with our source and not ended state
	orders, err := m.Storage.Orders.GetOrders(bson.M{"source":m.Source, "order_state":bson.M{"$nin":[]int64{t.ORDER_PAYED, t.ORDER_CANCELED}}})
	if err != nil {
		log.Printf("Error at retrieving current order ids: %v", err)
		return result
	}
	//and for each persisted order retrieve his info from external source
	for _, order := range orders {
		res, err := m._get_request("get_order_state", map[string]string{"order_id":order.OrderId}, true)
		if err != nil {
			log.Printf("Error at getting info from API by order id [%v]: %v", order.OrderId, err)
			continue
		}
		res_container := TMAPIOrderResponseWrapper{}
		err = json.Unmarshal(res, &res_container)
		if err != nil {
			log.Printf("Error at unmarshalling order response for order id [%v]: %v", order.OrderId, err)
			continue
		}
		if ok, message := res_container.Check(); !ok {
			log.Printf("Error at response result: %s", message)
			continue
		}
		response_order := res_container.Data
		result_order := t.Order{ID:response_order.OrderId}
		if state, ok := OrderStateMap[response_order.StateKind]; ok {
			result_order.State = state
		}else {
			log.Printf("Not imply state of response order [%v] state = %v", order.OrderId, res_container.Data.StateKind)
			continue
		}
		result_order.IDCar = response_order.CarId
		arrival_time, err := time.Parse("20060102150405", response_order.SourceTime)
		if err != nil {
			result_order.TimeArrival = &arrival_time
		}
		result = append(result, result_order)
	}
	return result
}

func (m *TaxiMasterAPI)Feedback(f t.Feedback) (ok bool, message string) {
	params := map[string]string{"phone":f.Phone, "rating":f.Rating, "text":f.FeedBackText, "order_id":f.IdOrder}
	res, err := m._post_request("save_client_feed_back", params, false)
	if err != nil {
		log.Printf("Error at sended request at fedback with params %+v", params)
		return ok, message
	}
	tm_res := TMAPIResponse{}
	err = json.Unmarshal(res, &tm_res)
	if err != nil {
		log.Printf("Error at unmarshaling feedback %v, [%s]", err, res)
	}
	ok, message = tm_res.Check();
	return ok, message
}


type TMAPICarInfo struct {
	CarId   int64 `json:"car_id"`
	Mark    string `json:"mark"`
	Model   string `json:"model"`
	GNumber string `json:"gos_number"`
	Color   string `json:"color"`
}
type TMAPICarInfoWrapper struct {
	TMAPIResponse
	Data struct {
			 CarsInfo []TMAPICarInfo `json:"cars_info"`
		 }`json:"data"`
}

func (m *TaxiMasterAPI)GetCarsInfo() []t.CarInfo {
	/**
	{
  "cars_info":[
    {
      "car_id":1,
      "code":"111",
      "name":"CAR_1",
      "gos_number":"111111",
      "color":"COLOR_1",
      "mark":"MARK_1",
      "model":"MODEL_1",
      "short_name":"SHORT_NAME_1",
      "production_year":2000,
      "is_locked":false,
      "order_params":[1,2]
    },

    get_cars_info
	 */
	result := []t.CarInfo{}
	res, err := m._get_request("get_cars_info", map[string]string{}, true)
	if err != nil {
		log.Printf("Error at requesting car info: %v", err)
		return result
	}
	response := TMAPICarInfoWrapper{}
	err = json.Unmarshal(res, &response)
	if ok, message := response.Check(); !ok {
		log.Printf("Error at response result: %s", message)
		return result
	}
	for _, resp_car_info := range response.Data.CarsInfo {
		elem := t.CarInfo{Color:resp_car_info.Color, ID:resp_car_info.CarId, Model:resp_car_info.Model, Number:resp_car_info.GNumber}
		result = append(result, elem)
	}
	return result
}

func (m *TaxiMasterAPI)AddressesSearch(query string) t.AddressPackage {
	return t.AddressPackage{}
}

func (m *TaxiMasterAPI)IsConnected() bool {
	return false
}