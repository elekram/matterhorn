package appBase

import (
	"html/template"
	"log"
	"net/http"

	appcfg "github.com/elekram/matterhorn/config"
	database "github.com/elekram/matterhorn/db"
)

func SignIn(cfg *appcfg.ConfigProperties, failedAuth bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		files := []string{
			"./cmd/web/templates/appbase-signin.tmpl",
		}

		ts := template.Must(template.ParseFiles(files...))

		err := ts.Execute(w, failedAuth)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", 500)
		}
	})
}

func Home(cfg *appcfg.ConfigProperties, db database.AppDb) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		files := []string{
			"./cmd/web/templates/appbase-home.tmpl",
		}

		ts, err := template.ParseFiles(files...)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", 500)
			return
		}

		err = ts.Execute(w, nil)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", 500)
		}
	})
}
