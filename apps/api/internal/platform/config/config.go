package config

import (
	"os"
	"strings"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	DatabaseURL    string
	RedisURL       string
	Port           string
	OTelEndpoint   string
	BaseDomain     string
	AllowedOrigins []string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		DatabaseURL:    envOrDefault("DATABASE_URL", "postgres://zenvikar:zenvikar@localhost:5432/zenvikar?sslmode=disable"),
		RedisURL:       envOrDefault("REDIS_URL", "localhost:6379"),
		Port:           envOrDefault("PORT", "8080"),
		OTelEndpoint:   envOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		BaseDomain:     envOrDefault("BASE_DOMAIN", "zenvikar.localhost"),
		AllowedOrigins: parseAllowedOrigins(os.Getenv("ALLOWED_ORIGINS")),
	}
}

// envOrDefault returns the value of the environment variable named by key,
// or the fallback value if the variable is not set or empty.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// parseAllowedOrigins splits a comma-separated string of origins into a slice.
// Returns ["*"] if the input is empty.
func parseAllowedOrigins(raw string) []string {
	if raw == "" {
		return []string{"*"}
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	if len(origins) == 0 {
		return []string{"*"}
	}
	return origins
}
