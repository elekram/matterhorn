package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	environment "github.com/elekram/matterhorn/config"
)

// func home(w http.ResponseWriter, r *http.Request) {
// 	w.Write([]byte("<b>Wassssup my dude?<b>"))
// }

type application struct {
	config     environment.Config
	logger     *log.Logger
	middleware []Middleware
}

func (app *application) status(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "application:", app.config.AppName)
	fmt.Fprintln(w, "status: online")
}

// func (app *application) handler(w http.ResponseWriter, r *http.Request) {
// 	w.Write([]byte("<b>Wassssup my dude?<b>"))
// }

func main() {
	config := environment.NewConfig()
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	app := &application{
		config: *config,
		logger: logger,
	}

	app.middleware = []Middleware{
		requestLogger,
		secureHeaders,
		disableCache,
		app.session,
	}

	serverTLSKeys, err := tls.LoadX509KeyPair(config.TLSPublicKey, config.TLSPrivateKey)
	if err != nil {
		app.logger.Fatalf("Error loading TLS public/private keys: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverTLSKeys},
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.Port),
		Handler:      app.router(),
		IdleTimeout:  time.Minute,
		TLSConfig:    tlsConfig,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	defer srv.Close()

	log.Printf("Serving on port %s 💁🏻", config.Port)
	log.Fatal(srv.ListenAndServeTLS("", ""))
}
