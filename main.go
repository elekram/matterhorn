package main

import (
	"log"
	"net/http"

	environment "github.com/elekram/matterhorn/config"
)

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<b>Wassssup Neegs?<b>"))
}

func main() {
	config := environment.NewConfig()

	mux := http.NewServeMux()
	mux.HandleFunc("/", home)
	mux.HandleFunc("GET /", home)

	log.Printf("Starting server on port %s", config.Port)
	err := http.ListenAndServe(":"+config.Port, mux)
	log.Fatal(err)
}
