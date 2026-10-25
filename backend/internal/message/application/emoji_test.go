package application

import "testing"

func TestIsSingleEmoji(t *testing.T) {
	valid := map[string]string{
		"thumbs up":     "\U0001F44D",
		"heart + VS16":  "\u2764\uFE0F",
		"skin tone":     "\U0001F44D\U0001F3FD",
		"family ZWJ":    "\U0001F468\u200D\U0001F469\u200D\U0001F467\u200D\U0001F466",
		"couple + kiss": "\U0001F469\U0001F3FD\u200D\u2764\uFE0F\u200D\U0001F48B\u200D\U0001F468\U0001F3FF",
		"flag":          "\U0001F1EA\U0001F1E8",
		"keycap":        "1\uFE0F\u20E3",
		"scotland flag": "\U0001F3F4\U000E0067\U000E0062\U000E0073\U000E0063\U000E0074\U000E007F",
		"rainbow flag":  "\U0001F3F3\uFE0F\u200D\U0001F308",
	}
	for name, s := range valid {
		if !isSingleEmoji(s) {
			t.Errorf("%s should be a single emoji", name)
		}
	}
	invalid := map[string]string{
		"empty":             "",
		"text":              "hi",
		"two emoji":         "\U0001F44D\U0001F44D",
		"emoji then text":   "\U0001F44D!",
		"digit alone":       "1",
		"zalgo":             "a\u0300\u0301\u0302\u0303",
		"zalgo on emoji":    "\U0001F44D\u0300\u0301",
		"dangling ZWJ":      "\U0001F468\u200D",
		"leading ZWJ":       "\u200D\U0001F468",
		"lone ZWJ":          "\u200D",
		"three indicators":  "\U0001F1EA\U0001F1E8\U0001F1E6",
		"unterminated tags": "\U0001F3F4\U000E0067",
		"bidi override":     "\u202E\U0001F44D",
		"invalid utf8":      "\xff",
		"space padded":      " \U0001F44D",
	}
	for name, s := range invalid {
		if isSingleEmoji(s) {
			t.Errorf("%s should be rejected", name)
		}
	}
}
