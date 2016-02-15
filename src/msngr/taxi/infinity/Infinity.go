package infinity

import (
	"encoding/json"
	"errors"

	"io/ioutil"
	"log"
	"net/http"

	"time"
	t "msngr/taxi"
	"fmt"
	"net/url"
	"net"
)

const (
	TRY_COUNT = 10
	INFINITY = "infinity"
)

type InfinityService struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	AvailableToClients bool   `json:"AvailableToClients"`
}
type InfinityServices struct {
	Rows []InfinityService `json:"rows"`
}
type InfinityCarsInfo struct {
	Rows []t.CarInfo `json:"rows"`
}

// infinity - Структура для работы с API infinity.
type infinity struct {
	ConnStrings   []string // Строка подключения к infinity API// default: http://109.202.25.248:8080/WebAPITaxi/
	LoginTime     time.Time
	Cookie        *http.Cookie
	LoginResponse struct {
			      Success   bool  `json:"success"`
			      IDClient  int64 `json:"idClient"`
			      Params    struct {
						ProtocolVersion            int    `json:"ProtocolVersion"`
						RefreshOrdersSeconds       int    `json:"RefreshOrdersSeconds"`
						LoginRegEx                 string `json:"LoginRegEx"`
						MyPhoneRegEx               string `json:"MyPhoneRegEx"`
						OurPhoneDisplay            string `json:"OurPhoneDisplay"`
						OurPhoneNumber             string `json:"OurPhoneNumber"`
						DefaultInfinityServiceID   int64  `json:"DefaultInfinityServiceID"`
						DefaultInfinityServiceName string `json:"DefaultInfinityServiceName"`
						DefaultRegionID            int64  `json:"DefaultRegionID"`
						DefaultRegionName          string `json:"DefaultRegionName"`
						DefaultDistrictID          string `json:"DefaultDistrictID"` // Can be null, so used as string here.
						DefaultDistrictName        string `json:"DefaultDistrictName"`
						DefaultCityID              int64  `json:"DefaultCityID"`
						DefaultCityName            string `json:"DefaultCityName"`
						DefaultPlaceID             string `json:"DefaultPlaceID"`    // Can be null, so used as string here.
						DefaultPlaceName           string `json:"DefaultPlaceName"`
					} `json:"params"`
			      SessionID string `json:"sessionid"`
		      }
	Services      []InfinityServices `json:"InfinityServices"`
	Config        t.TaxiAPIConfig
}

func get_domain(conn_string string) string {
	u, _ := url.Parse(conn_string)
	host, _, _ := net.SplitHostPort(u.Host)
	return host
}

func (i infinity) String() string {
	return fmt.Sprintf("\nInfinity API processing.\nConnection strings:%+v\nLogon?:%v, time:%v client id:%v\nid_service: %v", i.ConnStrings, i.LoginResponse.Success, i.LoginTime, i.LoginResponse.IDClient, i.Config.GetIdService())
}

func _initInfinity(config t.TaxiAPIConfig) *infinity {
	result := &infinity{}
	result.ConnStrings = config.GetConnectionStrings()
	result.Config = config

	logon := result.Login(config.GetLogin(), config.GetPassword())

	if !logon {
		go func() {
			res := result.ReLogin()
			if !res {
				log.Printf("can not connect to infinity %+v :(", result.ConnStrings)
			}
		}()
	}

	return result
}

func GetInfinityAPI(tc t.TaxiAPIConfig) t.TaxiInterface {
	instance := _initInfinity(tc)
	return instance
}

func GetInfinityAddressSupplier(tc t.TaxiAPIConfig) t.AddressSupplier {
	instance := _initInfinity(tc)
	return instance
}

// Login - Авторизация в сервисе infinity. Входные параметры: login:string; password:string.
// Возвращает true, если авторизация прошла успешно, false иначе.
// Устанавливает время авторизации в infinity.LoginTime при успешной авторизации.
func (p *infinity) Login(login, password string) bool {
	p.LoginResponse.Success = false
	body, err := p._request("Login", map[string]string{"l":login, "p":password, "app":"CxTaxiClient"})
	if err != nil {
		log.Printf("error at requst to infinity %v", err)
		return false
	}
	err = json.Unmarshal(body, &p.LoginResponse)
	if err != nil {
		log.Printf("error at unmarshalling json:%q \nerror: %v", string(body), err)
		return false
	}
	log.Printf("[login] self: %+q\n", p)
	if p.LoginResponse.Success {
		log.Println("[login] JSESSIONID: ", p.LoginResponse.SessionID)
		p.Cookie = &http.Cookie{
			Name:   "JSESSIONID",
			Value:  p.LoginResponse.SessionID,
			Path:   "/",
			Domain: get_domain(p.ConnStrings[0]),
		}
		p.LoginTime = time.Now()

		return true
	}
	return false
}

