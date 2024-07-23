package appcfg

import (
	"log"
	"os"
	"strconv"
)

type configProperties struct {
	AppName       string
	DevMode       bool
	Port          string
	TLSPublicKey  string
	TLSPrivateKey string
	SessionName   string
	SessionSecret string
	SessionSecure bool
	SessionMaxAge string
	MongoDb       string
	MongoUsername string
	MongoPassword string
}

var (
	Props *configProperties
)

func NewConfig() *configProperties {
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

	properties := configProperties{
		AppName:       getEnv("APPNAME", "NoAppName"),
		DevMode:       devMode,
		Port:          getEnv("PORT", "8443"),
		TLSPublicKey:  getEnv("TLS_PUBLICKEY", ""),
		TLSPrivateKey: getEnv("TLS_PRIVATEKEY", ""),
		SessionName:   getEnv("SESSION_NAME", "DummySessionName"),
		SessionSecret: getEnv("SESSION_SECRET", ""),
		SessionSecure: sessionSecure,
		SessionMaxAge: getEnv("SESSION_MAXAGE", ""),
		MongoDb:       getEnv("MONGO_DB", "dev_db"),
		MongoUsername: getEnv("MONGO_USERNAME", ""),
		MongoPassword: getEnv("MONGO_PASSWORD", ""),
	}

	return &properties
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
