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

	CONNECTION_ERROR = "Система обработки заказов такси не отвечает, попробуйте позже."
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
type InfinityAPI struct {
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
	TryingsCount  int
	InReloginNow  bool
	Name          string
}

func (i InfinityAPI) String() string {
	return fmt.Sprintf("\nInfinity API processing.\nConnection strings:%+v\nLogon?:%v, time:%v client id:%v\nid_service: %v", i.ConnStrings, i.LoginResponse.Success, i.LoginTime, i.LoginResponse.IDClient, i.Config.GetIdService())
}

func _initInfinity(config t.TaxiAPIConfig, name string) *InfinityAPI {
	result := &InfinityAPI{}
	result.ConnStrings = config.GetConnectionStrings()
	result.Config = config
	result.Name = name
	logon := result.Login()

	if !logon {
		go result.WaitForReLogin()
	}

	return result
}

func GetInfinityAPI(tc t.TaxiAPIConfig, name string) t.TaxiInterface {
	instance := _initInfinity(tc, name)
	return instance
}

func GetTestInfAPI(tc t.TaxiAPIConfig) interface{} {
	instance := _initInfinity(tc, "test")
	return instance
}

func GetInfinityAddressSupplier(tc t.TaxiAPIConfig, name string) t.AddressSupplier {
	instance := _initInfinity(tc, name + "_address_supplier")
	return instance
}

