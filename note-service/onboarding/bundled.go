package onboarding

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RegisterBundled publishes the trusted company dataset in Note's directory.
// Embeddings stay owned by RAG; Note owns citation lookup and the document ACL.
func RegisterBundled(ctx context.Context, db *pgxpool.Pool, directory string) error {
	if directory == "" {
		return nil
	}
	var manifest struct {
		Schema     int    `json:"schema_version"`
		Visibility string `json:"visibility"`
		Files      []struct {
			File, SHA256 string
			ID           string `json:"document_id"`
			Chunks       int    `json:"chunks"`
		} `json:"files"`
	}
	raw, err := os.ReadFile(filepath.Join(directory, "manifest.json"))
	if err != nil {
		return err
	}
	if err = json.Unmarshal(raw, &manifest); err != nil {
		return err
	}
	if manifest.Schema != 1 || manifest.Visibility != "company" || len(manifest.Files) == 0 {
		return fmt.Errorf("invalid bundled company manifest")
	}
	files, err := filepath.Glob(filepath.Join(directory, "*.json"))
	if err != nil {
		return err
	}
	if len(files) != len(manifest.Files)+1 {
		return fmt.Errorf("bundled manifest file count mismatch")
	}
	type document struct {
		ID     string `json:"document_id"`
		Source string `json:"source"`
		Chunks []struct {
			Index    int      `json:"chunk_index"`
			Headings []string `json:"headings"`
			Text     string   `json:"text"`
			Context  string   `json:"contextualized_text"`
		} `json:"chunks"`
	}
	docs := make([]document, 0, len(manifest.Files))
	names := map[string]bool{}
	ids := map[string]bool{}
	for _, entry := range manifest.Files {
		if filepath.Base(entry.File) != entry.File || strings.Contains(entry.File, "\\") || !strings.HasSuffix(entry.File, ".json") || entry.File == "manifest.json" || names[entry.File] || ids[entry.ID] || entry.ID == "" {
			return fmt.Errorf("invalid bundled manifest entry")
		}
		names[entry.File] = true
		ids[entry.ID] = true
		raw, err = os.ReadFile(filepath.Join(directory, entry.File))
		if err != nil {
			return err
		}
		if fmt.Sprintf("%x", sha256.Sum256(raw)) != entry.SHA256 {
			return fmt.Errorf("bundled checksum mismatch: %s", entry.File)
		}
		var doc document
		if err = json.Unmarshal(raw, &doc); err != nil {
			return err
		}
		if doc.ID != entry.ID || len(doc.Chunks) != entry.Chunks || len(doc.Chunks) == 0 {
			return fmt.Errorf("bundled document metadata mismatch: %s", entry.File)
		}
		sort.Slice(doc.Chunks, func(i, j int) bool { return doc.Chunks[i].Index < doc.Chunks[j].Index })
		for i, chunk := range doc.Chunks {
			if chunk.Index != i {
				return fmt.Errorf("invalid chunk order: %s", entry.File)
			}
		}
		docs = append(docs, doc)
	}
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(271411)"); err != nil {
		return err
	}
	for _, doc := range docs {
		var category, visibility string
		var team *string
		err = tx.QueryRow(ctx, "SELECT category,visibility,team_id FROM onboarding.knowledge_documents WHERE id=$1 FOR UPDATE", doc.ID).Scan(&category, &visibility, &team)
		if err != nil && err != pgx.ErrNoRows {
			return err
		}
		if err == nil && (category != "bundled_json" || visibility != "company" || team != nil) {
			return fmt.Errorf("bundled document ID conflicts with Note data: %s", doc.ID)
		}
		title := doc.Source
		if len(doc.Chunks[0].Headings) > 0 {
			title = doc.Chunks[0].Headings[0]
		}
		sections := make([]string, 0, len(doc.Chunks))
		for _, chunk := range doc.Chunks {
			text := chunk.Context
			if text == "" {
				text = chunk.Text
			}
			sections = append(sections, text)
		}
		content := strings.Join(sections, "\n\n")
		_, err = tx.Exec(ctx, `INSERT INTO onboarding.knowledge_documents(id,title,content,category,visibility,source_url,synthetic)
 VALUES($1,$2,$3,'bundled_json','company',$4,false)
 ON CONFLICT(id) DO UPDATE SET title=EXCLUDED.title,content=EXCLUDED.content,source_url=EXCLUDED.source_url,
 deleted_at=NULL,updated_at=CASE WHEN knowledge_documents.content IS DISTINCT FROM EXCLUDED.content OR knowledge_documents.title IS DISTINCT FROM EXCLUDED.title OR knowledge_documents.deleted_at IS NOT NULL THEN now() ELSE knowledge_documents.updated_at END`, doc.ID, title, content, doc.Source)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
