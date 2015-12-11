package sedi
import (
	"encoding/json"
	"errors"

	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	t "msngr/taxi"
	"msngr/configuration"
	"msngr"

)

const (
	MAX_TRY_COUNT = 5

	HOST_POSTFIX = "/handlers/sedi/api.ashx"
	AUTOCOMPLETE_POSTFIX = "/handlers/autocomplete.ashx"

	CAN_NOT_IMPLY_TARIFF = "Тариф не задан."

	CHECK = "проверьте его"
	CAN_NOT_ESTIMATE_COST = "Предложите свою стоимость за поездку"

	ORDER_INFO_DATE_FORMAT = time.RFC3339
	ORDER_REFRESH_TIME = 12
)

func isCanCheck(message string) bool {
	return strings.Contains(message, CHECK)
}

func isCanNotEstimateCost(message string) bool {
	return strings.Contains(message, CAN_NOT_ESTIMATE_COST)
}

type SediTariff struct {
	CostFull      int `json:"CostFull"`
	IsMinimumCost bool `json:"IsMinimumCost"`
	TariffId      int64 `json:"ID"`
}

type SediAPI struct {
	Host                 string
	apikey               string
	appkey               string
	userkey              string

	cookies              []*http.Cookie

	Info                 *SediLoginInfo
	City                 string
	SaleKeyword          string

	//some fields for hacking work
	UserTariffs          map[string]*SediTariff

	LastOrderResponse    *SediOrdersResponse
	LastOrderRequestTime time.Time

	Cars                 map[int64]t.CarInfo
}

func (s *SediAPI) String() string {
	return fmt.Sprintf("\nSEDI API: \n\tHost:%v, \n\tapi_key = %v, app_key = %v, user_key = %v \ninfo:%v \nUserTariffs: %v\nLastOrderResponseTime:%v, LastOrderResponse:%v\nCars:%v",
		s.Host,
		s.apikey, s.appkey, s.userkey,
		s.Info,
		s.UserTariffs,
		s.LastOrderRequestTime, s.LastOrderResponse,
		s.Cars,
	)
}
func prepareResponse(input []byte) []byte {
	reg := regexp.MustCompile("new Date\\((?P<time_stamp>\\d+)\\)")
	output := reg.ReplaceAll(input, []byte("$time_stamp.0"))
	//	reg = regexp.MustCompile("\\s")
	//	output = reg.ReplaceAll(output, []byte(""))
	return output
}

func NewSediAPI(api_data *configuration.ApiData) *SediAPI {
	s := SediAPI{
		Host:api_data.Host,
		City:api_data.City,
		SaleKeyword:api_data.SaleKeyword,

		apikey:api_data.ApiKey,
		appkey:api_data.AppKey,

		UserTariffs:map[string]*SediTariff{},
	}
	s.AuthorizeCustomer(api_data.Name, api_data.Phone)
	s.Info, _ = s.GetProfile()
	return &s
}

type SediResponse struct {
	Success bool `json:"Success"`
	Message string `json:"Message"`
}

func (s *SediAPI) updateCookies(cookies []*http.Cookie) {
	for _, new_cookie := range cookies {
		changed := false
		for i, cookie := range s.cookies {
			if new_cookie.Name == cookie.Name {
				s.cookies[i].Value = new_cookie.Value
				changed = true
			}
		}
		if !changed {
			s.cookies = append(s.cookies, new_cookie)
		}
	}
}

func (s *SediAPI) doReq(req *http.Request) ([]byte, error) {
	count := 0
	for {
		for _, cookie := range s.cookies {
			req.AddCookie(cookie)
		}
		client := &http.Client{}
		res, err := client.Do(req)
		if res == nil || err != nil {
			log.Println("SEDI response is: ", res, "; error is:", err, ". I will reconnect and will retrieve data again after 3s.")
			time.Sleep(3 * time.Second)
			count += 1
			if count > MAX_TRY_COUNT {
				return []byte{}, errors.New(fmt.Sprintf("Max tryings count ended and request not proceed... %#v", req))
			}
			continue
		}
		defer res.Body.Close()
		if msngr.DEBUG {
			log.Printf("SEDI COOKIE >>> %v", s.cookies)
			log.Printf("SEDI COOKIE <<< %+v", res.Cookies())
		}
		s.updateCookies(res.Cookies())

		body, err := ioutil.ReadAll(res.Body)
		body = prepareResponse(body)
		return body, err
	}
	return []byte{}, errors.New(fmt.Sprintf("Error at do request... %#v", req))
}

