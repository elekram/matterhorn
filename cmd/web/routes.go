package main

import (
	"embed"
	"net/http"

	"github.com/elekram/matterhorn/cmd/web/module"
)

//go:embed static/*
var content embed.FS

type handlers struct {
	handleSignIn     http.Handler
	handleOAuth      http.Handler
	handleAppBase    http.Handler
	handleModuleBase http.Handler
}

func (a *app) registerRouteHandlers() {
	h := handlers{
		handleSignIn:     handleSignIn(a.cfg),
		handleOAuth:      handleOAuth(a.oAuth2Config, a),
		handleAppBase:    handleAppBase(),
		handleModuleBase: module.Base(a.cfg, a.dbCon),
	}

	a.handlers = &h
}

func (s *app) registerRoutes() {
	s.router.Handle("GET /static/", http.StripPrefix(
		"/", http.FileServer(http.FS(content))))

	s.router.Handle("GET /signin", s.handlers.handleSignIn)
	s.router.Handle("GET /auth/oauth", s.handlers.handleOAuth)

	s.router.Handle("GET /", s.handlers.handleAppBase)
}
