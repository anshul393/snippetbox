package main

import (
	"html/template"
	"path/filepath"
	"time"

	db "github.com/anshul393/snippetbox/db/sqlc"
)

type User struct {
	Name   string
	Email  string
	Joined time.Time
}

type templateData struct {
	CurrentYear     int
	Snippet         *db.Snippet
	Snippets        []*db.Snippet
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
	User            *User
}

func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {

	cache := make(map[string]*template.Template)

	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts := template.New(name).Funcs(functions)

		ts, err := ts.ParseFiles("./ui/html/pages/base.tmpl.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("./ui/html/partials/*.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts

	}

	return cache, nil
}
