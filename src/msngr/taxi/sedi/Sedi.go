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
	"sync"

	t "msngr/taxi"
	s "msngr/taxi/set"
	cfg "msngr/configuration"
	g "msngr/taxi/geo"

	"msngr"
	"msngr/utils"
)

const (
	SEDI = "sedi"
	MAX_TRY_COUNT = 5

	HOST_POSTFIX = "/handlers/sedi/api.ashx"
	AUTOCOMPLETE_POSTFIX = "/handlers/autocomplete.ashx"

	CAN_NOT_IMPLY_TARIFF = "Тариф не задан."

	CHECK = "проверьте его"
	CAN_NOT_ESTIMATE_COST = "Предложите свою стоимость за поездку"

	ORDER_INFO_DATE_FORMAT = time.RFC3339
	ORDER_REFRESH_TIME = 11.0
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
	Host                  string
	apikey                string
	appkey                string
	userkey               string

	cookies               []*http.Cookie

	Info                  *SediLoginInfo
	SaleKeyword           string
	GeoOrbit              cfg.TaxiGeoOrbit
	City                  string
	//some fields for hacking work
	userTariffs           map[string]*SediTariff
	userOrders            map[int64]*SediOrderInfo
	lastSuccessUserOrders map[string]*t.NewOrderInfo
	cars                  map[int64]t.CarInfo
	addressKeys           s.Set

	//for taxi watching with not intersection
	lastOrderResponse     *SediOrdersResponse
	lastOrderRequestTime  time.Time

	l                     *sync.Mutex
}

func NewSediAPI(params cfg.TaxiApiParams) *SediAPI {
	s := SediAPI{
		Host:params.Data.Host,
		SaleKeyword:params.Data.SaleKeyword,
		City:params.Data.City,
		GeoOrbit:params.GeoOrbit,
		apikey:params.Data.ApiKey,
		appkey:params.Data.AppKey,

		//todo it must be more persistable...
		lastSuccessUserOrders:map[string]*t.NewOrderInfo{},
		userTariffs:map[string]*SediTariff{},
		userOrders:map[int64]*SediOrderInfo{},
		cars:map[int64]t.CarInfo{},

		addressKeys: s.NewSet(),

		l:&sync.Mutex{},
	}
	go login(&s, params.Data.Name, params.Data.Phone)
	s.cars[-1] = t.CarInfo{}
	return &s
}

func (s *SediAPI) getUserOrders() map[int64]*SediOrderInfo {
	s.l.Lock()
	defer s.l.Unlock()
	return s.userOrders
}

func (s *SediAPI) updateUserOrders(orders []SediOrderInfo) {
	s.l.Lock()
	for _, order := range orders {
		s.userOrders[order.OrderId] = &order
	}
	s.l.Unlock()
}

func (s *SediAPI) deleteUserOrder(order_id int64) {
	s.l.Lock()
	delete(s.userOrders, order_id)
	s.l.Unlock()
}

func (s *SediAPI) setCar(id int64, model, number, color string) {
	s.l.Lock()
	s.cars[id] = t.CarInfo{Model:model, Number:number, ID:id, Color:color}
	s.l.Unlock()
}

func (s *SediAPI) setLastOrderResponse(sor *SediOrdersResponse) {
	s.l.Lock()
	s.lastOrderResponse = sor
	s.lastOrderRequestTime = time.Now()
	s.l.Unlock()
}


func (s *SediAPI) getLastOrdersInfo() (*SediOrdersResponse, time.Time) {
	s.l.Lock()
	lr, lt := s.lastOrderResponse, s.lastOrderRequestTime
	s.l.Unlock()
	return lr, lt
}

func (s *SediAPI) String() string {
	s.l.Lock()
	res := fmt.Sprintf("\nSEDI API: \n\tHost:%v, \n\tapi_key = %v, app_key = %v, user_key = %v \ninfo:%v \nUserTariffs: %v\nLastOrderResponseTime:%v, LastOrderResponse:%v\nCars:%v\nAdressKeys:%+v",
		s.Host,
		s.apikey, s.appkey, s.userkey,
		s.Info,
		s.userTariffs,
		s.lastOrderRequestTime, s.lastOrderResponse,
		s.cars,
		s.addressKeys,
	)
	s.l.Unlock()
	return res
}

func prepareResponse(input []byte) []byte {
	reg := regexp.MustCompile("new Date\\((?P<time_stamp>\\d+)\\)")
	output := reg.ReplaceAll(input, []byte("$time_stamp.0"))
	//	reg = regexp.MustCompile("\\s")
	//	output = reg.ReplaceAll(output, []byte(""))
	return output
}

