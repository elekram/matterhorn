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
	ErrRetrievingProfile  = errors.New("error retrieving profile")
)

type store interface {
	getSession(id string) (session, bool)
	addSession(id string, p session)
}

type sessionMgr struct {
	sessionName             string
	sessiongSecret          string
	maxAge                  int
	secureSession           bool
	oAuthStateParemeterPool map[string]oAuthStateParemeters
	store                   store
}

type MemoryStore struct {
	store map[string]session
	mu    sync.Mutex
}

func (ms *MemoryStore) getSession(id string) (session, bool) {
	s, keyPresent := ms.store[id]

	if !keyPresent {
		return session{}, false
	}

	return s, true
}

func (ms *MemoryStore) addSession(id string, s session) {
	ms.mu.Lock()
	ms.store[id] = s
	ms.mu.Unlock()
}

type session struct {
	profile googleProfile
	expiry  time.Time
}

type oAuthStateParemeters struct {
	redirectUrl string
	expiresOn   time.Time
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

func newMemoryStore() store {
	ms := MemoryStore{
		store: map[string]session{},
		mu:    sync.Mutex{},
	}
	return &ms
}

func (s *sessionMgr) manageSession(app *app) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" &&
			strings.HasPrefix(r.RequestURI, "/static") {
			addCORSHeaders(w)
		}

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

			context, hasSession := s.store.getSession(cookieValue)
			if !hasSession {
				s.destroyCookie(w)
				app.handlers.handleSignIn.ServeHTTP(w, r)
				return
			}

			if time.Now().After(context.expiry) {
				s.destroyCookie(w)
				app.handlers.handleSignIn.ServeHTTP(w, r)
				return
			}

			if strings.HasPrefix(r.RequestURI, "/oauth2/redirect/google") {
				postSignInUrl := "/"
				http.Redirect(w, r, postSignInUrl, http.StatusTemporaryRedirect)
				return
			}

			app.router.ServeHTTP(w, r)
			return
		}

		if !s.cookieExists(r) &&
			strings.HasPrefix(r.RequestURI, "/oauth2/redirect/google") {
			oAuthStateId := r.FormValue("state")

			_, keyPresent := s.oAuthStateParemeterPool[oAuthStateId]
			if !keyPresent {
				http.Error(w, ErrBadOAuthRequest.Error(), http.StatusBadRequest)
				return
			}

			delete(s.oAuthStateParemeterPool, oAuthStateId)
			defer s.cleanUpExpiredOAuthParameterIds()

			gp, err := getGoogleProfile(r, app.oAuth2Config)
			if err != nil {
				fmt.Println(err)
				http.Error(w, ErrRetrievingProfile.Error(), http.StatusServiceUnavailable)
				return
			}

			newSessionId := s.generateId(30)

			sess := session{
				profile: gp,
				expiry:  time.Now().Add(time.Second * time.Duration(s.maxAge)),
			}

			s.store.addSession(newSessionId, sess)
			s.setCookie(w, newSessionId)

			postSignInUrl := "/"
			http.Redirect(w, r, postSignInUrl, http.StatusTemporaryRedirect)
			return
		}

		if !s.cookieExists(r) {
			app.handlers.handleSignIn.ServeHTTP(w, r)
			return
		}
	})
}

func newSession(
	sessionName,
	sessionSecret,
	maxAge string,
	secureSession bool,
	store store) *sessionMgr {

	ma, err := strconv.Atoi(maxAge)
	if err != nil {
		fmt.Println("New Session Error: maxAge not a number")
		panic(err)
	}

	_ = newMemoryStore()

	s := sessionMgr{
		sessionName:             sessionName,
		sessiongSecret:          sessionSecret,
		maxAge:                  ma,
		secureSession:           secureSession,
		oAuthStateParemeterPool: map[string]oAuthStateParemeters{},
		store:                   store,
	}

	return &s
}

func getGoogleProfile(
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

	var googleResponse map[string]interface{}

	err = json.NewDecoder(resp.Body).Decode(&googleResponse)
	if err != nil {
		return g, err
	}

	g = googleProfile{
		email:       googleResponse["email"].(string),
		name:        googleResponse["name"].(string),
		family_name: googleResponse["family_name"].(string),
		given_name:  googleResponse["given_name"].(string),
		hd:          googleResponse["hd"].(string),
		id:          googleResponse["id"].(string),
		picture:     googleResponse["picture"].(string),
	}

	return g, err
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

func (s *sessionMgr) cleanUpExpiredOAuthParameterIds() {
	expiredIds := []string{}
	for key, parameter := range s.oAuthStateParemeterPool {
		if time.Now().After(parameter.expiresOn) {
			expiredIds = append(expiredIds, key)
		}
	}

	for _, id := range expiredIds {
		delete(s.oAuthStateParemeterPool, id)
	}
}

func signCookie(cookieName, cookieVal, sessionSecret string) string {
	hmac := hmac.New(sha256.New, []byte(sessionSecret))
	hmac.Write([]byte(cookieName))
	hmac.Write([]byte(cookieVal))

	signature := hmac.Sum(nil)
	signedCookie := string(signature) + cookieVal

	return signedCookie
}

func getAuthenticatedCookieVal(
	cookieName,
	sessionSecret string,
	signedCookieValue []byte) ([]byte, error) {

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

func addCORSHeaders(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Vary", "Origin")
	w.Header().Add("Vary", "Access-Control-Request-Method")
	w.Header().Add("Vary", "Access-Control-Request-Headers")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Accept, token")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
}
