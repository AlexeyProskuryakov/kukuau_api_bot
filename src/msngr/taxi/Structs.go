package taxi

import (
	"msngr/utils"
	"msngr/db"

	"fmt"
	"time"
	"strings"
)

type Coordinates struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type AddressF struct {
	Coordinates Coordinates

	OSM_ID      int64  `json:"osm_id"`
	GID         string
	ID          int64  `json:"ID"`

	Name        string `json:"Name"`
	City        string `json:"City"`

	IDParent    int64  `json:"IDParent,omitempty"`
	ShortName   string `json:"ShortName,omitempty"`
	ItemType    int64  `json:"ItemType,omitempty"`
	FullName    string `json:"FullName"`
	IDRegion    int64  `json:"IDRegion"`
	IDDistrict  int64  `json:"IDDistrict"`
	IDCity      int64  `json:"IDCity"`
	IDPlace     int64  `json:"IDPlace"`
	Region      string `json:"Region,omitempty"`
	District    string `json:"District,omitempty"`
	Place       string `json:"Place,omitempty"`

	PostalCode  string
	HouseNumber string
}

func (a AddressF) String() string {
	return fmt.Sprintf("[%v]g[%v]o[%v] '%v' (%v) [%v] \n\tHouse:%v City:%v Region:%v District:%v Place:%v\nIDS:\tParent: %v, Region: %v, District: %v, City: %v, Place: %v\n",
		a.ID, a.GID, a.OSM_ID,
		a.Name, a.FullName, a.ShortName, a.HouseNumber, a.City, a.Region, a.District, a.Place,
		a.IDParent, a.IDRegion, a.IDDistrict, a.IDCity, a.IDPlace)
}

type AddressPackage struct {
	Rows *[]AddressF `json:"rows"`
}

