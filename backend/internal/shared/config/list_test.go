package config

import "testing"

func TestParseListDropsBlanksAndTrims(t *testing.T) {
	got := parseList(" 10.0.0.0/8, ,127.0.0.1,")
	if len(got) != 2 || got[0] != "10.0.0.0/8" || got[1] != "127.0.0.1" {
		t.Fatalf("got %v", got)
	}
	if parseList("") != nil {
		t.Fatal("empty input must give nil (trust no proxy)")
	}
}