func (s *SediAPI) SimpleGetRequest(q string, params map[string]string) ([]byte, error) {
	if s.apikey == "" || s.appkey == "" {
		panic("Before simple request api key and user key required")
	}
	req, err := http.NewRequest("GET", s.Host + HOST_POSTFIX, nil)
	if err != nil {
		log.Printf("SEDI GET request error in request")
	}
	values := req.URL.Query()
	for k, v := range params {
		values.Add(k, v)
	}
	values.Add("q", q)
	values.Add("apikey", s.apikey)
	values.Add("key", s.appkey)
	values.Add("userkey", s.userkey)
	req.URL.RawQuery = values.Encode()
	result, err := s.doReq(req)
	if msngr.DEBUG {
		log.Printf("SEDI >>> \n%v\n", req.URL)
		log.Printf("SEDI <<< \n%s\n", result)
	}
	resp := SediResponse{}
	json.Unmarshal(result, &resp)

	return result, err
}

type OrderStatus struct {
	Name string `json:"Name"`
	ID   string `json:"ID"`
}

type EnumsResponse struct {
	SediResponse
	OrderStatuses []OrderStatus `json:"OrderStatuses"`
}

func (s *SediAPI) GetOrderStatuses() ([]OrderStatus, error) {
	result := []OrderStatus{}
	res, err := s.SimpleGetRequest("get_enums", map[string]string{})
	if err != nil {
		log.Printf("SA Error at get order statuses")
		return result, err
	}
	response := EnumsResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		log.Printf("SA Error at unmarshal order status response")
		return result, err
	}
	if !response.Success {
		return result, errors.New(response.Message)
	}
	return response.OrderStatuses, nil
}

type ActivationKeyResponse struct {
	SediResponse
	Count uint
}

/*Send SMS to phone and return state of user in Sedi system       \
0 введенный номер не неизвестен, и для регистрации нового
пользователя необходимо будет запросить его имя

1 пользователь найден и для восстановления доступа
достаточно ввести только SMS-код

2 под этим номером зарегистрированы два разных пользователя
(заказчик и сотрудник), при вводе SMS-ключа необходимо
уточнить тип

Но нахуя все это знать читающему этот комментарий?
*/
func (s *SediAPI) GetActivationKey(phone string) (uint, error) {
	res, err := s.SimpleGetRequest("get_activation_key", map[string]string{"phone":phone})
	if err != nil {
		log.Printf("SA Error at get activation key")
		return 0, err
	}
	result := ActivationKeyResponse{}
	err = json.Unmarshal(res, &result)
	if err != nil {
		log.Printf("SA Error at unmarshall response from activation key")
		return 0, err
	}
	if !result.Success {
		log.Printf("SA Error in get activation key response: %v", err)
		return 0, errors.New(result.Message)
	}
	return result.Count, nil
}

type SediLoginInfo struct {
	ID                int64`json:"ID"`
	RegistrationDate  float64 `json:"RegistrationDate"`
	AccountID         int64 `json:"AccountID"`
	Login             string `json:"Login"`
	Name              string `json:"Name"`
	Nick              string `json:"Nick"`
	AnonymousOrder    bool `json:"AnonymousOrder"`
	IsCustomer        bool `json:"IsCustomer"`
	IsEmployee        bool `json:"IsEmployee"`
	AllowCustomerCost bool `json:"AllowCustomerCost"`
	AllowAuction      bool `json:"AllowAuction"`
	AllowCalc         bool `json:"AllowCalc"`
	Balance           struct {
						  AccountId int `json:"AccountID"`
						  Currency  string `json:"Currency"`
						  Credit    float64 `json:"Credit"`
						  Value     float64 `json:"Value"`
						  Locked    float64 `json:"Locked"`
					  } `json:"Balance"`
	Phones            []struct {
		Number string `json:"Number"`
		Type   string `json:"Type"`
	}`json:"Phones"`

	Affiliate         bool `json:"Affiliate"`
	AffilateOrders    bool `json:"AffilateOrders"`
	Promocode         int64 `json:"Promocode"`
	Owner             struct {
						  ID               int `json:"ID"`
						  Name             string `json:"Name"`
						  Url              string `json:"Url,omitempty"`
						  DispatcherPhones []string `json:"Phones"`
					  } `json:"Owner"`
}

