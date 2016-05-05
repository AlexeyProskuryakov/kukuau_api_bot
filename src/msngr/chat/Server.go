package chat

import (
	d "msngr/db"
	c "msngr/configuration"
	u "msngr/utils"
	ntf "msngr/notify"
	usrs "msngr/users"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/martini-contrib/render"
	"gopkg.in/mgo.v2/bson"

	"sort"

	"html/template"
	"time"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"log"
	"strings"
)

const (
	ALL = "all"
)

func getFuncMap(cName, cId, start_addr string) template.FuncMap {
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
func GetContacts(db *d.MainDb, to_name string) ([]usrs.Contact, error) {
	resp := []usrs.Contact{}
	err := db.Messages.MessagesCollection.Pipe([]bson.M{
		bson.M{"$match":bson.M{"to":to_name}},
		bson.M{"$group": bson.M{"_id":"$from", "unread_count":bson.M{"$sum":"$unread"}, "name":bson.M{"$first":"$from"}, "time":bson.M{"$max":"$time_stamp"}}}}).All(&resp)
	if err != nil {
		return resp, err
	}
	result := []usrs.Contact{}
	for i, cnt := range resp {
		user, _ := db.Users.GetUserById(cnt.ID)
		if user != nil {
			if user.ShowedName != "" {
				resp[i].Name = user.ShowedName
			} else {
				resp[i].Name = user.UserName
			}
			resp[i].Phone = user.Phone
			result = append(result, resp[i])
		}
	}
	sort.Sort(usrs.ByContactsLastMessageTime(result))
	return result, nil
}

func getRenderer(cName, cId, start_addr string) martini.Handler {
	renderer := render.Renderer(render.Options{
		Directory:"templates/chat",
		//Layout: "console/layout",
		Extensions: []string{".tmpl", ".html"},
		Charset: "UTF-8",
		IndentJSON: true,
		IndentXML: true,
		Funcs:[]template.FuncMap{
			getFuncMap(cName, cId, start_addr),
		},
	})
	return renderer
}

func getRendererTemplateDir(cName, cId, start_addr, template_dir string) martini.Handler {
	renderer := render.Renderer(render.Options{
		Directory:template_dir,
		//Layout: "console/layout",
		Extensions: []string{".tmpl", ".html"},
		Charset: "UTF-8",
		IndentJSON: true,
		IndentXML: true,
		Funcs:[]template.FuncMap{
			getFuncMap(cName, cId, start_addr),
		},
	})
	return renderer
}
func getBasicAuth(db *d.MainDb) martini.Handler {
	return auth.BasicFunc(func(username, password string) bool {
		usr, _ := db.Users.GetUser(bson.M{"user_name":username, "role":usrs.MANAGER})
		if usr != nil {
			return u.PHash(password) == usr.Password
		}
		return username == usrs.DEFAULT_USER && password == usrs.DEFAULT_PWD
	})
}

func GetMartini(cName, cId, start_addr string, db *d.MainDb) *martini.ClassicMartini {
	m := martini.Classic()
	m.Use(getRenderer(cName, cId, start_addr))
	m.Use(getBasicAuth(db))
	return m
}

//func GetMartiniTemplatesDir(cName, cId, start_addr, template_dir string, db *d.MainDb) *martini.ClassicMartini {
//	m := martini.Classic()
//	m.Use(getRendererTemplateDir(cName, cId, start_addr, template_dir))
//	m.Use(getBasicAuth(db))
//	return m
//}

func GetChatMainHandler(start_addr string, notifier *ntf.Notifier, db *d.MainDb, config c.ChatConfig) http.Handler {
	m := GetMartini(config.Name, config.CompanyId, start_addr, db)
	m.Get(start_addr, func(r render.Render, params martini.Params, req *http.Request) {
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

		var messages []d.MessageWrapper
		if with != ALL {
			user, _ := db.Users.GetUserById(with)
			if user != nil {
				messages, _ = db.Messages.GetMessages(bson.M{
					"$or":[]bson.M{
						bson.M{"from":user.UserId, "to":config.CompanyId},
						bson.M{"to":user.UserId, "from":config.CompanyId},
					},
				})
				for i, message := range messages {
					if message.From == user.UserId {
						messages[i].From = user.GetName()
					} else if message.To == user.UserId {
						messages[i].To = user.GetName()
					}
				}
			}
		}

		if strings.Contains(with, ALL) {
			messages, _ = db.Messages.GetMessages(bson.M{"to":with})
		}

		result_data["with"] = with
		result_data["messages"] = messages
		result_data["companyId"] = config.CompanyId

		if contacts, err := GetContacts(db, config.CompanyId); err == nil {
			result_data["contacts"] = contacts
		}
		r.HTML(200, "chat", result_data, render.HTMLOptions{Layout:"base"})

	})
	return m
}

func GetChatDeleteMessagesHandler(start_addr string, db *d.MainDb, config c.ChatConfig) http.Handler {
	m := GetMartini(config.Name, config.CompanyId, start_addr, db)
	m.Delete(start_addr, func(ren render.Render, req *http.Request) {
		type DeleteInfo struct {
			From string `json:"from"`
			To   string `json:"to"`
		}
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Printf("CS QE E: errror at reading req body %v", err)
			ren.JSON(500, map[string]interface{}{"error":err})
			return
		}
		dInfo := DeleteInfo{}
		err = json.Unmarshal(data, &dInfo)
		if err != nil {
			log.Printf("CS QE E: at unmarshal json messages %v\ndata:%s", err, data)
			ren.JSON(500, map[string]interface{}{"error":err})
			return
		}
		count, err := db.Messages.DeleteMessages(dInfo.From, dInfo.To)
		if err != nil {
			ren.JSON(500, map[string]interface{}{"error":err})
			return
		}
		ren.JSON(200, map[string]interface{}{"success":true, "deleted":count})
	})
	return m
}