func (p *infinity) IsConnected() bool {
	return p.LoginResponse.Success
}

func (p *infinity) ReLogin() bool {
	if p.Config.GetLogin() == "" && p.Config.GetPassword() == "" {
		panic(errors.New("ReLogin before login! I don't know login and password :( "))
	}
	sleep_time := time.Duration(1000)

	for count := 0; count < TRY_COUNT; count++ {
		result := p.Login(p.Config.GetLogin(), p.Config.GetPassword())
		if result {
			return result
		} else {
			log.Printf("Infinity: ReLogin is fail trying next after %+v", sleep_time)
			time.Sleep(sleep_time * time.Millisecond)
			sleep_time = time.Duration(float32(sleep_time) * 1.4)
		}
	}
	return false
}

// Ping возвращает true если запрос выполнен успешно и время сервера infinity в формате yyyy-MM-dd HH:mm:ss.
// Если запрос выполнен неуспешно возвращает false и пустую строку.
// Условие: пользователь должен быть авторизован.
//

func (p *infinity) _request(conn_suffix string, url_values map[string]string) ([]byte, error) {
	for i, connString := range p.ConnStrings {
		req, err := http.NewRequest("GET", connString + conn_suffix, nil)
		if err != nil {
			log.Printf("error at forming request %v, %#v\n error: %v", conn_suffix, url_values, err)
		}
		req.Header.Add("ContentType", "text/html;charset=UTF-8")
		values := req.URL.Query()
		for k, v := range url_values {
			values.Add(k, v)
		}

		req.URL.RawQuery = values.Encode()

		if p.Cookie != nil {
			req.AddCookie(p.Cookie)
		}
		client := &http.Client{Timeout:15 * time.Second}
		res, err := client.Do(req)
		if err != nil {
			log.Printf("Error at send request to infinity: %v", err)
			continue
		}
		defer func() {
			//delete from i position
			p.ConnStrings = append(p.ConnStrings[:i], p.ConnStrings[i + 1:]...)
			//paste to head
			p.ConnStrings = append(p.ConnStrings[:0], append([]string{connString}, p.ConnStrings[0:]...)...)
		}()
		if res != nil && res.Status == "403 Forbidden" {
			log.Println("INF response is: ", res, "; error is:", err, ". I will ReLogin and will retrieve data again after 3s.")
			time.Sleep(3 * time.Second)
			p.ReLogin()
			return p._request(conn_suffix, url_values)
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Printf("error at reading from response %v", err)
		}
		return body, nil
	}
	return nil, errors.New("Не могу подключится к АПИ такси :(")

}

// GetServices возвращает информацию об услугах доступных для заказа (filterField is set to true!)
func (p *infinity) GetServices() []InfinityService {
	var tmp []InfinityServices
	body, err := p._request("GetViewData", map[string]string{"params": "[{\"viewName\":\"Taxi.Services\",\"filterField\":{\"n\":\"AvailableToClients\",\"v\":true}}]"})
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		log.Printf("error in unmarshal json, %v", err)
	}
	return tmp[0].Rows
}

// GetCarsInfo возвращает информацию о машинах
func (p *infinity) GetCarsInfo() []t.CarInfo {
	var tmp []InfinityCarsInfo
	body, err := p._request("GetViewData", map[string]string{"params": "[{\"viewName\":\"Taxi.Cars.InfoEx\"}]"})
	if err != nil {
		return []t.CarInfo{}
	}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		log.Printf("error at unmarshal json from infinity %v", string(body))
		return []t.CarInfo{}
	}
	return tmp[0].Rows
}

