package main

import (
	"io"
	"log"
	"net/http"

	appcfg "github.com/elekram/matterhorn/config"
	database "github.com/elekram/matterhorn/db"
	"github.com/elekram/matterhorn/internal/bodymap"
	"golang.org/x/oauth2"
)

type app struct {
	cfg          *appcfg.ConfigProperties
	dbCon        database.AppDb
	router       *http.ServeMux
	oAuth2Config *oauth2.Config
	session      *sessionMgr
	logger       *log.Logger
	parseForm    bodymap.FormParser
	handlers     *handlers
}

func newAppServer(
	cfg *appcfg.ConfigProperties,
	logger *log.Logger,
	sessionMgr *sessionMgr,
	oAuth2Config *oauth2.Config,
	dbCon database.AppDb,
	parseForm func(io.ReadCloser) map[string]string,
) *app {

	router := &http.ServeMux{}

	a := app{
		cfg:          cfg,
		router:       router,
		dbCon:        dbCon,
		session:      sessionMgr,
		oAuth2Config: oAuth2Config,
		logger:       logger,
		parseForm:    parseForm,
	}

	return &a
}
