package taxi


type TaxiAPIConfig interface {
	GetHost() string
	GetConnectionStrings() []string
	GetLogin() string
	GetPassword() string
	GetIdService() string
}
