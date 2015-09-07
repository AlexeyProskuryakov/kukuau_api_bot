package infinity

import (
	"encoding/json"
	"errors"
	// "flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	// "strconv"
	"sync"
	"time"
)

type TaxiInterface interface {
	NewOrder(order NewOrder) (Answer, error)
	CancelOrder(order_id int64) (bool, string)
	CalcOrderCost(order NewOrder) (int, string)
	AddressesSearch(text string) FastAddress
	Orders() []Order
}

func warn(err error) {
	if err != nil {
		log.Println(err)
	}
}
func warnp(err error) {
	if err != nil {
		log.Println(err)
		panic(err)
	}
}

type InfinityApiParams struct {
	Host              string `json:"host"`
	Login             string `json:"login"`
	Password          string `json:"password"`
	ConnectionsString string `json:"connection_string"`
}

// infinity - Структура для работы с API infinity.
type infinity struct {
	Host       string
	ConnString string // Строка подключения к infinity API
	// default: http://109.202.25.248:8080/WebAPITaxi/
	LoginTime     time.Time
	Cookie        *http.Cookie
	LoginResponse struct {
		Success  bool  `json:"success"`
		IDClient int64 `json:"idClient"`
		Params   struct {
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
			DefaultPlaceID             string `json:"DefaultPlaceID"` // Can be null, so used as string here.
			DefaultPlaceName           string `json:"DefaultPlaceName"`
		} `json:"params"`
		SessionID string `json:"sessionid"`
	}
	Message struct {
		Success bool   `json:"isSuccess"`
		Content string `json:"content"`
	}
	Services []InfinityServices `json:"InfinityServices"`
}

// Global API variable
var instance *infinity
var fakeInstance *FakeInfinity
var once sync.Once

type InfinityMixin struct {
	API TaxiInterface
}

func initInfinity(conn_str, host, login, password string) *infinity {
	result := &infinity{}
	result.ConnString = conn_str
	result.Host = host
	result.Login(login, password)
	return result
}

func GetInfinityAPI(iap InfinityApiParams, isTest bool) TaxiInterface {
	if isTest {
		log.Println("return test API")
		once.Do(func() {
			fakeInstance = &FakeInfinity{}
		})
		return fakeInstance
	} else {
		log.Println("return REAL API AHTUNG!!!")
		once.Do(func() {
			instance = initInfinity(iap.ConnectionsString, iap.Host, iap.Login, iap.Login)
		})
		//todo this is coprocode realise with your brains!!!
		if instance == nil {
			instance = initInfinity(iap.ConnectionsString, iap.Host, iap.Login, iap.Login)
		}
		return instance
	}
}

type Answer struct {
	IsSuccess bool   `json:"isSuccess"`
	Message   string `json:"message"`

	Content struct {
		Id      int64  `json:"id"` // :7007330031,
		Name    string `json:"name"`
		Login   string `json:"login"`
		Number  int64  `json:"number"` // :406
		Cost    int    `json:"cost"`
		Details string `json:"details"`
	} `json:"content"`
}

type Destination struct {
	Lat          float64 `json:"lat"`           // : <Широта координаты адреса (при указании места на карте). Если указано, информация о доставке по указанию и адресе игнорируется>,
	Lon          float64 `json:"lon"`           // : <Долгота координаты адреса (при указании места на карте). Если указано, информация о доставке по указанию и адресе игнорируется>,"isByDirection" : <Заказ машины с указанием пункта назначения при подаче (если задано в true,информация о адресе игнонрируются)>,
	IdAddres     string  `json:"idAddress"`     // : <Идентификатор существующего описания адреса (адрес дома или объекта)>,
	IdRegion     int64   `json:"idRegion"`      // : <Идентификатор региона (Int64)>,
	IdDistrict   int64   `json:"idDistrict"`    // : <Идентификатор района (Int64)>,
	IdCity       int64   `json:"idCity"`        // : <Идентификатор города (Int64)>,
	IdPlace      int64   `json:"idPlace"`       // : <Идентификатор поселения (Int64)>,
	IdStreet     int64   `json:"idStreet"`      // : <Идентификатор улицы (Int64)>,
	House        string  `json:"house"`         // : <№ дома (строка)>,
	Building     string  `json:"building"`      // : <Строение (строка)>,
	Fraction     string  `json:"fraction"`      // : <Корпус (строка)>,
	Entrance     string  `json:"entrance"`      // : <Подъезд (строка)>,
	Apartament   string  `json:"apartment"`     // : <№ квартиры (строка)> ,
	IdFastAddres string  `json:"idFastAddress"` // // : <ID быстрого адреса. Дополнительное информационное поле, описывающее быстрый адрес, связанный с указанным адресом. Значение учитывается только при указании idAddress>
}