func (p *infinity) NewOrder(order t.NewOrderInfo) t.Answer {
	//because sedi use Id address and id from street autocomplete was equal idadress not idstreet
	order.Delivery.IdAddress = ""
	order.Destinations[0].IdAddress = ""

	order.IdService = p.Config.GetIdService()
	param, err := json.Marshal(order)
	if err != nil {
		log.Printf("INF NO error at marshal json to infinity %+v, %v", order, err)
		return t.Answer{IsSuccess:false, Message:fmt.Sprint(err)}
	}
	log.Printf("INF NEW ORDER (jsonified): \n%+v \nat INF:%+v", string(param), p)
	body, err := p._request("RemoteCall", map[string]string{"params": string(param), "method": "Taxi.WebAPI.NewOrder"})
	if err != nil {
		return t.Answer{IsSuccess:false, Message:fmt.Sprint(err)}
	}
	var ans t.Answer
	err = json.Unmarshal(body, &ans)
	if err != nil {
		log.Printf("INF NO error at unmarshal json from infinity %v", string(body))
		return t.Answer{IsSuccess:false, Message:fmt.Sprint(err)}
	}
	return ans
}

func (p *infinity) CalcOrderCost(order t.NewOrderInfo) (int, string) {
	//	order.Delivery.IdStreet = strconv.ParseInt(order.Delivery.IdAddress, 10, 64)
	//	order.Destinations[0].IdStreet = strconv.ParseInt(order.Destinations[0].IdAddress, 10, 64)

	order.IdService = p.Config.GetIdService()
	param, err := json.Marshal(order)
	if err != nil {
		log.Printf("error at marshal json from infinity %v", order)
		return -1, ""
	}
	body, err := p._request("RemoteCall", map[string]string{"params": string(param), "method": "Taxi.WebAPI.CalcOrderCost"})
	if err != nil {
		return 0, fmt.Sprint(err)
	}
	var tmp t.Answer
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		log.Printf("error at unmarshal json from infinity %v", string(body))
		return -1, ""
	}
	return tmp.Content.Cost, tmp.Content.Details
}

//{"phone":"89261234567","deliveryTime":"2015-07-15+07:00:00","deliveryMinutes":60,"idService":7006261161,"notes":"Хочется+комфортную+машину","markups":[7002780031,7004760103],"attributes":[1000113000,1000113002],"delivery":{"idRegion":7006803034,"idStreet":0,"house":"1","fraction":"1","entrance":"2","apartment":"30"},"destinations":{["lat":55.807898,"lon":37.785449,"idRegion":7006803034,"idPlace":7006803054,"idStreet":7006803054,"house":"12","entrance":"2","apartment":"30"}]}

type PrivateParams struct {
	Name  string `json:"name"`
	Login string `json:"login"`
}

//Taxi.WebAPI.Client.GetPrivateParams (Получение параметров клиента)
//Контент:
//Параметры личного кабинета клиента в виде JSON объекта: { "name" : <Имя клиента>, "login" : <Логин клиента> }
func (p *infinity) GetPrivateParams() (bool, string, string) {
	body, err := p._request("RemoteCall", map[string]string{"method": "Taxi.WebAPI.Client.GetPrivateParams"})
	var temp t.Answer
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("error at unmarshal json from infinity %v", string(body))
	}
	return temp.IsSuccess, temp.Content.Name, temp.Content.Login
}

/////////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////

//Taxi.WebAPI.Client.ChangePassword (Изменение пароля) Изменяет пароль клиента.
//Параметры:
//Новый пароль (строка)
func (p *infinity) ChangePassword(password string) (bool, string) {
	tmp, err := json.Marshal(password)
	if err != nil {
		log.Printf("error at marshal json to infinity %v", string(password))
		return false, fmt.Sprint(err)
	}
	body, err := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.ChangePassword"})
	var temp t.Answer
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("error at unmarshal json from infinity %v", string(body))
		return false, fmt.Sprint(err)
	}
	return temp.IsSuccess, temp.Message
}

//Taxi.WebAPI.Client.ChangeName (Изменение имени клиента) Изменяет имя клиента в системе.
//Параметры:
//Новое имя клиента (строка)
func (p *infinity) ChangeName(name string) (bool, string) {

	tmp, err := json.Marshal(name)
	if err != nil {
		log.Printf("error at marshal json to infinity %v", string(name))
		return false, fmt.Sprint(err)
	}

	body, err := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.ChangeName"})

	var temp t.Answer
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("error at unmarshal json from infinity %v", string(body))
		return false, fmt.Sprint(err)
	}
	return temp.IsSuccess, temp.Message
}

