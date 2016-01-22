package geo

import (
	"net/http"
	"log"
	"fmt"
	"encoding/json"
	"math"
	"errors"
	"regexp"
	"strings"

	u "msngr/utils"
	t "msngr/taxi"
	s "msngr/taxi/set"


)

var CC_REGEXP = regexp.MustCompilePOSIX("(ул(ица|\\.| )|пр(\\.|оспект|\\-кт)?|пер(\\.|еулок| )|г(ород|\\.|ор\\.| )|обл(асть|\\.| )|р(айон|\\-н )|^с )?")

func StreetsSearchController(w http.ResponseWriter, r *http.Request, i t.AddressSupplier) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	if r.Method == "GET" {

		params := r.URL.Query()
		query := params.Get("q")

		var results []DictItem
		if query != "" {
			if !i.IsConnected() {
				ans, _ := json.Marshal(map[string]string{"error":"true", "details":"service is not avaliable"})
				fmt.Fprintf(w, "%s", string(ans))
				return
			}
			log.Printf("connected. All ok. Start querying for: %+v", query)
			rows := i.AddressesAutocomplete(query).Rows
			if rows == nil {
				return
			}
//			log.Printf("was returned: %v rows", len(*rows))
			for _, nitem := range *rows {
				var item DictItem

				var key string
				if nitem.GID != "" {
					key = nitem.GID
				}else if nitem.OSM_ID != 0 {
					key = fmt.Sprint(nitem.OSM_ID)
				}else {
					key_raw, err := json.Marshal(nitem)
					key = string(key_raw)
					if err != nil {
						log.Printf("SSC: ERROR At unmarshal:%+v", err)
					}
				}
				item.Key = string(key)
				if nitem.ShortName != "" {
					item.Title = fmt.Sprintf("%v %v", nitem.Name, nitem.ShortName)
				}else {
					item.Title = nitem.Name
				}
				item.SubTitle = fmt.Sprintf("%v", u.FirstOf(nitem.Place, nitem.District, nitem.City, nitem.Region))
				results = append(results, item)
			}
		}
		ans, err := json.Marshal(results)
		if err != nil {
			log.Printf("SSC: ERROR At unmarshal:%+v", err)
		}
//		fmt.Fprintf(w, "%s", string(ans))
	}
}

func GetSetOfAddressF(a t.AddressF) s.Set {
	external_set := s.NewSet()
	nitem := a
	if nitem.ShortName == "пл" {
		nitem.Name = fmt.Sprintf("площадь %s", nitem.Name)
	}
	AddStringToSet(external_set, nitem.Name)
	AddStringToSet(external_set, nitem.Region)
	AddStringToSet(external_set, nitem.City)
	AddStringToSet(external_set, nitem.District)
	AddStringToSet(external_set, nitem.Place)
	return external_set
}


type DictItem struct {
	Key      string `json:"key"`
	Title    string `json:"title"`
	SubTitle string `json:"subtitle"`
}

type InPlace struct {
	StreetId   int64 `json:"ID"`
	RegionId   int64 `json:"IDRegion"`
	DistrictId *int64 `json:"IDDistrict"`
	CityId     *int64 `json:"IDCity"`
	PlaceId    *int64 `json:"IDPlace"`
}

func AddStringToSet(set s.Set, element string) (string, error) {
	result := ClearAddressString(element)
	if result != "" {
		set.Add(result)
		return result, nil
	}
	return element, errors.New(fmt.Sprintf("can not imply %+v ==> %+v", element, result))
}

func ClearAddressString(element string) (string) {
	result := strings.ToLower(element)
	result_raw := CC_REGEXP.ReplaceAllString(result, "")
	result = string(result_raw)
	result = strings.TrimSpace(result)
	return result
}


func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta / 2), 2)
}
// Distance function returns the distance (in meters) between two points of
// a given longitude and latitude relatively accurately (using a spherical
// approximation of the Earth) through the Haversin Distance Formula for
// great arc distance on a sphere with accuracy for small distances
// point coordinates are supplied in degrees and converted into rad. in the func
// distance returned is METERS!!!!!!
// http://en.wikipedia.org/wiki/Haversine_formula
func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2 - la1) + math.Cos(la1) * math.Cos(la2) * hsin(lo2 - lo1)
	result := 2 * r * math.Asin(math.Sqrt(h))
	return result
}
