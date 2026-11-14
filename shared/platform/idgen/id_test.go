package idgen

import (
	"testing"
)

func TestNew(t *testing.T) {
	if New() == New() {
		t.Fatal("duplicate identifiers")
	}
}
