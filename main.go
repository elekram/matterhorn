package main

import (
	"log"
	"net/http"

	environment "github.com/elekram/matterhorn/config"
)

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<b>Wassssup my dude?<b>"))
}

func main() {
	config := environment.NewConfig()

	mux := http.NewServeMux()
	mux.HandleFunc("/", home)
	mux.HandleFunc("GET /", home)

	log.Printf("Starting server on port %s", config.Port)
	err := http.ListenAndServeTLS(":"+config.Port,
		"./cert/8b9f9ffcb19fa503.crt",
		"./cert/_.cheltsec.vic.edu.au.key",
		mux,
	)
	log.Fatal(err)
}