//Taxi.WebAPI.Client.SendMessage (Отправка сообщения оператору) Отправляет операторам системы уведомление с сообщением данного клиента
//Параметры:
//Текст сообщения (строка)
func (p *infinity) WriteDispatcher(message string) (bool, string /*, string*/) {
	tmp, err := json.Marshal(message)
	if err != nil {
		log.Printf("INF WD error at marshal json to infinity %v", string(message))
		return false, fmt.Sprint(err)
	}
	params := map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.SendMessage"}
	//	log.Printf("INF: Write dispatcher: %+v", params)
	body, err := p._request("RemoteCall", params)
	var temp t.Answer
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("INF WD error at unmarshal json from infinity %v", string(body))
		return false, fmt.Sprint(err)
	}
	return temp.IsSuccess, temp.Message
}

func (p *infinity) CallbackRequest(phone string) (bool, string) {
	tmp, err := json.Marshal(phone)
	if err != nil {
		log.Printf("INF CBKR error at marshal json to infinity %v", string(phone))
		return false, fmt.Sprint(err)
	}
	//	log.Printf("Callback request (jsoned) %s", tmp)
	body, err := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.CallbackRequest"})
	var temp t.Answer
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("INF CBKR error at unmarshal json from infinity %v", string(body))
		return false, fmt.Sprint(err)
	}
	return temp.IsSuccess, temp.Message
}

//Taxi.WebAPI.Client.ClearHistory (Очистка истории заказов клиента)
//Отмечает закрытые заказы клиента как не видимые для личного кабинета (т.е. сама информация о заказе не удаляется)
func (p *infinity) ClearHistory() (bool, string) {
	body, err := p._request("RemoteCall", map[string]string{"method": "Taxi.WebAPI.Client.ClearHistory"})

	var temp t.Answer
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("error at unmarshal json from infinity %v", string(body))
		return false, fmt.Sprint(err)
	}
	return temp.IsSuccess, temp.Message
}

//Taxi.WebAPI.Client.CancelOrder (Отказ от заказа) Устанавливает для указанного заказа состояние «Отменен»
//Параметры:
//Идентификатор заказа (Int64)
func (p *infinity) CancelOrder(order int64) (bool, string, error) {
	tmp, err := json.Marshal(order)
	if err != nil {
		log.Printf("INF CO error at marshal json to infinity %v", string(order))
		return false, fmt.Sprint(err), err
	}

	body, err := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.CancelOrder"})
	if err != nil {
		return false, fmt.Sprint(err), err
	}
	var temp t.Answer
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("INF CO error at unmarshal json from infinity %v", string(body))
		return false, fmt.Sprint(err), err
	}
	return temp.IsSuccess, temp.Message, nil
}

//Taxi.WebAPI.Client.Feedback (Отправка отзыва о заказе)
//Указывает оценку и отзыв для указанного заказа, отправляя операторам системы уведомления об отзыве.
//Параметры:
//JSON объект: {
//"idOrder" : <Идентификатор заказа (Int64)>,
//"rating" : <Оценка (число)>,
//"notes" : <Текст отзыва>
//}
func (p *infinity) Feedback(inf t.Feedback) (bool, string) {
	tmp, err := json.Marshal(inf)
	if err != nil {
		log.Printf("INF FDBK error at marshal json to infinity %v", inf)
		return false, fmt.Sprint(err)
	}

	body, err := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.Feedback"})
	if err != nil {
		return false, fmt.Sprint(err)
	}
	var temp t.Answer
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("INF FDBK error at unmarshal json from infinity %v", string(body))
		return false, fmt.Sprint(err)
	}
	return temp.IsSuccess, temp.Message
}

//Taxi.WebAPI.Client.WhereIT (Отправка запроса «Клиент не видит машину»)
//Отправляет операторам системы уведомление «Клиент не видит машину»
//Параметры:
//Идентификатор заказа (Int64)
func (p *infinity) WhereIt(ID int64) (bool, string) {
	tmp, err := json.Marshal(ID)
	if err != nil {
		log.Printf("INF WHEREIT error at marshal json to infinity %v", string(ID))
		return false, fmt.Sprint(err)
	}
	body, err := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.WhereIT"})
	if err != nil {
		return false, fmt.Sprint(err)
	}
	var temp t.Answer
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("INF WHEREIT error at unmarshal json from infinity %v", string(body))
		return false, fmt.Sprint(err)
	}
	return temp.IsSuccess, temp.Message
}

