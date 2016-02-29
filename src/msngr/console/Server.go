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

	"strings"
	usrs "msngr/users"
	"time"
	"sort"
)

const (
	ALL = "all"
)

var START_DT = time.Now().Unix() - 3600 * 24 * 365 * 5

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

func GetUsersInfo(err_text string, db *d.MainDb) map[string]interface{} {
	result := map[string]interface{}{}
	users, e := db.Users.GetBy(bson.M{})
	if e != nil || err_text != "" {
		result["is_error"] = true
		if e != nil {
			result["error_text"] = e.Error()
		} else {
			result["error_text"] = err_text
		}
	}
	result["users"] = users
	return result
}

type ByContactsLastMessageTime []usrs.Contact

func (s ByContactsLastMessageTime) Len() int {
	return len(s)
}
func (s ByContactsLastMessageTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByContactsLastMessageTime) Less(i, j int) bool {
	return s[i].LastMessageTime > s[j].LastMessageTime

}

func send_messages_to_peoples(people []d.UserWrapper, ntf *ntf.Notifier, text string) {
	go func() {
		for _, user := range people {
			ntf.NotifyText(user.UserId, text)
		}
	}()
}

func GetContacts(db *d.MainDb, after int64) ([]usrs.Contact, error) {
	resp := []usrs.Contact{}
	err := db.Messages.Collection.Pipe([]bson.M{
		bson.M{"$match":bson.M{"time_stamp":bson.M{"$gt":after}}},
		bson.M{"$group": bson.M{"_id":"$from", "unread_count":bson.M{"$sum":"$unread"}, "name":bson.M{"$first":"$from"}, "time":bson.M{"$max":"$time_stamp"}}}}).All(&resp)
	if err != nil {
		return resp, err
	}
	result := []usrs.Contact{}
	for i, cnt := range resp {
		if cnt.Name == ME {
			continue
		}
		user, _ := db.Users.GetUserById(cnt.ID)
		if user != nil {
			resp[i].Name = user.UserName
			resp[i].Phone = user.Phone
			result = append(result, resp[i])
		}
	}
	sort.Sort(ByContactsLastMessageTime(result))
	//log.Printf("CONSOLE ADD CNTCTS :%+v", result)
	return result, nil
}

type Collocutor struct {
	InfoPresent bool
	Info        CollocutorInfo
	Name        string
	Phone       string
	Email       string
}
type OrdersInfo struct {
	SourceName string `bson:"_id"`
	Count      int `bson:"count"`
}
type CollocutorInfo struct {
	CountOrdersAll        int
	CountOrdersByProvider []OrdersInfo
}

