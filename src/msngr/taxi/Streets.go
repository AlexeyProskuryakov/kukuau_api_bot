package taxi

import (
	"net/http"
	"log"
	"fmt"
	"encoding/json"
	"msngr/utils"
)


func StreetsSearchController(w http.ResponseWriter, r *http.Request, i AddressSupplier) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	log.Println("Searching address...")
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

			rows := i.AddressesSearch(query).Rows
			for _, nitem := range rows {
				var item DictItem
				var err error
				t, err := json.Marshal(nitem)
				utils.CheckErr(err)
				item.Key = string(t)
				item.Title = fmt.Sprintf("%v %v", nitem.Name, nitem.ShortName)
				item.SubTitle = fmt.Sprintf("%v", utils.FirstOf(nitem.Place, nitem.District, nitem.City, nitem.Region))
				results = append(results, item)
			}
		}
		ans, err := json.Marshal(results)
		utils.CheckErr(err)
		fmt.Fprintf(w, "%s", string(ans))
	}
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

//helpers for forming
// destination and delivery on infinity results after street search request
func GetDeliveryHelper(info string, house string, entrance *string) Delivery {
	log.Printf("0 NO delivery marshalled: %+v\n and parameters: house: %+v, entrance: %q", info, house, entrance)
	in := InPlace{}
	err := json.Unmarshal([]byte(info), &in)
	utils.CheckErr(err)
	result := Delivery{IdStreet:in.StreetId, IdRegion:in.RegionId, House:house, Entrance:entrance,
		IdCity:in.CityId,
		IdPlace:in.PlaceId,
		IdDistrict:in.DistrictId,
	}
	log.Printf("1 NO delivery: %+v", result)
	return result
}

func GetDestinationHelper(info string, house string) Destination {
	log.Printf("0 NO destination marshalled: %+v", info)
	in := InPlace{}
	err := json.Unmarshal([]byte(info), &in)
	utils.CheckErr(err)
	result := Destination{IdStreet:in.StreetId, IdRegion:in.RegionId, House:house,
		IdCity:in.CityId,
		IdPlace:in.PlaceId,
		IdDistrict:in.DistrictId,
	}
	log.Printf("1 NO destination: %+v", result)
	return result
}
