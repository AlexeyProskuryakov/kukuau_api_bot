package taxi


type TaxiConfig interface {
	GetHost() string
	GetConnectionString() string
	GetLogin() string
	GetPassword() string
}

type ApiParams struct {
	Name              string `json:"name"`
	Host              *string `json:"host"`
	Login             *string `json:"login"`
	Password          *string `json:"password"`
	ConnectionsString *string `json:"connection_string"`
}

func (api *ApiParams) GetHost() string {
	return api.Host
}
func (api *ApiParams) GetConnectionString() string {
	return api.ConnectionsString
}
func (api *ApiParams) GetLogin() string {
	return api.Login
}
func (api *ApiParams) GetPassword() string {
	return api.Password
}

