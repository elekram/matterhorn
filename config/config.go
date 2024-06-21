package environment

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	AppName       string
	Port          string
	TLSPublicKey  string
	TLSPrivateKey string
	SessionName   string
	SessionSecret string
	SessionSecure bool
	SessionMaxAge string
}

func NewConfig() *Config {
	secure, err := strconv.ParseBool(getEnv("SESSION_SECURE", ""))
	if err != nil {
		log.Fatal("Error: ession environment variable")
	}

	config := Config{
		AppName:       getEnv("APPNAME", "NoName"),
		Port:          getEnv("PORT", "8443"),
		TLSPublicKey:  getEnv("TLS_PUBLICKEY", ""),
		TLSPrivateKey: getEnv("TLS_PRIVATEKEY", ""),
		SessionName:   getEnv("SESSION_NAME", "DummySessionName"),
		SessionSecret: getEnv("SESSION_SECRET", ""),
		SessionSecure: secure,
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