//Taxi.WebAPI.Client.Phones.Edit (Изменение/ Добавление телефона клиента)
//Изменяет телефон клиента, если в параметрах указан идентификатор существующего телефона данного
//клиента.
//Добавляет новый телефон клиента, если в параметрах отсутствует идентификатор существующего телефона.
//Параметры:
//JSON объект: {
//"id" : <Идентификатор телефона (Int64), необходим при редактировании>,
//"contact" : <Номер телефона (строка)>
//}

type phonesEdit struct {
	Id      int64  `json:"id"`
	Contact string `json:"contact"`
}

func (p *infinity) PhonesEdit(phone phonesEdit) (bool, string) {
	tmp, err := json.Marshal(phone)
	if err != nil {
		log.Printf("error at marshal json to infinity %+v", phone)
		return false, fmt.Sprint(err)
	}
	body, err := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.Phones.Edit"})
	var temp t.Answer
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("error at unmarshal json from infinity %v", string(body))
		return false, fmt.Sprint(err)
	}
	return temp.IsSuccess, temp.Message
}

//Taxi.WebAPI.Client.Phones.Remove (Удаление телефона клиента) Удаляет указанный телефон клиента.
//Параметры:
//Идентификатор телефона клиента (Int64)
func (p *infinity) PhonesRemove(phone int64) (bool, string) {
	tmp, err := json.Marshal(phone)
	if err != nil {
		log.Printf("error at marshal json to infinity %v", string(phone))
		return false, fmt.Sprint(err)
	}
	body, err := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.Phones.Remove"})
	var temp t.Answer
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("error at unmarshal json from infinity %v", string(body))
		return false, fmt.Sprint(err)
	}
	return temp.IsSuccess, temp.Message
}

//Taxi.WebAPI.Client.Addresses.Edit (Изменение/ Добавление адреса клиента)
//Изменяет «любимый» адрес клиента, если в параметрах указан идентификатор существующего элемента
//справочника, в противном случае будет добавлен новый адрес клиента.
type favorite struct {
	Id         int64  `json:"id"`         // <Идентификатор любимогo адреса (Int64)>,
	Name       string `json:"name"`       // <Наименование элемента (строка)>,
	ImageIndex int    `json:"imageIndex"` // <Индекс иконки, адреса (число)>,
	IdAddres   string `json:"idAddress"`  // <Идентификатор существующего описания адреса (адрес дома или объекта)>,
	IdRedion   int64  `json:"idRegion"`   // <Идентификатор региона (Int64)>,
	IdDistrict int64  `json:"idDistrict"` // <Идентификатор района (Int64)>,
	IdCity     int64  `json:"idCity"`     // <Идентификатор города (Int64)>,
	IdPlace    int64  `json:"idPlace"`    // <Идентификатор поселения (Int64)>,
	IdStreet   int64  `json:"idStreet"`   // <Идентификатор улицы (Int64)>,
	House      string `json:"house"`      // <No дома (строка)>,
	Building   string `json:"building"`   // <Строение (строка)>,
	Fracion    string `json:"fraction"`   // <Корпус (строка)>,
	Entrance   string `json:"entrance"`   // <Подъезд (строка)>,
	Apartament string `json:"apartment"`  // <No квартиры (строка)>
}

//Параметры idRegion, idDistrict, idCity, idStreet, house, building, fraction используются для создания нового
//описания адреса и не анализируются при указании параметра idAddress.
func (p *infinity) AddressesEdit(f favorite) (bool, string) {

	tmp, err := json.Marshal(f)
	if err != nil {
		log.Printf("error at marshal json to infinity %+v", f)
		return false, fmt.Sprint(err)
	}

	body, err := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.Addresses.Edit"})
	var temp t.Answer
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("error at unmarshal json from infinity %v", string(body))
		return false, fmt.Sprint(err)
	}

	return temp.IsSuccess, temp.Message
}

func (p *infinity) AddressesRemove(id int64) (bool, string) {
	body, err := p._request("RemoteCall", map[string]string{"params": string(id), "method": "Taxi.WebAPI.Client.Addresses.Remove"})
	var temp t.Answer
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("error at unmarshal json from infinity %v", string(body))
		return false, fmt.Sprint(err)
	}
	return temp.IsSuccess, temp.Message
}

/////////////////////////////
type Orders struct {
	Rows []t.Order `json:"rows"`
}

