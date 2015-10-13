package console

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/martini-contrib/render"
	"fmt"
)

var users = map[string]string{
	"alesha":"sederfes100500",
	"leha":"qwerty100500",
}

func Run(addr string) {
	m := martini.Classic()
	// authenticate every request
	martini.Env = martini.Dev

	m.Use(martini.Static("console_res/assets"))
	m.Use(
		render.Renderer(render.Options{
			Directory: "console_res/templates", // Specify what path to load the templates from.
			Layout: "layout.tmpl", // Specify a layout template. Layouts can call {{ yield }} to render the current template.
			Extensions: []string{".tmpl", ".html"}, // Specify extensions to load for templates.
			//			Funcs: []template.FuncMap{AppHelpers}, // Specify helper function maps for templates to access.
			//			Delims: render.Delims{"{[{", "}]}"}, // Sets delimiters to the specified strings.
			Charset: "UTF-8", // Sets encoding for json and html content-types. Default is "UTF-8".
			IndentJSON: true, // Output human readable JSON
			IndentXML: true, // Output human readable XML
			//			HTMLContentType: "application/xhtml+xml", // Output XHTML content type instead of default "text/html"
		}))
	m.Use(auth.BasicFunc(func(username, password string) bool {
		pwd, ok := users[username]
		return ok && pwd == password
	}))

	m.Get("/", func(user auth.User, r render.Render) {
		r.HTML(200, "layout", user)
	})
	m.Get("/console", func(user auth.User,  r render.Render)  {
		r.HTML(200, "console", user)
	})
	m.Get("/statistic", func(user auth.User) (int, string) {
		return 200, fmt.Sprintf("This is statistic! %v", user)
	})
	m.RunOnAddr(addr)
}
