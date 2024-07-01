package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var (
	ErrValueTooLong        = errors.New("cookie value too long")
	ErrInvalidValue        = errors.New("invalid cookie value")
	cookieUserValue string = "someuser@example.com"
)

var sessions = map[string]session{}

type session struct {
	username string
	expiry   time.Time
}

func (s session) isExpired() bool {
	return s.expiry.Before(time.Now())
}

func (app *application) session(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		println("[ Session middleware running... ]")

		cookie, err := r.Cookie(app.config.SessionName)
		if err != nil {
			if err != http.ErrNoCookie {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if err == http.ErrNoCookie {
				println("Cookie not found!")
				_ = cookie

				setCookie(w, r, *app)
				next.ServeHTTP(w, r)
				return
			}
		}
		fmt.Printf("Found Cookie\n")
		next.ServeHTTP(w, r)
	})
}

func setCookie(
	w http.ResponseWriter,
	r *http.Request,
	app application) {

	println("Setting cookie")
	fmt.Printf(app.config.SessionName + "\n")
	cookie := http.Cookie{
		Name:     app.config.SessionName,
		Value:    cookieUserValue,
		Path:     "/",
		MaxAge:   120,
		HttpOnly: true,
		Secure:   app.config.SessionSecure,
		SameSite: http.SameSiteLaxMode,
	}

	hmac := hmac.New(sha256.New, []byte(app.config.SessionSecret))
	hmac.Write([]byte(cookie.Name))
	hmac.Write([]byte(cookie.Value))

	signature := hmac.Sum(nil)

	sigAndCookieVal := string(signature) + cookie.Value
	cookie.Value = base64.URLEncoding.EncodeToString([]byte(sigAndCookieVal))

	http.SetCookie(w, &cookie)
}

func getCookie(w http.ResponseWriter, r *http.Request, app application) {
	println("this happened!")
	cookie, err := r.Cookie(app.config.SessionName)
	if err != nil {
		println(err)
	}

	signedCookieValue, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		println(err)
	}

	if len(signedCookieValue) < sha256.Size {
		println(ErrInvalidValue)
	}

	signature := signedCookieValue[:sha256.Size]
	value := signedCookieValue[sha256.Size:]

	mac := hmac.New(sha256.New, []byte(app.config.SessionSecret))
	mac.Write([]byte(app.config.SessionName))
	mac.Write([]byte(value))
	expectedSignature := mac.Sum(nil)

	if !hmac.Equal([]byte(signature), expectedSignature) {
		println("signature mismatch")
	} else {
		println("signatures match!!")
	}

	w.Write([]byte(cookie.Value + " sheeeeit!"))
}