func (p *InfinityAPI) Login() bool {
	p.LoginResponse.Success = false
	login := p.Config.GetLogin()
	pwd := p.Config.GetPassword()
	var body []byte
	var err error

	if p.Config.GetAPIData().ApiKey != "" {
		log.Printf("INF [%v] USE API KEY %v", p.Name, p.Config.GetAPIData().ApiKey)
		body, err = p._request("Login", map[string]string{"k":p.Config.GetAPIData().ApiKey, "app":"CxTaxiWebAPI"})
	} else if login != "" && pwd != "" {
		log.Printf("INF [%v] USE login and pass", p.Name)
		body, err = p._request("Login", map[string]string{"l":p.Config.GetLogin(), "p":p.Config.GetPassword(), "app":"CxTaxiClient"})
	} else {
		panic(fmt.Sprintf("Not api key not login and not pwd at [%v]", p.Name))
	}

	if err != nil {
		return false
	}
	err = json.Unmarshal(body, &p.LoginResponse)
	if err != nil {
		log.Printf("INF error at unmarshalling json at LOGIN:%q \nerror: %v", string(body), err)
		return false
	}
	if p.LoginResponse.Success {
		log.Printf("INF [%v] login SUCCESS! JSESSIONID: %v", p.Name, p.LoginResponse.SessionID)
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

func (p *InfinityAPI) IsConnected() bool {
	return p.LoginResponse.Success
}

func (p *InfinityAPI) Connect() {
	if p.InReloginNow {
		return
	}
	go p.WaitForReLogin()
}

func (p *InfinityAPI) WaitForReLogin() bool {
	/*Долгая функция которая будет стучаться 10 раз пока не ответят.
	*/
	if p.Config.GetLogin() == "" && p.Config.GetPassword() == "" {
		panic(errors.New("ReLogin before login! I don't know login and password :( "))
	}
	sleep_time := time.Duration(1000)
	p.InReloginNow = true
	log.Printf("INF Start wait for relogin")
	defer func() {
		log.Printf("INF Stop wait for relogin")
		p.InReloginNow = false
	}()

	for count := 0; count < TRY_COUNT; count++ {
		result := p.Login()
		if result {
			return result
		} else {
			log.Printf("INF: [%v] Login is fail trying next after %+v", p.Name, sleep_time)
			time.Sleep(sleep_time * time.Millisecond)
			sleep_time = time.Duration(float32(sleep_time) * 1.4)
		}
	}
	log.Printf("INF: Can not connect after %v times", TRY_COUNT)
	return false
}

func (p *InfinityAPI) _request(conn_suffix string, url_values map[string]string) ([]byte, error) {
	for i, connString := range p.ConnStrings {
		req, err := http.NewRequest("GET", connString + conn_suffix, nil)
		if err != nil {
			log.Printf("INF Error at forming request %v, %#v\n error: %v", conn_suffix, url_values, err)
			return nil, errors.New(CONNECTION_ERROR)
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
			log.Printf("INF Error at send request: %v", err)
			continue
		}
		defer func() {
			p.ConnStrings = append(p.ConnStrings[:i], p.ConnStrings[i + 1:]...)
			p.ConnStrings = append(p.ConnStrings[:0], append([]string{connString}, p.ConnStrings[0:]...)...)
		}()
		if res != nil && res.StatusCode != 200 {
			log.Printf("INF For %v [%v] > %v\n response is: %+v error is: %v", conn_suffix, url_values, connString, res, err)
			time.Sleep(time.Second * 5)
			p.Connect()
			return nil, errors.New(CONNECTION_ERROR)
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Printf("INF Error at reading from response %v", err)
			return body, errors.New(CONNECTION_ERROR)
		}
		log.Printf("INF [%v] OK > [%v]", p.Name, connString)
		return body, nil
	}
	return nil, errors.New(CONNECTION_ERROR)

}

func (p *InfinityAPI) GetCarsInfo() []t.CarInfo {
	var tmp []InfinityCarsInfo
	body, err := p._request("GetViewData", map[string]string{"params": "[{\"viewName\":\"Taxi.Cars.InfoEx\"}]"})
	if err != nil {
		log.Printf("INF GCI error at send req to infinity %v", err)
		return []t.CarInfo{}
	}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		log.Printf("INF GCI error at unmarshal json from infinity %v %v", string(body), err)
		return []t.CarInfo{}
	}
	return tmp[0].Rows
}

func (p *InfinityAPI) NewOrder(order t.NewOrderInfo) t.Answer {
	//because sedi use Id address and id from street autocomplete was equal idadress not idstreet
	order.Delivery.IdAddress = ""
	order.Destinations[0].IdAddress = ""

	order.IdService = p.Config.GetIdService()
	param, err := json.Marshal(order)
	if err != nil {
		log.Printf("INF NO error at marshal json to infinity %+v, %v", order, err)
		return t.Answer{IsSuccess:false, Message:fmt.Sprint(err)}
	}
	log.Printf("INF NEW ORDER (jsonified): \n%s\nat%+v", param, p)
	body, err := p._request("RemoteCall", map[string]string{"params": string(param), "method": "Taxi.WebAPI.NewOrder"})
	if err != nil {
		return t.Answer{IsSuccess:false, Message:fmt.Sprint(err)}
	}
	var ans t.Answer
	err = json.Unmarshal(body, &ans)
	log.Printf("INF NEW ORDER ANSER: \n%+v err? %v", ans, err)
	if err != nil {
		log.Printf("INF NO error at unmarshal json from infinity %v", string(body))
		p.Connect()
		return t.Answer{IsSuccess:false, Message:fmt.Sprint(err)}
	}
	return ans
}

func (p *InfinityAPI) CalcOrderCost(order t.NewOrderInfo) (int, string) {
	order.IdService = p.Config.GetIdService()
	param, err := json.Marshal(order)
	if err != nil {
		log.Printf("INF COC error at marshal json from infinity %+v %v", order, err)
		return -1, ""
	}
	body, err := p._request("RemoteCall", map[string]string{"params": string(param), "method": "Taxi.WebAPI.CalcOrderCost"})
	if err != nil {
		return 0, fmt.Sprint(err)
	}
	var tmp t.Answer
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		log.Printf("INFO COC error at unmarshal json from infinity %v %v", string(body), err)
		p.Connect()
		return -1, ""
	}
	return tmp.Content.Cost, tmp.Content.Details
}

func (p *InfinityAPI) WriteDispatcher(message string) (bool, string /*, string*/) {
	tmp, err := json.Marshal(message)
	if err != nil {
		log.Printf("INF WD error at marshal json to infinity %v %v", string(message), err)
		return false, fmt.Sprint(err)
	}
	params := map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.SendMessage"}
	//	log.Printf("INF: Write dispatcher: %+v", params)
	body, err := p._request("RemoteCall", params)
	var temp t.Answer
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("INF WD error at unmarshal json from infinity %v %v", string(body), err)
		return false, CONNECTION_ERROR
	}
	return temp.IsSuccess, temp.Message
}

func (p *InfinityAPI) CallbackRequest(phone string) (bool, string) {
	tmp, err := json.Marshal(phone)
	if err != nil {
		log.Printf("INF CBKR error at marshal json to infinity %v %v", string(phone), err)
		return false, fmt.Sprint(err)
	}
	//	log.Printf("Callback request (jsoned) %s", tmp)
	body, err := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.CallbackRequest"})
	var temp t.Answer
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("INF CBKR error at unmarshal json from infinity %v %v", string(body), err)
		return false, CONNECTION_ERROR
	}
	return temp.IsSuccess, temp.Message
}

