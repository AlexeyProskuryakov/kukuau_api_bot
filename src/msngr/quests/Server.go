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
	"errors"
	"fmt"
	"strings"
	"io/ioutil"
	"strconv"

	"time"
	"regexp"
)

var users = map[string]string{
	"alesha":"sederfes100500",
	"leha":"qwerty100500",
	"dima":"123",
}

type MessageUserInfo struct {
	Name    string
	LastKey string
	NextKey string
}
type ShowMessage struct {
	SID  string
	From MessageUserInfo
	Body string
}


func GetMessages(qs *QuestStorage) []ShowMessage {
	messages, _ := qs.GetMessages(bson.M{"answered":false, "is_key":false})
	s_users, _ := qs.GetAllUsers()
	s_user_map := map[string]QuestUserWrapper{}
	for _, s_user := range s_users {
		s_user_map[s_user.UserId] = s_user
	}
	out_messages := []ShowMessage{}
	for _, message := range messages {
		if u_info, ok := s_user_map[message.From]; ok {
			message.From = GetUserName(u_info, message.From)
			var last_key string
			var next_key string
			if _last_key, ok := u_info.LastKey[PROVIDER]; ok && _last_key != nil {
				last_key = *_last_key
				k_info, _ := qs.GetKeyInfo(*_last_key)
				if k_info != nil && k_info.NextKey != nil {
					nkp := k_info.NextKey
					next_key = *nkp
				}
			}

			out_message := ShowMessage{
				From:MessageUserInfo{
					Name:message.From,
					LastKey:last_key,
					NextKey:next_key,
				},
				Body:message.Body,
				SID:message.SID,
			}

			out_messages = append(out_messages, out_message)
		}
	}
	return out_messages
}

func ParseExportTxt(raw_data string, qs *QuestStorage) error {
	keys := strings.Fields(string(raw_data))
	for i, key := range keys {
		key_params := strings.Split(key, ";")
		next_key := key_params[2]
		var is_first bool
		if len(key_params) == 4 {
			is_first = key_params[3] == "true" || i == 0
		} else {
			is_first = i == 0
		}
		qs.AddKey(key_params[0], key_params[1], &next_key, is_first)
	}
	return nil
}

func ParseExportXlsx(xlf *xlsx.File, qs *QuestStorage, skip_row, skip_cell int) error {
	for _, sheet := range xlf.Sheets {
		if sheet != nil {
			sh_name := strings.TrimSpace(strings.ToLower(sheet.Name))
			if strings.HasSuffix(sh_name, "ключ") || strings.HasPrefix(sh_name, "ключ") {
				is_first := true
				for ir, row := range sheet.Rows {
					if row != nil && ir >= skip_row {
						key := row.Cells[skip_cell].Value
						description := row.Cells[skip_cell + 1].Value
						next_key_raw := row.Cells[skip_cell + 2].Value
						var next_key *string
						if next_key_raw != "" {
							next_key = &next_key_raw
						}
						if key != "" && description != ""{
							qs.AddKey(key, description, next_key, is_first)
						}
						is_first = false
					}
				}
			}
		}
	}
	return nil
}

