package quests

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/martini-contrib/render"

	c "msngr/configuration"
	"msngr/console"

	"log"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

var users = map[string]string{
	"alesha":"sederfes100500",
	"leha":"qwerty100500",
}

func Run(config c.QuestConfig, qs *QuestStorage) {
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

	get_result_map := func(user auth.User) map[string]interface{}{
		keys, _ := qs.GetAllKeys()
		messages, _ := qs.GetMessages(bson.M{"answered":false})
		result_map := map[string]interface{}{
			"user":console.IndexInfo{UserName:string(user)},
			"keys": keys,
			"messages":messages,
			"error_text":"",
			"is_error":false,
		}
		return result_map
	}

	get_result_error_map := func(user auth.User, error_info string) map[string]interface{}{
		keys, _ := qs.GetAllKeys()
		messages, _ := qs.GetMessages(bson.M{"answered":false})
		result_map := map[string]interface{}{
			"user":console.IndexInfo{UserName:string(user)},
			"keys": keys,
			"messages":messages,
			"error_text":error_info,
			"is_error":true,
		}
		return result_map
	}
	m.Get("/",func(user auth.User, render render.Render){
		render.HTML(200, "quests/index", get_result_map(user))
	})

	m.Post("/", func(user auth.User, render render.Render, request *http.Request){
		key := request.FormValue("key")
		answer := request.FormValue("answer")
		log.Printf("Key: %s\nAnswer: %s", key, answer)
		if key != "" && answer != ""{
			qs.AddKey(key, answer)
			render.HTML(200, "quests/index", get_result_map(user))
		} else {
			render.HTML(200, "quests/index", get_result_error_map(user, "Не валидные значения ключа и ответа."))
		}
	})

	log.Printf("Will start web server for quest at: %v", config.WebPort)
	m.RunOnAddr(config.WebPort)
}
