package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	cookieUserValue       string = "someuser@example.com"
	ErrValueTooLong              = errors.New("cookie value too long")
	ErrInvalidValue              = errors.New("invalid cookie value")
	ErrInvalidCookieValue        = errors.New("cookie failed intregity check")
)

type session struct {
	username string
	expiry   time.Time
}

var sessions = map[string]session{}

// func (s session) isExpired() bool {
// 	return s.expiry.Before(time.Now())
// }

func (app *application) session(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		println("[ Session middleware running... ]")

		println("URL: " + r.RequestURI)
		cookie, err := r.Cookie(app.config.SessionName)
		if err != nil {

			if err == http.ErrNoCookie {
				println("Cookie not found!")
				_ = cookie

				app.signin(w, r)
				// setCookie(w, r, app)
				// next.ServeHTTP(w, r)
				return
			}

			if err != http.ErrNoCookie {
				// something went wrong page goes here
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		if strings.ToLower(r.RequestURI) == "/signout" {
			fmt.Printf("%s", cookie.Expires)
			if cookie.Expires.Before(time.Now()) {
				println("Cookie expired nom non nom!!!")
			}

			destroyCookie(w, r, app)
			next.ServeHTTP(w, r)
			return
		}

		fmt.Printf("Found valid cookie\n")
		next.ServeHTTP(w, r)
	})
}

func destroyCookie(w http.ResponseWriter, r *http.Request, app *application) {
	println("destroiying cookie")
	sessionName := app.config.SessionName

	cookie := http.Cookie{
		Name:     sessionName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   app.config.SessionSecure,
	}

	http.SetCookie(w, &cookie)
}

func setCookie(w http.ResponseWriter, r *http.Request, app *application) {
	sessions[generateSessionId(30)] = session{
		username: "lee@cheltsec.vic.edu.au",
		expiry:   time.Now(),
	}

	println("Setting cookie")
	fmt.Printf(app.config.SessionName + "\n")

	sessionName := app.config.SessionName
	maxAge, err := strconv.Atoi(app.config.SessionMaxAge)

	if err != nil {
		println("session cookie: maxage not a number")
	}

	cookie := http.Cookie{
		Name:     sessionName,
		Value:    cookieUserValue,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   app.config.SessionSecure,
		SameSite: http.SameSiteLaxMode,
	}

	signedCookie := signCookie(cookie.Name, cookie.Value, app.config.SessionSecret)

	encodedCookieValue := base64.URLEncoding.EncodeToString([]byte(signedCookie))
	cookie.Value = encodedCookieValue

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

func signCookie(cookieName, cookieVal, sessionSecret string) string {
	hmac := hmac.New(sha256.New, []byte(sessionSecret))
	hmac.Write([]byte(cookieName))
	hmac.Write([]byte(cookieVal))

	signature := hmac.Sum(nil)
	signedCookie := string(signature) + cookieVal

	return signedCookie
}

func authenticateCookie(cookieName, cookieVal, sessionSecret string) (string, error) {
	signedCookieValue := cookieVal

	if len(signedCookieValue) < sha256.Size {
		return "", ErrInvalidCookieValue
	}

	signature := signedCookieValue[:sha256.Size]
	value := signedCookieValue[sha256.Size:]

	mac := hmac.New(sha256.New, []byte(sessionSecret))

	mac.Write([]byte(cookieName))
	mac.Write([]byte(value))
	expectedSignature := mac.Sum(nil)

	if !hmac.Equal([]byte(signature), expectedSignature) {
		return "", ErrInvalidCookieValue
	}

	return value, nil
}

func generateSessionId(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length+2)
	r.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}
