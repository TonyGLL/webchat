package config

import (
	"strings"
	"time"
)

// Config holds all configuration for the application.
// Values are read from environment variables.
type Config struct {
	Port               string
	DatabaseURL        string
	RedisURL           string
	JWTSecret          string
	CORSAllowedOrigins []string
	FrontendURL        string
	SMTPFrom           string
	SMTPPassword       string
	SMTPHost           string
	SMTPPort           string

	// Database pool tuning (DB_MAX_OPEN_CONNS, DB_MAX_IDLE_CONNS, DB_CONN_MAX_LIFETIME, DB_CONN_MAX_IDLE_TIME).
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration
}

// IsProduction returns true if the app environment is set to "prod" or "production".
func (c *Config) IsProduction() bool {
	env := Getenv("APP_ENV", "")
	return strings.ToLower(env) == "prod" || strings.ToLower(env) == "production"
}
