package main

import (
	"embed"
	"net/http"
)

//go:embed static/*
var content embed.FS

func (app *application) router() *http.ServeMux {
	mux := http.NewServeMux()

	// mux.HandleFunc("GET /v1/status", app.status)
	mux.HandleFunc("GET /", app.status)

	mux.Handle("GET /static/", http.StripPrefix(
		"/", http.FileServer(http.FS(content))))

	return mux
}
