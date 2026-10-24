package domain

import "testing"

func TestValidateUsername(t *testing.T) {
	for _, ok := range []string{"ana", "Ana_99", "a.b-c", "abc", "123456789012345678901234567890"} {
		if err := ValidateUsername(ok); err != nil {
			t.Errorf("%q should be valid: %v", ok, err)
		}
	}
	for _, bad := range []string{"", "ab", "victim@example.com", "has space", "ñandú", "a/b", "1234567890123456789012345678901"} {
		if err := ValidateUsername(bad); err == nil {
			t.Errorf("%q should be rejected", bad)
		}
	}
}

func TestNormalizeEmail(t *testing.T) {
	if got := NormalizeEmail("  Ana@Example.COM "); got != "ana@example.com" {
		t.Fatalf("got %q", got)
	}
}
