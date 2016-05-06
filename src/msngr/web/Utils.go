package web

import (
	"github.com/go-martini/martini"
	"net/http"
	"log"
	"time"
	"github.com/tealeg/xlsx"
	"strings"
	"regexp"
	"fmt"
	"encoding/json"
	"html/template"

	d "msngr/db"
)

func NonJsonLogger() martini.Handler {
	return func(res http.ResponseWriter, req *http.Request, c martini.Context, log *log.Logger) {
		//log.Printf("METHDO: %v, HEADERS: %+v", req.Method, req.Header)
		if req.Method == "GET" || req.Header.Get("Content-Type") != "application/json" {
			start := time.Now()
			addr := req.Header.Get("X-Real-IP")
			if addr == "" {
				addr = req.Header.Get("X-Forwarded-For")
				if addr == "" {
					addr = req.RemoteAddr
				}
			}

			log.Printf("Started %s %s for %s", req.Method, req.URL.Path, addr)

			rw := res.(martini.ResponseWriter)
			c.Next()
			log.Printf("Completed %v %s in %v\n", rw.Status(), http.StatusText(rw.Status()), time.Since(start))
		}
	}
}

func ParseExportXlsx(xlf *xlsx.File, skip_row, skip_cell int) ([][]string, error) {
	result := [][]string{}
	sheet_reg := regexp.MustCompile("([кК]омм?анда\\s*\\d+)|(.*ключ.*)|(^\\d+$)")
	for _, sheet := range xlf.Sheets {
		if sheet != nil {
			sh_name := strings.TrimSpace(strings.ToLower(sheet.Name))
			if sheet_reg.MatchString(sh_name) {
				log.Printf("Processing sheet: %v", sh_name)
				for ir, row := range sheet.Rows {
					if row != nil && ir >= skip_row && len(row.Cells) > skip_cell + 2 {
						log.Printf("Processing row: %+v, row cells: %+v len: %v", row, row.Cells, len(row.Cells))
						key := strings.ToLower(strings.TrimSpace(row.Cells[skip_cell].Value))
						description := strings.TrimSpace(row.Cells[skip_cell + 1].Value)
						next_key_raw := strings.ToLower(strings.TrimSpace(row.Cells[skip_cell + 2].Value))
						if key != "" && description != "" {
							pos := []string{key, description, next_key_raw}
							result = append(result, pos)
						}

					}
				}
			}
		}
	}
	return result, nil
}

type Flash struct {
	Message string
	Type    string
}

func (f *Flash) GetMessage() (string, string) {
	message := f.Message
	fType := f.Type
	f.Message = ""
	f.Type = ""
	return message, fType
}

func (f *Flash) SetMessage(s, t string) {
	f.Message = s
	f.Type = t
}

func GetFuncMap(cName, cId, start_addr string) template.FuncMap {
	return template.FuncMap{
		"eq_s":func(a, b string) bool {
			return a == b
		},
		"stamp_date":func(t time.Time) string {
			return t.Format(time.Stamp)
		},
		"header_name":func() string {
			return cName
		},
		"me":func() string {
			return cId
		},
		"prefix":func() string {
			return start_addr
		},
		"chat_with":func(with string) string {
			return fmt.Sprintf("%v?with=%v", start_addr, with)
		},
		"has_additional_data":func(msg d.MessageWrapper) bool {
			return len(msg.AdditionalData) > 0
		},
		"is_additional_data_valid":func(ad d.AdditionalDataElement) bool {
			return ad.Value != ""
		},
		"get_context":func(adF d.AdditionalFuncElement) string {
			data, _ := json.Marshal(adF.Context)
			return string(data)
		},
	}
}