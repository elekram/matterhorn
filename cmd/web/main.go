package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	appcfg "github.com/elekram/matterhorn/config"
)

func main() {
	cfg := appcfg.NewConfig()
	app := newAppServer(cfg)

	serverTLSKeys, err := tls.LoadX509KeyPair(app.cfg.TLSPublicKey, app.cfg.TLSPrivateKey)
	if err != nil {
		app.logger.Fatalf("Error loading TLS public/private keys: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverTLSKeys},
	}

	app.registerRoutes(app.router)

	handler := secureHeaders(
		disableCache(
			requestLogger(
				app.session.manageSession(app))))

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
