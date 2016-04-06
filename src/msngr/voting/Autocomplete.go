package voting

import (
	"net/http"
	"encoding/json"
	"log"
	"fmt"
	"regexp"

	st "msngr/structs"
	fs "github.com/renstrom/fuzzysearch/fuzzy"
	"sort"
)

var reg = regexp.MustCompile("[^A-Za-z0-9А-Яа-я]")

type ByFuzzyEquals struct {
	Data   []string
	Center string
}

func (s ByFuzzyEquals) Len() int {
	return len(s.Data)
}
func (s ByFuzzyEquals) Swap(i, j int) {
	s.Data[i], s.Data[j] = s.Data[j], s.Data[i]
}
func (s ByFuzzyEquals) Less(i, j int) bool {
	return fs.RankMatchFold(s.Data[i], s.Center) > fs.RankMatchFold(s.Data[j], s.Center)
}

func AutocompleteController(w http.ResponseWriter, r *http.Request, storage *VotingDataHandler, fieldName string) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if r.Method == "GET" {

		params := r.URL.Query()
		query := params.Get("q")

		var results []st.AutocompleteDictItem
		if query != "" {
			splitted := reg.Split(query, -1)
			strResult := []string{}

			res_map := map[string]bool{}

			for _, splitEl := range splitted {
				res, err := storage.TextFoundByCompanyField(splitEl, fieldName)
				if err != nil {
					log.Printf("Autocomplete controller ERROR at get data from db: %v", err)
					continue
				}
				for _, foundEl := range res {
					if _, ok := res_map[foundEl]; !ok {
						res_map[foundEl] = true
						strResult = append(strResult, foundEl)
					}
				}
			}
			by := ByFuzzyEquals{Data:strResult, Center:query}
			sort.Sort(by)
			for _, resEl := range by.Data {
				results = append(results, st.AutocompleteDictItem{Key:resEl, Title:resEl})
			}
		}
		ans, err := json.Marshal(results)
		if err != nil {
			log.Printf("AutocompleteController: ERROR At unmarshal:%+v", err)
		}
		fmt.Fprintf(w, "%s", string(ans))
	}
}
