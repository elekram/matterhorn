package module

import (
	"html/template"
	"log"
	"net/http"

	appcfg "github.com/elekram/matterhorn/config"
	database "github.com/elekram/matterhorn/db"
)

func Base(cfg *appcfg.ConfigProperties, db database.AppDb) http.Handler {
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
