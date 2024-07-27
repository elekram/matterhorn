package main

import (
	"embed"
	"log"
	"net/http"
	"os"

	appBase "github.com/elekram/matterhorn/cmd/web/appbase"
	appcfg "github.com/elekram/matterhorn/config"
	database "github.com/elekram/matterhorn/db"
	"go.mongodb.org/mongo-driver/mongo"
)

type server struct {
	cfg      *appcfg.ConfigProperties
	dbCon    *mongo.Client
	router   *http.ServeMux
	session  *sessionMgr
	logger   *log.Logger
	handlers *handlers
}

type handlers struct {
	handleSignIn http.Handler
}

//go:embed static/*
var content embed.FS

func newAppServer(cfg *appcfg.ConfigProperties) *server {

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	router := &http.ServeMux{}

	db := cfg.MongoDb
	dbUser := cfg.MongoUsername
	dbPassword := cfg.MongoPassword
	dbCon := database.NewConnection(db, dbUser, dbPassword)

	sessionMgr := newSession(cfg.SessionName, cfg.SessionSecret, 30, true)

	server := server{
		cfg:     cfg,
		router:  router,
		dbCon:   dbCon,
		session: sessionMgr,
		logger:  logger,
		handlers: &handlers{
			handleSignIn: appBase.SignIn(cfg, dbCon),
		},
	}

	return &server
}

func (s *server) registerRoutes(router *http.ServeMux) {
	router.Handle("GET /static/", http.StripPrefix(
		"/", http.FileServer(http.FS(content))))

	router.Handle("GET /", appBase.Home(s.cfg, s.dbCon))
	router.Handle("GET /signin", s.handlers.handleSignIn)
}
