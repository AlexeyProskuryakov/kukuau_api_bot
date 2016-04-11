package msngr

import (
	"regexp"
	fs "github.com/renstrom/fuzzysearch/fuzzy"
	"msngr/structs"
)

var AutocompleteSplitter = regexp.MustCompile("[^A-Za-z0-9А-Яа-я]")

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
	return fs.RankMatchFold(s.Center, s.Data[i]) > fs.RankMatchFold(s.Center, s.Data[j])
}

func ToAutocompleteItems(data []string) []structs.AutocompleteDictItem {
	results := []structs.AutocompleteDictItem{}
	for _, resEl := range data {
		results = append(results, structs.AutocompleteDictItem{Key:resEl, Title:resEl})
	}
	return results
}