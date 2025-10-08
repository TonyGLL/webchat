package config

import "strings"

// Config holds all configuration for the application.
// Values are read from environment variables.
type Config struct {
	Port               string
	DatabaseURL        string
	JWTSecret          string
	CORSAllowedOrigins []string
}

// IsProduction returns true if the app environment is set to "prod" or "production".
func (c *Config) IsProduction() bool {
	env := Getenv("APP_ENV", "")
	return strings.ToLower(env) == "prod" || strings.ToLower(env) == "production"
}
