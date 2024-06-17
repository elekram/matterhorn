package session

import (
	"log"
	"net/http"
	"os"
)

type Session struct {
	GetCookie func()
	SetCookie func()
}

func NewSession(h http.HandlerFunc) http.HandlerFunc {

	s := Session{}
	s.SetCookie = func() {
		println("setCookie")
	}

	s.GetCookie = func() {
		println("getCookie")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		log.SetOutput(os.Stdout) // logs go to Stderr by default
		log.Println(r.Method, r.URL)
		h.ServeHTTP(w, r) // call ServeHTTP on the original handler

	})
}

// func SetCookie() http.HandlerFunc {

// 		// Initialize a new cookie containing the string "Hello world!" and some
// 		// non-default attributes.
// 		cookie := http.Cookie{
// 			Name:     "exampleCookie",
// 			Value:    "Hello world!",
// 			Path:     "/",
// 			MaxAge:   3600,
// 			HttpOnly: true,
// 			Secure:   true,
// 			SameSite: http.SameSiteLaxMode,
// 		}

// 		// Use the http.SetCookie() function to send the cookie to the client.
// 		// Behind the scenes this adds a `Set-Cookie` header to the response
// 		// containing the necessary cookie data.
// 		http.SetCookie(w, &cookie)

// 		// Write a HTTP response as normal.
// 		w.Write([]byte("cookie set!"))
// }
