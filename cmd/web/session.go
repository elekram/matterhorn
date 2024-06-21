package main

import (
	"log"
	"net/http"
	"os"
)

type sessionOptions struct {
	name   string
	secret string
	secure bool
	maxAge int
}

func (app *application) session(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		setCookieHandler(w, r, *app)

		log.SetOutput(os.Stdout)
		log.Println(r.Method, r.URL)
		println("session middleware ran")
		next.ServeHTTP(w, r)
	})
}

func setCookieHandler(w http.ResponseWriter, r *http.Request, app application) {
	cookie := http.Cookie{
		Name:     app.config.SessionName,
		Value:    app.config.SessionSecret,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   app.config.SessionSecure,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)
	w.Write([]byte("cookie set!"))
}