func (ap AddressPackage) String() string {
	if ap.Rows != nil {
		var buf []string
		for i, row := range *ap.Rows {
			buf = append(buf, fmt.Sprintf("%v: %+v", i, row))
		}
		return fmt.Sprintf("Address Package:\n%s", strings.Join(buf, "\n"))
	} else {
		return fmt.Sprintf("Address Package: [empty]")
	}
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

type Destination struct {
	IdRegion      int64   `json:"idRegion"`                // : <Идентификатор региона (Int64)>,
	IdStreet      int64   `json:"idStreet"`                // : <Идентификатор улицы (Int64)>,

	House         string  `json:"house"`                   // : <№ дома (строка)>,
	Street        string
	City          string

	Lat           float64 `json:"lat,omitempty"`           // : <Широта координаты адреса (при указании места на карте). Если указано, информация о доставке по указанию и адресе игнорируется>,
	Lon           float64 `json:"lon,omitempty"`           // : <Долгота координаты адреса (при указании места на карте). Если указано, информация о доставке по указанию и адресе игнорируется>,"isByDirection" : <Заказ машины с указанием пункта назначения при подаче (если задано в true,информация о адресе игнонрируются)>,

	IdDistrict    *int64   `json:"idDistrict"`             // : <Идентификатор района (Int64)>,
	IdCity        *int64   `json:"idCity"`                 // : <Идентификатор города (Int64)>,
	IdPlace       *int64   `json:"idPlace"`                // : <Идентификатор поселения (Int64)>,

	Building      string  `json:"building,omitempty"`      // : <Строение (строка)>,
	Fraction      string  `json:"fraction,omitempty"`      // : <Корпус (строка)>,
	Entrance      string  `json:"entrance,omitempty"`      // : <Подъезд (строка)>,
	Apartment     string  `json:"apartment,omitempty"`     // : <№ квартиры (строка)> ,

	IdAddress     string  `json:"idAddress,omitempty"`     // : <Идентификатор существующего описания адреса (адрес дома или объекта)>,
	IdFastAddress string  `json:"idFastAddress,omitempty"` // : <ID быстрого адреса. Дополнительное информационное поле, описывающее быстрый адрес, связанный с указанным адресом. Значение учитывается только при указании idAddress>
}

func (d Destination) String() string {
	return fmt.Sprintf("\tDestination [%v] City: %v, Street: %v, House: %v, Entrance: %v\n",
		d.IdAddress,
		d.City,
		d.Street,
		d.House,
		d.Entrance,
	)
}

type Delivery Destination

func (d Delivery) String() string {
	return fmt.Sprintf("\tDelivery [%v] City: %v, Street: %v, House: %v, Entrance: %v\n",
		d.IdAddress,
		d.City,
		d.Street,
		d.House,
		d.Entrance,
	)
}

type NewOrderInfo struct {
								   //request
	Phone           string `json:"phone"`
	DeliveryTime    string `json:"deliveryTime,omitempty"`     //<Время подачи в формате yyyy-MM-dd HH:mm:ss>
	DeliveryMinutes int  `json:"deliveryMinutes"`              // <Количество минут до подачи (0-сейчас, но не менее минимального времени на подачу, указанного в настройках системы), не анализируется если задано поле deliveryTime >
	IdService       int64  `json:"idService"`                  //<Идентификатор услуги заказа (не может быть пустым)>
	Notes           string `json:"notes,omitempty"`            // <Комментарий к заказу>
	Markups         []string    `json:"markups,omitempty"`     //Markups           [2]int64 `json:"markups"`           // <Массив идентификаторов наценок заказа>
	Attributes      []int64      `json:"attributes,omitempty"` // <Массив идентификаторов дополнительных атрибутов заказа>
	Delivery        *Delivery      `json:"delivery"`           // Инфомация о месте подачи машины
	Destinations    []*Destination `json:"destinations"`       // Пункты назначения заказа (массив, не может быть пустым)
	IsNotCash       bool          `json:"isNotCash,omitempty"` //: // Флаг безналичного заказа <true или false (bool)>
}

func (o NewOrderInfo) String() string {
	return fmt.Sprintf("New Order Info for %s;" +
	"\n\t FROM %v " +
	"\n\t TO %v \n",
		o.Phone,
		o.Delivery,
		o.Destinations,
	)
}

type Order struct {
							    /**
							    Key fields is:
							    ID, State, Cost, TimeArrival, TimeDelivery, IDCar
							     */
	ID                int64  `json:"ID"`                // ID
	State             int    `json:"State"`             //Состояние заказа
	Cost              int    `json:"Cost"`              //Стоимость
	IsNotCash         bool   `json:"IsNotCash"`         //Безналичный заказ (bool)
	IsPrevious        int    `json:"IsPrevious"`        //Тип заказа (0 –активный, 1-предварительный, 2-предварительный ставший активным)
	LastStateTime     string `json:"LastStateTime"`     //Дата-Время последнего и\зменения состояния
	DeliveryTime      string `json:"DeliveryTime"`      //Требуемое время подачи машины
	Distance          int    `json:"Distance"`          //Расстояние км (если оно рассчитано системой)
	ArrivalTime       string `json:"TimeOfArrival"`     //Прогнозируемое время прибытия машины на заказ
	IDDeliveryAddress int64  `json:"IDDeliveryAddress"` //ID адреса подачи
	DeliveryStr       string `json:"DeliveryStr"`       //Адрес подачи в виде текста
	DestinationsStr   string `json:"DestinationsStr"`   //Пункты назначения в виде текста (с учетом настроек отображения в диспетчерской: Первый/Последний/Все)
	IDCar             int64  `json:"IDCar"`             //ID машины
	IDService         int64  `json:"IdService"`         //ID услуги
	Car               string `json:"Car"`               //Позывной машины
	Service           string `json:"Service"`           //Услуга
	Drivers           string `json:"Drivers"`           //ФИО Водителя

	TimeDelivery      *time.Time                        //Требуемое время подачи машины (в виде времени)
	TimeArrival       *time.Time                        //Прогнозируемое время прибытия машины на заказ (в виде времени)
}

func (o *Order) ToOrderData() *db.OrderData {
	odc, _ := utils.ToMap(o, "json")
	result := db.NewOrderData(odc)
	return &result
}

func (o Order) String() string {
	return fmt.Sprintf("Order [%v]\n\tState:%v, Cost:%v, TimeArrival:%+v, TimeDelivery:%+v, IdCar:%v",
		o.ID,
		o.State, o.Cost, o.TimeArrival, o.TimeDelivery, o.IDCar,
	)
}

type Markup struct {
	ID            int64 `json:"ID"`
	Name          string `json:"Name"`
	Type          int64 `json:"Type"`
	Value         int `json:"Value"`
	Accessibility int `json:"Accessibility"`
}

type AnswerContent struct {
	Id      int64  `json:"id"`     // :7007330031,
	Name    string `json:"name"`
	Login   string `json:"login"`
	Number  int64  `json:"number"` // :406
	Cost    int    `json:"cost"`
	Details string `json:"details"`
}

type Answer struct {
	IsSuccess bool   `json:"isSuccess"`
	Message   string `json:"message"`
	Content   AnswerContent `json:"content"`
}

type CarInfo struct {
	/**
	Key fields: color, model, number, id
	 */
	ID     int64   `json:"id"`
	Number string  `json:"Number"`
	Color  string  `json:"Color"`
	Model  string  `json:"Model"`
	Lat    float64 `json:"Lat"`
	Lon    float64 `json:"Lon"`
}

func (car CarInfo) String() string {
	var result string
	if car.Number != "" && car.Color != "" && car.Model != "" {
		result = fmt.Sprintf("%v %v с номером %v ", car.Color, car.Model, car.Number)
	} else if car.Model != "" && car.Number != "" {
		result = fmt.Sprintf("%v %v", car.Model, car.Number)
	} else if car.Model != "" || car.Number != "" {
		result = fmt.Sprintf("%v %v", car.Model, car.Number)
	} else {
		result = "не определена"
	}
	return result
}

type Feedback struct {
	Phone        string
	IdOrder      int64  `json:"idOrder"`
	Rating       int    `json:"rating"`
	FeedBackText string `json:"notes"`
}

const (
	ORDER_CREATED = 1
	ORDER_ASSIGNED = 2
	ORDER_CAR_SET_OUT = 3
	ORDER_CLIENT_WAIT = 4
	ORDER_IN_PROCESS = 5
	ORDER_DOWNTIME = 6

	ORDER_PAYED = 7
	ORDER_CANCELED = 9
	ORDER_NOT_CREATED = 13

	ORDER_NOT_PAYED = 8
	ORDER_FIXED = 12

	ORDER_ERROR = -1
)

var InfinityStatusesName = map[int]string{
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

func IsOrderNotActual(state int) bool {
	return utils.In(state, []int{0, ORDER_PAYED, ORDER_CANCELED, ORDER_NOT_CREATED, ORDER_NOT_PAYED})
}