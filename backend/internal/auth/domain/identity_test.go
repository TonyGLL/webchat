package domain

import (
	"errors"
	"testing"
)

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

func TestReservedUsernames(t *testing.T) {
	for _, bad := range []string{"admin", "Admin", "ADMINISTRATOR", "ad.min", "a_d_m_i_n", "moderator", "support", "system", "root", "WebChat"} {
		if err := ValidateUsername(bad); !errors.Is(err, ErrReservedUsername) {
			t.Errorf("%q should be reserved, got %v", bad, err)
		}
	}
	for _, ok := range []string{"adminton", "ana", "my_support_team"} {
		if err := ValidateUsername(ok); err != nil {
			t.Errorf("%q should be allowed: %v", ok, err)
		}
	}
}

func TestNormalizeDisplayName(t *testing.T) {
	good := map[string]string{
		"  Ana  María ":  "Ana María",
		"Jose\u0301":     "Jos\u00e9", // NFC composes
		"Ana\u00a0Maria": "Ana Maria", // no-break space becomes a plain space
		"Nguyễn Văn":     "Nguyễn Văn",
		"李":              "李",
	}
	for in, want := range good {
		if got, err := NormalizeDisplayName(in); err != nil || got != want {
			t.Errorf("NormalizeDisplayName(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for name, bad := range map[string]string{
		"empty": "", "blank": "   ", "newline": "Ana\nB", "tab": "Ana\tB", "nul": "A\x00B",
		"bidi override": "Ana\u202Eevil", "bidi isolate": "\u2066Ana", "zero width space": "A\u200BB",
		"zero width joiner": "A\u200DB", "BOM": "\uFEFFAna", "line separator": "A\u2028B", "invalid utf8": "A\xffB",
	} {
		if got, err := NormalizeDisplayName(bad); !errors.Is(err, ErrInvalidDisplayName) {
			t.Errorf("%s: got %q, %v; want ErrInvalidDisplayName", name, got, err)
		}
	}
}
