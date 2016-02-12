package web

import (
	"github.com/go-martini/martini"
	"net/http"
	"log"
	"time"
	"github.com/tealeg/xlsx"
"strings"
	"regexp"
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
			if sheet_reg.MatchString(sh_name){
				for ir, row := range sheet.Rows {
					if row != nil && ir >= skip_row {
						key := row.Cells[skip_cell].Value
						description := row.Cells[skip_cell + 1].Value
						next_key_raw := row.Cells[skip_cell + 2].Value
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

func ValidateExport(keys [][]string){

}