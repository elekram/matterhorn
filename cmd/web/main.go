package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	env "github.com/elekram/matterhorn/config"
)

var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)

func main() {
	env.Config = env.NewConfig()

	serverTLSKeys, err := tls.LoadX509KeyPair(env.Config.TLSPublicKey, env.Config.TLSPrivateKey)
	if err != nil {
		logger.Fatalf("Error loading TLS public/private keys: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverTLSKeys},
	}

	mux := secureHeaders(disableCache(requestLogger(router())))

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", env.Config.Port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		TLSConfig:    tlsConfig,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	defer srv.Close()

	log.Printf("Serving on port %s 💁🏻", env.Config.Port)
	log.Fatal(srv.ListenAndServeTLS("", ""))
}
