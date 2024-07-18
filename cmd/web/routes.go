package main

import (
	"embed"
	"net/http"

	"github.com/elekram/matterhorn/cmd/web/app"
)

//go:embed static/*
var content embed.FS

func router() *http.ServeMux {
	mux := *http.NewServeMux()

	mux.HandleFunc("GET /", app.Root)
	mux.HandleFunc("GET /signin", app.SignIn)

	mux.Handle("GET /static/", http.StripPrefix(
		"/", http.FileServer(http.FS(content))))

	return &mux
}
