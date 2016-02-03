package quests

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/martini-contrib/render"
	"github.com/tealeg/xlsx"

	c "msngr/configuration"

	ntf "msngr/notify"

	"log"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"fmt"
	"strings"
	"io/ioutil"
	"strconv"
	"regexp"
	"html/template"
	"encoding/json"
	"time"
	"msngr/utils"
)

var users = map[string]string{
	"alesha":"sederfes100500",
	"leha":"qwerty100500",
	"dima":"123",
}

const (
	ALL = "all"
	ALL_TEAM_MEMBERS = "all_team_members"
)

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

func send_messages_to_peoples(people []TeamMember, ntf *ntf.Notifier, text string) {
	go func() {
		for _, user := range people {
			ntf.NotifyText(user.UserId, text)
		}
	}()
}

func Run(config c.QuestConfig, qs *QuestStorage, ntf *ntf.Notifier) {

	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Layout: "quests/layout",
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
			key, err := qs.AddKey(start_key, description, next_key)
			log.Printf("QW is error? %v key: %v", err, key)
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
	m.Get("/chat", func(render render.Render, params martini.Params, req *http.Request) {
		var with string
		result_data := map[string]interface{}{}
		query := req.URL.Query()
		for key, value := range query {
			if key == "with" && len(value) > 0 {
				with = value[0]
				log.Printf("QS: with found is: %v", with)
				break
			}
		}
		type Collocutor struct {
			IsTeam bool
			IsMan  bool
			IsAll  bool
			Info   interface{}
			Name   string
		}
		collocutor := Collocutor{}

		var messages []Message

		if with != ALL && with != ALL_TEAM_MEMBERS {
			if team, _ := qs.GetTeamByName(with); team != nil {
				type TeamInfo struct {
					FoundedKeys []string
					Members     []TeamMember
					AllKeys     []Key
				}

				collocutor.Name = team.Name
				collocutor.IsTeam = true
				members, _ := qs.GetMembersOfTeam(team.Name)
				keys, _ := qs.GetKeys(bson.M{"for_team":team.Name})

				collocutor.Info = TeamInfo{FoundedKeys:team.FoundKeys, Members:members, AllKeys:keys}

				messages, _ = qs.GetMessages(bson.M{
					"$or":[]bson.M{
						bson.M{"from":team.Name},
						bson.M{"to":team.Name},
					},
				})
			}else {
				if peoples, _ := qs.GetPeoples(bson.M{"user_id":with}); len(peoples) > 0 {
					man := peoples[0]
					collocutor.IsMan = true
					collocutor.Name = man.Name
					collocutor.Info = man

					messages, _ = qs.GetMessages(bson.M{
						"$or":[]bson.M{
							bson.M{"from":man.UserId},
							bson.M{"to":man.UserId},
						},
					})
					for i, _ := range messages {
						if messages[i].From != ME {
							messages[i].From = man.Name
						}
					}
				} else {
					with = "all"
				}
			}
		}

		if strings.HasPrefix(with, "all") {
			collocutor.IsAll = true
			collocutor.Name = with
			messages, _ = qs.GetMessages(bson.M{"to":with})
		}

		result_data["with"] = with
		result_data["collocutor"] = collocutor
		result_data["messages"] = messages

		all_teams, _ := qs.GetAllTeams()
		if contacts, err := qs.GetContacts(all_teams); err == nil {
			result_data["contacts"] = contacts
		}
		render.HTML(200, "quests/chat", result_data)
	})

	m.Post("/send_message", func(render render.Render, req *http.Request) {
		from := req.FormValue("from")
		to := req.FormValue("to")
		text := req.FormValue("chat-form-message")
		if from != "" && to != "" && text != "" {
			if to == "all" {
				peoples, _ := qs.GetPeoples(bson.M{})
				send_messages_to_peoples(peoples, ntf, text)
			} else if to == "all_team_members" {
				peoples, _ := qs.GetAllTeamMembers()
				send_messages_to_peoples(peoples, ntf, text)
			} else {
				team, _ := qs.GetTeamByName(to)
				if team == nil {
					man, _ := qs.GetManByUserId(to)
					if man != nil {
						ntf.NotifyText(man.UserId, text)
					}
				}else {
					peoples, _ := qs.GetMembersOfTeam(team.Name)
					send_messages_to_peoples(peoples, ntf, text)
				}
			}
			qs.StoreMessage(from, to, text, false)
			log.Printf("QS: will answered all messages from %v by %v", from, to)
			qs.SetMessagesAnswered(to, from)

		} else {
			render.Redirect("/chat")
		}
		render.Redirect(fmt.Sprintf("/chat?with=%v", to))
	})

	m.Post("/new_messages", func(render render.Render, req *http.Request) {
		type NewMessagesReq struct {
			For   string `json:"m_for"`
			After int64 `json:"after"`
		}
		q := NewMessagesReq{}
		request_body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			render.JSON(500, map[string]interface{}{"ok":false, "detail":"can not read request body"})
			return
		}
		err = json.Unmarshal(request_body, &q)
		if err != nil {
			render.JSON(500, map[string]interface{}{"ok":false, "detail":fmt.Sprintf("can not unmarshal request body %v \n %s", err, request_body)})
			return
		}

		messages, err := qs.GetMessages(bson.M{"from":q.For, "time_stamp":bson.M{"$gt":q.After}})
		log.Printf("and  find: %+v", messages)
		if err != nil {
			render.JSON(500, map[string]interface{}{"ok":false, "detail":fmt.Sprintf("error in db: %v", err)})
			return
		}
		render.JSON(200, map[string]interface{}{"messages":messages, "next_":time.Now().Unix()})
	})

	m.Post("/new_contacts", func(render render.Render, req *http.Request) {
		type NewContactsReq struct {
			After int64 `json:"after"`
			Exist []string `json:"exist"`
		}
		cr := NewContactsReq{}
		request_body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			render.JSON(500, map[string]interface{}{"ok":false, "detail":"can not read request body"})
			return
		}
		err = json.Unmarshal(request_body, &cr)
		if err != nil {
			render.JSON(500, map[string]interface{}{"ok":false, "detail":fmt.Sprintf("can not unmarshal request body %v \n %s", err, request_body)})
			return
		}
		contacts, err := qs.GetContactsAfter(cr.After)
		if err != nil {
			log.Printf("err :%v", err)
			render.JSON(500, map[string]interface{}{"ok":false, "detail":fmt.Sprintf("db err body %v \n %s", err)})
			return
		}
		new_contacts := []Contact{}
		old_contacts := []Contact{}

		for _, contact := range contacts {
			if utils.InS(contact.ID, cr.Exist) {
				old_contacts = append(old_contacts, contact)
			}else {
				new_contacts = append(new_contacts, contact)
			}
		}
		render.JSON(200, map[string]interface{}{
			"ok":true,
			"new":new_contacts,
			"old":old_contacts,
			"next_":time.Now().Unix(),
		})

	})

	log.Printf("Will start web server for quest at: %v", config.WebPort)
	m.RunOnAddr(config.WebPort)
}
