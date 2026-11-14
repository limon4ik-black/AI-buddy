package security

import (
	"testing"
)

func TestTokenHelpers(t *testing.T) {
	if Equal("", "") || Equal("a", "b") || !Equal("a", "a") {
		t.Fatal("secret comparison")
	}
	if TokenHash("test") == "test" || TokenHash("test") != TokenHash("test") {
		t.Fatal("token hashing")
	}
}
