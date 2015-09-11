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
	// "os"
	// "strconv"
	"sync"
	"time"
)

type TaxiInterface interface {
	NewOrder(order NewOrder) (Answer, error)
	CancelOrder(order_id int64) (bool, string)
	CalcOrderCost(order NewOrder) (int, string)
	Orders() []Order
	Feedback(f Feedback) (bool, string)
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
	//for
	curent_credentials struct {
		login    string
		password string
	}
}

// Global API variable
var instance *infinity
var fakeInstance *FakeInfinity
var once sync.Once

type InfinityMixin struct {
	API TaxiInterface
}

func _initInfinity(conn_str, host, login, password string) *infinity {
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
			instance = _initInfinity(iap.ConnectionsString, iap.Host, iap.Login, iap.Login)
		})
		//todo this is coprocode realise with your brains!!!
		if instance == nil {
			instance = _initInfinity(iap.ConnectionsString, iap.Host, iap.Login, iap.Login)
		}
		return instance
	}
}

func GetRealInfinityAPI(iap InfinityApiParams) *infinity {
	if instance == nil {
		instance = _initInfinity(iap.ConnectionsString, iap.Host, iap.Login, iap.Login)
	}
	return instance
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
	Attributes   [2]int64      `json:"attributes"`   // <Массив идентификаторов дополнительных атрибутов заказа>
	Delivery     Delivery      `json:"delivery"`     // Инфомация о месте подачи машины
	Destinations []Destination `json:"destinations"` // Пункты назначения заказа (массив, не может быть пустым)
	IsNotCash    bool          `json:"isNotCash"`    //: // Флаг безналичного заказа <true или false (bool)>
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
		p.curent_credentials.login = login
		p.curent_credentials.password = password
		return true
	}
	return false
}

