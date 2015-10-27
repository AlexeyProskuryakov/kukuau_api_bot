package taxi
import (
	"msngr/utils"
	"msngr/db"
	"fmt"
)

type AddressF struct {
	GID        string
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

type AddressPackage struct {
	Rows *[]AddressF `json:"rows"`
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
	IdRegion      int64   `json:"idRegion"`                 // : <Идентификатор региона (Int64)>,
	IdStreet      int64   `json:"idStreet"`                 // : <Идентификатор улицы (Int64)>,
	House         string  `json:"house"`                    // : <№ дома (строка)>,

	Lat           *float64 `json:"lat,omitempty"`           // : <Широта координаты адреса (при указании места на карте). Если указано, информация о доставке по указанию и адресе игнорируется>,
	Lon           *float64 `json:"lon,omitempty"`           // : <Долгота координаты адреса (при указании места на карте). Если указано, информация о доставке по указанию и адресе игнорируется>,"isByDirection" : <Заказ машины с указанием пункта назначения при подаче (если задано в true,информация о адресе игнонрируются)>,

	IdDistrict    *int64   `json:"idDistrict"`              // : <Идентификатор района (Int64)>,
	IdCity        *int64   `json:"idCity"`                  // : <Идентификатор города (Int64)>,
	IdPlace       *int64   `json:"idPlace"`                 // : <Идентификатор поселения (Int64)>,

	Building      *string  `json:"building,omitempty"`      // : <Строение (строка)>,
	Fraction      *string  `json:"fraction,omitempty"`      // : <Корпус (строка)>,
	Entrance      *string  `json:"entrance,omitempty"`      // : <Подъезд (строка)>,
	Apartment     *string  `json:"apartment,omitempty"`     // : <№ квартиры (строка)> ,

	IdAddress     *string  `json:"idAddress,omitempty"`     // : <Идентификатор существующего описания адреса (адрес дома или объекта)>,
	IdFastAddress *string  `json:"idFastAddress,omitempty"` // : <ID быстрого адреса. Дополнительное информационное поле, описывающее быстрый адрес, связанный с указанным адресом. Значение учитывается только при указании idAddress>
}

type Delivery struct {

	IdRegion      int64 `json:"idRegion"`                   // <Идентификатор региона (Int64)>,
	IdStreet      int64  `json:"idStreet"`                  // : <Идентификатор улицы (Int64)>,
	House         string `json:"house"`                     // : <№ дома (строка)>,

	Lat           *float64 `json:"lat,omitempty"`           // : <Широта координаты адреса (при указании места на карте). Если указано, информация о адресе игнорируется>,
	Lon           *float64 `json:"lon,omitempty"`           // : <Долгота координаты адреса (при указании места на карте). Если указано, информация о адресе игнорируется>,

	Building      *string `json:"building,omitempty"`       // : <Строение (строка)>,
	Fraction      *string `json:"fraction,omitempty"`       // : <Корпус (строка)>,
	Entrance      *string `json:"entrance,omitempty"`       // : <Подъезд (строка)>,
	Apartment     *string `json:"apartment,omitempty"`      // : <№ квартиры (строка)>,


	IdPlace       *int64 `json:"idPlace"`                   //IdPlace       int64  `json:"idPlace"`       //: <Идентификатор поселения (Int64)>,
	IdCity        *int64 `json:"idCity"`                    //IdCity        int64  `json:"idCity"`        // : <Идентификатор города (Int64)>,
	IdDistrict    *int64 `json:"idDistrict"`                //IdDistrict    int64  `json:"idDistrict"`    // : <Идентификатор района (Int64)>,

	IdAddress     *string `json:"idAddress,omitempty"`      // <Идентификатор существующего описания адреса (адрес дома или объекта)>,
	IdFastAddress *string  `json:"idFastAddress,omitempty"` //: <ID быстрого адреса. Дополнительное информационное поле, описывающее быстрый адрес, связанный с указанным адресом. Значение учитывается только при указании idAddress>

}

type NewOrder struct {
																 //request
	Phone           string `json:"phone"`
	DeliveryTime    *string `json:"deliveryTime,omitempty"`      //<Время подачи в формате yyyy-MM-dd HH:mm:ss>
	DeliveryMinutes int64  `json:"deliveryMinutes"`              // <Количество минут до подачи (0-сейчас, но не менее минимального времени на подачу, указанного в настройках системы), не анализируется если задано поле deliveryTime >
	IdService       string  `json:"idService"`                    //<Идентификатор услуги заказа (не может быть пустым)>
	Notes           *string `json:"notes,omitempty"`             // <Комментарий к заказу>
																 //Markups           [2]int64 `json:"markups"`           // <Массив идентификаторов наценок заказа>
	Attributes      *[2]int64      `json:"attributes,omitempty"` // <Массив идентификаторов дополнительных атрибутов заказа>
	Delivery        Delivery      `json:"delivery"`              // Инфомация о месте подачи машины
	Destinations    []Destination `json:"destinations"`          // Пункты назначения заказа (массив, не может быть пустым)
	IsNotCash       *bool          `json:"isNotCash,omitempty"`  //: // Флаг безналичного заказа <true или false (bool)>
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

func (o *Order) ToOrderData() db.OrderData {
	odc, _ := utils.ToMap(o, "json")
	return db.NewOrderData(odc)
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

func (car CarInfo) String() string {
	return fmt.Sprintf("%v %v с номером %v", car.Color, car.Model, car.Number)
}

type Feedback struct {
	IdOrder int64  `json:"idOrder"`
	Rating  int    `json:"rating"`
	Notes   string `json:"notes"`
}

const (
	ORDER_CREATED = 1
	ORDER_ASSIGNED = 2
	ORDER_CAR_SET_OUT = 3
	ORDER_CLIENT_WAIT = 4
	ORDER_IN_PROCESS = 5

	ORDER_PAYED = 7
	ORDER_CANCELED = 9
	ORDER_NOT_CREATED = 13

	ORDER_NOT_PAYED = 8
	ORDER_FIXED = 12

//	ID_SERVICE = 5001753333
)
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


func IsOrderNotAvailable(state int) bool {
	return utils.In(state, []int{0, 7, 9, 13})
}