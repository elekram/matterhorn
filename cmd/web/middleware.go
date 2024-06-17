package main

import (
	"log"
	"net/http"
	"os"
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

func secureHeaders(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")
		println("secure headers ran")
		next.ServeHTTP(w, r)
	})
}

func requestLogger(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.SetOutput(os.Stdout) // logs go to Stderr by default
		log.Println(r.Method, r.URL)
		println("logmiddleware ran")
		next.ServeHTTP(w, r) // call ServeHTTP on the original handler
	})
}