func (l SediLoginInfo) String() string {
	return fmt.Sprintf(" ID: %v, AccountID: %v, Login: %v, Name: %v, Nick: %v" +
	"\n Customer?: %v, Employee?: %v, AllowCustomerCost?: %v, AllowAuction?: %v, AllowCalc?: %v" +
	"\n Balance: %+v\n Phones: %+v\n Affiliate?: %v, AffilateOrders?: %v, Promocode: %v\n Owner: %+v",
		l.ID, l.AccountID, l.Login, l.Name, l.Nick,
		l.IsCustomer, l.IsEmployee, l.AllowCustomerCost, l.AllowAuction, l.AllowCalc,
		l.Balance, l.Phones, l.Affiliate, l.AffilateOrders, l.Promocode, l.Owner)
}

type SediLoginInfoResponse struct {
	SediResponse
	LoginInfo SediLoginInfo `json:"LoginInfo"`
}

func (s *SediAPI) auth(q string, prms map[string]string) (*SediLoginInfoResponse, error) {
	res, err := s.SimpleGetRequest(q, prms)

	if err != nil {
		log.Printf("SA AUTH Error at %v [%v] %v", q, prms, err)
		return nil, err
	}
	result := SediLoginInfoResponse{}
	err = json.Unmarshal(res, &result)
	if err != nil {
		log.Printf("SA AUTH Error at unmarshall response from %v [%v] key %v\n%s", q, prms, err, res)
		return nil, err
	}
	if !result.Success {
		log.Printf("SA AUTH Error in %v [%v] is: %v, message: %v", q, prms, err, result.Message)
		return nil, errors.New(result.Message)
	}
	return &result, nil
}

func (s *SediAPI) ActivationKey(sms_code string) (*SediLoginInfo, error) {
	resp_info, err := s.auth("activation_key", map[string]string{"actkey":sms_code})
	if resp_info != nil {
		return &resp_info.LoginInfo, err
	}
	return nil, err
}

func (s *SediAPI) GetProfile() (*SediLoginInfo, error) {
	resp_info, err := s.auth("get_profile", map[string]string{})
	if resp_info != nil {
		return &resp_info.LoginInfo, err
	}
	return nil, err
}

func (s *SediAPI) AuthorizeCustomer(name, phone string) (*SediLoginInfo, error) {
	resp_info, err := s.auth("authorize_customer", map[string]string{"name":name, "phone":phone})
	if resp_info.Success {
		for _, cookie := range s.cookies {
			if cookie.Name == "auth" {
				s.userkey = cookie.Value
				break
			}
		}
		return &resp_info.LoginInfo, err
	}
	return nil, err
}

func (s *SediAPI) prepareOrderParams(order t.NewOrderInfo) map[string]string {
	params := map[string]string{
		"city0":s.City,
		"street0":order.Delivery.Street,
		"house0":order.Delivery.House,
	}
	if order.Delivery.IdAddress != nil {
		params["addrid0"] = *order.Delivery.IdAddress
	}

	for i, destination := range order.Destinations {
		address_number := i + 1
		params[fmt.Sprintf("city%v", address_number)] = s.City
		params[fmt.Sprintf("street%v", address_number)] = destination.Street
		params[fmt.Sprintf("house%v", address_number)] = destination.House
		if destination.IdAddress != nil {
			params[fmt.Sprintf("addrid%v", address_number)] = *destination.IdAddress
		}
	}

	if order.DeliveryTime != nil {
		t, _ := time.Parse("2006-01-02 15:04:05", *order.DeliveryTime)
		params["date"] = "new Date(" + fmt.Sprintf("%v", t) + ")"
		params["ordertype"] = "preliminary"
	} else {
		params["date"] = "new Date(" + fmt.Sprintf("%v", time.Now().Unix() + 60 * (1 + order.DeliveryMinutes)) + ")"
		params["ordertyep"] = "rush"
	}

	if s.SaleKeyword != "" {
		params["keyword"] = s.SaleKeyword
	}
	return params
}