type Delivery struct {
	//Lat           float64 `json:"lat"`           // : <Широта координаты адреса (при указании места на карте). Если указано, информация о адресе игнорируется>,
	//Lon           float64 `json:"lon"`           // : <Долгота координаты адреса (при указании места на карте). Если указано, информация о адресе игнорируется>,
	//IdAddress     string `json:"idAddress"`      // <Идентификатор существующего описания адреса (адрес дома или объекта)>,
	IdRegion int64 `json:"idRegion"` // <Идентификатор региона (Int64)>,
	//IdDistrict    int64  `json:"idDistrict"`    // : <Идентификатор района (Int64)>,
	//IdCity        int64  `json:"idCity"`        // : <Идентификатор города (Int64)>,
	//IdPlace       int64  `json:"idPlace"`       //: <Идентификатор поселения (Int64)>,
	IdStreet int64  `json:"idStreet"` // : <Идентификатор улицы (Int64)>,
	House    string `json:"house"`    // : <№ дома (строка)>,
	//Building      string `json:"building"`      // : <Строение (строка)>,
	Fracion    string `json:"fraction"`  // : <Корпус (строка)>,
	Entrance   string `json:"entrance"`  // : <Подъезд (строка)>,
	Apartament string `json:"apartment"` // : <№ квартиры (строка)>,
	//IdFastAddress string `json:"idFastAddress"` //: <ID быстрого адреса. Дополнительное информационное поле, описывающее быстрый адрес, связанный с указанным адресом. Значение учитывается только при указании idAddress>
}

type NewOrder struct {
	//request
	Phone           string `json:"phone"`
	DeliveryTime    string `json:"deliveryTime"`    //<Время подачи в формате yyyy-MM-dd HH:mm:ss>
	DeliveryMinutes int64  `json:"deliveryMinutes"` // <Количество минут до подачи (0-сейчас, но не менее минимального времени на подачу, указанного в настройках системы), не анализируется если задано поле deliveryTime >
	IdService       int64  `json:"idService"`       //<Идентификатор услуги заказа (не может быть пустым)>
	Notes           string `json:"notes"`           // <Комментарий к заказу>
	//Markups           [2]int64 `json:"markups"`           // <Массив идентификаторов наценок заказа>
	Attributes [2]int64 `json:"attributes"` // <Массив идентификаторов дополнительных атрибутов заказа>
	// Инфомация о месте подачи машины
	Delivery Delivery `json:"delivery"`
	// Пункты назначения заказа (массив, не может быть пустым)
	Destinations []Destination `json:"destinations"`
	// Флаг безналичного заказа
	IsNotCash bool `json:"isNotCash"` //: <true или false (bool)>
}

type Order struct {
	ID                int64  `json:"ID"`                // ID
	State             int    `json:"State"`             //Состояние заказа
	Cost              int    `json:"Cost"`              //Стоимость
	IsNotCash         bool   `json:"IsNotCash"`         //Безналичный заказ (bool)
	IsPrevious        int    `json:"IsPrevious"`        //Тип заказа (0 –активный, 1-предварительный, 2-предварительный ставший активным)
	LastStateTime     string `json:"LastStateTime"`     //Дата-Время последнего и\зменения состояния
	DeliveryTime      string `json:"DeliveryTime"`      //Требуемое время подачи машины
	Distance          int    `json:"Distance"`          //Расстояние км (если оно рассчитано системой)
	TimeOfArrival     string `json:"TimeOfArrival"`     //Прогнозируемое время прибытия машины на заказ
	IDDeliveryAddress int64  `json:"IDDeliveryAddress"` //ID адреса подачи
	DeliveryStr       string `json:"DeliveryStr"`       //Адрес подачи в виде текста
	DestinationsStr   string `json:"DestinationsStr"`   //Пункты назначения в виде текста (с учетом настроек отображения в диспетчерской: Первый/Последний/Все)
	IDCar             int64  `json:"IDCar"`             //ID машины
	IDService         int64  `json:"IdService"`         //ID услуги
	Car               string `json:"Car"`               //Позывной машины
	Service           string `json:"Service"`           //Услуга
	Drivers           string `json:"Drivers"`           //ФИО Водителя
}

