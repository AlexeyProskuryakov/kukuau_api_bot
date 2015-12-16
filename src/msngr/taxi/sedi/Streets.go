package sedi
import (
	"log"
	"encoding/json"
	"strconv"
	t "msngr/taxi"
	"net/http"
	"msngr"
)

type SediAddressF struct {
	Name string `json:"v"`
	City string `json:"c"`
	Type string `json:"t"`
	Id   int64 `json:"n"`
	Geo  struct {
			 Lat float64 `json:"lat"`
			 Lon float64 `json:"lon"`
		 } `json:"g"`
}

type SediAutocompleteResponse []SediAddressF

func (s SediAutocompleteResponse) ToAddressPackage() t.AddressPackage {
	result := t.AddressPackage{}
	rows := []t.AddressF{}
	for _, s_addr_f := range s {
		rows = append(rows, t.AddressF{
			ID:s_addr_f.Id,
			Name:s_addr_f.Name,
			City:s_addr_f.City,
			Coordinates:t.Coordinates{Lat:s_addr_f.Geo.Lat, Lon:s_addr_f.Geo.Lon}})
	}
	result.Rows = &rows
	return result
}

func (s *SediAPI) AddressesAutocomplete(text string) t.AddressPackage {
	result := t.AddressPackage{}

	req, err := http.NewRequest("GET", s.Host + AUTOCOMPLETE_POSTFIX, nil)
	//	req, err := http.NewRequest("GET", "http://api.sedi.ru" + AUTOCOMPLETE_POSTFIX, nil)
	if err != nil {
		log.Printf("SEDI GET request error in request")
	}
	values := req.URL.Query()

	values.Add("q", "addr")

	values.Add("apikey", s.apikey)
	values.Add("types", "street,object")
	//	values.Add("key", s.appkey)
	//	values.Add("userkey", s.userkey)
	//	values.Add("streetobj", text)
	//	values.Add("city", s.City)
	lat, lon := GeoToString(s.GeoOrbit.Lat, s.GeoOrbit.Lon)
	values.Add("lat", lat)
	values.Add("lon", lon)
	values.Add("radius", strconv.FormatFloat(s.GeoOrbit.Radius, 'f', 0, 64))

	values.Add("search", text)

	req.URL.RawQuery = values.Encode()
	log.Printf("SEDI >>> %v\n", req.URL)

	res, err := s.doReq(req)

	log.Printf("SEDI <<< \n%s\n", res)
	if err != nil {
		log.Printf("SEDI AUTOCMPLETE ERROR: %v", err)
		return result
	}
	response_object := SediAutocompleteResponse{}
	err = json.Unmarshal(res, &response_object)
	if err != nil {
		log.Printf("SEDI AUTOCOMPLETE UNMARSHALL ERROR: %v\nres:[%s]", err, res)
	}
	for _, object := range response_object {
		s.l.Lock()
		s.addressKeys.Add(strconv.FormatInt(object.Id, 10))
		s.l.Unlock()
	}
	return response_object.ToAddressPackage()
}

func (s *SediAPI) GetExternalInfo(key, name string) (*t.AddressF, error) {
	if msngr.DEBUG {
		log.Printf("SEDI get external info : [%v] %v ", key, name)
	}
	s.l.Lock()
	s.addressKeys.Add(key)
	s.l.Unlock()
	id_key, err := strconv.ParseInt(key, 10, 64)
	if err != nil {
		log.Printf("SEDI GET EXTERNAL INFO ERROR: %v", err)
		return nil, err
	}
	return &t.AddressF{ID:id_key, Name:name}, nil
}

func (s *SediAPI) IsHere(key string) bool {
	if msngr.DEBUG {
		s.l.Lock()
		result := s.addressKeys.Contains(key)
		s.l.Unlock()
		log.Printf("SEDI CONTAINS? %v", result)
	}
	return true
}

type SediGeoCodingResult struct {
	SediResponse
	Addresses []SediAddress `json:"Addresses"`
}

func (sgcr SediGeoCodingResult) toAddressPackage() t.AddressPackage {
	rows := []t.AddressF{}
	for _, address_res := range sgcr.Addresses {
		rows = append(rows, t.AddressF{
			ID:address_res.ID,
			City:address_res.CityName,
			Name:address_res.StreetName,
			PostalCode:address_res.PostalCode,
			HouseNumber:address_res.HouseNumber,
			Coordinates: t.Coordinates{Lat:address_res.GeoPoint.Lat, Lon:address_res.GeoPoint.Lon},
		})
	}
	result := t.AddressPackage{}
	result.Rows = &rows
	return result
}

func (s *SediAPI) GeoCoding(lat, lon float64) t.AddressPackage {
	result := t.AddressPackage{}

	lat_s, lon_s := GeoToString(lat, lon)
	res, err := s.getRequest("get_address", map[string]string{"lat":lat_s, "lon":lon_s})
	if err != nil {
		log.Printf("SEDI GEOCODING ERROR: %v", err)
		return result
	}
	result_object := SediGeoCodingResult{}
	err = json.Unmarshal(res, &result_object)
	if err != nil {
		log.Printf("SEDI GEOCODING ERROR UNMARSHAL: %v \n[%s]", err, res)
		return result
	}
	return result_object.toAddressPackage()
}