type SediNewOrderResponse struct {
	SediResponse
	OrderId int64 `json:"ObjectId"`
}


func (s *SediAPI)NewOrder(order t.NewOrderInfo) t.Answer {
	params := s.prepareOrderParams(order)
	order_tariff := &SediTariff{}
	if tariff, ok := s.UserTariffs[order.Phone]; ok {
		order_tariff = tariff
	}else {
		if cost, message := s.CalcOrderCost(order); cost < 0 {
			return t.Answer{IsSuccess:false, Message:message}
		}
		tariff, ok = s.UserTariffs[order.Phone]
		order_tariff = tariff
	}
	params["tariff"] = fmt.Sprintf("%v", order_tariff.TariffId)
	params["cost"] = fmt.Sprintf("%v", order_tariff.CostFull)
	params["name"] = "KUKUAPIBOT"
	params["phone"] = order.Phone

	res, err := s.SimpleGetRequest("add_order", params)
	if err != nil {
		log.Printf("SEDI NEW ORDER ERROR :%v\n params: %+v", err, params)
	}
	new_order_response := SediNewOrderResponse{}
	err = json.Unmarshal(res, &new_order_response)
	if err != nil {
		log.Printf("SEDI ERROR at unmarshal new order response %v\nresponse:[%s]", err, res)
	}
	result := t.Answer{IsSuccess:new_order_response.Success, Message:new_order_response.Message}
	if !new_order_response.Success {
		if isCanCheck(new_order_response.Message) {
			updated_order, err := checkOrderAddress(res, &order)
			if err == nil {
				return s.NewOrder(*updated_order)
			}
		}
		log.Printf("SEDI ERROR at create order %v", new_order_response.Message)
		return result
	}
	result.Content.Cost = order_tariff.CostFull
	result.Content.Id = new_order_response.OrderId
	return result
}

type SediAddress struct {
	ID           int64 `json:"Id"`
	CityName     string `json:"CityName"`
	StreetName   string `json:"StreetName"`
	HouseNumber  string `json:"HouseNumber"`
	PostalCode   string `json:"PostalCode"`
	LocalityName string `json:"LocalityName"`
	GeoPoint     struct {
					 Lat float64 `json:"Latitude"`
					 Lon float64 `json:"Longtitude"`
				 } `json:"GeoPoint"`
}

type SediCalcResponse struct {
	SediResponse
	Tariffs   []SediTariff  `json:"Tariffs"`
	Addresses []SediAddress `json:"Addresses"`
	Duration  int `json:"Duration"`
	Distance  float32 `json:"Distance"`

}

func (csr SediCalcResponse) String() string {
	return fmt.Sprintf("SEDI Calc order response. Ok? %v, %s\n\tDuration: %v, Distance: %v, TariffsCount: %v, AdressesCount: %v\n\tTariffs:%+v\n\tAddresses:%+v", csr.Success, csr.Message, csr.Duration, csr.Distance, len(csr.Tariffs), len(csr.Addresses), csr.Tariffs, csr.Addresses)
}

