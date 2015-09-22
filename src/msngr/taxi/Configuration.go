package taxi


type TaxiConfig interface {
	GetHost() string
	GetConnectionString() string
	GetLogin() string
	GetPassword() string
}

type ApiParams struct {
	Name string `json:"name"`
	Data struct {
			 Host              string `json:"host"`
			 Login             string `json:"login"`
			 Password          string `json:"password"`
			 ConnectionsString string `json:"connection_string"`
		 } `json:"data"`

}

func (api ApiParams) GetHost() string {
	return api.Data.Host
}
func (api ApiParams) GetConnectionString() string {
	return api.Data.ConnectionsString
}
func (api ApiParams) GetLogin() string {
	return api.Data.Login
}
func (api ApiParams) GetPassword() string {
	return api.Data.Password
}