func Run(addr string, notifier *ntf.Notifier, db *d.MainDb, cs c.ConfigStorage, qs *quests.QuestStorage, ntf *ntf.Notifier) {
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
			render.HTML(200, "console/new_keys", GetKeysInfo("Файл имеет не то расширение :(", qs))
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
		log.Printf("CONSOLE WEB will show keys")
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
		log.Printf("CONSOLE WEB was delete all keys")
		qs.Steps.RemoveAll(bson.M{})
		render.Redirect("/new_keys")
	})

	r.Get("/chat", func(render render.Render, params martini.Params, req *http.Request) {
		var with string
		result_data := map[string]interface{}{}
		query := req.URL.Query()
		for key, value := range query {
			if key == "with" && len(value) > 0 {
				with = value[0]
				log.Printf("CONSOLE CHAT: [with] == [%v]", with)
				break
			}
		}
		if with == "" {
			with = ALL
		}
		db.Messages.SetMessagesRead(with)
		collocutor := Collocutor{}

		var messages []d.MessageWrapper
		if with != ALL {
			user, _ := db.Users.GetUserById(with)
			if user != nil {
				messages, _ = db.Messages.GetMessages(bson.M{
					"$or":[]bson.M{
							bson.M{"from":user.UserId},
							bson.M{"to":user.UserId},
					},
				})
				for i, _ := range messages {
					if messages[i].From != ME {
						messages[i].From = user.UserName
					}
				}
				collocutor.Name = user.UserName
				collocutor.Phone = user.Phone
				collocutor.Email = user.Email

				order_count, _ := db.Orders.Collection.Find(bson.M{"whom":user.UserId}).Count()
				resp := []OrdersInfo{}
				err := db.Orders.Collection.Pipe([]bson.M{
					bson.M{"$match":bson.M{"whom":user.UserId}},
					bson.M{"$group": bson.M{"_id":"$source", "count":bson.M{"$sum":1}}}}).All(&resp)

				if err == nil {
					collocutor.Info.CountOrdersByProvider = resp
				}
				collocutor.Info.CountOrdersAll = order_count
				collocutor.InfoPresent = true
			}
		}

		if strings.Contains(with, ALL) {
			//log.Printf("CONSOLE WEB CHAT: getting messages for all")
			messages, _ = db.Messages.GetMessages(bson.M{"to":with})
		}
		result_data["collocutor"] = collocutor
		result_data["with"] = with
		result_data["messages"] = messages

		if contacts, err := GetContacts(db, START_DT); err == nil {
			result_data["contacts"] = contacts
		}
		render.HTML(200, "console/chat", result_data)
	})

	r.Post("/send_message", func(render render.Render, req *http.Request) {
		from := req.FormValue("from")
		to := req.FormValue("to")
		text := req.FormValue("chat-form-message")
		log.Printf("CONSOLE SM: %v -> %v [%v]", from, to, text)
		if from != "" && to != "" && text != "" {
			if to == ALL {
				peoples, _ := db.Users.GetBy(bson.M{})
				//log.Printf("CONSOLE SM: will send [%v] to all %v peoples", text, len(peoples))
				send_messages_to_peoples(peoples, ntf, text)
				db.Messages.StoreMessage(from, to, text, u.GenId())

			} else if to == "all_hash_writers" {
				go func() {
					peoples, _ := db.Users.GetBy(bson.M{"last_marker":bson.M{"$exists":true}})
					send_messages_to_peoples(peoples, ntf, text)
				}()

			} else {
				user, _ := db.Users.GetUserById(to)
				if user != nil {
					ntf.NotifyText(to, text)
				}
				db.Messages.SetMessagesAnswered(to, from)
			}
			log.Printf("CONSOLE SM: will answered all messages from %v by %v", to, from)

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
		query := bson.M{"time_stamp":bson.M{"$gt":q.After}}
		if q.For == "" {
			q.For = ALL
			query["to"] = ME
		} else {
			query["from"] = q.For
		}

		messages, err := db.Messages.GetMessages(query)
		if err != nil {
			render.JSON(500, map[string]interface{}{"ok":false, "detail":fmt.Sprintf("error in db: %v", err)})
			return
		}
		result := []d.MessageWrapper{}
		for i, msg := range messages {
			if msg.From != ME {
				u, _ := db.Users.GetUserById(msg.From)
				messages[i].From = u.UserName
				result = append(result, messages[i])
			}
		}

		render.JSON(200, map[string]interface{}{"messages":result, "next_":time.Now().Unix()})
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
		log.Printf("CONSOLE WEB NC Ask: %+v", cr)
		contacts, err := GetContacts(db, cr.After)
		if err != nil {
			render.JSON(500, map[string]interface{}{"ok":false, "detail":fmt.Sprintf("db err body %v \n %s", err)})
			return
		}
		new_contacts := []usrs.Contact{}
		old_contacts := []usrs.Contact{}

		for _, contact := range contacts {
			if u.InS(contact.ID, cr.Exist) {
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

	r.Get("/users", func(render render.Render, req *http.Request) {
		render.HTML(200, "console/users", GetUsersInfo("", db))
	})

	r.Post("/add_user", func(user auth.User, render render.Render, request *http.Request) {
		u_id := request.FormValue("user-id")
		u_name := request.FormValue("user-name")
		u_phone := request.FormValue("user-phone")
		u_email := request.FormValue("user-e-mail")
		u_role := request.FormValue("user-role")
		u_pwd := request.FormValue("user-pwd")

		log.Printf("CONSOLE WEB add user [%s]  '%s' +%s %s |%v| {%s}", u_id, u_name, u_phone, u_email, u_role, u_pwd)
		if u_name != "" && u_id != "" {
			db.Users.AddUserObject(d.UserWrapper{UserId:u_id, UserName:u_name, Email:u_email, Phone:u_phone, Role:u_role, Password:u.PHash(u_pwd), LastUpdate:time.Now()})
			render.Redirect("/users")
		} else {
			render.HTML(200, "console/users", GetUsersInfo("Невалидные значения имени и (или) идентификатора добавляемого пользователя", db))
		}
	})

	r.Post("/delete_user/:id", func(params martini.Params, render render.Render) {
		uid := params["id"]
		err := db.Users.Collection.Remove(bson.M{"user_id":uid})
		log.Printf("CONSOLE WEB will delete user %v (%v)", uid, err)
		render.Redirect("/users")
	})

	r.Post("/update_user/:id", func(params martini.Params, render render.Render, request *http.Request) {
		u_id := params["id"]
		u_name := request.FormValue("user-name")
		u_phone := request.FormValue("user-phone")
		u_email := request.FormValue("user-e-mail")
		u_role := request.FormValue("user-role")
		u_pwd := request.FormValue("user-pwd")

		upd := bson.M{}
		if u_name != "" {
			upd["user_name"] = u_name
		}
		if u_email != "" {
			upd["email"] = u_email
		}
		if u_phone != "" {
			upd["phone"] = u_phone
		}
		if u_role != "" {
			upd["role"] = u_role
		}
		if u_pwd != "" {
			upd["password"] = u.PHash(u_pwd)
		}
		db.Users.Collection.Update(bson.M{"user_id":u_id}, bson.M{"$set":upd})
		log.Printf("CONSOLE WEB update user [%s]  '%s' +%s %s |%v| {%v}", u_id, u_name, u_phone, u_email, u_role, u_pwd)
		render.Redirect("/users")
	})

	r.Get("/delete_chat/:between", func(params martini.Params, render render.Render, req *http.Request) {
		between := params["between"]
		db.Messages.Collection.RemoveAll(bson.M{"$or":[]bson.M{bson.M{"from":between}, bson.M{"to":between}}})
		render.Redirect(fmt.Sprintf("/chat?with=%v", between))
	})

	r.Get("/profiles", func(render render.Render){
		render.HTML(200, "console/profiles", map[string]interface{}{})
	})

	m.Action(r.Handle)
	m.RunOnAddr(addr)
}