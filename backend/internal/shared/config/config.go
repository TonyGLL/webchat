package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// NewConfig creates a new Config struct and loads values from environment variables.
// It also loads from a .env file if the path is provided.
func NewConfig(envFile string) (*Config, error) {
	// Load .env file if path is specified. It's not an error if it doesn't exist.
	if envFile != "" {
		_ = godotenv.Load(envFile)
	}

	cfg := &Config{
		Port:               Getenv("PORT", "8080"),
		DatabaseURL:        Getenv("DATABASE_URL", ""),
		RedisURL:           Getenv("REDIS_URL", ""),
		JWTSecret:          Getenv("JWT_SECRET", "default-secret"),
		CORSAllowedOrigins: parseCorsOrigins(Getenv("CORS_ALLOWED_ORIGINS", "")),
		SMTPFrom:           Getenv("SMTP_FROM", ""),
		SMTPPassword:       Getenv("SMTP_PASSWORD", ""),
		SMTPHost:           Getenv("SMTP_HOST", ""),
		SMTPPort:           Getenv("SMTP_PORT", "587"),
	}

	// If DATABASE_URL is not set, try to construct it from individual DB environment variables
	if cfg.DatabaseURL == "" {
		dbUser := Getenv("DB_USER", "")
		dbPassword := Getenv("DB_PASSWORD", "")
		dbHost := Getenv("DB_HOST", "")
		dbPort := Getenv("DB_PORT", "")
		dbName := Getenv("DB_NAME", "")

		if dbUser == "" || dbHost == "" || dbPort == "" || dbName == "" {
			return nil, fmt.Errorf("DATABASE_URL is not set and individual DB_USER, DB_HOST, DB_PORT, DB_NAME environment variables are not fully provided")
		}
		cfg.DatabaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
	}

	return cfg, nil
}

// Getenv retrieves the value of the environment variable named by the key,
// or returns the provided fallback value if the variable is not set.
func Getenv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// parseCorsOrigins takes a comma-separated string and splits it into a slice of strings.
func parseCorsOrigins(s string) []string {
	if s == "" {
		return nil // Return nil to let the CORS middleware use its default behavior
	}
	origins := strings.Split(s, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}
	return origins
}

// LoadConfig is kept for compatibility but the main entry point is NewConfig.
// It's useful for pre-loading before other packages initialize.
func LoadConfig(path string) error {
	err := godotenv.Load(path)
	if err != nil {
		return fmt.Errorf("error loading .env file from path %s: %w", path, err)
	}
	return nil
}
