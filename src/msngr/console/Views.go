package console

import (
	"net/http"
	"github.com/martini-contrib/render"
	"io/ioutil"
	"fmt"
	"msngr/structs"
	"encoding/json"

	cfg "msngr/configuration"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"regexp"
	"log"
	"github.com/tealeg/xlsx"
	"strconv"
	w "msngr/web"
	"msngr/quests"
	"gopkg.in/mgo.v2/bson"
	d "msngr/db"
	"time"
	u "msngr/utils"
)

func ConfigurationView(request *http.Request, render render.Render, cs cfg.ConfigurationStorage) {
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
}

func EnsureWorkWithKeys(r martini.Router, qs *quests.QuestStorage) martini.Router {
	//todo add group and refactor normal
	r.Post("/load/up", func(render render.Render, request *http.Request) {
		xlsFileReg := regexp.MustCompile(".+\\.xlsx?")
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
				qs.AddStep(prel[0], prel[1], prel[2])
			}
		} else {
			render.HTML(200, "console/new_keys", GetKeysInfo("Файл имеет не то расширение :(", qs))
		}

		render.Redirect("/new_keys")
	})

	r.Get("/new_keys", func(r render.Render) {
		log.Printf("CONSOLE WEB will show keys")
		r.HTML(200, "new_keys", GetKeysInfo("", qs), render.HTMLOptions{Layout:"base"})
	})

	r.Post("/add_key", func(user auth.User, render render.Render, request *http.Request) {
		start_key := request.FormValue("start-key")
		next_key := request.FormValue("next-key")
		description := request.FormValue("description")

		log.Printf("CONSOLE WEB add key %s -> %s -> %s", start_key, description, next_key)
		if start_key != "" && description != "" {
			key, err := qs.AddStep(start_key, description, next_key)
			log.Printf("QW is error? %v key: %v", err, key)
			render.Redirect("/new_keys")
		} else {
			render.HTML(200, "console/new_keys", GetKeysInfo("Невалидные значения ключа или ответа", qs))
		}
	})

	r.Post("/delete_key/:key", func(params martini.Params, render render.Render) {
		key := params["key"]
		err := qs.DeleteStep(key)
		log.Printf("CONSOLE WEB will delete %v (%v)", key, err)
		render.Redirect("/new_keys")
	})

	r.Post("/update_key/:key", func(params martini.Params, render render.Render, request *http.Request) {
		key_id := params["key"]

		start_key := request.FormValue("start-key")
		next_key := request.FormValue("next-key")
		description := request.FormValue("description")

		err := qs.UpdateStep(key_id, start_key, description, next_key)
		log.Printf("CONSOLE WEB was update key %s %s %s %s\n err? %v", key_id, start_key, description, next_key, err)
		render.Redirect("/new_keys")
	})

	r.Get("/delete_key_all", func(render render.Render) {
		log.Printf("CONSOLE WEB was delete all keys")
		qs.Steps.RemoveAll(bson.M{})
		render.Redirect("/new_keys")
	})

	return r
}

func EnsureWorkWithUsers(r martini.Router, db *d.MainDb) martini.Router{
	r.Group("/users", func(r martini.Router) {
		r.Get("", func(r render.Render, req *http.Request) {
			r.HTML(200, "users", GetUsersInfo("", db), render.HTMLOptions{Layout:"base"})
		})

		r.Post("/add", func(user auth.User, render render.Render, request *http.Request) {
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
				render.HTML(200, "users", GetUsersInfo("Невалидные значения имени и (или) идентификатора добавляемого пользователя", db))
			}
		})

		r.Post("/delete/:id", func(params martini.Params, render render.Render) {
			uid := params["id"]
			err := db.Users.Collection.Remove(bson.M{"user_id":uid})
			log.Printf("CONSOLE WEB will delete user %v (%v)", uid, err)
			render.Redirect("/users")
		})

		r.Post("/update/:id", func(params martini.Params, render render.Render, request *http.Request) {
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
	})
	return r
}