func (csr SediCalcResponse) GetMinCost() (int, *SediTariff) {
	costs := map[SediTariff]int{}
	for _, tarif := range csr.Tariffs {
		if tarif.IsMinimumCost {
			return tarif.CostFull, &tarif
		}else {
			costs[tarif] = tarif.CostFull
		}
	}
	if len(costs) > 0 {
		min := int(^uint(0) >> 1)
		for _, cost := range costs {
			if cost < min && cost != 0 {
				min = cost
			}
		}
		for tarif, cost := range costs {
			if min == cost {
				return cost, &tarif
			}
		}
	}
	return 0, nil
}
/*
Ебучий костыль для адреса. Иначе никак (по крайней мере на такую голову...).

*/
func checkAddressEquals(address_one SediAddress, house string, street string) bool {
	st1 := strings.TrimSpace(t.CC_REGEXP.ReplaceAllString(street, ""))
	st2 := strings.TrimSpace(t.CC_REGEXP.ReplaceAllString(address_one.StreetName, ""))
	return st1 == st2 && strings.TrimSpace(address_one.HouseNumber) == strings.TrimSpace(house)

}
func checkOrderAddress(res []byte, order *t.NewOrderInfo) (*t.NewOrderInfo, error) {
	response_object := SediGeoCodingResult{}
	err := json.Unmarshal(res, &response_object)
	if err != nil {
		log.Printf("SEDI CHECK ORDER ADDRESS UNMARSHALL ERROR %v \n%s", err, res)
		return nil, err
	}

	for _, resp_address := range response_object.Addresses {
		if checkAddressEquals(resp_address, order.Delivery.House, order.Delivery.Street) {
			address_id_str := strconv.FormatInt(resp_address.ID, 10)
			order.Delivery.IdAddress = &address_id_str
			continue
		}
		for i, dest_address := range order.Destinations {
			if checkAddressEquals(resp_address, dest_address.House, dest_address.Street) {
				address_id_str := strconv.FormatInt(resp_address.ID, 10)
				order.Destinations[i].IdAddress = &address_id_str
				break
			}
		}
	}
	return order, nil
}

func (s *SediAPI)CalcOrderCost(order t.NewOrderInfo) (int, string) {
	if len(order.Destinations) < 1 {
		return 0, "No destinations at order"
	}
	params := s.prepareOrderParams(order)
	res, err := s.SimpleGetRequest("Calccost", params)
	response_result := SediCalcResponse{}
	err = json.Unmarshal(res, &response_result)
	if err != nil {
		return -1, fmt.Sprintf("Error unmarshal: %+v \nFor response: %s", err, res)
	}
	log.Printf("Response costs:\n%v", response_result)
	if !response_result.Success {
		if isCanCheck(response_result.Message) {
			updated_order, err := checkOrderAddress(res, &order)
			if err == nil {
				return s.CalcOrderCost(*updated_order)
			}
		} else if isCanNotEstimateCost(response_result.Message) {
			return -1, response_result.Message
		}

		return -1, fmt.Sprintf("%v", response_result.Message)
	}
	cost, tariff := response_result.GetMinCost()
	if tariff != nil {
		s.UserTariffs[order.Phone] = tariff
		return cost, "OK"
	}
	return -1, CAN_NOT_IMPLY_TARIFF
}

type SediAddressF struct {
	Name string `json:"v"`
	City string `json:"c"`
	Type string `json:"t"`
	Id   int64 `json:"n"`
	Geo  struct {
			 Lat float64 `json:"lat"`
			 Lon float64 `json:"lon"`
		 } `json:"g"`
}

type SediAutocompleteResponse []SediAddressF

func (s SediAutocompleteResponse) ToAddressPackage() t.AddressPackage {
	result := t.AddressPackage{}
	rows := []t.AddressF{}
	for _, s_addr_f := range s {
		rows = append(rows, t.AddressF{ID:s_addr_f.Id, Name:s_addr_f.Name, City:s_addr_f.City})
	}
	result.Rows = &rows
	return result
}

func (s *SediAPI) AddressesSearch(text string) t.AddressPackage {
	result := t.AddressPackage{}

	req, err := http.NewRequest("GET", s.Host + AUTOCOMPLETE_POSTFIX, nil)
//	req, err := http.NewRequest("GET", "http://api.sedi.ru" + AUTOCOMPLETE_POSTFIX, nil)
	if err != nil {
		log.Printf("SEDI GET request error in request")
	}
	values := req.URL.Query()

	values.Add("q", "addr")

	values.Add("apikey", s.apikey)
	values.Add("key", s.appkey)
	values.Add("userkey", s.userkey)

	values.Add("streetobj", text)
	values.Add("city", s.City)

	req.URL.RawQuery = values.Encode()
	log.Printf("SEDI >>> %v\n", req.URL)

	res, err := s.doReq(req)

	log.Printf("SEDI <<< \n%s\n", res)
	if err != nil {
		log.Printf("SEDI AUTOCMPLETE ERROR: %v", err)
		return result
	}
	response_object := SediAutocompleteResponse{}
	err = json.Unmarshal(res, &response_object)
	if err != nil {
		log.Printf("SEDI AUTOCOMPLETE UNMARSHALL ERROR: %v\nres:[%s]", err, res)
	}
	return response_object.ToAddressPackage()
}

