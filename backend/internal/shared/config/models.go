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

	// TrustedProxies (TRUSTED_PROXIES, comma-separated IPs/CIDRs) lists the reverse proxies whose
	// X-Forwarded-For is believed. Empty: no proxy is trusted and the client IP is the TCP peer.
	TrustedProxies []string

	// Database pool tuning (DB_MAX_OPEN_CONNS, DB_MAX_IDLE_CONNS, DB_CONN_MAX_LIFETIME, DB_CONN_MAX_IDLE_TIME).
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration
}

// appEnv is the normalised APP_ENV value.
func appEnv() string { return strings.ToLower(strings.TrimSpace(Getenv("APP_ENV", ""))) }

// IsDevelopment is true only when APP_ENV is explicitly dev, development, local or test.
func (c *Config) IsDevelopment() bool {
	switch appEnv() {
	case "dev", "development", "local", "test":
		return true
	}
	return false
}

// IsProduction is secure by default: everything except an explicit dev/development/local/test
// APP_ENV counts as production (Secure refresh cookie, Gin release mode), so a deployment that
// forgets the variable is not left with the weaker settings.
func (c *Config) IsProduction() bool { return !c.IsDevelopment() }

// EnvWarning describes a configuration situation worth a startup warning, or returns "".
func (c *Config) EnvWarning() string {
	switch env := appEnv(); {
	case c.IsDevelopment():
		return "APP_ENV=" + env + ": development mode: the refresh cookie is not marked Secure and Gin runs in debug mode; never use this in production"
	case env == "":
		return "APP_ENV is not set: assuming production (Secure refresh cookie, Gin release mode); set APP_ENV=dev to run over plain http locally"
	}
	return ""
}
