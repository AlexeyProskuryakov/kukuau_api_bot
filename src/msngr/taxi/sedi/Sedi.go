package sedi
import (
	"log"
	"fmt"
	"errors"
	"io/ioutil"
	"time"
	"net/http"
	"encoding/json"
)

const (
	MAX_TRY_COUNT = 5
	HOST_POSTFIX = "/handlers/sedi/api.ashx"

)

type SediAPI struct {
	Host    string
	Cookies []*http.Cookie
	Info    *LoginInfo
}

func NewSediAPI(host string) *SediAPI {
	s := SediAPI{Host:host}
	return &s
}

type SediResponse struct {
	Success bool `json:"Success"`
	Message string `json:"Message"`
}


func (s *SediAPI) doReq(req *http.Request) ([]byte, error) {
	count := 0
	for {
		client := &http.Client{}
		if len(s.Cookies) > 0 {
			req.AddCookie(s.Cookies[0])
		}
		res, err := client.Do(req)
		if res == nil || err != nil {
			log.Println("SA response is: ", res, "; error is:", err, ". I will reconnect and will retrieve data again after 3s.")
			time.Sleep(3 * time.Second)
			count += 1
			if count > MAX_TRY_COUNT {
				return []byte{}, errors.New(fmt.Sprintf("Max tryings count ended and request not proceed... %#v", req))
			}
			continue
		}
		defer res.Body.Close()
		s.Cookies = res.Cookies()
		//d
		for _, cookie := range s.Cookies {
			log.Printf("cookie: %+v\n", *cookie)
		}
		log.Printf("Cookes count: %v", len(s.Cookies))
		//ed
		body, err := ioutil.ReadAll(res.Body)
		return body, err
	}
	return []byte{}, errors.New(fmt.Sprintf("Error at do request... %#v", req))
}


func (s *SediAPI) SimpleGetRequest(q string, params map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", s.Host + HOST_POSTFIX, nil)
	if err != nil {
		log.Printf("SEDI GET request error in request")
	}
	values := req.URL.Query()
	for k, v := range params {
		values.Add(k, v)
	}
	values.Add("q", q)
	req.URL.RawQuery = values.Encode()

	result, err := s.doReq(req)
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
/*
{
 "LoginInfo": {
	 "ID": -1, // ID пользователя
	 "AccountID": -1, // ID лицевого счета
	 "Login": "", // логин
	 "Name": "", // имя
	 "Nick": "", // позывной
	 "Patronymic": "", // отчество
	 "SecondName": "", // фамилия
	 "Email": "",
	 "Photo": false, // загружено ли фото
	 "AnonymousOrder": false, // разрешены ли заказы без регистрации
	 "IsAnonymous": true, // это анонимный пользователь
	 "IsCustomer": false, // это заказчик
	 "IsEmployee": false, // это сотрудник
	 "AllowCustomerCost": false, // разрешено ли задавать свою стоимость заказа
	 "AllowAuction": false, // разрешено ли регистрировать аукционы
	 "AllowCalc": true, // разрешен ли расчет стоимости,
	 "UnconfirmedRating": true, // разрешены ли оценки без указания заказа,
	 "Balance": {
		 "Currency": "", // валюта счета
		 "Credit": 0, // допустимый кредит
		 "Value": 0 // баланс лицевого счета
	 },
	 "Phones": [], // номера телефонов
	 "Addresses": [ // адрес регистрации
		 {
		 "Id": 457484,
		 "CountryName": "Russia",
		 "CityName": "Санкт-Петербург",
		 "LocalityName": "",
		 "PostalCode": "195269",
		 "StreetName": "",
		 "ObjectName": "",
		 "HouseNumber": "",
		 "ApartmentNumber": "",
		 "EntranceNumber": 0,
		 "GeoPoint": {
			 "Latitude": 60.034689,
			 "Longitude": 30.40019
				}
	 	}
 		],
	 "Affiliate": false, // подключена ли партнёрская программа
	 "AffilateOrders": false, // является ли пользователь ее участником
	 "Promocode": -1, // промокод партнерской программы
	 "Owner": { // информация о группе
		 "ID": 179,
		 "Name": "Ваша Цена",
		 "Url": "taxivc.ru",
		 "Phones": [ // телефоны диспетчеров
				"84996410064"
				]
			}
 		},
 "Success": true // успешность выполнения
}

*/
type LoginInfo struct {
	ID                string `json:"ID"`
	AccountID         string `json:"AccountID"`
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
						  Currency string `json:"Currency"`
						  Credit   int `json:"Credit"`
						  Value    int `json:"Value"`
					  } `json:"Balance"`
	Phones            []string `json:"Phones"`
	Affiliate         bool `json:"Affiliate"`
	AffilateOrders    bool `json:"AffilateOrders"`
	Promocode         int64 `json:"Promocode"`
	Owner             struct {
						  ID               string `json:"ID"`
						  Name             string `json:"Name"`
						  Url              string `json:"Url"`
						  DispatcherPhones []string `json:"Phones"`
					  } `json:"Owner"`
}

