package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	appcfg "github.com/elekram/matterhorn/config"
	"golang.org/x/oauth2"
)

func signIn(cfg *appcfg.ConfigProperties, failedAuth bool) http.Handler {
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

func handleOAuth(OAuthCfg *oauth2.Config, a *app) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stateId := a.session.generateId(30) 
		a.session.oAuthIdPool[stateId] = oAuthIdPool{
			expiry: time.Now().Add(time.Second * time.Duration(300)),
		}

		fmt.Println("1: " + stateId)

		url := OAuthCfg.AuthCodeURL(stateId, oauth2.AccessTypeOffline)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})
}

func handleOAuthCallback(OAuthCfg *oauth2.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")

		// Exchanging the code for an access token
		token, err := OAuthCfg.Exchange(context.Background(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Creating an HTTP client to make authenticated request using the access key.
		// This client method also regenerate the access key using the refresh key.
		client := OAuthCfg.Client(context.Background(), token)

		// Getting the user public details from google API endpoint
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Closing the request body when this function returns.
		// This is a good practice to avoid memory leak
		defer resp.Body.Close()

		var v any

		// Reading the JSON body using JSON decoder
		err = json.NewDecoder(resp.Body).Decode(&v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// sending the user public value as a response. This is may not be a good practice,
		// but for demonstration, I think it serves the need.
		fmt.Fprintf(w, "%v", v)
	})
}
