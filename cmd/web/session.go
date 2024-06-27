package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
)

var (
	ErrValueTooLong = errors.New("cookie value too long")
	ErrInvalidValue = errors.New("invalid cookie value")
)

func (app *application) session(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		setCookie(w, r, *app, func() {
			getCookie(w, r, *app)
		})

		log.SetOutput(os.Stdout)
		log.Println(r.Method, r.URL)
		println("session middleware ran")
		next.ServeHTTP(w, r)
	})
}

func writeCookie(w http.ResponseWriter, cookie http.Cookie) error {
	// Encode the cookie value using base64.
	cookie.Value = base64.URLEncoding.EncodeToString([]byte(cookie.Value))

	// Check the total length of the cookie contents. Return the ErrValueTooLong
	// error if it's more than 4096 bytes.
	if len(cookie.String()) > 4096 {
		return ErrValueTooLong
	}

	// Write the cookie as normal.
	http.SetCookie(w, &cookie)

	return nil
}

func setCookie(
	w http.ResponseWriter,
	r *http.Request,
	app application,
	_f func()) error {

	cookie := http.Cookie{
		Name:     app.config.SessionName,
		Value:    "SomeValueasdasd",
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   app.config.SessionSecure,
		SameSite: http.SameSiteLaxMode,
	}

	hmac := hmac.New(sha256.New, []byte(app.config.SessionSecret))

	//hmac.Write([]byte(cookie.Name))
	hmac.Write([]byte(cookie.Value))
	signature := hmac.Sum(nil)

	//cookie.Value = base64.URLEncoding.EncodeToString([]byte(app.config.SessionSecret))
	fmt.Printf("%x", signature)
	http.SetCookie(w, &cookie)
	w.Write([]byte(cookie.Value))
	return nil
}

func getCookie(w http.ResponseWriter, r *http.Request, app application) {
	cookie, err := r.Cookie(app.config.SessionName)
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			http.Error(w, "cookie not found", http.StatusBadRequest)
		default:
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}
		return
	}

	w.Write([]byte(cookie.Value))
}
