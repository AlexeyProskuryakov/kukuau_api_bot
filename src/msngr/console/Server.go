package console

import (
	"net/http"
	"encoding/json"
	"log"
	"strings"
	"time"
	"sort"
	"io/ioutil"
	"fmt"
	"html/template"

	d "msngr/db"
	c "msngr/configuration"
	u "msngr/utils"
	ntf "msngr/notify"
	usrs "msngr/users"
	w "msngr/web"
	"msngr/quests"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/martini-contrib/render"
	"gopkg.in/mgo.v2/bson"
	"os"
	"io"
	"regexp"
	"msngr/voting"
)

const (
	ALL = "all"
)

func GetKeysInfo(err_text string, qs *quests.QuestStorage) map[string]interface{} {
	var keys []quests.Step
	var e error
	result := map[string]interface{}{}

	keys, e = qs.GetAllSteps()

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

func GetContacts(db *d.MainDb) ([]usrs.Contact, error) {
	resp := []usrs.Contact{}
	err := db.Messages.Collection.Pipe([]bson.M{
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

var TAG_REGEXP = regexp.MustCompile(`<\/?[^/bruia]([^>]*)>`)

func ProfileTextTagClear(p *Profile) *Profile {
	p.ShortDescription = strings.Replace(p.ShortDescription, "<div>", "<br>", -1)
	p.TextDescription = strings.Replace(p.TextDescription, "<div>", "<br>", -1)
	p.ShortDescription = TAG_REGEXP.ReplaceAllString(p.ShortDescription, "")
	p.TextDescription = TAG_REGEXP.ReplaceAllString(p.TextDescription, "")
	return p
}

func Run(addr string, db *d.MainDb, qs *quests.QuestStorage, vdh *voting.VotingDataHandler, ntf *ntf.Notifier, cfg c.Configuration) {
	m := martini.New()
	m.Use(w.NonJsonLogger())

	m.Use(martini.Recovery())

	m.Use(render.Renderer(render.Options{
		Directory:"templates/console",
		//Layout: "console/layout",
		Extensions: []string{".tmpl", ".html"},
		Charset: "UTF-8",
		IndentJSON: true,
		IndentXML: true,
		Funcs:[]template.FuncMap{
			template.FuncMap{
				"eq_s":func(a, b string) bool {
					return a == b
				},
				"stamp_date":func(t time.Time) string {
					return t.Format(time.Stamp)
				},
				"chat_with":func(with string) string {
					result := fmt.Sprintf("/chat?with=%v", with)
					log.Printf("chat with: %v", result)
					return result
				},
				"me":func() string {
					return ME
				},
				"is_message_":func(msg d.MessageWrapper, attrName string) bool {
					return msg.IsAttrPresent(attrName)
				},
				"has_additional_data":func(msg d.MessageWrapper) bool {
					return len(msg.AdditionalData) > 0
				},
				"is_additional_data_valid":func(ad d.AdditionalDataElement) bool{
					return ad.Value != ""
				},
			},
		},
	}))
	m.Use(auth.BasicFunc(func(username, password string) bool {
		usr, _ := db.Users.GetUser(bson.M{"user_name":username, "role":usrs.MANAGER})
		if usr != nil {
			return u.PHash(password) == usr.Password
		}
		return username == usrs.DEFAULT_USER && password == usrs.DEFAULT_PWD
	}))

	m.Use(martini.Static("static"))

	r := martini.NewRouter()

	r.Get("/", func(user auth.User, r render.Render) {
		r.HTML(200, "index", map[string]interface{}{}, render.HTMLOptions{Layout:"base"})
	})

	r.Group("/profile", func(r martini.Router) {
		type ProfileId struct {
			Id string `json:"id"`
		}
		pg_conf := cfg.Main.PGDatabase
		ph, err := NewProfileDbHandler(pg_conf.ConnString)
		if err != nil {
			panic(err)
		}

		r.Get("", func(render render.Render) {
			render.HTML(200, "profile", map[string]interface{}{})
		})
		r.Get("/all", func(render render.Render) {
			profiles, err := ph.GetAllProfiles()
			if err != nil {
				log.Printf("CS Error at getting all profiles: %v", err)
				render.JSON(500, map[string]interface{}{"success":false, "error":err})
			}
			render.JSON(200, map[string]interface{}{
				"success":true,
				"profiles":profiles,
			})
		})
		r.Get("/link_types", func(render render.Render) {
			render.JSON(200, map[string]interface{}{"data":ph.GetContactLinkTypes()})
		})
		r.Post("/read", func(render render.Render, params martini.Params, req *http.Request) {
			data, err := ioutil.ReadAll(req.Body)
			if err != nil {
				log.Printf("error at reading post data %v", err)
			}
			log.Printf("CS READ data: %s", data)
			info := ProfileId{}
			err = json.Unmarshal(data, &info)
			if err != nil {
				log.Printf("CS READ error at unmarshal read data %v", err)
				render.JSON(500, map[string]interface{}{"error":err, "success":false})
				return
			}
			profile, err := ph.GetProfile(info.Id)
			if err != nil {
				log.Printf("CS READ error at unmarshal read data %v", err)
				render.JSON(500, map[string]interface{}{"error":err, "success":false})
				return
			}
			out, err := json.Marshal(profile)
			if err != nil {
				log.Printf("CS READ error at marshal data to out")
				render.JSON(500, map[string]interface{}{"error":err, "success":false})
			}
			render.JSON(200, map[string]interface{}{"success":true, "data":out})
		})

		r.Post("/create", func(render render.Render, params martini.Params, req *http.Request) {
			data, err := ioutil.ReadAll(req.Body)
			if err != nil {
				log.Printf("error at reading post data %v", err)
			}
			log.Printf("CS CREATE data: %s", data)
			profile := &Profile{}
			err = json.Unmarshal(data, profile)
			if err != nil {
				log.Printf("CS CREATE error at unmarshal data at create profile %v", err)
				render.JSON(500, map[string]interface{}{"error":err, "success":false})
				return
			}
			profile = ProfileTextTagClear(profile)
			log.Printf("CS CREATE profile: %+v", profile)
			profile, err = ph.InsertNewProfile(profile)
			if err != nil {
				log.Printf("CS CREATE DB are not available")
				render.JSON(500, map[string]interface{}{"error":err, "success":false})
			}
			out, err := json.Marshal(profile)
			if err != nil {
				log.Printf("CS CREATE error at marshal data to out")
				render.JSON(500, map[string]interface{}{"error":err, "success":false})
			}
			render.JSON(200, map[string]interface{}{"success":true, "data":out})
		})

		r.Post("/update", func(render render.Render, params martini.Params, req *http.Request) {
			data, err := ioutil.ReadAll(req.Body)
			log.Printf("CS UPDATE data: %s", data)
			if err != nil {
				log.Printf("CS UPDATE error at reading post data %v", err)
				render.JSON(500, map[string]interface{}{"error":err, "success":false})
				return
			}
			profile := &Profile{}
			err = json.Unmarshal(data, profile)
			if err != nil {
				log.Printf("CS UPDATE error at unmarshal data at create profile %v", err)
				render.JSON(500, map[string]interface{}{"error":err, "success":false})
				return
			}
			profile = ProfileTextTagClear(profile)
			log.Printf("CS UPDATE profile: %+v", profile)
			err = ph.UpdateProfile(profile)
			if err != nil {
				render.JSON(500, map[string]interface{}{"error":err, "success":false})
				return
			}
			render.JSON(200, map[string]interface{}{"success":true})
		})

		r.Post("/delete", func(render render.Render, params martini.Params, req *http.Request) {
			data, err := ioutil.ReadAll(req.Body)
			if err != nil {
				log.Printf("CS DELETE error at reading post data %v", err)
				render.JSON(500, map[string]interface{}{"error":err, "success":false})
				return
			}
			log.Printf("CS DELETE data: %s", data)

			info := ProfileId{}
			err = json.Unmarshal(data, &info)
			if err != nil {
				log.Printf("CS DELETE error at unmarshal delete data %v", err)
				render.JSON(500, map[string]interface{}{"error":err, "success":false})
				return
			}
			ph.DeleteProfile(info.Id)
			render.JSON(200, map[string]interface{}{"success":true})
		})
		r.Post("/upload_img/:profile_id", func(render render.Render, params martini.Params, req *http.Request) {
			profile_id := params["profile_id"]
			path := fmt.Sprintf("%v/%v", cfg.Console.ProfileImgPath, profile_id)
			file, handler, err := req.FormFile("img_file")
			defer file.Close()
			if err != nil {
				log.Printf("CS error at forming file %v", err)
				render.JSON(500, map[string]interface{}{"error":err, "success":false})
				return
			}

			if !strings.Contains(handler.Header.Get("Content-Type"), "image") {
				render.JSON(200, map[string]interface{}{"error":"Вы загружаете не картинку", "success":false})
				return
			}

			err = os.Mkdir(path, 0777)
			if err != nil {
				log.Printf("CS warn at mkdir %v", err)
			}
			file_path := fmt.Sprintf("%v/%v", path, handler.Filename)
			f, err := os.OpenFile(file_path, os.O_WRONLY | os.O_CREATE, 0664)
			defer f.Close()
			if err != nil {
				log.Printf("CS error at open file %v", err)
				render.JSON(500, map[string]interface{}{"error":err, "success":false})
				return
			}
			log.Printf("CS will save file at: [%v]", file_path)
			io.Copy(f, file)
			file_url := fmt.Sprintf("%v/%v/%v", cfg.Console.ProfileImgServer, profile_id, handler.Filename)
			log.Printf("CS will form image at: [%v]", file_url)

			profile, _ := ph.GetProfile(profile_id)
			if profile == nil {
				profile = &Profile{UserName:profile_id}
			}
			profile.ImageURL = file_url
			ph.UpdateProfile(profile)

			render.JSON(200, map[string]interface{}{"success":true, "url":file_url})
		})
		r.Get("/all_groups", func(ren render.Render) {
			groups, err := ph.GetAllGroups()
			if err != nil {
				log.Printf("CS error at groups retrieve: %v", err)
				ren.JSON(500, map[string]interface{}{"error":err, "success":false})
				return
			}
			log.Printf("CS forming next groups: %+v", groups)
			ren.JSON(200, map[string]interface{}{"success":true, "groups":groups})
		})
	})

	r.Group("/chat", func(r martini.Router) {

		r.Get("", func(r render.Render, params martini.Params, req *http.Request) {
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
				messages, _ = db.Messages.GetMessages(bson.M{"to":with})
			}
			result_data["collocutor"] = collocutor
			result_data["with"] = with
			result_data["messages"] = messages

			if contacts, err := GetContacts(db); err == nil {
				result_data["contacts"] = contacts
			}
			log.Printf("CS result data :%+v", result_data)
			r.HTML(200, "chat", result_data, render.HTMLOptions{Layout:"base"})
		})

		r.Post("/send", func(render render.Render, req *http.Request) {
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
			if message.From != "" && message.To != "" && message.Body != "" {
				if message.To == ALL {
					peoples, _ := db.Users.GetBy(bson.M{})
					ntf.SendMessageToPeople(peoples, message.Body)

				} else if message.To == "all_hash_writers" {
					peoples, _ := db.Users.GetBy(bson.M{"last_marker":bson.M{"$exists":true}})
					ntf.SendMessageToPeople(peoples, message.Body)

				} else {
					user, _ := db.Users.GetUserById(message.To)
					if user != nil {
						db.Messages.SetMessagesRead(user.UserId)
						go ntf.NotifyText(message.To, message.Body)
					}

				}
				if err != nil {
					render.JSON(500, map[string]interface{}{"error":err})
				}
			} else {
				render.Redirect("/chat")
			}
			render.JSON(200, map[string]interface{}{"ok":true, "message":d.NewMessageForWeb(message.From, message.To, message.Body)})
		})

		r.Post("/messages_read", func(render render.Render, req *http.Request) {
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
			err = db.Messages.SetMessagesRead(readed.From)
			if err != nil {
				log.Printf("CS QE E: at unmarshal json messages %v\ndata:%s", err, data)
				render.JSON(500, map[string]interface{}{"error":err})
				return
			}
			render.JSON(200, map[string]interface{}{"ok":true})
		})

		r.Post("/messages", func(render render.Render, req *http.Request) {
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
					if u.ShowedName != "" {
						messages[i].From = u.ShowedName
					} else {
						messages[i].From = u.UserName
					}
					result = append(result, messages[i])
				}
			}
			render.JSON(200, map[string]interface{}{"messages":result, "next_":time.Now().Unix()})
		})

		r.Post("/contacts", func(render render.Render, req *http.Request) {
			type NewContactsReq struct {
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
			contacts, err := GetContacts(db)
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
		r.Delete("/delete_messages", func(params martini.Params, ren render.Render, req *http.Request) {
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

		r.Post("/contacts_change", func(render render.Render, req *http.Request) {
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
	})

	r.Get("/vote_result", func(ren render.Render) {
		votes, err := vdh.GetTopVotes(-1)
		if err != nil{
			log.Printf("CS ERROR at retrieving votes %v", err)
		}
		ren.HTML(200, "vote_result", map[string]interface{}{"votes":votes}, render.HTMLOptions{Layout:"base"})
	})

	r = EnsureWorkWithKeys(r, qs)
	r = EnsureWorkWithUsers(r, db)

	r.Get("/statistic/taxi", func(render render.Render) {
		render.HTML(200, "console/statistic", map[string]interface{}{"providers":[]string{"academ"}})
	})

	m.Action(r.Handle)
	m.RunOnAddr(addr)
}
