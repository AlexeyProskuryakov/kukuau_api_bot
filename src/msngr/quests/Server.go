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
	w "msngr/web"
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

var keys_cache []Step

func GetKeysInfo(err_text string, qs *QuestStorage) map[string]interface{} {
	var e error
	result := map[string]interface{}{}

	keys, e := qs.GetAllKeys()

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
	m := martini.New()
	m.Use(w.NonJsonLogger())
	m.Use(martini.Recovery())
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

	r := martini.NewRouter()

	r.Get("/", func(user auth.User, render render.Render) {
		render.HTML(200, "quests/index", map[string]interface{}{})
	})

	r.Get("/new_keys", func(render render.Render) {
		render.HTML(200, "quests/new_keys", GetKeysInfo("", qs))
	})

	r.Post("/add_key", func(user auth.User, render render.Render, request *http.Request) {
		start_key := strings.TrimSpace(request.FormValue("start-key"))
		next_key := strings.TrimSpace(request.FormValue("next-key"))
		description := request.FormValue("description")

		log.Printf("QUESTS WEB add key %s -> %s -> %s", start_key, description, next_key)
		if start_key != "" && description != "" {
			key, err := qs.AddKey(start_key, description, next_key)
			if key != nil &&err != nil {
				render.HTML(200, "quests/new_keys", GetKeysInfo("Такой ключ уже существует. Используйте изменение ключа если хотите его изменить.", qs))
				return
			}
		} else {

			render.HTML(200, "quests/new_keys", GetKeysInfo("Невалидные значения ключа или ответа", qs))
			return
		}
		render.Redirect("/new_keys")
	})

	r.Post("/delete_key/:key", func(params martini.Params, render render.Render) {
		key := params["key"]
		err := qs.DeleteKey(key)
		log.Printf("QUESTS WEB will delete %v (%v)", key, err)
		render.Redirect("/new_keys")
	})

	r.Post("/update_key/:key", func(params martini.Params, render render.Render, request *http.Request) {
		key_id := params["key"]

		start_key := strings.TrimSpace(request.FormValue("start-key"))
		next_key := strings.TrimSpace(request.FormValue("next-key"))
		description := request.FormValue("description")

		err := qs.UpdateKey(key_id, start_key, description, next_key)
		log.Printf("QUESTS WEB was update key %s %s %s %s\n err? %v", key_id, start_key, description, next_key, err)
		render.Redirect("/new_keys")
	})

	r.Get("/delete_key_all", func(render render.Render) {
		qs.Steps.RemoveAll(bson.M{})
		render.Redirect("/new_keys")
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
	r.Get("/chat", func(render render.Render, params martini.Params, req *http.Request) {
		var with string
		result_data := map[string]interface{}{}
		query := req.URL.Query()
		for key, value := range query {
			if key == "with" && len(value) > 0 {
				with = value[0]
				log.Printf("QSERV: with found is: %v", with)
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
					AllKeys     []Step
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

	r.Post("/send_message", func(render render.Render, req *http.Request) {
		from := req.FormValue("from")
		to := req.FormValue("to")
		text := req.FormValue("chat-form-message")
		if from != "" && to != "" && text != "" {
			if to == "all" {
				peoples, _ := qs.GetPeoples(bson.M{})
				log.Printf("QSERV: will send [%v] to all %v peoples", text, len(peoples))
				send_messages_to_peoples(peoples, ntf, text)
			} else if to == "all_team_members" {
				peoples, _ := qs.GetAllTeamMembers()
				log.Printf("QSERV: will send [%v] to all team members %v peoples", text, len(peoples))
				send_messages_to_peoples(peoples, ntf, text)
			} else {
				team, _ := qs.GetTeamByName(to)
				if team == nil {
					man, _ := qs.GetManByUserId(to)
					if man != nil {
						log.Printf("QSERV: will send [%v] to %v", text, man.UserId)
						ntf.NotifyText(man.UserId, text)
					}
				}else {
					peoples, _ := qs.GetMembersOfTeam(team.Name)
					log.Printf("QSERV: will send [%v] to team members of %v team to %v peoples", text, team.Name, len(peoples))
					send_messages_to_peoples(peoples, ntf, text)
				}
			}
			qs.StoreMessage(from, to, text, false)
			log.Printf("QSERV: will answered all messages from %v by %v", to, from)
			qs.SetMessagesAnswered(to, from)

		} else {
			render.Redirect("/chat")
		}
		render.Redirect(fmt.Sprintf("/chat?with=%v", to))
	})

	r.Post("/new_messages", func(render render.Render, req *http.Request) {
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
		if err != nil {
			render.JSON(500, map[string]interface{}{"ok":false, "detail":fmt.Sprintf("error in db: %v", err)})
			return
		}

		for i, message := range messages {
			team, _ := qs.GetTeamByName(message.From)
			if team != nil {
				messages[i].From = team.Name
			}else {
				man, _ := qs.GetManByUserId(message.From)
				if man != nil {
					messages[i].From = man.Name
				}

			}
		}

		render.JSON(200, map[string]interface{}{"messages":messages, "next_":time.Now().Unix()})
	})

	r.Post("/new_contacts", func(render render.Render, req *http.Request) {
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
	r.Get("/manage", func(render render.Render, req *http.Request) {
		render.HTML(200, "quests/manage", map[string]interface{}{})
	})
	r.Post("/delete_all", func(render render.Render, req *http.Request) {
		//1. Steps or keys:
		si, _ := qs.Steps.RemoveAll(bson.M{})
		//2 Peoples
		pi, _ := qs.Peoples.UpdateAll(bson.M{"is_passerby":false, "team_name":bson.M{"$exists":true}, "team_sid":bson.M{"$exists":true}}, bson.M{"$set":bson.M{"is_passerby":true}, "$unset":bson.M{"team_name":"", "team_sid":""}})
		//3 teams
		ti, _ := qs.Teams.RemoveAll(bson.M{})
		render.JSON(200, map[string]interface{}{
			"ok":true,
			"steps_removed":si.Removed,
			//"steps_removed":0,
			"peoples_updated":pi.Updated,
			//"peoples_updated":0,
			"teams_removed":ti.Removed,
			//"teams_removed":0,

		})
	})

	log.Printf("Will start web server for quest at: %v", config.WebPort)

	//m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	m.RunOnAddr(config.WebPort)
}
