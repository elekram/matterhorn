package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var (
	ErrValueTooLong       = errors.New("cookie value too long")
	ErrInvalidValue       = errors.New("invalid cookie value")
	ErrInvalidCookieValue = errors.New("cookie failed intregity check")
)

type sessionMgr struct {
	authenticated  bool
	sessionName    string
	sessiongSecret string
	maxAge         int
	useMemoryStore bool
	secureSession  bool
	memoryStore    map[string]session
}

type session struct {
	username string
	expiry   time.Time
}

func (s *sessionMgr) manageSession(app *server) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("uri=" + r.RequestURI)

		if strings.Contains(r.RequestURI, strings.ToLower("/static")) {
			app.router.ServeHTTP(w, r)
		}

		if !s.authenticated && r.Method == "POST" && r.RequestURI == strings.ToLower("/auth") {
			app.handlers.handleSignIn.ServeHTTP(w, r)
			return
		}

		fmt.Println("WOOOOOOOOOO!")

		if !s.authenticated {
			// app.SignIn(w, r)
			return
		}

		if !s.cookieExists(r) {
			s.setCookie(w, r)
		}
	})
}

func newSession(sessionName, sessionSecret string, maxAge int, secureSession bool) *sessionMgr {
	s := sessionMgr{
		authenticated:  false,
		sessionName:    sessionName,
		sessiongSecret: sessionSecret,
		maxAge:         maxAge,
		secureSession:  secureSession,
		useMemoryStore: true,
		memoryStore:    map[string]session{},
	}

	return &s
}

func (s *sessionMgr) setCookie(w http.ResponseWriter, r *http.Request) {

	if s.useMemoryStore {
		fmt.Println("Browser did not send cookie")
		fmt.Println("Create new SessionId and it to the store")
		newSessionId := generateSessionId(30)

		s.memoryStore[newSessionId] = session{
			username: "lee@cheltsec.vic.edu.au",
			expiry:   time.Now().Add(time.Minute),
		}

		cookie := http.Cookie{
			Name:     s.sessionName,
			Value:    "",
			Path:     "/",
			MaxAge:   s.maxAge,
			HttpOnly: true,
			Secure:   s.secureSession,
			SameSite: http.SameSiteLaxMode,
		}

		signedCookie := signCookie(cookie.Name, cookie.Value, s.sessiongSecret)

		encodedCookieValue := base64.URLEncoding.EncodeToString([]byte(signedCookie))
		cookie.Value = encodedCookieValue

		http.SetCookie(w, &cookie)

		return
	}

}

func (s *sessionMgr) cookieExists(r *http.Request) bool {
	fmt.Println(s.sessionName)

	_, err := r.Cookie(s.sessionName)
	if err != nil {
		if err == http.ErrNoCookie {
			fmt.Println("no cookie")
			return false
		}
	}
	fmt.Println("cookie!!!!")
	return true
}

// func (s *sessionMgr) getCookie(w http.ResponseWriter, r *http.Request) {
// 	println("Get cookie")
// 	cookie, err := r.Cookie(s.sessionName)
// 	if err != nil {
// 		println(err)
// 	}

// 	w.Write([]byte(cookie.Value + " no cookie"))
// 	return

// 	signedCookieValue, err := base64.URLEncoding.DecodeString(cookie.Value)
// 	if err != nil {
// 		println(err)
// 	}

// 	if len(signedCookieValue) < sha256.Size {
// 		println(ErrInvalidValue)
// 	}

// 	signature := signedCookieValue[:sha256.Size]
// 	value := signedCookieValue[sha256.Size:]

// 	mac := hmac.New(sha256.New, []byte(appcfg.Props.SessionSecret))
// 	mac.Write([]byte(appcfg.Props.SessionName))
// 	mac.Write([]byte(value))
// 	expectedSignature := mac.Sum(nil)

// 	if !hmac.Equal([]byte(signature), expectedSignature) {
// 		println("signature mismatch")
// 	} else {
// 		println("signatures match!!")
// 	}

// 	w.Write([]byte(cookie.Value + " sheeeeit!"))
// }

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

func (s *sessionMgr) destroyCookie(w http.ResponseWriter, r *http.Request) {
	println("destroiying cookie")
	sessionName := s.sessionName

	cookie := http.Cookie{
		Name:     sessionName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   s.secureSession,
	}

	http.SetCookie(w, &cookie)
}

func generateSessionId(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length+2)
	r.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}
