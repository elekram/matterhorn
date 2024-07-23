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

var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)

func main() {
	appcfg.Props = appcfg.NewConfig()

	serverTLSKeys, err := tls.LoadX509KeyPair(appcfg.Props.TLSPublicKey, appcfg.Props.TLSPrivateKey)
	if err != nil {
		logger.Fatalf("Error loading TLS public/private keys: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverTLSKeys},
	}

	db := appcfg.Props.MongoDb
	dbUser := appcfg.Props.MongoUsername
	dbPassword := appcfg.Props.MongoPassword

	database.DBCon = database.NewConnection(db, dbUser, dbPassword)

	mux := secureHeaders(disableCache(requestLogger(router())))

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", appcfg.Props.Port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		TLSConfig:    tlsConfig,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	defer srv.Close()

	log.Printf("Serving on port %s 💁🏻", appcfg.Props.Port)
	log.Fatal(srv.ListenAndServeTLS("", ""))
}
