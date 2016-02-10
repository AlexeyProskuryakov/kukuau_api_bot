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
)

func Run(addr string, notifier *ntf.Notifier,  db *d.MainDb, cs c.ConfigStorage) {
	m := martini.Classic()

	martini.Env = martini.Dev
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
	m.Get("/", func() {})

	m.Post("/configuration", func(request *http.Request, render render.Render) {
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

	m.RunOnAddr(addr)
}