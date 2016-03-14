package taxi

type ConnectInterface interface {
	IsConnected() bool
	Connect()
}

type TaxiInterface interface {
	/**
	Сия херь переопределяется остальными таксоматорскими АПИшками...
	 */
	ConnectInterface
	NewOrder(order NewOrderInfo) Answer //создани нового заказа
	CalcOrderCost(order NewOrderInfo) (int, string) //рассчет стоимости заказа
	CancelOrder(order_id int64) (bool, string, error) //отмена заказа
	//For watching by current orders;
	Orders() []Order //запрос списка текущих заказов
	//For feedback by order info:
	Feedback(f Feedback) (bool, string)
	//for cars with mapping id in CarInfo with orders car.id
	GetCarsInfo() []CarInfo
	//additional
	WriteDispatcher(message string) (bool, string) //написать диспетчеру
	CallbackRequest(phone string) (bool, string) //запросить обратный звонок
	WhereIt(order_id int64) (bool, string) //оповестить что клиент не видит автомобиль
	Markups() []Markup//наценки

}

type AddressSupplier interface {
	//For adress autocomplete
	ConnectInterface
	AddressesAutocomplete(query string) AddressPackage
}

type AddressHandler interface {
	ConnectInterface
	GetExternalInfo(key, name string) (*AddressF, error)
	IsHere(key string) bool
}

type AddressGeoSupplier interface {
	GeoCoding(lat, lon float64) AddressPackage
}

type ExternalApiMixin struct {
	API TaxiInterface
}




