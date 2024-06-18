package main

import (
	"log"
	"net/http"
	"os"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

func (app *application) use(h http.HandlerFunc, m ...Middleware) http.HandlerFunc {
	if len(m) < 1 {
		return h
	}

	wrappedHandler := h

	// ensures middleware runs in order as per ...Middleware slice
	for i := len(m) - 1; i >= 0; i-- {
		wrappedHandler = m[i](wrappedHandler)
	}
	return wrappedHandler
}

func disableCache(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private, no-cache, no-store, must-revalidate, proxy-revalidate")
		w.Header().Set("Expires", "0")
		w.Header().Set("Surrogate-Control", "max-age=0")
		println("cache disabled")
		next.ServeHTTP(w, r)
	})
}

func secureHeaders(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-DNS-Prefetch-Control", "off")
		w.Header().Set("X-Download-Options", "noopen")
		w.Header().Set("Strict-Transport-Security", "max-age=15552000; includeSubDomains")

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
