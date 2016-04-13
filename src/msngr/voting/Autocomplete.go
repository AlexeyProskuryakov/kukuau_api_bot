package voting

import (
	"net/http"
	"encoding/json"
	"log"
	"fmt"


	st "msngr/structs"
	fs "github.com/renstrom/fuzzysearch/fuzzy"
	m "msngr"
	"sort"
)

func AutocompleteController(w http.ResponseWriter, r *http.Request, storage *VotingDataHandler, fieldName string, additionalVariants []string) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if r.Method == "GET" {

		params := r.URL.Query()
		query := params.Get("q")

		var results []st.AutocompleteDictItem
		if query != "" {
			splitted := m.AutocompleteSplitter.Split(query, -1)
			strResult := []string{}

			res_map := map[string]bool{}

			for _, splitEl := range splitted {
				for _, variant := range additionalVariants{
					if fs.RankMatchFold(splitEl, variant) > 0 {
						strResult = append(strResult, variant)
					}
				}
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
				log.Printf("res_map: %+v", res_map)
			}
			by := m.ByFuzzyEquals{Data:strResult, Center:query}
			sort.Sort(by)
			results = m.ToAutocompleteItems(by.Data)
		}
		ans, err := json.Marshal(results)
		if err != nil {
			log.Printf("AutocompleteController: ERROR At unmarshal:%+v", err)
		}
		fmt.Fprintf(w, "%s", string(ans))
	}
}