func login(s *SediAPI, name, phone string) {
	for {
		_, err := s.AuthorizeCustomer(name, phone)
		if err == nil {
			break
		} else {
			log.Printf("SEDI Will try to reconnect to API server...")
		}
		time.Sleep(5 * time.Second)
	}
	profile, _ := s.GetProfile()
	s.Info = profile
}


type SediResponse struct {
	Success bool `json:"Success"`
	Message string `json:"Message"`
}

func (s *SediAPI) IsConnected() bool {
	s.l.Lock()
	res := s.userkey != ""
	s.l.Unlock()
	return res
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
		s.updateCookies(res.Cookies())

		body, err := ioutil.ReadAll(res.Body)
		body = prepareResponse(body)
		return body, err
	}
	return []byte{}, errors.New(fmt.Sprintf("Error at do request... %#v", req))
}

func (s *SediAPI) getRequest(q string, params map[string]string) ([]byte, error) {
	s.l.Lock()
	defer s.l.Unlock()

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
	res, err := s.getRequest("get_enums", map[string]string{})
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

/*Send SMS to phone and return state of user in Sedi system
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
	res, err := s.getRequest("get_activation_key", map[string]string{"phone":phone})
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
	res, err := s.getRequest(q, prms)

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
	if err != nil {
		log.Printf("SEDI can not auth :( For name:%v, phone:%v", name, phone)
		return nil, err
	}
	if resp_info.Success {
		for _, cookie := range s.cookies {
			if cookie.Name == "auth" {
				s.l.Lock()
				s.userkey = cookie.Value
				s.l.Unlock()
				break
			}
		}
		return &resp_info.LoginInfo, err
	}
	return nil, err
}

func GeoToString(lat, lon float64) (string, string) {
	f := func(in float64) string {
		return strconv.FormatFloat(in, 'f', 6, 64)
	}
	return f(lat), f(lon)
}

func (s *SediAPI) prepareOrderParams(order t.NewOrderInfo) map[string]string {
	params := map[string]string{
		"street0":order.Delivery.Street,
		"house0":order.Delivery.House,
		"city0":utils.FirstOf(order.Delivery.City, s.City).(string),
	}
	if order.Delivery.IdAddress != "" {
		params["addrid0"] = order.Delivery.IdAddress
	}

	for i, destination := range order.Destinations {
		if destination == nil {
			continue
		}
		address_number := i + 1
		params[fmt.Sprintf("street%v", address_number)] = destination.Street
		params[fmt.Sprintf("house%v", address_number)] = destination.House
		params[fmt.Sprintf("city%v", address_number)] = utils.FirstOf(destination.City, s.City).(string)

		if destination.IdAddress != "" {
			params[fmt.Sprintf("addrid%v", address_number)] = destination.IdAddress
		}
		if destination.Lat != 0.0 && destination.Lon != 0.0 {
			params[fmt.Sprintf("lat%v", address_number)], params[fmt.Sprintf("lon%v", address_number)] = GeoToString(destination.Lat, destination.Lon)
		}
	}

	if order.DeliveryTime != "" {
		t, _ := time.Parse("2006-01-02 15:04:05", order.DeliveryTime)
		params["date"] = "new Date(" + fmt.Sprintf("%v", t) + ")"
		params["ordertype"] = "preliminary"
	} else {
		params["date"] = "new Date(" + fmt.Sprintf("%v", time.Now().Unix() + 60 * (1 + int64(order.DeliveryMinutes))) + ")"
		params["ordertype"] = "rush"
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

func (s *SediAPI) getLastSuccessOrder(phone string) (*t.NewOrderInfo, bool) {
	s.l.Lock()
	res, ok := s.lastSuccessUserOrders[phone]
	s.l.Unlock()
	return res, ok
}

func (s *SediAPI) getUserTariff(phone string) (*SediTariff, bool) {
	s.l.Lock()
	res, ok := s.userTariffs[phone];
	s.l.Unlock()
	return res, ok
}
var CS_REGEXP = regexp.MustCompilePOSIX("[\\+\\(](.*)")

func getNormalStreetName(street string) string {
	return CS_REGEXP.ReplaceAllString(street, "")
}

func ensureOrderDelivery(order *t.NewOrderInfo) t.NewOrderInfo {
	/**
	Устанавливаем названия улиц для места отправления. Нельзя удалять слово улица проспект переулок и прочую ересь, но нужно удалить "+" и все что после него
	 */
	new_order := *order
	new_order.Delivery.Street = getNormalStreetName(order.Delivery.Street)
	new_order.Delivery.IdAddress = ""
	return new_order
}