// Login - Авторизация в сервисе infinity. Входные параметры: login:string; password:string.
// Возвращает true, если авторизация прошла успешно, false иначе.
// Устанавливает время авторизации в infinity.LoginTime при успешной авторизации.
func (p *infinity) Login(login, password string) bool {
	p.LoginResponse.Success = false
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"Login", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")

	values := req.URL.Query()
	values.Add("l", login)
	values.Add("p", password)
	values.Add("app", "CxTaxiClient")
	req.URL.RawQuery = values.Encode()
	log.Println(req.URL)
	res, err := client.Do(req)
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)
	err = json.Unmarshal(body, &p.LoginResponse)
	warnp(err)
	log.Printf("[login] self: %+q\n", p)
	if p.LoginResponse.Success {
		log.Println("[login] JSESSIONID: ", p.LoginResponse.SessionID)
		// log.Printf("[login] self: %+q\n", p)
		p.Cookie = &http.Cookie{
			Name:   "JSESSIONID",
			Value:  p.LoginResponse.SessionID,
			Path:   "/",
			Domain: "109.202.25.248",
		}
		p.LoginTime = time.Now()
		return true
	}
	return false
}

// Ping возвращает true если запрос выполнен успешно и время сервера infinity в формате yyyy-MM-dd HH:mm:ss.
// Если запрос выполнен неуспешно возвращает false и пустую строку.
// Условие: пользователь должен быть авторизован.
func (p *infinity) Ping() (bool, string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"RemoteCall", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()
	values.Add("method", "Taxi.WebAPI.GetDateTime")
	req.URL.RawQuery = values.Encode()

	//log.Println(req.URL)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	//log.Println(string(body))
	err = json.Unmarshal(body, &p.Message)
	warnp(err)
	return p.Message.Success, p.Message.Content
}

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
	Rows []InfinityCarInfo `json:"rows"`
}
type InfinityCarInfo struct {
	ID       int64   `json:"id"`
	Callsign string  `json:"Callsign"`
	State    int     `json:"State"`
	Number   string  `json:"Number"`
	Color    string  `json:"Color"`
	Model    string  `json:"Model"`
	Driver   string  `json:"Driver"`
	Lat      float64 `json:"Lat"`
	Lon      float64 `json:"Lon"`
}

// GetServices возвращает информацию об услугах доступных для заказа (filterField is set to true!)
func (p *infinity) GetServices() []InfinityService {
	var tmp []InfinityServices
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"GetViewData", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()
	values.Add("params", "[{\"viewName\":\"Taxi.Services\",\"filterField\":{\"n\":\"AvailableToClients\",\"v\":true}}]")
	req.URL.RawQuery = values.Encode()
	//log.Println(values)
	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	//log.Println(string(body))
	warnp(err)
	err = json.Unmarshal(body, &tmp)
	warnp(err)
	return tmp[0].Rows
}

// GetCarsInfo возвращает информацию о машинах
func (p *infinity) GetCarsInfo() []InfinityCarInfo {
	var tmp []InfinityCarsInfo
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"GetViewData", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()
	values.Add("params", "[{\"viewName\":\"Taxi.Cars.InfoEx\"}]")
	req.URL.RawQuery = values.Encode()
	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	//log.Println(string(body))
	warnp(err)
	err = json.Unmarshal(body, &tmp)
	warnp(err)
	return tmp[0].Rows
}

func (p *infinity) NewOrder(order NewOrder) (Answer, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"RemoteCall", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()
	values.Add("method", "Taxi.WebAPI.NewOrder")

	param, err := json.Marshal(order)
	warn(err)
	values.Add("params", string(param))

	//log.Println(string(param), err)

	req.URL.RawQuery = values.Encode()

	log.Println(req.URL)
	log.Printf("[New Order] inf cookie: %+v \n[%+v]", p.Cookie, p)
	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	log.Println(string(body))

	var ans Answer
	err = json.Unmarshal(body, &ans)
	warnp(err)
	return ans, err
}

