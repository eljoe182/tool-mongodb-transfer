package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	EnvOriginURI = "MONGODB_URI_ORIGIN"
	EnvDestURI   = "MONGODB_URI_DESTINATION"
)

// Config holds MongoDB connection URIs for origin and destination.
type Config struct {
	OriginURI string
	DestURI   string
}

// Load reads required URIs from getenv. Empty values are treated as missing.
func Load(getenv func(string) string) (Config, error) {
	if getenv == nil {
		getenv = os.Getenv
	}

	origin := strings.TrimSpace(getenv(EnvOriginURI))
	dest := strings.TrimSpace(getenv(EnvDestURI))

	var missing []string
	if origin == "" {
		missing = append(missing, EnvOriginURI)
	}
	if dest == "" {
		missing = append(missing, EnvDestURI)
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return Config{OriginURI: origin, DestURI: dest}, nil
}

// LoadFromEnv loads an optional .env file, then reads process environment variables.
func LoadFromEnv() (Config, error) {
	_ = godotenv.Load()
	return Load(os.Getenv)
}
