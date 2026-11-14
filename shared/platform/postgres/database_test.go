package postgres

import (
	"context"
	"testing"

	"github.com/limon4ik-black/AI-buddy/shared/migrations"
)

func TestInvalidURL(t *testing.T) {
	if _, err := Open(context.Background(), "invalid://broken", "info", migrations.Info); err == nil {
		t.Fatal("invalid database URL accepted")
	}
}
