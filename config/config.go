package environment

import (
	"os"
)

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

type Config struct {
	AppName string
	Port    string
}

func NewConfig() *Config {
	config := Config{
		AppName: getEnv("APPNAME", "NoName"),
		Port:    getEnv("PORT", "8443"),
	}

	return &config
}
