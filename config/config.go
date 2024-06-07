package environment

import (
	"os"
)

type Config struct {
	AppName       string
	Port          string
	TLSPublicKey  string
	TLSPrivateKey string
}

func NewConfig() *Config {
	config := Config{
		AppName:       getEnv("APPNAME", "NoName"),
		Port:          getEnv("PORT", "8443"),
		TLSPublicKey:  getEnv("TLS_PUBLICKEY", ""),
		TLSPrivateKey: getEnv("TLS_PRIVATEKEY", ""),
	}

	return &config
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
