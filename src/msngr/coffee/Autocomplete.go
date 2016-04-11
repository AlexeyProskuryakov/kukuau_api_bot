package coffee

import (
	"net/http"
	"encoding/json"
	"log"
	"fmt"

	st "msngr/structs"
	m "msngr"
)

func AutocompleteController(w http.ResponseWriter, r *http.Request, storage *CoffeeConfigHandler, fieldName, companyName string) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	if r.Method == "GET" {

		params := r.URL.Query()
		query := params.Get("q")

		var results []st.AutocompleteDictItem
		if query != "" {
			coffeeConfig, err := storage.GetConfig(companyName)
			if err != nil {
				log.Printf("coffee autocomplete retrieve data ERROR: %v", err)
			}
			result := coffeeConfig.Autocomplete(query, fieldName)
			results = m.ToAutocompleteItems(result)
		}
		ans, err := json.Marshal(results)
		if err != nil {
			log.Printf("AutocompleteController: ERROR At unmarshal:%+v", err)
		}
		fmt.Fprintf(w, "%s", string(ans))
	}
}
