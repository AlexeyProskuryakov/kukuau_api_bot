package taxi

type ConnectInterface interface {
	IsConnected() bool
}

type TaxiInterface interface {
	ConnectInterface
	NewOrder(order NewOrderInfo) Answer //создани нового заказа
	CalcOrderCost(order NewOrderInfo) (int, string) //рассчет стоймости заказа
	CancelOrder(order_id int64) (bool, string) //отмена заказа
	Orders() []Order //запрос списка текущих заказов
	Feedback(f Feedback) (bool, string) //отправка отзыва

	GetCarsInfo() []CarInfo //запрос списка автомобилей

	WriteDispatcher(message string) (bool, string) //написать диспетчеру
	CallbackRequest(phone string) (bool, string) //запросить обратный звонок
	WhereIt(order_id int64) (bool, string) //оповестить что клиент не видит автомобиль


}

type AddressSupplier interface {
	ConnectInterface
	AddressesSearch(query string) AddressPackage
}


type ExternalApiMixin struct {
	API TaxiInterface
}




