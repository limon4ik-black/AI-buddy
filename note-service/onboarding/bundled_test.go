package onboarding

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBundledDocumentsAreRegisteredAndTamperingIsRejected(t *testing.T) {
	db := identityDB(t)
	ctx := context.Background()
	if err := Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join("..", "..", "rag-service", "data", "knowledge", "v1")
	if err := RegisterBundled(ctx, db, ""); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := RegisterBundled(ctx, db, directory); err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err := db.QueryRow(ctx, "SELECT count(*) FROM onboarding.knowledge_documents WHERE category='bundled_json' AND visibility='company'").Scan(&count); err != nil || count != 23 {
		t.Fatalf("registered=%d err=%v", count, err)
	}
	// A bundled source must not overwrite a document owned by another directory.
	if _, err := db.Exec(ctx, "UPDATE onboarding.knowledge_documents SET category='internal' WHERE id=(SELECT id FROM onboarding.knowledge_documents LIMIT 1)"); err != nil {
		t.Fatal(err)
	}
	if err := RegisterBundled(ctx, db, directory); err == nil {
		t.Fatal("conflicting source overwritten")
	}
	for _, body := range []string{"invalid JSON", `{}`, `{"schema_version":1,"visibility":"company","files":[]}`} {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		if err := RegisterBundled(ctx, db, dir); err == nil {
			t.Fatal("invalid manifest accepted")
		}
	}
	if err := RegisterBundled(ctx, db, filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("missing manifest accepted")
	}
	// Tamper with a real JSON without changing its expected digest.
	temp := t.TempDir()
	files, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range files {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(temp, entry.Name()), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	raw, _ := os.ReadFile(filepath.Join(temp, "manifest.json"))
	var manifest struct{ Files []struct{ File string } }
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(temp, manifest.Files[0].File), []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := RegisterBundled(ctx, db, temp); err == nil {
		t.Fatal("tampered data accepted")
	}
}