func (p *infinity) CalcOrderCost(order NewOrder) (int, string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"RemoteCall", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()
	values.Add("method", "Taxi.WebAPI.CalcOrderCost")

	param, err := json.Marshal(order)
	warn(err)
	values.Add("params", string(param))

	//log.Println(string(param), err)

	req.URL.RawQuery = values.Encode()

	log.Println(req.URL)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	log.Println(string(body))

	var tmp Answer
	err = json.Unmarshal(body, &tmp)
	//log.Println(tmp)
	warnp(err)
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
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"RemoteCall", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()
	values.Add("method", "Taxi.WebAPI.Client.GetPrivateParams")
	req.URL.RawQuery = values.Encode()
	//log.Println(req.URL)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	var temp Answer

	err = json.Unmarshal(body, &temp)
	return temp.IsSuccess, temp.Content.Name, temp.Content.Login
}

/////////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////

//Taxi.WebAPI.Client.ChangePassword (Изменение пароля) Изменяет пароль клиента.
//Параметры:
//Новый пароль (строка)
func (p *infinity) ChangePassword(password string) (bool, string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"RemoteCall", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()
	values.Add("method", "Taxi.WebAPI.Client.ChangePassword")
	tmp, err := json.Marshal(password)
	warnp(err)
	values.Add("params", string(tmp))
	req.URL.RawQuery = values.Encode()

	//log.Println(req.URL)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	var temp Answer
	err = json.Unmarshal(body, &temp)
	warnp(err)
	return temp.IsSuccess, temp.Message
}

//Taxi.WebAPI.Client.ChangeName (Изменение имени клиента) Изменяет имя клиента в системе.
//Параметры:
//Новое имя клиента (строка)
func (p *infinity) ChangeName(name string) (bool, string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"RemoteCall", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()
	values.Add("method", "Taxi.WebAPI.Client.ChangeName")
	tmp, err := json.Marshal(name)
	warnp(err)
	values.Add("params", string(tmp))
	req.URL.RawQuery = values.Encode()

	//log.Println(req.URL)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	var temp Answer
	err = json.Unmarshal(body, &temp)
	warnp(err)
	return temp.IsSuccess, temp.Message
}

//Taxi.WebAPI.Client.SendMessage (Отправка сообщения оператору) Отправляет операторам системы уведомление с сообщением данного клиента
//Параметры:
//Текст сообщения (строка)
func (p *infinity) SendMessage(message string) (bool, string /*, string*/) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"RemoteCall", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()
	values.Add("method", "Taxi.WebAPI.Client.SendMessage")
	tmp, err := json.Marshal(message)
	warnp(err)
	values.Add("params", string(tmp))
	req.URL.RawQuery = values.Encode()

	//log.Println(req.URL)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	var temp Answer
	err = json.Unmarshal(body, &temp)
	warnp(err)
	return temp.IsSuccess, temp.Message
}

func (p *infinity) CallbackRequest(phone string) (bool, string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"RemoteCall", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()
	values.Add("method", "Taxi.WebAPI.Client.CallbackRequest")
	tmp, err := json.Marshal(phone)
	warnp(err)
	values.Add("params", string(tmp))
	req.URL.RawQuery = values.Encode()

	//log.Println(req.URL)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	var temp Answer
	err = json.Unmarshal(body, &temp)
	warnp(err)
	return temp.IsSuccess, temp.Message
}

//Taxi.WebAPI.Client.ClearHistory (Очистка истории заказов клиента)
//Отмечает закрытые заказы клиента как не видимые для личного кабинета (т.е. сама информация о заказе не удаляется)
func (p *infinity) ClearHistory() (bool, string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"RemoteCall", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()
	values.Add("method", "Taxi.WebAPI.Client.ClearHistory")
	req.URL.RawQuery = values.Encode()

	//log.Println(req.URL)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	var temp Answer
	err = json.Unmarshal(body, &temp)
	warnp(err)
	return temp.IsSuccess, temp.Message
}

