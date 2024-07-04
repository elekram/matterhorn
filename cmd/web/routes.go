package main

import (
	"embed"
	"net/http"
)

//go:embed static/*
var content embed.FS

func (app *application) router() *http.ServeMux {

	// mux.HandleFunc("GET /v1/status", app.status)
	app.mux.HandleFunc("GET /", app.home)
	app.mux.HandleFunc("GET /signout", app.signout)

	app.mux.Handle("GET /static/", http.StripPrefix(
		"/", http.FileServer(http.FS(content))))

	return &app.mux
}
