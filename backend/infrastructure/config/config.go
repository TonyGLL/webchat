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
		DatabaseURL:        Getenv("DATABASE_URL", ""), // No default, should be set
		JWTSecret:          Getenv("JWT_SECRET", "default-secret"),
		CORSAllowedOrigins: parseCorsOrigins(Getenv("CORS_ALLOWED_ORIGINS", "")),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is not set")
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