//Taxi.WebAPI.Client.CancelOrder (Отказ от заказа) Устанавливает для указанного заказа состояние «Отменен»
//Параметры:
//Идентификатор заказа (Int64)
func (p *infinity) CancelOrder(order int64) (bool, string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"RemoteCall", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()
	values.Add("method", "Taxi.WebAPI.Client.CancelOrder")

	tmp, err := json.Marshal(order)
	warnp(err)
	values.Add("params", string(tmp))

	req.URL.RawQuery = values.Encode()

	//log.Println(req.URL)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	var temp Answer
	err = json.Unmarshal(body, &temp)
	warnp(err)
	return temp.IsSuccess, temp.Message
}

//Taxi.WebAPI.Client.Feedback (Отправка отзыва о заказе)
//Указывает оценку и отзыв для указанного заказа, отправляя операторам системы уведомления об отзыве.
//Параметры:
//JSON объект: {
//"idOrder" : <Идентификатор заказа (Int64)>,
//"rating" : <Оценка (число)>,
//"notes" : <Текст отзыва>
//}

type feedback struct {
	IdOrder int64  `json:"idOrder"`
	Rating  int    `json:"rating"`
	Notes   string `json:"notes"`
}

func (p *infinity) Feedback(inf feedback) (bool, string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"RemoteCall", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()
	values.Add("method", "Taxi.WebAPI.Client.Feedback")

	tmp, err := json.Marshal(inf)
	warnp(err)
	values.Add("params", string(tmp))

	req.URL.RawQuery = values.Encode()

	//log.Println(req.URL)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	var temp Answer
	err = json.Unmarshal(body, &temp)
	warnp(err)
	return temp.IsSuccess, temp.Message
}

//Taxi.WebAPI.Client.WhereIT (Отправка запроса «Клиент не видит машину»)
//Отправляет операторам системы уведомление «Клиент не видит машину»
//Параметры:
//Идентификатор заказа (Int64)
func (p *infinity) WhereIT(ID int64) (bool, string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"RemoteCall", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()
	values.Add("method", "Taxi.WebAPI.Client.WhereIT")

	tmp, err := json.Marshal(ID)
	warnp(err)
	values.Add("params", string(tmp))

	req.URL.RawQuery = values.Encode()

	//log.Println(req.URL)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	var temp Answer
	err = json.Unmarshal(body, &temp)
	warnp(err)
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
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"RemoteCall", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()
	values.Add("method", "Taxi.WebAPI.Client.Phones.Edit")

	tmp, err := json.Marshal(phone)
	warnp(err)
	values.Add("params", string(tmp))

	req.URL.RawQuery = values.Encode()

	//log.Println(req.URL)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	var temp Answer
	//log.Println(string(body))
	err = json.Unmarshal(body, &temp)
	warnp(err)
	return temp.IsSuccess, temp.Message
}

//Taxi.WebAPI.Client.Phones.Remove (Удаление телефона клиента) Удаляет указанный телефон клиента.
//Параметры:
//Идентификатор телефона клиента (Int64)
func (p *infinity) PhonesRemove(phone int64) (bool, string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"RemoteCall", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()
	values.Add("method", "Taxi.WebAPI.Client.Phones.Remove")

	tmp, err := json.Marshal(phone)
	warnp(err)
	values.Add("params", string(tmp))

	req.URL.RawQuery = values.Encode()

	//log.Println(req.URL)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	var temp Answer
	err = json.Unmarshal(body, &temp)
	warnp(err)
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
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"RemoteCall", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()
	values.Add("method", "Taxi.WebAPI.Client.Addresses.Edit")

	tmp, err := json.Marshal(f)
	warnp(err)
	values.Add("params", string(tmp))

	req.URL.RawQuery = values.Encode()

	//log.Println(req.URL)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	var temp Answer
	err = json.Unmarshal(body, &temp)
	warnp(err)
	return temp.IsSuccess, temp.Message
}

func (p *infinity) AddressesRemove(id int64) (bool, string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"RemoteCall", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()
	values.Add("method", "Taxi.WebAPI.Client.Addresses.Remove")

	tmp, err := json.Marshal(id)
	warnp(err)
	values.Add("params", string(tmp))

	req.URL.RawQuery = values.Encode()

	//log.Println(req.URL)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	var temp Answer
	err = json.Unmarshal(body, &temp)
	warnp(err)
	return temp.IsSuccess, temp.Message
}

/////////////////////////////

