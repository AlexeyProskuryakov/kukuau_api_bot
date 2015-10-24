package taxi

type ConnectInterface interface {
	IsConnected() bool
}

type TaxiInterface interface {
	ConnectInterface
	NewOrder(order NewOrder) Answer
	CancelOrder(order_id int64) (bool, string)
	CalcOrderCost(order NewOrder) (int, string)
	Orders() []Order
	Feedback(f Feedback) (bool, string)
	GetCarsInfo() []CarInfo
}

type AddressSupplier interface {
	ConnectInterface
	AddressesSearch(query string) AddressPackage
}


type ExternalApiMixin struct {
	API TaxiInterface
}




