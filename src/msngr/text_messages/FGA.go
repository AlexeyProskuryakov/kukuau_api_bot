package text_messages
import (
	"bufio"
	"os"
	"log"
	"encoding/json"
	"msngr/utils"
	"path"
	"math/rand"
	"time"
	"strings"
	"regexp"
	"io/ioutil"
)

const (
	PATH = "fga_res"
	CONTENT = "fga.txt"
	SUBSTITUTIONS = "substitutes.json"
)

type SubstitutionsStruct struct {
	Rules map[string][]string `json:"rules"`
}

func (ss SubstitutionsStruct) GetRandomResult(key string) string {
	if subst, ok := ss.Rules[key]; ok {
		r1 := rand.New(rand.NewSource(time.Now().UnixNano()))
		index := r1.Intn(len(subst))
		return subst[index]
	}
	return ""
}

func RegSplit(text string, delimeter string) []string {
	reg := regexp.MustCompile(delimeter)
	indexes := reg.FindAllStringIndex(text, -1)
	laststart := 0
	result := make([]string, len(indexes) + 1)
	for i, element := range indexes {
		result[i] = text[laststart:element[0]]
		laststart = element[1]
	}
	result[len(indexes)] = text[laststart:len(text)]
	return result
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func NewFGA() TextMessageSupplier {
	fga_path := utils.FoundFile(PATH)
	if fga_path == nil {
		log.Printf("Can not found path of FGA resources")
		return &FGA{}
	}
	contentFn := path.Join(*fga_path, CONTENT)
	content, err := readLines(contentFn)
	if err != nil {
		log.Printf("FGA Can not load fga content file")
		return &FGA{}
	}

	substitutesFn := path.Join(*fga_path, SUBSTITUTIONS)
	substitutions_raw, err := ioutil.ReadFile(substitutesFn)
	if err != nil {
		log.Printf("FGA Can not load substitutions config file")
		return &FGA{}
	}
	subst := SubstitutionsStruct{}
	err = json.Unmarshal(substitutions_raw, &subst)
	if err != nil {
		log.Printf("FGA Error at unmarshal substitutions file")
	}

	result := &FGA{Content:content, Substitutes:subst}
	return result
}

type FGA struct {
	Content     []string
	Substitutes SubstitutionsStruct
}

func (fga *FGA) GenerateMessage() string {
	r1 := rand.New(rand.NewSource(time.Now().UnixNano()))
	c_index := r1.Intn(len(fga.Content))
	advise := fga.Content[c_index]

	adv_low := strings.ToLower(advise)
	adv_low_words := RegSplit(adv_low, "[,! ]+")
	for _, word := range adv_low_words {
		if _, ok := fga.Substitutes.Rules[word]; ok {
			replace := fga.Substitutes.GetRandomResult(word)
			advise = strings.Replace(advise, word, replace, -1)
		}
	}
	return advise
}

