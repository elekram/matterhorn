package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"text/template"
	"time"

	appcfg "github.com/elekram/matterhorn/config"
	"golang.org/x/oauth2"
)

func handleSignIn(cfg *appcfg.ConfigProperties) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		files := []string{
			"./cmd/web/templates/_signin.tmpl",
		}

		ts := template.Must(template.ParseFiles(files...))

		data := struct {
			AppName string
		}{
			AppName: cfg.AppName,
		}

		err := ts.Execute(w, data)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", 500)
		}
	})
}

func handleOAuth(OAuthCfg *oauth2.Config, a *app) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stateId := a.session.generateId(10)
		a.session.oAuthStateParemeterPool[stateId] = oAuthStateParemeters{
			redirectUrl: "",
			expiresOn:   time.Now().Add(time.Second * time.Duration(300)),
		}

		url := OAuthCfg.AuthCodeURL(stateId, oauth2.AccessTypeOffline)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})
}

func handleAppBase() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "404 Error: handler for %s not found", html.EscapeString(r.URL.Path))
			return
		}

		fmt.Println("test")

		files := []string{
			"./cmd/web/templates/_app-base.tmpl",
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
