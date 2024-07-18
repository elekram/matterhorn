package env

import (
	"log"
	"os"
	"strconv"
)

type config struct {
	AppName       string
	DevMode       bool
	Port          string
	TLSPublicKey  string
	TLSPrivateKey string
	SessionName   string
	SessionSecret string
	SessionSecure bool
	SessionMaxAge string
}

var (
	Config *config
)

func NewConfig() *config {
	dm := getEnv("DEV_MODE", "true")
	devMode, err := strconv.ParseBool(dm)
	if err != nil {
		log.Fatal("Error: parsebool failed")
	}

	ss := getEnv("SESSION_SECURE", "")
	sessionSecure, err := strconv.ParseBool(ss)
	if err != nil {
		log.Fatal("Error: parsebool failed")
	}

	config := config{
		AppName:       getEnv("APPNAME", "NoName"),
		DevMode:       devMode,
		Port:          getEnv("PORT", "8443"),
		TLSPublicKey:  getEnv("TLS_PUBLICKEY", ""),
		TLSPrivateKey: getEnv("TLS_PRIVATEKEY", ""),
		SessionName:   getEnv("SESSION_NAME", "DummySessionName"),
		SessionSecret: getEnv("SESSION_SECRET", ""),
		SessionSecure: sessionSecure,
		SessionMaxAge: getEnv("SESSION_MAXAGE", ""),
	}

	return &config
}

func getEnv(key, defaultValue string) string {
	value, ok := os.LookupEnv(key)

	if !ok {
		log.Printf("Key: %s", key)
		log.Fatal("missing key")
	}

	if len(value) > 0 {
		return value
	}

	return defaultValue
}
