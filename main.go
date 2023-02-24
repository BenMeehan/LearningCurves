package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"facette.io/natsort"
	"github.com/julienschmidt/httprouter"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
)

type M map[string]interface{}

var tmpl *template.Template
var dt = make(map[string][]string)

var topics = ""

func getCards() {
	for k, _ := range dt {
		t := `<div class="card mx-4">
				<div class="col">
					<div class="card-body">
					<h5 class="card-title">` + strings.Title(k) + `</h5>
					<a class="btn btn-primary" href="/c/` + k + `/01.intro">Learn</a>
					</div>
				</div>
			  </div>`
		topics += t
	}
}

func getDirectories() []string {
	var dirs []string
	entries, err := os.ReadDir("tuts")
	if err != nil {
		log.Fatal(err)
	}

	for _, e := range entries {
		dirs = append(dirs, e.Name())
	}
	return dirs
}

func getFiles(dir string) []string {
	var files []string
	var filenames []string
	fs, err := ioutil.ReadDir("tuts/" + dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range fs {
		name := f.Name()
		filenames = append(filenames, name)
	}
	natsort.Sort(filenames)
	for _, name := range filenames {
		files = append(files, `<li class="list-group-item">`+`<a href="/c/`+dir+"/"+name[:len(name)-3]+`">`+name[:len(name)-3]+`</a></li>`)
	}
	return files
}

func prePopulate() {
	dirs := getDirectories()
	for _, d := range dirs {
		dt[d] = getFiles(d)
	}
}

func pageHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	parser := parser.NewWithExtensions(extensions)

	course := ps.ByName("course")
	article := ps.ByName("article")

	dat, err := os.ReadFile("tuts/" + course + "/" + article + ".md")
	if err != nil {
		fmt.Fprint(w, err.Error())
	}
	md := dat
	html := markdown.ToHTML(md, parser, nil)

	p := M{"title": strings.Title(article[3:]), "content": template.HTML(html), "list": template.HTML(strings.Join(dt[course], "\n"))}
	err = tmpl.ExecuteTemplate(w, "page", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func homeHanlder(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	p := M{"topics": template.HTML(topics)}
	err := tmpl.ExecuteTemplate(w, "home", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	prePopulate()
	getCards()
	var err error
	tmpl, err = template.ParseGlob("templates/*")
	if err != nil {
		panic(err)
	}
	router := httprouter.New()
	router.ServeFiles("/assets/*filepath", http.Dir("assets"))
	router.GET("/", homeHanlder)
	router.GET("/c/:course/:article", pageHandler)
	log.Println("listening on port ", 8080)
	log.Fatal(http.ListenAndServe(":8080", router))
}
