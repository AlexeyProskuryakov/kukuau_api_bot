package main

import (
	"log"
	"os"
	"text/template"
	"path/filepath"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

type View struct {
	Content string
}

func main() {
	res, _ := filepath.Glob("templates\\temp\\*.html")
	log.Printf("%+v", res)
	var templates = template.Must(template.ParseGlob("templates\\temp\\*.html"))
	err := templates.ExecuteTemplate(
		os.Stdout,
		"indexPage",
		View{Content:"foo"},
	)
	if err != nil {
		log.Printf("err: %v", err)
		return
	}

	m := martini.Classic()
	// render html templates from templates directory
	m.Use(render.Renderer(render.Options{
		Directory:"templates/temp",
		Extensions:[]string{".html"},
	}))

	m.Get("/", func(r render.Render) {
		r.HTML(200, "head", View{Content:"tututrutu"}, render.HTMLOptions{Layout:"indexPage"})
	})

	m.Run()
}

