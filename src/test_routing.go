package main

import (
	"net/http"

	"msngr/db"
	"msngr/configuration"
	"msngr/web"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

func formTestMartini() *martini.Martini {
	config := configuration.ReadTestConfigInRecursive()
	mainDb := db.NewMainDb(config.Main.Database.ConnString, config.Main.Database.Name)
	m := web.NewSessionAuthorisationHandler(mainDb)
	m.Use(render.Renderer(render.Options{
		Directory:"templates/auth",
		Extensions: []string{".tmpl", ".html"},
		Charset: "UTF-8",

	}))
	router := martini.NewRouter()

	router.Get("/", func(r render.Render) {
		r.HTML(200, "login", nil)
	})
	router.Group("/test", func(r martini.Router) {
		r.Get("/:id", func(ren render.Render, prms martini.Params) {
			id := prms["id"]
			ren.JSON(200, map[string]interface{}{"id":id})
		})
		r.Get("/new", func(ren render.Render) {
			ren.JSON(200, nil)
		})
		r.Get("/update/:id", func(ren render.Render, prms martini.Params) {
			id := prms["id"]
			ren.JSON(200, map[string]interface{}{"id":id})
		})
		r.Get("/delete/:id", func(ren render.Render, prms martini.Params) {
			id := prms["id"]
			ren.JSON(200, map[string]interface{}{"id":id})
		})
	})
	m.Action(router.Handle)
	return m
}

func main() {
	mux := http.NewServeMux()
	http.HandleFunc("/t", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("foooooooo"))
	})
	http.Handle("/t/route/*", formTestMartini())
	server := &http.Server{
		Addr: ":9292",
	}
	server.ListenAndServe()
}
