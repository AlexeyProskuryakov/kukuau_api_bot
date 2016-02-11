package console

import (
	"net/http"

	"encoding/json"

	"io/ioutil"
	"fmt"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/martini-contrib/render"
	"msngr/structs"

	"html/template"

	d "msngr/db"
	c "msngr/configuration"
	u "msngr/utils"
	ntf "msngr/notify"
	"gopkg.in/mgo.v2/bson"
	"msngr/quests"
	w "msngr/web"
	"log"
	"regexp"
	"github.com/tealeg/xlsx"
	"strconv"
)

func GetKeysInfo(err_text string, qs *quests.QuestStorage) map[string]interface{} {
	var keys []quests.Step
	var e error
	result := map[string]interface{}{}

	keys, e = qs.GetAllKeys()

	if e != nil || err_text != "" {
		result["is_error"] = true
		if e != nil {
			result["error_text"] = e.Error()
		} else {
			result["error_text"] = err_text
		}
	}
	result["keys"] = keys
	return result
}
func Run(addr string, notifier *ntf.Notifier, db *d.MainDb, cs c.ConfigStorage, qs *quests.QuestStorage) {
	m := martini.New()
	m.Use(w.NonJsonLogger())

	m.Use(martini.Recovery())
	m.Use(render.Renderer(render.Options{
		Layout: "console/layout",
		Extensions: []string{".tmpl", ".html"},
		Charset: "UTF-8",
		IndentJSON: true,
		IndentXML: true,
		Funcs:[]template.FuncMap{
			template.FuncMap{
				"eq_s":func(a, b string) bool {
					return a == b
				},
			},
		},
	}))

	m.Use(auth.BasicFunc(func(username, password string) bool {
		usr, _ := db.Users.GetUser(bson.M{"user_name":username, "role":MANAGER})
		if usr != nil {
			return u.PHash(password) == usr.Password
		}
		return username == default_user && password == default_pwd
	}))

	m.Use(martini.Static("static"))
	r := martini.NewRouter()

	r.Get("/", func(user auth.User, render render.Render) {
		render.HTML(200, "console/index", map[string]interface{}{})
	})

	xlsFileReg := regexp.MustCompile(".+\\.xlsx?")

	r.Post("/load/up", func(render render.Render, request *http.Request) {
		file, header, err := request.FormFile("file")

		log.Printf("Form file information: file: %+v \nheader:%v, %v\nerr:%v", file, header.Filename, header.Header, err)

		if err != nil {
			render.HTML(200, "quests/new_keys", GetKeysInfo(fmt.Sprintf("Ошибка загрузки файлика: %v", err), qs))
			return
		}
		defer file.Close()

		data, err := ioutil.ReadAll(file)
		if err != nil {
			render.HTML(200, "quests/new_keys", GetKeysInfo(fmt.Sprintf("Ошибка загрузки файлика: %v", err), qs))
			return
		}

		if xlsFileReg.MatchString(header.Filename) {
			xlFile, err := xlsx.OpenBinary(data)
			log.Printf("file: %+v, err: %v", xlFile, err)
			if err != nil || xlFile == nil {
				render.HTML(200, "quests/new_keys", GetKeysInfo(fmt.Sprintf("Ошибка обработки файлика: %v", err), qs))
				return
			}
			skip_rows, _ := strconv.Atoi(request.FormValue("skip-rows"))
			skip_cols, _ := strconv.Atoi(request.FormValue("skip-cols"))

			parse_res, _ := w.ParseExportXlsx(xlFile, skip_rows, skip_cols)
			for _, prel := range parse_res {
				qs.AddKey(prel[0], prel[1], prel[2])
			}
		} else {
			render.HTML(200, "quests/new_keys", GetKeysInfo("Файл имеет не то расширение :(", qs))
		}

		render.Redirect("/new_keys")
	})
	r.Post("/configuration", func(request *http.Request, render render.Render) {
		input, err := ioutil.ReadAll(request.Body)
		defer request.Body.Close()
		if err != nil {
			render.JSON(500, map[string]interface{}{"Error":fmt.Sprintf("Can not read request body. Because: %v", err)})
			return
		}
		type CommandInfo struct {
			Provider string `json:"provider"`
			Name     string `json:"name"`
			Command  structs.OutCommand `json:"command"`
		}
		commandInfo := CommandInfo{}
		err = json.Unmarshal(input, &commandInfo)
		if err != nil {
			render.JSON(500, map[string]interface{}{"Error":fmt.Sprintf("Can not unmarshall input to command info. Because: %v", err)})
			return
		}
		cs.SaveCommand(commandInfo.Provider, commandInfo.Name, commandInfo.Command)
		render.JSON(200, map[string]interface{}{"OK":true})
	})

	r.Get("/new_keys", func(render render.Render) {
		render.HTML(200, "console/new_keys", GetKeysInfo("", qs))
	})

	r.Post("/add_key", func(user auth.User, render render.Render, request *http.Request) {
		start_key := request.FormValue("start-key")
		next_key := request.FormValue("next-key")
		description := request.FormValue("description")

		log.Printf("CONSOLE WEB add key %s -> %s -> %s", start_key, description, next_key)
		if start_key != "" && description != "" {
			key, err := qs.AddKey(start_key, description, next_key)
			log.Printf("QW is error? %v key: %v", err, key)
			render.Redirect("/new_keys")
		} else {
			render.HTML(200, "console/new_keys", GetKeysInfo("Невалидные значения ключа или ответа", qs))
		}
	})

	r.Post("/delete_key/:key", func(params martini.Params, render render.Render) {
		key := params["key"]
		err := qs.DeleteKey(key)
		log.Printf("CONSOLE WEB will delete %v (%v)", key, err)
		render.Redirect("/new_keys")
	})

	r.Post("/update_key/:key", func(params martini.Params, render render.Render, request *http.Request) {
		key_id := params["key"]

		start_key := request.FormValue("start-key")
		next_key := request.FormValue("next-key")
		description := request.FormValue("description")

		err := qs.UpdateKey(key_id, start_key, description, next_key)
		log.Printf("CONSOLE WEB was update key %s %s %s %s\n err? %v", key_id, start_key, description, next_key, err)
		render.Redirect("/new_keys")
	})

	r.Get("/delete_key_all", func(render render.Render) {
		qs.Steps.RemoveAll(bson.M{})
		render.Redirect("/new_keys")
	})

	m.Action(r.Handle)
	m.RunOnAddr(addr)
}