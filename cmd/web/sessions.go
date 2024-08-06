package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

var (
	ErrValueTooLong       = errors.New("cookie value too long")
	ErrInvalidValue       = errors.New("invalid cookie value")
	ErrInvalidCookieValue = errors.New("cookie failed intregity check")
	ErrBadOAuthRequest    = errors.New("state id check failed")
)

type oAuthIdPool struct {
	expiry time.Time
}

type sessionMgr struct {
	sessionName    string
	sessiongSecret string
	maxAge         int
	useMemoryStore bool
	secureSession  bool
	oAuthIdPool    map[string]oAuthIdPool
	context        session
	memoryStore    *memoryStore
}

type memoryStore struct {
	store map[string]session
	mu    sync.Mutex
}

type session struct {
	profile googleProfile
	expiry  time.Time
}

type googleProfile struct {
	email       string
	name        string
	family_name string
	given_name  string
	hd          string
	id          string
	picture     string
}

func (s *sessionMgr) manageSession(app *app) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.RequestURI, "/static") ||
			strings.HasPrefix(r.RequestURI, "/auth/oauth") {
			app.router.ServeHTTP(w, r)
			return
		}

		if s.cookieExists(r) {
			cookieValue, err := s.getCookie(r)
			if err != nil {
				fmt.Println(err)
				return
			}

			context, keyPresent := s.memoryStore.store[cookieValue]
			if !keyPresent {
				s.destroyCookie(w)
				app.handlers.handleSignIn.ServeHTTP(w, r)
				return
			}

			s.context = context

			if time.Now().After(s.context.expiry) {
				fmt.Println("session expired")
				s.destroyCookie(w)
				app.handlers.handleSignIn.ServeHTTP(w, r)
				return
			}

			fmt.Println("session active")
			app.router.ServeHTTP(w, r)
			return
		}

		if !s.cookieExists(r) &&
			r.Method == "GET" &&
			strings.HasPrefix(r.RequestURI, "/oauth2/redirect/google") {
			stateId := r.FormValue("state")

			fmt.Println("2: " + stateId)

			_, keyPresent := s.oAuthIdPool[stateId]
			if !keyPresent {
				http.Error(w, ErrBadOAuthRequest.Error(), http.StatusBadRequest)
				return
			}

			defer s.clearIdPool()

			googleContext, err := HandleOAuthCallback(w, r, app.oAuth2Config)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			newSessionId := s.generateId(30)

			_, whileStoreHasKey := s.memoryStore.store[newSessionId]
			for whileStoreHasKey {
				newSessionId = s.generateId(30)
			}

			s.memoryStore.mu.Lock()
			s.memoryStore.store[newSessionId] = session{
				profile: googleContext,
				expiry:  time.Now().Add(time.Second * time.Duration(s.maxAge)),
			}
			s.memoryStore.mu.Unlock()
			fmt.Println(s.memoryStore.store)

			s.setCookie(w, newSessionId)
			app.handlers.handleHomePage.ServeHTTP(w, r)
			return
		}

		if !s.cookieExists(r) {
			app.handlers.handleSignIn.ServeHTTP(w, r)
			return
		}
	})
}

func HandleOAuthCallback(
	w http.ResponseWriter,
	r *http.Request,
	OAuthCfg *oauth2.Config) (googleProfile, error) {
	g := googleProfile{}
	code := r.URL.Query().Get("code")

	token, err := OAuthCfg.Exchange(context.Background(), code)
	if err != nil {
		return g, err
	}

	client := OAuthCfg.Client(context.Background(), token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return g, err
	}

	defer resp.Body.Close()

	var response map[string]interface{}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return g, err
	}

	g = googleProfile{
		email:       response["email"].(string),
		name:        response["name"].(string),
		family_name: response["family_name"].(string),
		given_name:  response["given_name"].(string),
		hd:          response["hd"].(string),
		id:          response["id"].(string),
		picture:     response["picture"].(string),
	}

	return g, err
}

func (s *sessionMgr) clearIdPool() {
	expiredIds := []string{}
	for key, val := range s.oAuthIdPool {
		if time.Now().After(val.expiry) {
			expiredIds = append(expiredIds, key)
		}
	}

	for _, val := range expiredIds {
		delete(s.oAuthIdPool, val)
	}
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
		oAuthIdPool:    map[string]oAuthIdPool{},
		useMemoryStore: true,
		memoryStore:    &ms,
		context: session{
			profile: googleProfile{},
			expiry:  time.Time{},
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

func (s *sessionMgr) setCookie(w http.ResponseWriter, sessionId string) {

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

func (s *sessionMgr) generateId(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length+2)
	r.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}
