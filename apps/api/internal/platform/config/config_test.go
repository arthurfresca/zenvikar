package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear any env vars that might be set
	envVars := []string{"DATABASE_URL", "REDIS_URL", "PORT", "OTEL_EXPORTER_OTLP_ENDPOINT", "APP_ENV", "JWT_SECRET", "JWT_TTL_MINUTES", "BASE_DOMAIN", "ALLOWED_ORIGINS"}
	for _, key := range envVars {
		t.Setenv(key, "")
	}

	cfg := Load()

	if cfg.DatabaseURL != "postgres://zenvikar:zenvikar@localhost:5432/zenvikar?sslmode=disable" {
		t.Errorf("unexpected DatabaseURL: %s", cfg.DatabaseURL)
	}
	if cfg.RedisURL != "localhost:6379" {
		t.Errorf("unexpected RedisURL: %s", cfg.RedisURL)
	}
	if cfg.Port != "8080" {
		t.Errorf("unexpected Port: %s", cfg.Port)
	}
	if cfg.OTelEndpoint != "localhost:4317" {
		t.Errorf("unexpected OTelEndpoint: %s", cfg.OTelEndpoint)
	}
	if cfg.Environment != "development" {
		t.Errorf("unexpected Environment: %s", cfg.Environment)
	}
	if cfg.JWTSecret != "dev-only-change-me" {
		t.Errorf("unexpected JWTSecret: %s", cfg.JWTSecret)
	}
	if cfg.JWTTTLMinutes != 120 {
		t.Errorf("unexpected JWTTTLMinutes: %d", cfg.JWTTTLMinutes)
	}
	if cfg.BaseDomain != "zenvikar.localhost" {
		t.Errorf("unexpected BaseDomain: %s", cfg.BaseDomain)
	}
	if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != "*" {
		t.Errorf("unexpected AllowedOrigins: %v", cfg.AllowedOrigins)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://custom:pass@db:5432/mydb")
	t.Setenv("REDIS_URL", "redis:6380")
	t.Setenv("PORT", "9090")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "otel:4318")
	t.Setenv("APP_ENV", "staging")
	t.Setenv("JWT_SECRET", "super-secret")
	t.Setenv("JWT_TTL_MINUTES", "45")
	t.Setenv("BASE_DOMAIN", "example.com")
	t.Setenv("ALLOWED_ORIGINS", "https://example.com, https://app.example.com")

	cfg := Load()

	if cfg.DatabaseURL != "postgres://custom:pass@db:5432/mydb" {
		t.Errorf("unexpected DatabaseURL: %s", cfg.DatabaseURL)
	}
	if cfg.RedisURL != "redis:6380" {
		t.Errorf("unexpected RedisURL: %s", cfg.RedisURL)
	}
	if cfg.Port != "9090" {
		t.Errorf("unexpected Port: %s", cfg.Port)
	}
	if cfg.OTelEndpoint != "otel:4318" {
		t.Errorf("unexpected OTelEndpoint: %s", cfg.OTelEndpoint)
	}
	if cfg.Environment != "staging" {
		t.Errorf("unexpected Environment: %s", cfg.Environment)
	}
	if cfg.JWTSecret != "super-secret" {
		t.Errorf("unexpected JWTSecret: %s", cfg.JWTSecret)
	}
	if cfg.JWTTTLMinutes != 45 {
		t.Errorf("unexpected JWTTTLMinutes: %d", cfg.JWTTTLMinutes)
	}
	if cfg.BaseDomain != "example.com" {
		t.Errorf("unexpected BaseDomain: %s", cfg.BaseDomain)
	}
	if len(cfg.AllowedOrigins) != 2 || cfg.AllowedOrigins[0] != "https://example.com" || cfg.AllowedOrigins[1] != "https://app.example.com" {
		t.Errorf("unexpected AllowedOrigins: %v", cfg.AllowedOrigins)
	}
}

func TestParseAllowedOrigins(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty string", "", []string{"*"}},
		{"single origin", "https://example.com", []string{"https://example.com"}},
		{"multiple origins", "https://a.com,https://b.com", []string{"https://a.com", "https://b.com"}},
		{"with spaces", " https://a.com , https://b.com ", []string{"https://a.com", "https://b.com"}},
		{"only commas", ",,", []string{"*"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAllowedOrigins(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d origins, got %d: %v", len(tt.expected), len(result), result)
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("origin[%d]: expected %q, got %q", i, tt.expected[i], v)
				}
			}
		})
	}
}

func TestEnvOrDefault(t *testing.T) {
	t.Run("returns env value when set", func(t *testing.T) {
		os.Setenv("TEST_CONFIG_VAR", "custom_value")
		defer os.Unsetenv("TEST_CONFIG_VAR")

		if got := envOrDefault("TEST_CONFIG_VAR", "default"); got != "custom_value" {
			t.Errorf("expected custom_value, got %s", got)
		}
	})

	t.Run("returns default when not set", func(t *testing.T) {
		os.Unsetenv("TEST_CONFIG_VAR_MISSING")

		if got := envOrDefault("TEST_CONFIG_VAR_MISSING", "fallback"); got != "fallback" {
			t.Errorf("expected fallback, got %s", got)
		}
	})
}

func TestConfigIsProduction(t *testing.T) {
	cfg := &Config{Environment: "production"}
	if !cfg.IsProduction() {
		t.Fatalf("expected production mode")
	}

	cfg.Environment = "Production"
	if !cfg.IsProduction() {
		t.Fatalf("expected case-insensitive production mode")
	}

	cfg.Environment = "development"
	if cfg.IsProduction() {
		t.Fatalf("did not expect development mode to be production")
	}
}

func TestEnvIntOrDefault(t *testing.T) {
	t.Setenv("JWT_TTL_MINUTES", "60")
	if got := envIntOrDefault("JWT_TTL_MINUTES", 120); got != 60 {
		t.Fatalf("expected 60, got %d", got)
	}

	t.Setenv("JWT_TTL_MINUTES", "")
	if got := envIntOrDefault("JWT_TTL_MINUTES", 120); got != 120 {
		t.Fatalf("expected fallback 120, got %d", got)
	}

	t.Setenv("JWT_TTL_MINUTES", "invalid")
	if got := envIntOrDefault("JWT_TTL_MINUTES", 120); got != 120 {
		t.Fatalf("expected fallback for invalid value, got %d", got)
	}

	t.Setenv("JWT_TTL_MINUTES", "-10")
	if got := envIntOrDefault("JWT_TTL_MINUTES", 120); got != 120 {
		t.Fatalf("expected fallback for non-positive value, got %d", got)
	}
}
