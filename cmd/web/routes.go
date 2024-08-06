package main

import (
	"embed"
	"net/http"

	appBase "github.com/elekram/matterhorn/cmd/web/appbase"
)

//go:embed static/*
var content embed.FS

type handlers struct {
	handleSignIn        http.Handler
	handleHomePage      http.Handler
	handleOAuthCallback http.Handler
	handleOAuth         http.Handler
}

func (a *app) registerRouteHandlers() {
	h := handlers{
		handleSignIn:        signIn(a.cfg, false),
		handleOAuth:         handleOAuth(a.oAuth2Config, a),
		handleOAuthCallback: handleOAuthCallback(a.oAuth2Config),
		handleHomePage:      appBase.Home(a.cfg, a.dbCon),
	}

	a.handlers = &h
}

func (s *app) registerRoutes() {
	s.router.Handle("GET /static/", http.StripPrefix(
		"/", http.FileServer(http.FS(content))))

	s.router.Handle("GET /signin", s.handlers.handleSignIn)
	s.router.Handle("GET /auth/oauth", s.handlers.handleOAuth)
	s.router.Handle("GET /oauth2/redirect/google", s.handlers.handleOAuthCallback)

	s.router.Handle("GET /", s.handlers.handleHomePage)
}