//Taxi.Orders (Заказы: активные и предварительные)
func (p *infinity) Orders() []Order {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"GetViewData", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()

	values.Add("params", "[{\"viewName\": \"Taxi.Orders\"}]")

	req.URL.RawQuery = values.Encode()

	log.Println(req.URL)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	var temp []Order
	err = json.Unmarshal(body, &temp)
	log.Println(body)
	warnp(err)
	return temp
}

//Taxi.Orders.Closed.ByDates (История заказов: По датам)
func (p *infinity) OrdersClosedByDates() []Order {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"GetViewData", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()

	values.Add("params", "[{\"viewName\": \"Taxi.Orders.Closed.ByDates\"}]")

	req.URL.RawQuery = values.Encode()

	log.Println(req.URL)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	var temp []Order
	log.Println(body)
	err = json.Unmarshal(body, &temp)
	warnp(err)
	return temp
}

//Taxi.Orders.Closed.LastN (История заказов: Последние)
func (p *infinity) OrdersClosedlastN() []Order {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"GetViewData", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()

	values.Add("params", "[{\"viewName\": \"Taxi.Orders.Closed.LastN\"}]")

	req.URL.RawQuery = values.Encode()

	log.Println(req.URL)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	var temp []Order
	log.Println(body)
	err = json.Unmarshal(body, &temp)
	warnp(err)
	return temp
}

//Taxi.Destinations.ByActiveOrder (Пункты назначения: Активные заказы)
//Taxi.Destinations.ByClosedOrder (Пункты назначения: Закрытые заказы (история))

//Taxi.Markups (Список доступных наценок)
func (p *infinity) Markups() []Order {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"GetViewData", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()

	values.Add("params", "[{\"viewName\": \"Taxi.Markups\"}]")

	req.URL.RawQuery = values.Encode()

	log.Println(req.URL)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	var temp []Order
	log.Println(body)
	err = json.Unmarshal(body, &temp)
	warnp(err)
	return temp
}

//Taxi.Services (Список услуг)
//Taxi.ClientPhones (Телефоны клиента)
//Taxi.Cars.Info (Дополнительная информация о машине)
//Taxi.CarAttributes (Список атрибутов машины)

/////////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////
//Taxi.FastAddresses.Search (Поиск быстрых адресов) Доступность: Личный кабинет + Заказ с сайта
//Поля:  ID  Name  IDType  IDAddress  Apartment  Entrance  StrAddress  AddrDescription  Type
//Наименование
//Тип адреса/быстрого адреса
//ID адреса
//No квартиры (строка)
//Подъезд
//Фактический адрес
//Описание адреса
//Тип адреса/быстрого адреса в виде строки

type FastAddress struct {
	Rows []struct {
		ID         int64  `json:"ID"`
		IDParent   int64  `json:"IDParent,omitempty"`
		Name       string `json:"Name"`
		ShortName  string `json:"ShortName,omitempty"`
		ItemType   int64  `json:"ItemType,omitempty"`
		FullName   string `json:"FullName"`
		IDRegion   int64  `json:"IDRegion"`
		IDDistrict int64  `json:"IDDistrict"`
		IDCity     int64  `json:"IDCity"`
		IDPlace    int64  `json:"IDPlace"`
		Region     string `json:"Region,omitempty"`
		District   string `json:"District,omitempty"`
		City       string `json:"City"`
		Place      string `json:"Place,omitempty"`
	} `json:"rows"`
}

type Address struct {
	ID         int64  `json:"ID"`
	IDParent   int64  `json:"IDParent,omitempty"`
	Name       string `json:"Name"`
	ShortName  string `json:"ShortName,omitempty"`
	ItemType   int64  `json:"ItemType,omitempty"`
	FullName   string `json:"FullName"`
	IDRegion   int64  `json:"IDRegion"`
	IDDistrict int64  `json:"IDDistrict"`
	IDCity     int64  `json:"IDCity"`
	IDPlace    int64  `json:"IDPlace"`
	Region     string `json:"Region,omitempty"`
	District   string `json:"District,omitempty"`
	City       string `json:"City"`
	Place      string `json:"Place,omitempty"`
}