func (l LoginInfo) String() string {
	return fmt.Sprintf(" ID: %v, AccountID: %v, Login: %v, Name: %v, Nick: %v" +
	"\n Customer?: %v, Employee?: %v, AllowCustomerCost?: %v, AllowAuction?: %v, AllowCalc?: %v" +
	"\n Balance: %+v\n Phones: %+v\n Affiliate?: %v, AffilateOrders?: %v, Promocode: %v\n Owner: %+v",
		l.ID, l.AccountID, l.Login, l.Name, l.Nick,
		l.IsCustomer, l.IsEmployee, l.AllowCustomerCost, l.AllowAuction, l.AllowCalc,
		l.Balance, l.Phones, l.Affiliate, l.AffilateOrders, l.Promocode, l.Owner)
}

type LoginInfoResponse struct {
	SediResponse
	LoginInfo LoginInfo `json:"LoginInfo"`
}

func (s *SediAPI) auth(q string, prms map[string]string) (*LoginInfoResponse, error) {
	res, err := s.SimpleGetRequest(q, prms)
	if err != nil {
		log.Printf("SA AUTH Error at %v [%v] %v", q, prms, err)
		return nil, err
	}
	result := LoginInfoResponse{}
	err = json.Unmarshal(res, &result)
	if err != nil {
		log.Printf("SA AUTH Error at unmarshall response from %v [%v] key %v\n%s", q, prms, err, res)
		return nil, err
	}
	if !result.Success {
		log.Printf("SA AUTH Error in %v [%v] is: %v, message: %v", q, prms, err, result.Message)
		return nil, errors.New(result.Message)
	}
	log.Printf("API AUTH INFO %v [%v] :\n%+v\n", q, prms, result.LoginInfo)
	return &result, nil
}
/*
Retrieving LoginInformation from sms
*/
func (s *SediAPI) ActivationKey(sms_code string) (*LoginInfo, error) {
	resp_info, err := s.auth("activation_key", map[string]string{"actkey":sms_code})
	if resp_info != nil {
		return &resp_info.LoginInfo, err
	}
	return nil, err
}

func (s *SediAPI) GetProfile() (*LoginInfo, error) {
	resp_info, err := s.auth("get_profile", map[string]string{})
	if resp_info != nil {
		return &resp_info.LoginInfo, err
	}
	return nil, err
}

func (s *SediAPI) AuthoriseCustomer(name, phone string) (*LoginInfo, error) {
	resp_info, err := s.auth("authorise_customer", map[string]string{"name":name, "phone":phone})
	if resp_info != nil {
		return &resp_info.LoginInfo, err
	}
	return nil, err
}

/*
Вопросы:
1) Почему не приходит смс код при запросе get_activation_key на сервере test2.sedi.ru? (Номер +79811064022)
2) В каких случаях требуется авторизация через смс?
3) Почему в ответе на get_profile приходит невалидный json? В поле LogInfo.RegistrationDate отправляется "new Date(...)" когда ожидается дата в текстовом или строковом виде?
4) Почему при запросе authorise_customer c параметрами  name: "Михаил Егоренков", phone:"+79612183729" сервер отвечает ошибкой и сообщением "Применение не поддерживается"?
5) Что нужно сделать чтобы завести сотрудника (или некоторый иной аккаунт) для возможности иметь функциональность:
	1) Рассчитать стоймость заказа
	2) Создать заказ, отправить его на биржу и отменить.
	3) Отследить состояние заказа, и иметь информацию об автомобиле исполнителя, времени в течении которого исполнитель приедет к клиенту
	4) Отправить отзыв по заказу (поставить оценку исполнителю)

*/