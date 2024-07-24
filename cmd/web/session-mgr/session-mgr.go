package sessionmgr

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	appcfg "github.com/elekram/matterhorn/config"
)

var (
	SessionMgr            *sessionMgr
	ErrValueTooLong       = errors.New("cookie value too long")
	ErrInvalidValue       = errors.New("invalid cookie value")
	ErrInvalidCookieValue = errors.New("cookie failed intregity check")
)

type sessionMgr struct {
	sessionName    string
	maxAge         int
	useMemoryStore bool
	secureSession  bool
	sessions       map[string]session
}

type session struct {
	username string
	expiry   time.Time
}

var memoryStore = map[string]session{}

func (SessionMgr sessionMgr) ManageSession(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !cookieExists(r) {
			setCookie(w, r)
		}

	})
}

func NewSession(sessionName string, maxAge int, secureSession bool) *sessionMgr {
	sessionMgr := sessionMgr{
		sessionName:    sessionName,
		maxAge:         maxAge,
		useMemoryStore: false,
		secureSession:  secureSession,
		sessions:       map[string]session{},
	}

	return &sessionMgr
}

func setCookie(w http.ResponseWriter, r *http.Request) {
	sessionName := SessionMgr.sessionName

	if SessionMgr.useMemoryStore {
		fmt.Println("Browser did not send cookie")
		fmt.Println("Create new SessionId and it to the store")
		newSessionId := generateSessionId(30)

		memoryStore[newSessionId] = session{
			username: "lee@cheltsec.vic.edu.au",
			expiry:   time.Now().Add(time.Minute),
		}

		cookie := http.Cookie{
			Name:     sessionName,
			Value:    "",
			Path:     "/",
			MaxAge:   SessionMgr.maxAge,
			HttpOnly: true,
			Secure:   SessionMgr.secureSession,
			SameSite: http.SameSiteLaxMode,
		}

		signedCookie := signCookie(cookie.Name, cookie.Value, appcfg.Props.SessionSecret)

		encodedCookieValue := base64.URLEncoding.EncodeToString([]byte(signedCookie))
		cookie.Value = encodedCookieValue

		http.SetCookie(w, &cookie)

		return
	}

}

func cookieExists(r *http.Request) bool {
	println("Get cookie")
	_, err := r.Cookie(SessionMgr.sessionName)
	if err != nil {
		if err == http.ErrNoCookie {
			return false
		}
	}
	return true
}

func getCookie(w http.ResponseWriter, r *http.Request) {
	println("Get cookie")
	cookie, err := r.Cookie(SessionMgr.sessionName)
	if err != nil {
		println(err)
	}

	w.Write([]byte(cookie.Value + " no cookie"))
	return

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

func generateSessionId(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length+2)
	r.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}
