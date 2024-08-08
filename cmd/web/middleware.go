package main

import (
	"log"
	"net/http"
	"os"
)

func disableCache(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private, no-cache, no-store, must-revalidate, proxy-revalidate")
		w.Header().Set("Expires", "0")
		w.Header().Set("Surrogate-Control", "max-age=0")
		// println("[ Disabling cache... ]")
		next.ServeHTTP(w, r)
	})
}

// Quick and dirty secure headers function
func secureHeaders(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-DNS-Prefetch-Control", "off")
		w.Header().Set("X-Download-Options", "noopen")
		// w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Strict-Transport-Security", "max-age=15552000; includeSubDomains")

		// println("[ Securing headers... ]")
		next.ServeHTTP(w, r)
	})
}

func requestLogger(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.SetOutput(os.Stdout)
		log.Println(r.Method, r.URL)
		// println("[ Log middleware ran.. ]")
		if r.URL.Path == "/foo" {
			return
		}

		next.ServeHTTP(w, r)
	})
}
