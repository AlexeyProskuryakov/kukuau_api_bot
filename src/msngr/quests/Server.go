package quests

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/martini-contrib/render"
	"github.com/tealeg/xlsx"

	c "msngr/configuration"

	"msngr/notify"
	"msngr/structs"
	"msngr/utils"

	"log"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"fmt"
	"strings"
	"io/ioutil"
	"strconv"
	"regexp"
)

var users = map[string]string{
	"alesha":"sederfes100500",
	"leha":"qwerty100500",
	"dima":"123",
}

func ParseExportXlsx(xlf *xlsx.File, qs *QuestStorage, skip_row, skip_cell int) error {
	for _, sheet := range xlf.Sheets {
		if sheet != nil {
			sh_name := strings.TrimSpace(strings.ToLower(sheet.Name))
			if strings.HasSuffix(sh_name, "ключ") || strings.HasPrefix(sh_name, "ключ") {

				for ir, row := range sheet.Rows {
					if row != nil && ir >= skip_row {
						key := row.Cells[skip_cell].Value
						description := row.Cells[skip_cell + 1].Value
						next_key_raw := row.Cells[skip_cell + 2].Value
						if key != "" && description != "" {
							qs.AddKey(key, description, next_key_raw)
						}

					}
				}
			}
		}
	}
	return nil
}

var keys_cache []Key

func get_keys_info(err_text string, qs *QuestStorage) map[string]interface{} {
	var keys []Key
	var e error
	result := map[string]interface{}{}
	if err_text == "" {
		keys, e = qs.GetAllKeys()
	} else {
		keys = keys_cache
	}
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

func get_messages_info(err_text string, qs *QuestStorage) map[string]interface{} {
	teams, _ := qs.GetAllTeams()
	result := map[string]interface{}{}
	if err_text != "" {
		result["is_error"] = true
		result["error_text"] = err_text
	}
	result["teams"] = teams
	return result
}

func Run(config c.QuestConfig, qs *QuestStorage, ntf *msngr.Notifier) {
	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Layout: "quests/layout",
		Extensions: []string{".tmpl", ".html"},
		Charset: "UTF-8",
		IndentJSON: true,
		IndentXML: true,
	}))

	m.Use(auth.BasicFunc(func(username, password string) bool {
		pwd, ok := users[username]
		return ok && pwd == password
	}))

	m.Use(martini.Static("static"))

	m.Get("/", func(user auth.User, render render.Render) {
		render.HTML(200, "quests/index", map[string]interface{}{})
	})

	m.Get("/new_keys", func(render render.Render) {
		render.HTML(200, "quests/new_keys", get_keys_info("", qs))
	})

	m.Post("/add_key", func(user auth.User, render render.Render, request *http.Request) {
		start_key := request.FormValue("start-key")
		next_key := request.FormValue("next-key")
		description := request.FormValue("description")

		log.Printf("QUESTS WEB add key %s -> %s -> %s", start_key, description, next_key)
		if start_key != "" && description != "" {
			err := qs.AddKey(start_key, description, next_key)
			log.Printf("QW is error? %v", err)
			render.Redirect("/new_keys")
		} else {
			render.HTML(200, "quests/new_keys", get_keys_info("Невалидные значения ключа или ответа", qs))
		}
	})

	m.Post("/delete_key/:key", func(params martini.Params, render render.Render) {
		key := params["key"]
		err := qs.DeleteKey(key)
		log.Printf("QUESTS WEB will delete %v (%v)", key, err)
		render.Redirect("/new_keys")
	})

	m.Post("/update_key/:key", func(params martini.Params, render render.Render, request *http.Request) {
		key_id := params["key"]

		start_key := request.FormValue("start-key")
		next_key := request.FormValue("next-key")
		description := request.FormValue("description")

		err := qs.UpdateKey(key_id, start_key, description, next_key)
		log.Printf("QUESTS WEB was update key %s %s %s %s\n err? %v", key_id, start_key, description, next_key, err)
		render.Redirect("/new_keys")
	})

	m.Get("/delete_key_all", func(render render.Render) {
		qs.Keys.RemoveAll(bson.M{})
		render.Redirect("/new_keys")
	})

	m.Get("/messages", func(render render.Render) {
		//fake
		teams, _ := qs.GetAllTeams()

		render.HTML(200, "quests/messages", map[string]interface{}{
			"teams":teams,
		})
	})

	m.Get("/messages/:team_id", func(render render.Render, params martini.Params) {

	})

	m.Post("/message_answer_all", func(render render.Render, request *http.Request) {
		users, _ := qs.GetAllTeamMembers()
		answer := request.FormValue("message_all")
		if answer != "" {
			go func() {
				for _, user := range users {
					ntf.Notify(structs.OutPkg{To:user.UserId,
						Message: &structs.OutMessage{
							ID: utils.GenId(),
							Type: "chat",
							Body: answer,
						}})
				}
			}()
		}else {
			render.HTML(200, "quests/messages", get_messages_info("Ответ не может быть пустым", qs))
		}
		render.Redirect("/messages")
	})

	m.Get("/messages/new_count/:after", func(render render.Render, params martini.Params) {
		//after_input, err := strconv.ParseInt(params["after"], 10, 64)
		//if err != nil {
		//	render.JSON(200, map[string]interface{}{"error":err.Error()})
		//}
		//messages, err := qs.GetMessages(bson.M{
		//	"answered":false,
		//	"is_key":false,
		//	"time":bson.M{"$gte":after_input},
		//})
		//
		//if err != nil {
		//	render.JSON(200, map[string]interface{}{"error":err.Error()})
		//}else {
		//	render.JSON(200, map[string]interface{}{"error":false, "count":len(messages) })
		//}
	})

	xlsFileReg := regexp.MustCompile(".+\\.xlsx?")

	m.Post("/load/up", func(render render.Render, request *http.Request) {
		file, header, err := request.FormFile("file")

		log.Printf("Form file information: file: %+v \nheader:%v, %v\nerr:%v", file, header.Filename, header.Header, err)

		if err != nil {
			render.HTML(200, "quests/new_keys", get_keys_info(fmt.Sprintf("Ошибка загрузки файлика: %v", err), qs))
			return
		}
		defer file.Close()

		data, err := ioutil.ReadAll(file)
		if err != nil {
			render.HTML(200, "quests/new_keys", get_keys_info(fmt.Sprintf("Ошибка загрузки файлика: %v", err), qs))
			return
		}

		if xlsFileReg.MatchString(header.Filename) {
			xlFile, err := xlsx.OpenBinary(data)
			log.Printf("file: %+v, err: %v", xlFile, err)
			if err != nil || xlFile == nil {
				render.HTML(200, "quests/new_keys", get_keys_info(fmt.Sprintf("Ошибка обработки файлика: %v", err), qs))
				return
			}
			skip_rows, _ := strconv.Atoi(request.FormValue("skip-rows"))
			skip_cols, _ := strconv.Atoi(request.FormValue("skip-cols"))

			ParseExportXlsx(xlFile, qs, skip_rows, skip_cols)
		} else {
			render.HTML(200, "quests/new_keys", get_keys_info("Файл имеет не то расширение :(", qs))
		}

		render.Redirect("/new_keys")
	})

	m.Get("/chat", func(render render.Render){
		render.HTML(200, "quests/chat", map[string]interface{}{})
	})

	log.Printf("Will start web server for quest at: %v", config.WebPort)
	m.RunOnAddr(config.WebPort)
}
