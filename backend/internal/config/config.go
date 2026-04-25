package config

import "os"

// Config contains runtime configuration for the backend service.
type Config struct {
	Port      string
	JWTSecret string
	AdminUser string
	AdminPass string
}

// LoadFromEnv reads configuration from environment variables with safe defaults.
func LoadFromEnv() Config {
	return Config{
		Port:      readOrDefault("PORT", "8080"),
		JWTSecret: readOrDefault("JWT_SECRET", "change-me-in-production"),
		AdminUser: readOrDefault("ADMIN_USER", "admin"),
		AdminPass: readOrDefault("ADMIN_PASS", "admin123"),
	}
}

func readOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
