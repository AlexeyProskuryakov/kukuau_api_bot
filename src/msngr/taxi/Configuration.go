package taxi

import "msngr/configuration"

type TaxiAPIConfig interface {
	GetHost() string
	GetConnectionStrings() []string
	GetLogin() string
	GetPassword() string
	GetIdService() int64
	GetAPIData() configuration.ApiData
}