func GetUserName(u_info QuestUserWrapper, default_res string) string {
	if u_info.Name != "" && u_info.Phone != "" {
		return fmt.Sprintf("%v (%v)", u_info.Name, u_info.Phone)
	} else if u_info.Name != "" && u_info.EMail != "" {
		return fmt.Sprintf("%v (%v)", u_info.Name, u_info.EMail)
	} else if u_info.Name != "" {
		return u_info.Name
	} else if u_info.Phone != "" {
		return u_info.Phone
	} else if u_info.EMail != "" {
		return u_info.EMail
	} else {
		return default_res
	}
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


	m.Get("/", func(user auth.User, render render.Render) {
		render.HTML(200, "quests/index", map[string]interface{}{})
	})

	get_result_error_map := func(error_info string) map[string]interface{} {
		keys, _ := qs.GetAllKeys()
		result_map := map[string]interface{}{
			"keys": keys,
			"error_text":error_info,
			"is_error":true,
		}
		return result_map
	}

	get_keys_map := func() map[string]interface{} {
		keys, err := qs.GetAllKeys()
		if err != nil {
			log.Printf("Error for load keys: %v", err)
		}
		result_map := map[string]interface{}{
			"keys": keys,
		}
		return result_map
	}

	m.Get("/new_keys", func(render render.Render) {
		render.HTML(200, "quests/new_keys", get_keys_map())
	})


	m.Post("/add_key", func(user auth.User, render render.Render, request *http.Request) {
		key := request.FormValue("key")
		next_key_raw := request.FormValue("next-key")
		description := request.FormValue("description")
		is_first_raw := request.FormValue("is-first")
		log.Printf("QUEST: adding key: is first raw: %+v", is_first_raw)
		var is_first bool
		if is_first_raw == "on" {
			is_first = true
		}
		log.Printf("QUEST: key: %s\nAnswer: %s ", key, description)
		if key != "" && description != "" {
			var next_key *string
			if next_key_raw == "" {
				next_key = nil
			} else {
				next_key = &next_key_raw
			}
			qs.AddKey(key, description, next_key, is_first)
			render.Redirect("/new_keys")
		} else {
			render.HTML(200, "quests/keys_new", get_result_error_map("Невалидные значения ключа, ответа или позиции."))
		}
	})

	m.Post("/delete_key/:key", func(params martini.Params, render render.Render) {
		key := params["key"]
		err := qs.DeleteKey(key)
		log.Printf("QUESTS WEB will delete %v (%v)", key, err)
		render.Redirect("/new_keys")
	})

	m.Get("/users_keys", func(render render.Render) {
		users_keys, err := qs.GetMessages(bson.M{"answered":false, "is_key":true})
		if err != nil {
			log.Printf("Error at getting users keys %v", err)
		}
		for i, mk := range users_keys {
			mk.Time = time.Unix(mk.TimeStamp, 0)
			users_keys[i] = mk
		}

		result_map := map[string]interface{}{
			"keys":users_keys,
		}
		render.HTML(200, "quests/users_keys", result_map)
	})

	m.Get("/messages", func(render render.Render) {
		out_messages := GetMessages(qs)
		result_map := map[string]interface{}{
			"messages":out_messages,
			"error_text":"",
			"is_error":false,
		}
		render.HTML(200, "quests/messages", result_map)
	})


	ensure_messages_error := func(err error) map[string]interface{} {
		messages := GetMessages(qs)
		return map[string]interface{}{
			"error_text":err.Error(),
			"is_error":true,
			"messages":messages,
		}
	}

	m.Get("/user_messages/:id", func(params martini.Params, render render.Render) {
		message, _ := qs.GetMessage(params["id"])
		user_messages, _ := qs.GetMessages(bson.M{"from":message.From, "is_key":false})
		user, _ := qs.GetUserInfo(message.From, PROVIDER)

		render.HTML(200, "quests/user_messages", map[string]interface{}{
			"messages":user_messages,
			"user":GetUserName(user.User, message.From),
		})
	})

	m.Post("/message_answer/:id", func(params martini.Params, user auth.User, render render.Render, request *http.Request) {
		answer := request.FormValue("message_answer")
		log.Printf("Operator was answer: %s", answer)
		if answer != "" {
			message, err := qs.GetMessage(params["id"])
			if err != nil {
				render.HTML(200, "quests/messages", ensure_messages_error(err))
			}
			go func() {
				ntf.Notify(structs.OutPkg{To:message.From,
					Message: &structs.OutMessage{
						ID: utils.GenId(),
						Type: "chat",
						Body: answer,
					}})
			}()
			qs.SetMessageAnswer(message.ID)
			render.Redirect("/messages")
		} else {
			render.HTML(200, "quests/messages", ensure_messages_error(errors.New("Сообщение не может быть пустым.")))
		}
	})

	m.Post("/message_answer_all", func(render render.Render, request *http.Request) {
		users, err := qs.GetSubscribedUsers()
		if err != nil {
			render.HTML(200, "quests/messages", ensure_messages_error(err))
		}
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
			render.HTML(200, "quests/messages", ensure_messages_error(errors.New("Сообщение не может быть пустым.")))
		}
		render.Redirect("/messages")
	})

	m.Get("/messages/new_count/:after", func(render render.Render, params martini.Params) {
		after_input, err := strconv.ParseInt(params["after"], 10, 64)
		if err != nil {
			render.JSON(200, map[string]interface{}{"error":err.Error()})
		}
		messages, err := qs.GetMessages(bson.M{
			"answered":false,
			"is_key":false,
			"time":bson.M{"$gte":after_input},
		})

		if err != nil {
			render.JSON(200, map[string]interface{}{"error":err.Error()})
		}else {
			render.JSON(200, map[string]interface{}{"error":false, "count":len(messages) })
		}
	})

	m.Get("/load/klichat_quest_keys.txt", func(render render.Render) {
		var str_buff string
		keys, err := qs.GetAllKeys()
		if err != nil {
			err_message := []byte(fmt.Sprintf("Error at getting all keys: %v", err.Error()))
			render.Data(500, err_message)
		}
		for _, key := range keys {
			var next_key string
			if key.NextKey != nil {
				next_key_p := key.NextKey
				next_key = *next_key_p
			}

			str_buff += fmt.Sprintf("%s;%s;%s;%v\r\n", key.Key, strings.TrimSpace(key.Description), next_key, key.IsFirst)
		}
		render.Data(200, []byte(str_buff))
	})


	xlsFileReg := regexp.MustCompile(".+\\.xlsx?")

	m.Post("/load/up", func(render render.Render, request *http.Request) {
		file, header, err := request.FormFile("file")

		log.Printf("Form file information: file: %+v \nheader:%v, %v\nerr:%v", file, header.Filename, header.Header, err)

		if err != nil {
			render.HTML(200, "quests/keys_new", get_result_error_map(fmt.Sprintf("Ошибка загрузки файлика: %v", err)))
			return
		}
		defer file.Close()

		data, err := ioutil.ReadAll(file)
		if err != nil {
			render.HTML(200, "quests/keys_new", get_result_error_map(fmt.Sprintf("Ошибка загрузки файлика: %v", err)))
			return
		}

		if xlsFileReg.MatchString(header.Filename) {
			xlFile, err := xlsx.OpenBinary(data)
			log.Printf("file: %+v, err: %v", xlFile, err)
			if err != nil || xlFile == nil {
				render.HTML(200, "quests/new_keys", get_result_error_map(fmt.Sprintf("Ошибка обработки файлика: %v", err)))
				return
			}
			skip_rows, _ := strconv.Atoi(request.FormValue("skip-rows"))
			skip_cols, _ := strconv.Atoi(request.FormValue("skip-cols"))

			ParseExportXlsx(xlFile, qs, skip_rows, skip_cols)
		} else {
			raw_data := string(data)
			log.Printf("Result: %s", raw_data)
			ParseExportTxt(raw_data, qs)
		}

		render.Redirect("/new_keys")
	})

	log.Printf("Will start web server for quest at: %v", config.WebPort)
	m.RunOnAddr(config.WebPort)
}
