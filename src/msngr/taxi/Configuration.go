package taxi


type TaxiAPIConfig interface {
	GetHost() string
	GetConnectionString() string
	GetLogin() string
	GetPassword() string
	GetIdService() string
}

type ApiParams struct {
	Name string `json:"name"`
	Data struct {
			 Host              string `json:"host"`
			 Login             string `json:"login"`
			 Password          string `json:"password"`
			 ConnectionsString string `json:"connection_string"`
			 IdService         string `json:"id_service"`
			 BearerToken       string `json:"bearer_token"`
		 } `json:"data"`
	Fake struct {
			 SendedStates []int `json:"sended_states"`
			 SleepTime    int `json:"sleep_time"`
		 } `json:"fake"`
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
func (api ApiParams) GetIdService() string {
	return api.Data.IdService
}

type TaxiConfig struct {
	Api         ApiParams `json:"api"`
	DictUrl     string `json:"dict_url"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Information struct {
					Phone string `json:"phone"`
					Text  string `json:"text"`
				} `json:"information"`
	GeoOrbit    TaxiGeoOrbit `json:"geo_orbit"`
}