package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	appcfg "github.com/elekram/matterhorn/config"
	database "github.com/elekram/matterhorn/db"
)

func main() {
	cfg := appcfg.NewConfig()
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	appDb := database.NewConnection(
		cfg.MongoDb,
		cfg.MongoUsername,
		cfg.MongoPassword)

	sessionMgr := newSession(
		cfg.SessionName,
		cfg.SessionSecret,
		"120",
		true,
		newMongoStore(appDb))

	oAuth2Conf := newOAuthConfig(cfg)

	app := newAppServer(
		cfg,
		logger,
		sessionMgr,
		oAuth2Conf,
		appDb)

	app.registerRouteHandlers()
	app.registerRoutes()

	handler := secureHeaders(
		disableCache(
			requestLogger(
				app.session.manageSession(app))))

	serverTLSKeys, err := tls.LoadX509KeyPair(app.cfg.TLSPublicKey, app.cfg.TLSPrivateKey)
	if err != nil {
		app.logger.Fatalf("Error loading TLS public/private keys: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverTLSKeys},
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", app.cfg.Port),
		Handler:      handler,
		IdleTimeout:  time.Minute,
		TLSConfig:    tlsConfig,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	defer srv.Close()

	log.Printf("Serving on port %s 💁🏻", app.cfg.Port)
	log.Fatal(srv.ListenAndServeTLS("", ""))
}