func GetChatSendHandler(start_addr string, notifier *ntf.Notifier, db *d.MainDb, config c.ChatConfig, cs *ChatStorage) http.Handler {
	m := GetMartini(config.Name, config.CompanyId, start_addr, db)
	m.Post(start_addr, func(render render.Render, req *http.Request) {
		type MessageFromF struct {
			From string `json:"from"`
			To   string `json:"to"`
			Body string `json:"body"`
		}

		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Printf("CS QE E: errror at reading req body %v", err)
			render.JSON(500, map[string]interface{}{"error":err})
			return
		}
		message := MessageFromF{}
		err = json.Unmarshal(data, &message)
		if err != nil {
			log.Printf("CS QE E: at unmarshal json messages %v\ndata:%s", err, data)
			render.JSON(500, map[string]interface{}{"error":err})
			return
		}
		log.Printf("CS message to send: %v", message)
		var messageSID string
		if message.From != "" && message.To != "" && message.Body != "" {
			if message.To == ALL {
				peoples, _ := cs.GetUsersOfCompany(config.CompanyId)
				notifier.SendMessageToPeople(peoples, message.Body)
			} else {
				user, _ := db.Users.GetUserById(message.To)
				if user != nil {
					db.Messages.SetMessagesRead(user.UserId)
					_, resultMessage, _ := notifier.NotifyText(message.To, message.Body)
					resultMessage, _ = db.Messages.GetMessageByMessageId(resultMessage.MessageID)
					messageSID = resultMessage.SID
				}
				db.Messages.SetMessagesAnswered(message.To, config.CompanyId, config.CompanyId)
			}
		} else {
			render.Redirect("/chat")
		}
		render.JSON(200, map[string]interface{}{"ok":true, "message":d.NewMessageForWeb(messageSID, message.From, message.To, message.Body, )})
	})
	return m
}
func GetChatMessageReadHandler(start_addr string, notifier *ntf.Notifier, db *d.MainDb, config c.ChatConfig) http.Handler {
	m := GetMartini(config.Name, config.CompanyId, start_addr, db)
	m.Post(start_addr, func(render render.Render, req *http.Request) {
		type Readed struct {
			From string `json:"from"`
		}
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Printf("CS QE E: errror at reading req body %v", err)
			render.JSON(500, map[string]interface{}{"error":err})
			return
		}
		readed := Readed{}
		err = json.Unmarshal(data, &readed)
		if err != nil {
			log.Printf("CS QE E: at unmarshal json messages %v\ndata:%s", err, data)
			render.JSON(500, map[string]interface{}{"error":err})
			return
		}
		err = db.Messages.SetAllMessagesRead(readed.From, config.CompanyId)
		if err != nil {
			log.Printf("CS QE E: at unmarshal json messages %v\ndata:%s", err, data)
			render.JSON(500, map[string]interface{}{"error":err})
			return
		}
		render.JSON(200, map[string]interface{}{"ok":true})
	})
	return m
}

