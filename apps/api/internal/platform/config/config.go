package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	DatabaseURL        string
	RedisURL           string
	Port               string
	OTelEndpoint       string
	Environment        string
	JWTSecret          string
	JWTTTLMinutes      int
	GoogleClientID     string
	GoogleClientSecret string
	FacebookAppID      string
	FacebookAppSecret  string
	APIPublicURL       string
	BaseDomain         string
	AllowedOrigins     []string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		DatabaseURL:        envOrDefault("DATABASE_URL", "postgres://zenvikar:zenvikar@localhost:5432/zenvikar?sslmode=disable"),
		RedisURL:           envOrDefault("REDIS_URL", "localhost:6379"),
		Port:               envOrDefault("PORT", "8080"),
		OTelEndpoint:       envOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		Environment:        strings.ToLower(envOrDefault("APP_ENV", "development")),
		JWTSecret:          envOrDefault("JWT_SECRET", "dev-only-change-me"),
		JWTTTLMinutes:      envIntOrDefault("JWT_TTL_MINUTES", 120),
		GoogleClientID:     envOrDefault("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: envOrDefault("GOOGLE_CLIENT_SECRET", ""),
		FacebookAppID:      envOrDefault("FACEBOOK_APP_ID", ""),
		FacebookAppSecret:  envOrDefault("FACEBOOK_APP_SECRET", ""),
		APIPublicURL:       envOrDefault("API_PUBLIC_URL", "http://api.zenvikar.localhost"),
		BaseDomain:         envOrDefault("BASE_DOMAIN", "zenvikar.localhost"),
		AllowedOrigins:     parseAllowedOrigins(os.Getenv("ALLOWED_ORIGINS")),
	}
}

// IsProduction reports whether the app is running in production mode.
func (c *Config) IsProduction() bool {
	return strings.EqualFold(strings.TrimSpace(c.Environment), "production")
}

// envOrDefault returns the value of the environment variable named by key,
// or the fallback value if the variable is not set or empty.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// envIntOrDefault returns the integer value of the environment variable named
// by key, or fallback when parsing fails or the variable is not set.
func envIntOrDefault(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
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
