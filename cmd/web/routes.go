package main

import (
	"net/http"
)

func (app *application) router() *http.ServeMux {

	router := http.NewServeMux()
	router.HandleFunc("GET /v1/status", app.status)
	router.HandleFunc("/", app.use(app.status, app.middleware...))

	return router
}
