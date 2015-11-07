package taxi


type TaxiAPIConfig interface {
	GetHost() string
	GetConnectionString() string
	GetLogin() string
	GetPassword() string
	GetIdService() string
}
