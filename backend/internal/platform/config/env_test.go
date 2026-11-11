package config

import (
	"strings"
	"testing"
)

func TestConfiguration(t *testing.T) {
	t.Setenv("BUDDY_TEST_SETTING", "set")
	if Env("BUDDY_TEST_SETTING", "fallback") != "set" || Env("UNSET_BUDDY_SETTING", "fallback") != "fallback" {
		t.Fatal("environment")
	}
	if _, err := Secrets("MISSING_BUDDY_SECRET"); err == nil {
		t.Fatal("missing secret accepted")
	}
	t.Setenv("BUDDY_TEST_SECRET", strings.Repeat("s", 32))
	values, err := Secrets("BUDDY_TEST_SECRET")
	if err != nil || values["BUDDY_TEST_SECRET"] == "" {
		t.Fatal(values, err)
	}
	t.Setenv("JWT_PRIVATE_SEED", "")
	if _, err := JWT(true); err == nil {
		t.Fatal("missing signing seed accepted")
	}
}