//Taxi.t.Orders (Заказы: активные и предварительные)
func (p *infinity) Orders() []t.Order {
	body, err := p._request("GetViewData", map[string]string{"params": "[{\"viewName\": \"Taxi.Orders\"}]"})
	if err != nil {
		log.Print("INF ORDRS error at connection to inf at orders")
		p.ReLogin()
		return []t.Order{}
	}
	temp := []Orders{}
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("INF ORDRS error at unmarshal json from infinity %s", string(body))
		p.ReLogin()
		return []t.Order{}
	}
	result := []t.Order{}
	for _, order := range temp[0].Rows {
		if arrival, err := time.Parse("2006-01-02 15:04:05", order.ArrivalTime); err == nil {
			order.TimeArrival = &arrival
		}
		if delivery, err := time.Parse("2006-01-02 15:04:05", order.DeliveryTime); err == nil {
			order.TimeDelivery = &delivery
		}
		result = append(result, order)
	}
	return result
}

func (p *infinity) OrdersClosedByDates() []t.Order {
	body, err := p._request("GetViewData", map[string]string{"params": "[{\"viewName\": \"Taxi.Orders.Closed.ByDates\"}]"})
	temp := []Orders{}
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("error at unmarshal json from infinity %s", string(body))
		return []t.Order{}
	}
	return temp[0].Rows
}

func (p *infinity) OrdersClosedlastN() []t.Order {
	body, err := p._request("GetViewData", map[string]string{"params": "[{\"viewName\": \"Taxi.Orders.Closed.LastN\"}]"})

	var temp []t.Order
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("error at unmarshal json from infinity %s", string(body))
		return []t.Order{}
	}
	return temp
}

//Taxi.Markups (Список доступных наценок)
func (p *infinity) Markups() []t.Markup {
	type MarkupsResult struct {
		Rows []t.Markup `json:"rows"`
	}
	var temp []MarkupsResult

	body, err := p._request("GetViewData", map[string]string{"params": "[{\"viewName\": \"Taxi.Markups\"}]"})
	if err != nil {
		log.Print("INF MRKPS error at connection to inf at markups %( ")
		return []t.Markup{}
	}

	//log.Printf("INF: markups result: %s", body)
	err = json.Unmarshal(body, &temp)
	//log.Printf("INF: markups result: %+v", temp)
	if err != nil {
		log.Printf("INF MRKPS error at unmarshal json at  %s, %v", string(body), err)
		return []t.Markup{}
	}
	return temp[0].Rows
}

//Taxi.Services (Список услуг)
//Taxi.ClientPhones (Телефоны клиента)
//Taxi.Cars.Info (Дополнительная информация о машине)
//Taxi.CarAttributes (Список атрибутов машины)

/////////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////
//Taxi.t.FastAddresses.Search (Поиск быстрых адресов) Доступность: Личный кабинет + Заказ с сайта
//Поля:  ID  Name  IDType  IDAddress  Apartment  Entrance  StrAddress  AddrDescription  Type
//Наименование
//Тип адреса/быстрого адреса
//ID адреса
//No квартиры (строка)
//Подъезд
//Фактический адрес
//Описание адреса
//Тип адреса/быстрого адреса в виде строки


func (p *infinity) AddressesAutocomplete(text string) t.AddressPackage {
	body, err := p._request("GetViewData", map[string]string{"params": "[{\"viewName\": \"Taxi.Addresses.Search\", \"params\": [{\"n\": \"SearchText\", \"v\": \"" + text + "\"}]}]"})
	if err != nil {
		log.Printf("INF AA error at connection to inf %( ")
		return t.AddressPackage{}
	}
	var temp []t.AddressPackage
	//log.Printf("INF ADDRESS AUTOCOMPLETE FOR %s \nRETRIEVE THIS:%s", text, string(body))
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("INF AA  error at unmarshal json from infinity %s", string(body))
		return t.AddressPackage{}
	}
	return temp[0]
}

//Taxi.ClientAddresses (Адреса клиента)
func (p *infinity) ClientAddresses() t.AddressPackage {
	body, err := p._request("GetViewData", map[string]string{"params": "[{\"viewName\": \"Taxi.ClientAddresses\"}]"})
	var temp []t.AddressPackage
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("error at unmarshal json from infinity %s", string(body))
		return t.AddressPackage{}
	}
	return temp[0]
}