type SediGeoCodingResult struct {
	SediResponse
	Addresses []SediAddress `json:"Addresses"`
}

func (sgcr SediGeoCodingResult) toAddressPackage() t.AddressPackage {
	rows := []t.AddressF{}
	for _, address_res := range sgcr.Addresses {
		rows = append(rows, t.AddressF{
			ID:address_res.ID,
			City:address_res.CityName,
			Name:address_res.StreetName,
			PostalCode:address_res.PostalCode,
			HouseNumber:address_res.HouseNumber,
			Coordinates: t.Coordinates{Lat:address_res.GeoPoint.Lat, Lon:address_res.GeoPoint.Lon},
		})
	}
	result := t.AddressPackage{}
	result.Rows = &rows
	return result
}
func (s *SediAPI) GeoCoding(lat, lon float64) t.AddressPackage {
	result := t.AddressPackage{}

	lat_s := strconv.FormatFloat(lat, 'f', 6, 64)
	lon_s := strconv.FormatFloat(lon, 'f', 6, 64)
	res, err := s.SimpleGetRequest("get_address", map[string]string{"lat":lat_s, "lon":lon_s})
	if err != nil {
		log.Printf("SEDI GEOCODING ERROR: %v", err)
		return result
	}
	result_object := SediGeoCodingResult{}
	err = json.Unmarshal(res, &result_object)
	if err != nil {
		log.Printf("SEDI GEOCODING ERROR UNMARSHAL: %v \n[%s]", err, res)
		return result
	}
	return result_object.toAddressPackage()
}

func (s *SediAPI)CancelOrder(order_id int64) (bool, string) {
	res, err := s.SimpleGetRequest("cancel_order", map[string]string{"orderid":strconv.FormatInt(order_id, 10)})
	if err != nil {
		log.Printf("SEDI CANCEL ORDER ERROR %v", err)
		return false, fmt.Sprintf("Ошибка! %v", err)
	}
	response := SediResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		log.Printf("SEDI UNMARSHAL CACNEL ERRROR %v \nresult: %s", err, res)
	}
	return response.Success, response.Message
}
var STATES_MAPPING = map[string]int{
	"unknown":  1,
	"driverwaitcustomer":  4,
	"fromexternalsystemisrefused": 9,
	"waitexecute": 2,
	"doneok": 7,
	"donenodriver": 9,
	"donenoclient": 9,
	"completedinexternalsystem": 9,
	"new": 1,
	"ineditmode": 1,
	"cancelled": 9,
	"donefail": 9,
	"customernotifiedaboutdriverwait": 4,
	"executornotfoundonrbt": 9,
	"executornotfound": 9,
	"waittaxi": 2,
	"waitcomplete": 9,
	"waitconfirminwayfromdriver": 2,
	"trycancel": 9,
	"inway": 3,
	"nearcustomer": 3,
}


