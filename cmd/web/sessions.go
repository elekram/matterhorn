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
	"sync"
	"time"
)

var (
	ErrValueTooLong       = errors.New("cookie value too long")
	ErrInvalidValue       = errors.New("invalid cookie value")
	ErrInvalidCookieValue = errors.New("cookie failed intregity check")
)

type sessionMgr struct {
	sessionName    string
	sessiongSecret string
	maxAge         int
	useMemoryStore bool
	secureSession  bool
	context        session
	memoryStore    *memoryStore
}

type memoryStore struct {
	store map[string]session
	mu    sync.Mutex
}

type session struct {
	username string
	expiry   time.Time
}

func (s *sessionMgr) manageSession(app *app) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// pass through static content
		if strings.Contains(r.RequestURI, strings.ToLower("/static")) {
			app.router.ServeHTTP(w, r)
			return
		}

		if s.cookieExists(r) {
			cookieVal, err := s.getCookie(r)
			if err != nil {
				fmt.Println(err)
				return
			}

			context, keyPresent := s.memoryStore.store[cookieVal]
			if !keyPresent {
				s.destroyCookie(w)
				app.handlers.handleSignIn.ServeHTTP(w, r)
				return
			}

			s.context = context

			if time.Now().Before(s.context.expiry) {
				fmt.Println("session active")
				app.router.ServeHTTP(w, r)
				return
			}

			if time.Now().After(s.context.expiry) {
				fmt.Println("session expired")
				s.destroyCookie(w)
				app.handlers.handleSignIn.ServeHTTP(w, r)
				return
			}
		}

		if !s.cookieExists(r) &&
			r.Method == "POST" &&
			strings.ToLower(r.RequestURI) == "/auth" {

			email := strings.ToLower(r.FormValue("email"))
			password := r.FormValue("password")

			if email == "mark@mark.com" && password == "password" {
				var newSessionId = generateSessionId(30)

				_, ifKeyIsPresent := s.memoryStore.store[newSessionId]
				for ifKeyIsPresent {
					newSessionId = generateSessionId(30)
				}

				s.memoryStore.mu.Lock()
				s.memoryStore.store[newSessionId] = session{
					username: email,
					expiry:   time.Now().Add(time.Second * time.Duration(s.maxAge)),
				}
				s.memoryStore.mu.Unlock()

				s.setCookie(w, r, newSessionId)
				app.handlers.handleHomePage.ServeHTTP(w, r)
				return
			}
		}

		if !s.cookieExists(r) {
			app.handlers.handleSignIn.ServeHTTP(w, r)
			return
		}
	})
}

func newSession(sessionName, sessionSecret, maxAge string, secureSession bool) *sessionMgr {
	ms := memoryStore{
		store: map[string]session{},
		mu:    sync.Mutex{},
	}

	ma, err := strconv.Atoi(maxAge)
	if err != nil {
		fmt.Println("New Session Error: maxAge not a number")
		panic(err)
	}

	s := sessionMgr{
		sessionName:    sessionName,
		sessiongSecret: sessionSecret,
		maxAge:         ma,
		secureSession:  secureSession,
		useMemoryStore: true,
		memoryStore:    &ms,
		context: session{
			username: "",
			expiry:   time.Time{},
		},
	}

	return &s
}

func (s *sessionMgr) cookieExists(r *http.Request) bool {
	_, err := r.Cookie(s.sessionName)
	if err != nil {
		if err == http.ErrNoCookie {
			return false
		}
	}
	return true
}

func (s *sessionMgr) setCookie(w http.ResponseWriter, r *http.Request, sessionId string) {

	if s.useMemoryStore {
		signedValues := signCookie(s.sessionName, sessionId, s.sessiongSecret)
		encodedCookieValue := base64.URLEncoding.EncodeToString([]byte(signedValues))

		cookie := http.Cookie{
			Name:     s.sessionName,
			Value:    encodedCookieValue,
			Path:     "/",
			MaxAge:   s.maxAge,
			HttpOnly: true,
			Secure:   s.secureSession,
			SameSite: http.SameSiteLaxMode,
		}

		http.SetCookie(w, &cookie)
		return
	}
}

func (s *sessionMgr) getCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(s.sessionName)
	if err != nil {
		println(err)
	}

	decodedSignedCookie, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		fmt.Println("Error cookie decode")
		fmt.Println(err)
	}

	authenticCookieValue, err := getAuthenticatedCookieVal(
		s.sessionName,
		s.sessiongSecret,
		decodedSignedCookie)
	if err != nil {
		fmt.Println("cookie failed integrity check!")
		return "", err
	}

	return string(authenticCookieValue), nil
}

func signCookie(cookieName, cookieVal, sessionSecret string) string {
	hmac := hmac.New(sha256.New, []byte(sessionSecret))
	hmac.Write([]byte(cookieName))
	hmac.Write([]byte(cookieVal))

	signature := hmac.Sum(nil)
	signedCookie := string(signature) + cookieVal

	return signedCookie
}

func getAuthenticatedCookieVal(cookieName, sessionSecret string, signedCookieValue []byte) ([]byte, error) {
	if len(signedCookieValue) < sha256.Size {
		return nil, ErrInvalidCookieValue
	}

	signature := signedCookieValue[:sha256.Size]
	value := signedCookieValue[sha256.Size:]

	mac := hmac.New(sha256.New, []byte(sessionSecret))

	mac.Write([]byte(cookieName))
	mac.Write([]byte(value))
	expectedSignature := mac.Sum(nil)

	if !hmac.Equal([]byte(signature), expectedSignature) {
		return nil, ErrInvalidCookieValue
	}

	return value, nil
}

func (s *sessionMgr) destroyCookie(w http.ResponseWriter) {
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