func (p *infinity) AddressesSearch(text string) FastAddress {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"GetViewData", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()

	values.Add("params", "[{\"viewName\": \"Taxi.Addresses.Search\", \"params\": [{\"n\": \"SearchText\", \"v\": \""+text+"\"}]}]")

	req.URL.RawQuery = values.Encode()

	log.Println(req.URL)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	var temp []FastAddress
	log.Println(string(body))
	err = json.Unmarshal(body, &temp)

	//log.Println(temp)
	warnp(err)
	return temp[0]
}

//Taxi.ClientAddresses (Адреса клиента)
func (p *infinity) ClientAddresses() FastAddress {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+"GetViewData", nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()

	values.Add("params", "[{\"viewName\": \"Taxi.ClientAddresses\"}]")

	//109.202.25.248:8080/WebAPITaxi/GetViewData?params=[{"viewName": "Taxi.Addresses.Search", "params": [{"n": "SearchText", "v": "Никола"}]}]
	req.URL.RawQuery = values.Encode()

	log.Println(values)

	req.AddCookie(p.Cookie)
	//log.Println("Cookies in request? ", req.Cookies())
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)

	var temp []FastAddress
	err = json.Unmarshal(body, &temp)
	warnp(err)
	return temp[0]
}

//Taxi.Addresses.ByNameAndType (Объекты адреса по типу и наименованию)
//Taxi.Addresses.Search (Поиск улиц/объектов)
//Taxi.Terminal.Info (Информация для пользователя)
//Taxi.OrderMarkups.ByClosedOrder (Наценки: Закрытые заказы (история))
//Taxi.OrderMarkups.ByActiveOrder (Наценки: Активные заказы)
//Taxi.OrderAttributes.ByClosedOrder (Атрибуты машины: Закрытые заказы)
//Taxi.OrderAttributes.ByActiveOrder (Атрибуты машины: Активные заказы)
//Taxi.Cars.InfoEx (Дополнительная расширенная информация о машине)

//////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////

type logfile struct {
	m sync.Mutex
	t time.Time
	f *os.File
	n string
}

func (l *logfile) SetName() *os.File {
	l.m.Lock()
	if l.n != "" {
		l.f.Close()
	}
	l.t = time.Now()
	l.n = fmt.Sprintf("%d_%d_%d.log", l.t.Day(), int(l.t.Month()), l.t.Year())
	l.m.Unlock()
	f, err := os.OpenFile(l.n, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	log.SetOutput(f)
	log.Println("Log start point")
	l.f = f
	return f
}
func (l *logfile) Close() {
	l.m.Lock()
	l.f.Close()
	l.m.Unlock()
}

type DictItem struct {
	Value string `json:"value"`
	Text  string `json:"text"`
}

func StreetsSearchHandler(w http.ResponseWriter, r *http.Request, i TaxiInterface) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	log.Println("Searching address...")
	if r.Method == "GET" {
		params := url.Values{}
		params = r.URL.Query()
		log.Println(params)
		query := params.Get("q")
		log.Println(query)
		var results []DictItem
		if query != "" {
			rows := i.AddressesSearch(query).Rows
			for _, nitem := range rows {
				var item DictItem
				var err error
				t, err := json.Marshal(nitem)
				item.Value = string(t)
				warn(err)
				item.Text = fmt.Sprintf("%v %v", nitem.Name, nitem.ShortName)
				results = append(results, item)
			}
		}
		ans, err := json.Marshal(results)
		warn(err)
		fmt.Fprintf(w, "%s", string(ans))
	}
}

//helpers for forming destionation and delivery on infinity results after street search request
func H_get_delivery(info string, house string) (d Delivery) {
	//"{\"ID\":5009756374,\"IDParent\":5009755360,\"Name\":\"Николаева\",\"ShortName\":\"ул\",\"ItemType\":5,\"FullName\":\" ул Николаева\",\"IDRegion\":5009755359,\"IDDistrict\":0,\"IDCity\":5009755360,\"IDPlace\":0,\"Region\":\"обл Новосибирская\",\"City\":\"г Новосибирск\"}"
	err := json.Unmarshal([]byte(info), &d)
	warn(err)
	d.House = house
	return
}

func H_get_destination(info string, house string) (d Destination) {
	err := json.Unmarshal([]byte(info), &d)
	warn(err)
	d.House = house
	return
}

// func main() {
// 	//var flagDBCons int
// 	var flagPort int
// 	var flagNoLog bool

