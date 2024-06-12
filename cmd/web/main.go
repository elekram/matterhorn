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

type Middleware func(http.HandlerFunc) http.HandlerFunc

func (app *application) use(handler http.HandlerFunc, m ...Middleware) http.HandlerFunc {

	if len(m) < 1 {
		return handler
	}

	wrappedHandler := handler

	for i := len(m) - 1; i >= 0; i-- {
		wrappedHandler = m[i](wrappedHandler)
	}
	return wrappedHandler
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<b>Wassssup my dude?<b>"))
}

func LogMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		log.SetOutput(os.Stdout) // logs go to Stderr by default
		log.Println(r.Method, r.URL)
		h.ServeHTTP(w, r) // call ServeHTTP on the original handler

	})
}

type application struct {
	config     environment.Config
	logger     *log.Logger
	middleware []Middleware
}

func (app *application) status(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "application:", app.config.AppName)
	fmt.Fprintln(w, "status: online")
}

func test(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("hello from test() middleware")
	})
}

func (app *application) handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<b>Wassssup my dude?<b>"))
}

func main() {
	config := environment.NewConfig()
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	app := &application{
		config: *config,
		logger: logger,
		middleware: []Middleware{
			LogMiddleware,
		},
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

	log.Printf("Starting server on port %s", config.Port)
	log.Fatal(srv.ListenAndServeTLS("", ""))
}
