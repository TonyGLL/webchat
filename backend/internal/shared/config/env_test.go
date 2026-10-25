package config

import (
	"os"
	"testing"
)

// Secure by default: only an explicit dev-like APP_ENV turns the production behaviour off.
func TestIsProductionIsSecureByDefault(t *testing.T) {
	tests := []struct {
		env      string
		unset    bool
		wantProd bool
		wantWarn bool
	}{
		{unset: true, wantProd: true, wantWarn: true},
		{env: "", wantProd: true, wantWarn: true},
		{env: "prod", wantProd: true},
		{env: "Production", wantProd: true},
		{env: "staging", wantProd: true},
		{env: "prodution", wantProd: true}, // a typo must not downgrade security
		{env: "dev", wantProd: false, wantWarn: true},
		{env: " LOCAL ", wantProd: false, wantWarn: true},
		{env: "development", wantProd: false, wantWarn: true},
		{env: "test", wantProd: false, wantWarn: true},
	}
	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			if tt.unset {
				t.Setenv("APP_ENV", "x") // registers the restore
				os.Unsetenv("APP_ENV")
			} else {
				t.Setenv("APP_ENV", tt.env)
			}
			c := &Config{}
			if got := c.IsProduction(); got != tt.wantProd {
				t.Errorf("IsProduction() = %v, want %v", got, tt.wantProd)
			}
			if got := c.EnvWarning() != ""; got != tt.wantWarn {
				t.Errorf("EnvWarning() = %q, want warning=%v", c.EnvWarning(), tt.wantWarn)
			}
		})
	}
}
