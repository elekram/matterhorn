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

	appcfg "github.com/elekram/matterhorn/config"
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

func Session(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		println("[ Session middleware running... ]")

		println("URL: " + r.RequestURI)
		cookie, err := r.Cookie(appcfg.Props.SessionName)
		if err != nil {

			if err == http.ErrNoCookie {
				println("Cookie not found!")
				_ = cookie

				// app.signin(w, r)
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

			destroyCookie(w, r)
			next.ServeHTTP(w, r)
			return
		}

		fmt.Printf("Found valid cookie\n")
		next.ServeHTTP(w, r)
	})
}

func destroyCookie(w http.ResponseWriter, r *http.Request) {
	println("destroiying cookie")
	sessionName := appcfg.Props.SessionName

	cookie := http.Cookie{
		Name:     sessionName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   appcfg.Props.SessionSecure,
	}

	http.SetCookie(w, &cookie)
}

func setCookie(w http.ResponseWriter, r *http.Request) {
	sessions[generateSessionId(30)] = session{
		username: "lee@cheltsec.vic.edu.au",
		expiry:   time.Now(),
	}

	println("Setting cookie")
	fmt.Printf(appcfg.Props.SessionName + "\n")

	sessionName := appcfg.Props.SessionName
	maxAge, err := strconv.Atoi(appcfg.Props.SessionMaxAge)

	if err != nil {
		println("session cookie: maxage not a number")
	}

	cookie := http.Cookie{
		Name:     sessionName,
		Value:    cookieUserValue,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   appcfg.Props.SessionSecure,
		SameSite: http.SameSiteLaxMode,
	}

	signedCookie := signCookie(cookie.Name, cookie.Value, appcfg.Props.SessionSecret)

	encodedCookieValue := base64.URLEncoding.EncodeToString([]byte(signedCookie))
	cookie.Value = encodedCookieValue

	http.SetCookie(w, &cookie)
}

func getCookie(w http.ResponseWriter, r *http.Request) {
	println("this happened!")
	cookie, err := r.Cookie(appcfg.Props.SessionName)
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

	mac := hmac.New(sha256.New, []byte(appcfg.Props.SessionSecret))
	mac.Write([]byte(appcfg.Props.SessionName))
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