func get_messages(between1, between2 string, db *d.MainDb) ([]d.MessageWrapper, error) {
	query := bson.M{"unread":1}
	if between1 == "" {
		query["$or"] = []bson.M{bson.M{"to":between2}, bson.M{"from":between2}}
	} else {
		query["$or"] = []bson.M{bson.M{"from":between1, "to":between2}, bson.M{"to":between1, "from":between2}}
	}

	messages, err := db.Messages.GetMessages(query)
	result := []d.MessageWrapper{}
	if err != nil {
		log.Printf("CS unread messages: error at retrieve messages %v", err)
		return result, err
	}
	for i, msg := range messages {
		if msg.From == between2 {
			u, _ := db.Users.GetUserById(msg.To)
			messages[i].ToName = u.GetName()
			messages[i].FromName = messages[i].From
		} else {
			u, _ := db.Users.GetUserById(msg.From)
			messages[i].FromName = u.GetName()
		}
		result = append(result, messages[i])
	}
	return result, nil
}
func GetChatUnreadMessagesHandler(start_addr string, notifier *ntf.Notifier, db *d.MainDb, config c.ChatConfig) http.Handler {
	m := GetMartini(config.Name, config.CompanyId, start_addr, db)
	m.Post(start_addr, func(render render.Render, req *http.Request) {
		type NewMessagesReq struct {
			For string `json:"m_for"`
		}
		nmReq := NewMessagesReq{}
		request_body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			render.JSON(500, map[string]interface{}{"ok":false, "detail":"can not read request body"})
			return
		}
		err = json.Unmarshal(request_body, &nmReq)
		if err != nil {
			render.JSON(500, map[string]interface{}{"ok":false, "detail":fmt.Sprintf("can not unmarshal request body %v \n %s", err, request_body)})
			return
		}
		result, err := get_messages(nmReq.For, config.CompanyId, db)
		if err != nil {
			render.JSON(500, map[string]interface{}{"ok":false, "detail":fmt.Sprintf("error in db: %v", err)})
		}
		render.JSON(200, map[string]interface{}{"messages":result})
	})
	return m
}
func GetChatContactsHandler(start_addr string, notifier *ntf.Notifier, db *d.MainDb, config c.ChatConfig) http.Handler {
	m := GetMartini(config.Name, config.CompanyId, start_addr, db)
	m.Post(start_addr, func(render render.Render, req *http.Request) {
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
		contacts, err := GetContacts(db, config.CompanyId)
		if err != nil {
			render.JSON(500, map[string]interface{}{"ok":false, "detail":fmt.Sprintf("db err body %v", err)})
			return
		}
		new_contacts := []usrs.Contact{}
		old_contacts := []usrs.Contact{}

		for _, contact := range contacts {
			if u.InS(contact.ID, cr.Exist) {
				if contact.NewMessagesCount > 0 {
					old_contacts = append(old_contacts, contact)
				}
			} else {
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
	return m
}

func GetChatContactsChangeHandler(start_addr string, notifier *ntf.Notifier, db *d.MainDb, config c.ChatConfig) http.Handler {
	m := GetMartini(config.Name, config.CompanyId, start_addr, db)
	m.Post(start_addr, func(render render.Render, req *http.Request) {
		type NewContactName struct {
			Id      string `json:"id"`
			NewName string `json:"new_name"`
		}
		ncn := NewContactName{}
		request_body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			render.JSON(500, map[string]interface{}{"ok":false, "detail":"can not read request body"})
			return
		}
		err = json.Unmarshal(request_body, &ncn)
		if err != nil {
			render.JSON(500, map[string]interface{}{"ok":false, "detail":fmt.Sprintf("can not unmarshal request body %v \n %s", err, request_body)})
			return
		}
		err = db.Users.SetUserShowedName(ncn.Id, ncn.NewName)
		if err != nil {
			render.JSON(500, map[string]interface{}{"ok":false, "detail":err})
			return
		}
		render.JSON(200, map[string]interface{}{"ok":true})

	})
	return m
}

func GetChatConfigHandler(start_addr, prefix string, db *d.MainDb, config c.ChatConfig) http.Handler {
	m := GetMartini(config.Name, config.CompanyId, prefix, db)
	m.Get(start_addr, func(ren render.Render, req *http.Request) {
		ren.HTML(200, "config", map[string]interface{}{}, render.HTMLOptions{Layout:"base"})
	})
	return m
}