func ensureOrderDestinations(order *t.NewOrderInfo) t.NewOrderInfo {
	/**
	Устанавливаем названия улиц для места прибытия. Нельзя удалять слово улица проспект переулок и прочую ересь, но нужно удалить "+" и все что после него
	 */
	new_order := *order
	new_order.Destinations = make([]*t.Destination, len(order.Destinations))
	for i, p_dest := range order.Destinations {
		new_order.Destinations[i] = p_dest
		new_order.Destinations[i].Street = getNormalStreetName(p_dest.Street)
		new_order.Destinations[i].IdAddress = ""
	}
	return new_order
}


func (s *SediAPI)NewOrder(order t.NewOrderInfo) t.Answer {
	log.Printf("SEDI NEW ORDER. Input new order info details is here: \nFrom: %v %v %v\nTo: %v %v %v",
		order.Delivery.House, order.Delivery.Street, order.Delivery.City,
		order.Destinations[0].House, order.Destinations[0].Street, order.Destinations[0].City,
	)

	_order := order
	params := s.prepareOrderParams(_order)

	if (order.Delivery.Lat == 0.0 || order.Delivery.Lon == 0.0) {
		/**
		Ебучий седи. То есть либо ты даешь ему ид адреса, либо ебись с названиями.
		Вообще исходя из того что мы в любом случае перед этим методом получаем инфу из автокомплита, то алгоритм такой:
	 	*/
		//Извлекаем название улицы. Нельзя удалять слово улица проспект переулок и прочую ересь, но нужно удалить "+" либо "(" и все что после него
		order_with_string_delivery := ensureOrderDelivery(&_order)
		order_with_string_delivery_and_dest := ensureOrderDestinations(&order_with_string_delivery)
		//Пробуем создать заказ с обновленной улицей.
		log.Printf("SEDI NEW ORDER! try street like...\n%v", order_with_string_delivery_and_dest)
		se_cost, _ := s.CalcOrderCost(order_with_string_delivery_and_dest)
		if se_cost == -1 {
			log.Printf("SEDI NEW ORDER! not dest and del:( Maybe only del?\n%v", order_with_string_delivery)
			se_cost, _ = s.CalcOrderCost(order_with_string_delivery)
			if se_cost == -1 {
				order_with_string_dest := ensureOrderDestinations(&_order)
				log.Printf("SEDI NEW ORDER! not del:( May be only dest?\n%v", order_with_string_dest)
				se_cost, _ = s.CalcOrderCost(order_with_string_dest)
			}
		}
		//Если стоимость рассчиталась то заказ валиден и должен быть и последний заказец с самыми валидными данными.
		if __order, ok := s.getLastSuccessOrder(order.Phone); ok && se_cost != -1 {
			//вот его-то и берем
			_order = *__order
			log.Printf("SEDI NEW ORDER! USING STREET EQUALS ORDER! \n%v\n\t%v", _order, __order)
			params = s.prepareOrderParams(_order) //и параметрый нахуй все по новой...
		} else {
			//иначе видимо был поиск по объектам и идем по идентификаторам адресов которые были в
			//начальном ордере и отправляем.
			log.Printf("SEDI NEW ORDER! Can not found last success order/ Then i think that orders del/dest is in map objects \n%v", _order)
		}
	} else {
		//ну а если координаты то вообще на все поебать.
		params["lat0"], params["lon0"] = GeoToString(order.Delivery.Lat, order.Delivery.Lon)
	}

	log.Printf("SEDI NEW ORDER FINAL ORDER INFORMATION \nFROM: Street: %v, House: %v, City: %v, Id: %v;\nTO: Street: %v, House: %v, City: %v, Id: %v;",
		_order.Delivery.Street, _order.Delivery.House, _order.Delivery.City, _order.Delivery.IdAddress,
		_order.Destinations[0].Street, _order.Destinations[0].House, _order.Destinations[0].City, _order.Destinations[0].IdAddress,
	)

	//ебаные тарифы... тарифы должны быть сохранены после успешных калькуляций
	order_tariff := &SediTariff{}
	if tariff, ok := s.getUserTariff(_order.Phone); ok {
		order_tariff = tariff
	} else { //и если они не сохранены, то калькулируем чтобы сохранить
		if cost, message := s.CalcOrderCost(_order); cost == -1 {
			return t.Answer{IsSuccess:false, Message:message}
		}
		order_tariff, _ = s.getUserTariff(_order.Phone)
	}
	params["tariff"] = fmt.Sprintf("%v", order_tariff.TariffId)
	params["cost"] = fmt.Sprintf("%v", order_tariff.CostFull)
	params["name"] = _order.Phone
	params["phone"] = _order.Phone

	res, err := s.getRequest("add_order", params)
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
			log.Printf("SEDI ... Interested is why it was exeuted this string...(566)")
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

func (s SediAddress) String() string {
	return fmt.Sprintf("\n\tSEDI Address [%v]: \n\t\tCity: %v, Street: %v, House: %v, PostalCode: %v, LocalityName: %v, \n\t\tGeo point:%v\n", s.ID,
		s.CityName, s.StreetName, s.HouseNumber, s.PostalCode, s.LocalityName, s.GeoPoint)
}

type SediCalcResponse struct {
	SediResponse
	Tariffs   []SediTariff  `json:"Tariffs"`
	Addresses []SediAddress `json:"Addresses"`
	Duration  int `json:"Duration"`
	Distance  float32 `json:"Distance"`

}

func (csr SediCalcResponse) String() string {
	return fmt.Sprintf("SEDI Calc order response. Ok? %v, %s" +
	"\n\tDuration: %v, Distance: %v, TariffsCount: %v, AdressesCount: %v" +
	"\n\tTariffs:%+v" +
	"\n\tAddresses:%+v",
		csr.Success, csr.Message,
		csr.Duration, csr.Distance, len(csr.Tariffs), len(csr.Addresses),
		csr.Tariffs, csr.Addresses)
}

func (csr SediCalcResponse) GetMinCost() (int, *SediTariff) {
	/*
	Возвращат минимальную стоимость (либо та которая имеется с признаком самамя минимальная либо
	просто минимальную из предложенных)
			*/
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

func checkAddressEquals(address_one SediAddress, house, street string) bool {
	/*
	Сравниваем оставляя только введеное название улиц. Дома лишь лоувятся
	 */
	pre_str := func(street string) string {
		return strings.TrimSpace(g.CC_REGEXP.ReplaceAllString(strings.ToLower(street), ""))
	}
	pre_house := func(house string) string {
		return strings.ToLower(strings.TrimSpace(house))
	}
	st1, st2 := pre_str(street), pre_str(address_one.StreetName)
	h1, h2 := pre_house(address_one.HouseNumber), pre_house(house)

	result := st1 == st2 && h1 == h2

	log.Printf("SEDI CHECKING ADDRESSES EQUALS : %v  == ? %v Street: %v, House: %v", address_one, result, street, house)
	return result
}

func checkAndUpdateOrderAddress(res []byte, order *t.NewOrderInfo) (*t.NewOrderInfo, error) {
	/**
	Проверка и обновление адресов создаваемого заказа.
	Так как седи может отвечать ошибкой и предлагать варианты адресов то мы просматриваем
	то что пришло от седи и то что уже есть.
	В итоге возвращается информация о заказе с проставленными идентификатоами адресов отправления и прибытия которые
	которые дало седи или просто тот же самый ордер если в ответе от седи нет инфы об адресах или эта информация не равнозначна
	 */
	response_object := SediGeoCodingResult{}
	err := json.Unmarshal(res, &response_object)
	if err != nil {
		log.Printf("SEDI CHECK ORDER ADDRESS UNMARSHALL ERROR %v \n%s", err, res)
		return nil, err
	}

	for _, resp_address := range response_object.Addresses {
		if checkAddressEquals(resp_address, order.Delivery.House, order.Delivery.Street) {
			address_id_str := strconv.FormatInt(resp_address.ID, 10)
			order.Delivery.IdAddress = address_id_str
			continue
		}
		for i, dest_address := range order.Destinations {
			if checkAddressEquals(resp_address, dest_address.House, dest_address.Street) {
				address_id_str := strconv.FormatInt(resp_address.ID, 10)
				order.Destinations[i].IdAddress = address_id_str
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
	if order.Delivery.Lat != 0.0 && order.Delivery.Lon != 0.0 {
		params["lat0"], params["lon0"] = GeoToString(order.Delivery.Lat, order.Delivery.Lon)
	}
	log.Printf("Request cost:\n%v", order)
	res, err := s.getRequest("Calccost", params)
	response_result := SediCalcResponse{}
	err = json.Unmarshal(res, &response_result)
	if err != nil {
		return -1, fmt.Sprintf("SEDI Calc order Error unmarshal: %+v \nFor response: %s", err, res)
	}
	log.Printf("Response costs:\n%v\n>>>\n%v", order, response_result)
	if !response_result.Success {
		if isCanCheck(response_result.Message) {
			updated_order, err := checkAndUpdateOrderAddress(res, &order)
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
		s.l.Lock()
		s.userTariffs[order.Phone] = tariff
		s.lastSuccessUserOrders[order.Phone] = &order
		s.l.Unlock()
		return cost, "OK"
	}
	return -1, CAN_NOT_IMPLY_TARIFF
}

func (s *SediAPI)CancelOrder(order_id int64) (bool, string) {
	res, err := s.getRequest("cancel_order", map[string]string{"orderid":strconv.FormatInt(order_id, 10)})
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
type SediProperty struct {
	Key   string `json:"Key"`
	Type  string `json:"Type"`
	Value struct {
			  Id   int64 `json:"ID"`
			  Name string `json:"Name"`
		  } `json:"Value"`
	Id    int64 `json:"ID"`
	Name  string `json:"Name"`
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
			Props  *struct {
				Properties []SediProperty `json:"Properties"`
			}
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

func (sor SediOrdersResponse) IsContains(order_id int64) bool {
	for _, order := range sor.Orders {
		if order.OrderId == order_id {
			return true
		}
	}
	return false
}

func (s *SediAPI) toInternalOrders(sor *SediOrdersResponse) []t.Order {
	result := []t.Order{}
	for _, order := range sor.Orders {
		order_state, ok := STATES_MAPPING[order.Status.Id]
		if !ok {
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
			car := order.Driver.Car
			id_car := car.Id
			int_order.IDCar = id_car

			var car_color string
			if car.Props != nil {
				for _, prop := range car.Props.Properties {
					if prop.Key == "color" {
						car_color = prop.Value.Name
						break
					}
				}
			}
			s.setCar(id_car, car.Name, car.Number, car_color)
		}
		result = append(result, int_order)
	}
	return result
}

func (s *SediAPI)Orders() []t.Order {
	result := []t.Order{}
	response := &SediOrdersResponse{}
	lr, lt := s.getLastOrdersInfo()
	if time.Now().Sub(lt).Seconds() < ORDER_REFRESH_TIME && lr != nil {
		response = lr
	} else {
		res, err := s.getRequest("get_orders", map[string]string{})
		if err != nil {
			log.Printf("SEDI ORDERS INFO ERROR: %v", err)
			return result
		}
		err = json.Unmarshal(res, response)
		if err != nil {
			log.Printf("SEDI ORDERS INFO UNMARSHALL ERROR %v, \nres %s", err, res)
			return result
		}
		if !response.Success {
			log.Printf("SEDI ORDERS INFO ERROR %v", response.Message)
			return result
		}
		s.updateUserOrders(response.Orders)
		order_ids := []string{}
		for order_id, _ := range s.getUserOrders() {
			if !response.IsContains(order_id) {
				order_ids = append(order_ids, strconv.FormatInt(order_id, 10))
				s.deleteUserOrder(order_id)
			}
		}
		if len(order_ids) > 0 {
			order_ids_param := strings.Join(order_ids, ",")
			log.Printf("SEDI getting additional info for orders: %+v", order_ids_param)
			_, lt = s.getLastOrdersInfo()

			wait_time := ORDER_REFRESH_TIME / 2
			log.Printf("SEDI ORDERS WILL SLEEP before get orders response time is arrive...%v", wait_time)
			time.Sleep(time.Duration(wait_time) * time.Second)

			add_res, err := s.getRequest("get_orders", map[string]string{"orderids":order_ids_param})
			if err != nil {
				log.Printf("SEDI ORDERS ADDITIONAL INFO ERROR %v", err)
				return s.toInternalOrders(response)
			}
			add_response := &SediOrdersResponse{}
			err = json.Unmarshal(add_res, &add_response)
			if err != nil {
				log.Printf("SEDi ORDERS ADDITIONAL INFO UNMARSHAL ERROR: %v\n%s", err, add_res)
				return s.toInternalOrders(response)
			}
			log.Printf("SEDI ORDERS EXTEND ORDERS %+v", add_response.Orders)
			response.Orders = append(response.Orders, add_response.Orders...)
		}
		s.setLastOrderResponse(response)
	}
	return s.toInternalOrders(response)
}
func (s *SediAPI)Feedback(f t.Feedback) (bool, string) {
	res, err := s.getRequest("set_rating", map[string]string{
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
	s.l.Lock()
	cars := s.cars
	s.l.Unlock()
	result := make([]t.CarInfo, len(cars), len(cars))
	idx := 0
	for _, car_info := range cars {
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