func (p *infinity) reconnect() {
	if p.curent_credentials.login == "" && p.curent_credentials.password == "" {
		panic(errors.New("reconnect before connect! I don't know login and password :( "))
	}
	sleep_time := time.Duration(1000)
	for {
		result := p.Login(p.curent_credentials.login, p.curent_credentials.password)
		if result {
			break
		} else {
			time.Sleep(sleep_time * time.Millisecond)
			sleep_time = time.Duration(float32(sleep_time) * 1.4)
		}
	}
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
		p.reconnect()
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

func (p *infinity) _request(conn_suffix string, url_values map[string]string) []byte {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.ConnString+conn_suffix, nil)
	warnp(err)
	req.Header.Add("ContentType", "text/html;charset=UTF-8")
	values := req.URL.Query()
	for k, v := range url_values {
		values.Add(k, v)
	}

	req.URL.RawQuery = values.Encode()
	req.AddCookie(p.Cookie)
	res, err := client.Do(req)
	if res.Status == "403 Forbidden" {
		err = errors.New("Ошибка авторизации infinity! (Возможно не установлены cookies)")
		p.reconnect()
	}
	warnp(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	warnp(err)
	return body
}

// GetServices возвращает информацию об услугах доступных для заказа (filterField is set to true!)
func (p *infinity) GetServices() []InfinityService {
	var tmp []InfinityServices

	body := p._request("GetViewData", map[string]string{"params": "[{\"viewName\":\"Taxi.Services\",\"filterField\":{\"n\":\"AvailableToClients\",\"v\":true}}]"})
	err := json.Unmarshal(body, &tmp)
	warnp(err)
	return tmp[0].Rows
}

// GetCarsInfo возвращает информацию о машинах
func (p *infinity) GetCarsInfo() []InfinityCarInfo {
	var tmp []InfinityCarsInfo
	body := p._request("GetViewData", map[string]string{"params": "[{\"viewName\":\"Taxi.Cars.InfoEx\"}]"})
	err := json.Unmarshal(body, &tmp)
	warnp(err)
	return tmp[0].Rows
}

func (p *infinity) NewOrder(order NewOrder) (Answer, error) {
	param, err := json.Marshal(order)
	warnp(err)
	body := p._request("RemoteCall", map[string]string{"params": string(param), "method": "Taxi.WebAPI.NewOrder"})
	var ans Answer
	err = json.Unmarshal(body, &ans)
	warnp(err)
	return ans, err
}

func (p *infinity) CalcOrderCost(order NewOrder) (int, string) {
	param, err := json.Marshal(order)
	warnp(err)
	body := p._request("RemoteCall", map[string]string{"params": string(param), "method": "Taxi.WebAPI.CalcOrderCost"})
	var tmp Answer
	err = json.Unmarshal(body, &tmp)
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

	body := p._request("RemoteCall", map[string]string{"method": "Taxi.WebAPI.Client.GetPrivateParams"})
	var temp Answer
	err := json.Unmarshal(body, &temp)
	warnp(err)
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
	warnp(err)
	body := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.ChangePassword"})
	var temp Answer
	err = json.Unmarshal(body, &temp)
	warnp(err)
	return temp.IsSuccess, temp.Message
}

//Taxi.WebAPI.Client.ChangeName (Изменение имени клиента) Изменяет имя клиента в системе.
//Параметры:
//Новое имя клиента (строка)
func (p *infinity) ChangeName(name string) (bool, string) {

	tmp, err := json.Marshal(name)
	warnp(err)

	body := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.ChangeName"})

	var temp Answer
	err = json.Unmarshal(body, &temp)
	warnp(err)
	return temp.IsSuccess, temp.Message
}

//Taxi.WebAPI.Client.SendMessage (Отправка сообщения оператору) Отправляет операторам системы уведомление с сообщением данного клиента
//Параметры:
//Текст сообщения (строка)
func (p *infinity) SendMessage(message string) (bool, string /*, string*/) {
	tmp, err := json.Marshal(message)
	warnp(err)

	body := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.SendMessage"})

	var temp Answer
	err = json.Unmarshal(body, &temp)
	warnp(err)
	return temp.IsSuccess, temp.Message
}

func (p *infinity) CallbackRequest(phone string) (bool, string) {
	tmp, err := json.Marshal(phone)
	warnp(err)
	body := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.CallbackRequest"})
	var temp Answer
	err = json.Unmarshal(body, &temp)
	warnp(err)
	return temp.IsSuccess, temp.Message
}

//Taxi.WebAPI.Client.ClearHistory (Очистка истории заказов клиента)
//Отмечает закрытые заказы клиента как не видимые для личного кабинета (т.е. сама информация о заказе не удаляется)
func (p *infinity) ClearHistory() (bool, string) {
	body := p._request("RemoteCall", map[string]string{"method": "Taxi.WebAPI.Client.ClearHistory"})

	var temp Answer
	err := json.Unmarshal(body, &temp)
	warnp(err)
	return temp.IsSuccess, temp.Message
}

//Taxi.WebAPI.Client.CancelOrder (Отказ от заказа) Устанавливает для указанного заказа состояние «Отменен»
//Параметры:
//Идентификатор заказа (Int64)
func (p *infinity) CancelOrder(order int64) (bool, string) {
	tmp, err := json.Marshal(order)
	warnp(err)

	body := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.CancelOrder"})

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

type Feedback struct {
	IdOrder int64  `json:"idOrder"`
	Rating  int    `json:"rating"`
	Notes   string `json:"notes"`
}

func (p *infinity) Feedback(inf Feedback) (bool, string) {
	tmp, err := json.Marshal(inf)
	warnp(err)

	body := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.Feedback"})

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
	tmp, err := json.Marshal(ID)
	warnp(err)

	body := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.WhereIT"})

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
	tmp, err := json.Marshal(phone)
	warnp(err)
	body := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.Phones.Edit"})
	var temp Answer
	err = json.Unmarshal(body, &temp)
	warnp(err)
	return temp.IsSuccess, temp.Message
}

//Taxi.WebAPI.Client.Phones.Remove (Удаление телефона клиента) Удаляет указанный телефон клиента.
//Параметры:
//Идентификатор телефона клиента (Int64)
func (p *infinity) PhonesRemove(phone int64) (bool, string) {
	tmp, err := json.Marshal(phone)
	warnp(err)
	body := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.Phones.Remove"})
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

	tmp, err := json.Marshal(f)
	warnp(err)

	body := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.Addresses.Edit"})
	var temp Answer
	err = json.Unmarshal(body, &temp)
	warnp(err)

	return temp.IsSuccess, temp.Message
}

func (p *infinity) AddressesRemove(id int64) (bool, string) {
	tmp, err := json.Marshal(id)
	warnp(err)

	body := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.Addresses.Remove"})
	var temp Answer
	err = json.Unmarshal(body, &temp)
	warnp(err)
	return temp.IsSuccess, temp.Message
}

/////////////////////////////

//Taxi.Orders (Заказы: активные и предварительные)
func (p *infinity) Orders() []Order {
	body := p._request("GetViewData", map[string]string{"params": "[{\"viewName\": \"Taxi.Orders\"}]"})

	var temp []Order
	err := json.Unmarshal(body, &temp)
	warnp(err)
	return temp
}

//Taxi.Orders.Closed.ByDates (История заказов: По датам)
func (p *infinity) OrdersClosedByDates() []Order {
	body := p._request("GetViewData", map[string]string{"params": "[{\"viewName\": \"Taxi.Orders.Closed.ByDates\"}]"})

	var temp []Order
	err := json.Unmarshal(body, &temp)
	warnp(err)
	return temp
}

//Taxi.Orders.Closed.LastN (История заказов: Последние)
func (p *infinity) OrdersClosedlastN() []Order {
	body := p._request("GetViewData", map[string]string{"params": "[{\"viewName\": \"Taxi.Orders.Closed.LastN\"}]"})

	var temp []Order
	err := json.Unmarshal(body, &temp)
	warnp(err)
	return temp
}

//Taxi.Destinations.ByActiveOrder (Пункты назначения: Активные заказы)
//Taxi.Destinations.ByClosedOrder (Пункты назначения: Закрытые заказы (история))

//Taxi.Markups (Список доступных наценок)
func (p *infinity) Markups() []Order {
	body := p._request("GetViewData", map[string]string{"params": "[{\"viewName\": \"Taxi.Markups\"}]"})

	var temp []Order
	err := json.Unmarshal(body, &temp)
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

	body := p._request("GetViewData", map[string]string{"params": "[{\"viewName\": \"Taxi.Addresses.Search\", \"params\": [{\"n\": \"SearchText\", \"v\": \"" + text + "\"}]}]"})

	var temp []FastAddress
	err := json.Unmarshal(body, &temp)
	warnp(err)
	return temp[0]
}

//Taxi.ClientAddresses (Адреса клиента)
func (p *infinity) ClientAddresses() FastAddress {
	body := p._request("GetViewData", map[string]string{"params": "[{\"viewName\": \"Taxi.ClientAddresses\"}]"})
	var temp []FastAddress
	err := json.Unmarshal(body, &temp)
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
type DictItem struct {
	Value string `json:"value"`
	Text  string `json:"text"`
}

func StreetsSearchController(w http.ResponseWriter, r *http.Request, i *infinity) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	log.Println("Searching address...")
	if r.Method == "GET" {
		params := url.Values{}
		params = r.URL.Query()
		// log.Println(params)
		query := params.Get("q")
		// log.Println(query)
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

///////////////////////////////////////////////////////////////////////////

type CarsCache struct {
	cars map[int64]InfinityCarInfo
	api  *infinity
}

func _create_cars_map(i *infinity) map[int64]InfinityCarInfo {
	cars_map := make(map[int64]InfinityCarInfo)
	cars_info := i.GetCarsInfo()
	for _, info := range cars_info {
		cars_map[info.ID] = info
	}

	return cars_map
}

func NewCarsCache(i *infinity) *CarsCache {
	cars_map := _create_cars_map(i)
	handler := CarsCache{cars: cars_map, api: i}
	return &handler
}

func (ch *CarsCache) CarInfo(car_id int64) *InfinityCarInfo {
	key, ok := ch.cars[car_id]
	if !ok {
		ch.cars = _create_cars_map(ch.api)
		key, ok = ch.cars[car_id]
		if !ok {
			return nil
		}
	}
	return &key
}

var StatusesMap = map[int]string{
	1:  "Не распределен",
	2:  "Назначен",
	3:  "Выехал",
	4:  "Ожидание клиента",
	5:  "Выполнение",
	6:  "Простой",
	7:  "Оплачен",
	8:  "Не оплачен",
	9:  "Отменен",
	11: "Запланирована машина",
	12: "Зафиксирован",
	13: "Не создан",
	14: "Горящий заказ",
	15: "Не подтвержден",
}

func IsOrderNotAvaliable(state int) bool {
	if state == 9 || state == 13 || state == 7 {
		return true
	}
	return false
}

const (
	ORDER_PAYED    = 7
	ORDER_CANCELED = 9
)