// 	flag.IntVar(&flagPort, "port", 80, "Bind to custom port.")
// 	flag.BoolVar(&flagNoLog, "nolog", false, "Turn off writing log")

// 	flag.Parse()
// 	t1 := time.Now()
// 	log.Println(fmt.Sprintf("%ds", (24*3600 - t1.Hour()*3600 + t1.Minute()*60 + t1.Second())))
// 	var log2file logfile
// 	if !flagNoLog {

// 		log2file.SetName()
// 		defer log2file.Close()
// 	}
// 	go func() {
// 		t1 := time.Now()
// 		leftTime, _ := time.ParseDuration(fmt.Sprintf("%ds", (24*3600 - t1.Hour()*3600 + t1.Minute()*60 + t1.Second())))
// 		timer := time.NewTimer(leftTime) // Seconds left until a new Day
// 		<-timer.C
// 		log2file.SetName()
// 	}()

// 	api_params := InfinityApiParams{}
// 	InfinityAPI := GetInfinityAPI(api_params)

// 	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		controlHandler(w, r, "localhost:"+strconv.Itoa(flagPort), InfinityAPI)
// 	})
// 	http.HandleFunc("/streets/", func(w http.ResponseWriter, r *http.Request) {
// 		StreetsSearchHandler(w, r, *InfinityAPI)
// 	})
// 	InfinityAPI.ConnString = "http://109.202.25.248:8080/WebAPITaxi/"
// 	InfinityAPI.Host = "109.202.25.248:8080"
// 	status := InfinityAPI.Login("test1", "test1")
// 	if status {
// 		log.Println("Установлено соединение с infinity")
// 	}

// 	status, serverTime := InfinityAPI.Ping()
// 	if status {
// 		log.Println("Время сервера: ", serverTime)
// 		log.Println(InfinityAPI.GetServices())
// 		log.Println(InfinityAPI.GetCarsInfo())
// 	} else {
// 		log.Println("Проблема со связью с сервером infinity")
// 	}

// 	log.Println("Started at localhost:", flagPort)
// 	http.ListenAndServe(":"+strconv.Itoa(flagPort), nil)
// }

////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////
type FakeInfinity struct {
	orders []Order
}

func (inf *FakeInfinity) NewOrder(order NewOrder) (ans Answer, e error) {
	saved_order := Order{
		ID:    int64(len(inf.orders) + 1),
		State: 0,
		Cost:  100500,
	}

	inf.orders = append(inf.orders, saved_order)

	ans = Answer{
		IsSuccess: true,
		Message:   "test order was formed",
	}
	ans.Content.Id = saved_order.ID
	log.Println("FA now i have orders: ", len(inf.orders))
	return
}

func (inf *FakeInfinity) Orders() []Order {
	return inf.orders
}

func (inf *FakeInfinity) CancelOrder(order_id int64) (bool, string) {
	log.Println("FA order was canceled", order_id)
	for i, order := range inf.orders {
		if order.ID == order_id {
			inf.orders[i].State = 7
			return true, "test order was cancelled"
		}
	}
	return false, "order not found :( "
}

func (p *FakeInfinity) CalcOrderCost(order NewOrder) (int, string) {
	log.Println("FA calulate cost for order: ", order)
	return 100500, "Good cost!"
}

func (p *FakeInfinity) AddressesSearch(text string) FastAddress {
	// 	type FastAddress struct {
	// 	Rows []struct {
	// 		ID         int64  `json:"ID"`
	// 		IDParent   int64  `json:"IDParent,omitempty"`
	// 		Name       string `json:"Name"`
	// 		ShortName  string `json:"ShortName,omitempty"`
	// 		ItemType   int64  `json:"ItemType,omitempty"`
	// 		FullName   string `json:"FullName"`
	// 		IDRegion   int64  `json:"IDRegion"`
	// 		IDDistrict int64  `json:"IDDistrict"`
	// 		IDCity     int64  `json:"IDCity"`
	// 		IDPlace    int64  `json:"IDPlace"`
	// 		Region     string `json:"Region,omitempty"`
	// 		District   string `json:"District,omitempty"`
	// 		City       string `json:"City"`
	// 		Place      string `json:"Place,omitempty"`
	// 	} `json:"rows"`
	// }
	log.Println("FA return empty address")
	return FastAddress{}
}
