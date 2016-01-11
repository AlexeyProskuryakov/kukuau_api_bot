package quests

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/martini-contrib/render"

	c "msngr/configuration"

	"msngr/notify"
	"msngr/structs"


	"log"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"msngr/utils"
	"errors"
)

var users = map[string]string{
	"alesha":"sederfes100500",
	"leha":"qwerty100500",
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

	get_result_error_map := func(user auth.User, error_info string) map[string]interface{} {
		keys, _ := qs.GetAllKeys()
		result_map := map[string]interface{}{
			"keys": keys,
			"error_text":error_info,
			"is_error":true,
		}
		return result_map
	}

	get_result_map := func(user auth.User) map[string]interface{} {
		keys, _ := qs.GetAllKeys()
		result_map := map[string]interface{}{
			"keys": keys,
			"error_text":"",
			"is_error":false,
		}
		return result_map
	}

	m.Get("/new_keys", func(user auth.User, render render.Render) {
		render.HTML(200, "quests/new_keys", get_result_map(user))
	})


	m.Post("/add_key", func(user auth.User, render render.Render, request *http.Request) {
		key := request.FormValue("key")
		next_key_raw := request.FormValue("next-key")
		description := request.FormValue("description")
		is_first_raw := request.FormValue("is-first")
		log.Printf("QUEST: adding key: is first raw: %+v", is_first_raw)
		var is_first bool
		if is_first_raw == "on"{
			is_first = true
		}
		log.Printf("QUEST: key: %s\nAnswer: %s ", key, description)
		if key != "" && description != ""{
			var next_key *string
			if next_key_raw == ""{
				next_key = nil
			} else {
				next_key = &next_key_raw
			}
			qs.AddKey(key, description, next_key, is_first)
			render.Redirect("/new_keys")
		} else {
			render.HTML(200, "quests/keys_new", get_result_error_map(user, "Не валидные значения ключа, ответа или позиции."))
		}
	})

	m.Post("/delete_key/:key", func(params martini.Params, render render.Render) {
		key := params["key"]
		err := qs.DeleteKey(key)
		log.Printf("QUESTS WEB will delete %v (%v)", key, err)
		render.Redirect("/new_keys")
	})

	m.Get("/users_keys", func(render render.Render){
		users_keys, _ := qs.GetMessages(bson.M{"data.answered":false, "is_key":true})
		result_map := map[string]interface{}{
			"keys":users_keys,
		}
		render.HTML(200, "quests/users_keys", result_map)
	})

	m.Get("/messages", func(user auth.User, render render.Render) {
		messages, _ := qs.GetMessages(bson.M{"data.answered":false, "is_key":false})
		result_map := map[string]interface{}{
			"messages":messages,
			"error_text":"",
			"is_error":false,
		}
		render.HTML(200, "quests/messages", result_map)
	})


	ensure_messages_error := func(err error) map[string]interface{} {
		messages, _ := qs.GetMessages(bson.M{"data.answered":false})
		return map[string]interface{}{
			"error_text":err.Error(),
			"is_error":true,
			"messages":messages,
		}
	}

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

	m.Post("/message_all", func(render render.Render, request *http.Request) {
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

	log.Printf("Will start web server for quest at: %v", config.WebPort)
	m.RunOnAddr(config.WebPort)
}
