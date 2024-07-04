package main

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	environment "github.com/elekram/matterhorn/config"
	oauth2 "golang.org/x/oauth2"
)

var Google = oauth2.Endpoint{
	AuthURL:       "https://accounts.google.com/o/oauth2/auth",
	TokenURL:      "https://oauth2.googleapis.com/token",
	DeviceAuthURL: "https://oauth2.googleapis.com/device/code",
}

type application struct {
	mux    http.ServeMux
	config environment.Config
	logger *log.Logger
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	templateFiles := []string{
		"./cmd/web/templates/home.page.tmpl",
		"./cmd/web/templates/_layout.tmpl",
		"./cmd/web/templates/footer.partial.tmpl",
	}

	type data struct {
		Vegetable string
	}

	d := data{
		Vegetable: "Potatos",
	}

	ts, err := template.ParseFiles(templateFiles...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = ts.Execute(w, d)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}

}

func (app *application) status(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<!DOCTYPE html><html lang='en'><head><meta charset='UTF-8'><link rel='icon' href='data:,'></head><body>Status Finished Running</body></html>"))
	println("status handler ran")
}

func (app *application) signout(w http.ResponseWriter, r *http.Request) {
	println("Cookie Destroyed!")
	w.Write([]byte("<!DOCTYPE html><html lang='en'><head><meta charset='UTF-8'><link rel='icon' href='data:,'></head><body>Signed out!</body></html>"))
}

func (app *application) signin(w http.ResponseWriter, r *http.Request) {
	println("Cookie Destroyed!")
	w.Write([]byte("<!DOCTYPE html><html lang='en'><head><meta charset='UTF-8'><link rel='icon' href='data:,'></head><body>SIGN IN PAGE!</body></html>"))
}

func main() {

	config := environment.NewConfig()
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	app := &application{
		mux:    *http.NewServeMux(),
		config: *config,
		logger: logger,
	}

	serverTLSKeys, err := tls.LoadX509KeyPair(config.TLSPublicKey, config.TLSPrivateKey)
	if err != nil {
		app.logger.Fatalf("Error loading TLS public/private keys: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverTLSKeys},
	}

	mux := secureHeaders(disableCache(requestLogger(app.router())))

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.Port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		TLSConfig:    tlsConfig,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	defer srv.Close()

	log.Printf("Serving on port %s 💁🏻", config.Port)
	log.Fatal(srv.ListenAndServeTLS("", ""))
}