type SediOrderInfo struct {
	OrderId  int64 `json:"ID"`

	Status   struct {
				 Name string `json:"Name"`
				 Id   string `json:"ID"`
			 } `json:"Status"`

	Date     float64 `json:"Date"`
	Cost     float64 `json:"Cost"`
	Currency string `json:"Currency"`
	Tariff   struct {
				 Id   int64 `json:"ID"`
				 Name string `json:"Name"`
			 } `json:"Tariff"`

	Driver   *struct {
		Geo   struct {
				  Lat float64 `json:"Lat"`
				  Lon float64 `json:"Lon"`
			  } `json:"Geo"`
		Car   *struct {
			Number string `json:"Number"`
			Id     int64 `json:"ID"`
			Name   string `json:"Name"`
		} `json:"car,omitempty"`
		Phone string `json:"Phone"`
		Id    int64 `json:"ID"`
		Name  string `json:"Name"`
	} `json:"Driver,omitempty"`

	Rating   *struct {
		Comment string `json:"Comment"`
		Rate    int `json:"Rate"`
	} `json:"Rating,omitempty"`

	Route    struct {
				 Length float64 `json:"Length"`
			 } `json:"Rout"`
}
type SediOrdersResponse struct {
	SediResponse
	Orders []SediOrderInfo `json:"Orders"`
}

func (s *SediAPI) toInternalOrders(sor SediOrdersResponse) []t.Order {
	result := []t.Order{}
	for _, order := range sor.Orders {
		order_state, ok := STATES_MAPPING[order.Status.Id]
		if !ok {
			log.Printf("SEDI WARNING for order %v \ncan not recognize state [%v]", order, order.Status)
			continue
		}
		time_delivery := time.Unix(int64(order.Date), 0)
		int_order := t.Order{
			ID:order.OrderId,
			State: order_state,
			Cost: int(order.Cost),
			TimeDelivery: &time_delivery,
		}

		if order.Driver != nil && order.Driver.Car != nil {
			id_car := order.Driver.Car.Id
			int_order.IDCar = id_car
			s.Cars[id_car] = t.CarInfo{Model:order.Driver.Car.Name, Number:order.Driver.Car.Number, ID:id_car}
		}
		result = append(result, int_order)
	}
	return result
}

func (s *SediAPI)Orders() []t.Order {
	result := []t.Order{}
	response := SediOrdersResponse{}

	if time.Now().Sub(s.LastOrderRequestTime).Seconds() > ORDER_REFRESH_TIME && s.LastOrderResponse != nil {
		response = *s.LastOrderResponse
	} else {
		res, err := s.SimpleGetRequest("get_orders", map[string]string{})
		if err != nil {
			log.Printf("SEDI ORDERS INFO ERROR: %v", err)
			return result
		}
		err = json.Unmarshal(res, &response)
		if err != nil {
			log.Printf("SEDI ORDERS INFO UNMARSHALL ERROR %v, \nres %s", err, res)
			return result
		}
	}
	if !response.Success {
		log.Printf("SEDI ORDERS INFO ERROR %v", response.Message)
		return result
	}

	s.LastOrderResponse = &response
	s.LastOrderRequestTime = time.Now()
	return s.toInternalOrders(response)
}
func (s *SediAPI)Feedback(f t.Feedback) (bool, string) {
	res, err := s.SimpleGetRequest("set_rating", map[string]string{
		"orderid":strconv.FormatInt(f.IdOrder, 10),
		"rating":strconv.Itoa(f.Rating),
		"comment":f.FeedBackText})
	if err != nil {
		log.Printf("SEDI FEEDBACK ERROR: %v", err)
		return false, ""
	}
	response := SediResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		log.Printf("SEDI FEEDBACK UNMARSHAL ERROR %v\nres %s", err, res)
	}
	return response.Success, response.Message
}

func (s *SediAPI)GetCarsInfo() []t.CarInfo {
	result := make([]t.CarInfo, len(s.Cars), len(s.Cars))
	idx := 0
	for _, car_info := range s.Cars {
		result[idx] = car_info
		idx++
	}
	return result
}

func (s *SediAPI)WriteDispatcher(message string) (bool, string) {
	log.Printf("SEDI WARN WriteDispatcher IS NOT IMPLEMENT")
	return false, ""
} //написать диспетчеру
func (s *SediAPI)CallbackRequest(phone string) (bool, string) {
	log.Printf("SEDI WARN CallbackRequest IS NOT IMPLEMENT")
	return false, ""
} //запросить обратный звонок
func (s *SediAPI)WhereIt(order_id int64) (bool, string) {
	log.Printf("SEDI WARN WhereIt IS NOT IMPLEMENT")
	return false, ""
}//оповестить что клиент не видит автомобиль