func (p *InfinityAPI) CancelOrder(order int64) (bool, string, error) {
	tmp, err := json.Marshal(order)
	if err != nil {
		log.Printf("INF CO error at marshal json to infinity %v %v", string(order), err)
		return false, fmt.Sprint(err), err
	}

	body, err := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.CancelOrder"})
	if err != nil {
		return false, fmt.Sprint(err), err
	}
	var temp t.Answer
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("INF CO error at unmarshal json from infinity %v %v", string(body), err)
		return false, CONNECTION_ERROR, err
	}
	return temp.IsSuccess, temp.Message, nil
}

func (p *InfinityAPI) Feedback(inf t.Feedback) (bool, string) {
	tmp, err := json.Marshal(inf)
	if err != nil {
		log.Printf("INF FDBK error at marshal json to infinity %v %v", inf, err)
		return false, fmt.Sprint(err)
	}

	body, err := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.Feedback"})
	if err != nil {
		return false, fmt.Sprint(err)
	}
	var temp t.Answer
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("INF FDBK error at unmarshal json from infinity %v %v", string(body), err)
		return false, CONNECTION_ERROR
	}
	return temp.IsSuccess, temp.Message
}

func (p *InfinityAPI) WhereIt(ID int64) (bool, string) {
	tmp, err := json.Marshal(ID)
	if err != nil {
		log.Printf("INF WHEREIT error at marshal json to infinity %v %v", ID, err)
		return false, fmt.Sprint(err)
	}
	body, err := p._request("RemoteCall", map[string]string{"params": string(tmp), "method": "Taxi.WebAPI.Client.WhereIT"})
	if err != nil {
		return false, fmt.Sprint(err)
	}
	var temp t.Answer
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("INF WHEREIT error at unmarshal json from infinity %v %v", string(body), err)
		return false, CONNECTION_ERROR
	}
	return temp.IsSuccess, temp.Message
}

func (p *InfinityAPI) Orders() []t.Order {
	body, err := p._request("GetViewData", map[string]string{"params": "[{\"viewName\": \"Taxi.Orders\"}]"})
	if err != nil {
		log.Printf("INF ORDRS error to connect at orders retreieve %v", err)
		return []t.Order{}
	}
	type Orders struct {
		Rows []t.Order `json:"rows"`
	}
	temp := []Orders{}
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("INF ORDRS error at unmarshal json from infinity %s %v", string(body), err)
		p.Connect()
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

func (p *InfinityAPI) Markups() []t.Markup {
	type MarkupsResult struct {
		Rows []t.Markup `json:"rows"`
	}
	var temp []MarkupsResult

	body, err := p._request("GetViewData", map[string]string{"params": "[{\"viewName\": \"Taxi.Markups\"}]"})
	if err != nil {
		log.Print("INF MRKPS error at connection to inf at markups %( %v", err)
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

func (p *InfinityAPI) AddressesAutocomplete(text string) t.AddressPackage {
	body, err := p._request("GetViewData", map[string]string{"params": "[{\"viewName\": \"Taxi.Addresses.Search\", \"params\": [{\"n\": \"SearchText\", \"v\": \"" + text + "\"}]}]"})
	if err != nil {
		log.Printf("INF AA error at connection to inf %v ( ", err)
		return t.AddressPackage{}
	}
	var temp []t.AddressPackage
	//log.Printf("INF ADDRESS AUTOCOMPLETE FOR %s \nRETRIEVE THIS:%s", text, string(body))
	err = json.Unmarshal(body, &temp)
	if err != nil {
		log.Printf("INF AA error at unmarshal json from infinity %s", string(body))
		p.Connect()
		return t.AddressPackage{}
	}
	return temp[0]
}

func get_domain(conn_string string) string {
	u, _ := url.Parse(conn_string)
	host, _, _ := net.SplitHostPort(u.Host)
	return host
}

// GetServices возвращает информацию об услугах доступных для заказа (filterField is set to true!)
// хуй ее знает зачем понадобилось оставить но я плюшкин
func (p *InfinityAPI) GetServices() []InfinityService {
	var tmp []InfinityServices
	body, err := p._request("GetViewData", map[string]string{"params": "[{\"viewName\":\"Taxi.Services\",\"filterField\":{\"n\":\"AvailableToClients\",\"v\":true}}]"})
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		log.Printf("error in unmarshal json, %v", err)
	}
	return tmp[0].Rows
}