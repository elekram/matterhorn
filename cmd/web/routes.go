package main

import (
	"net/http"
)

func (app *application) router() http.Handler {

	router := http.NewServeMux()
	router.HandleFunc("GET /v1/status", app.use(app.status, app.middleware...))
	router.HandleFunc("GET /", app.use(app.status, app.middleware...))

	return router
}
