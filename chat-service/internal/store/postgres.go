package store

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/limon4ik-black/AI-buddy/chat-service/internal/chat"
	"github.com/limon4ik-black/AI-buddy/chat-service/migrations"
)

type Postgres struct{ Pool *pgxpool.Pool }

// Apply the single MVP migration atomically; the lock also serializes parallel startups.
func (p *Postgres) Migrate(ctx context.Context) error {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	schema := p.Pool.Config().ConnConfig.RuntimeParams["search_path"]
	if schema == "" {
		schema = "public"
	}
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err = tx.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS "+quoted+"; SET LOCAL search_path TO "+quoted); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(8793201)"); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, migrations.Initial); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, migrations.Context); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (p *Postgres) Create(ctx context.Context, user uuid.UUID) (chat.Created, error) {
	var c chat.Created
	err := p.Pool.QueryRow(ctx, "INSERT INTO chats(id,user_id) VALUES($1,$2) RETURNING id,title,created_at", uuid.New(), user).Scan(&c.ID, &c.Title, &c.CreatedAt)
	return c, err
}
func (p *Postgres) List(ctx context.Context, user uuid.UUID) ([]chat.Summary, error) {
	rows, err := p.Pool.Query(ctx, "SELECT id,title,updated_at FROM chats WHERE user_id=$1 ORDER BY updated_at DESC,id", user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]chat.Summary, 0)
	for rows.Next() {
		var c chat.Summary
		if err := rows.Scan(&c.ID, &c.Title, &c.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}
func (p *Postgres) History(ctx context.Context, user, id uuid.UUID) ([]chat.Message, error) {
	tx, err := p.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var exists bool
	if err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM chats WHERE id=$1 AND user_id=$2)", id, user).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, chat.ErrNotFound
	}
	rows, err := tx.Query(ctx, "SELECT id,role,content,created_at FROM messages WHERE chat_id=$1 ORDER BY id", id)
	if err != nil {
		return nil, err
	}
	result := make([]chat.Message, 0)
	for rows.Next() {
		var m chat.Message
		if err = rows.Scan(&m.ID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		result = append(result, m)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	return result, tx.Commit(ctx)
}
func (p *Postgres) Delete(ctx context.Context, user, id uuid.UUID) error {
	tag, err := p.Pool.Exec(ctx, "DELETE FROM chats WHERE id=$1 AND user_id=$2", id, user)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return chat.ErrNotFound
	}
	return nil
}
func (p *Postgres) Append(ctx context.Context, user, id uuid.UUID, role, content string) (chat.Message, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return chat.Message{}, err
	}
	defer tx.Rollback(ctx)
	// Lock the parent against concurrent deletion and update the activity timestamp atomically.
	var found uuid.UUID
	err = tx.QueryRow(ctx, "SELECT id FROM chats WHERE id=$1 AND user_id=$2 FOR UPDATE", id, user).Scan(&found)
	if errors.Is(err, pgx.ErrNoRows) {
		return chat.Message{}, chat.ErrNotFound
	}
	if err != nil {
		return chat.Message{}, err
	}
	var m chat.Message
	err = tx.QueryRow(ctx, "INSERT INTO messages(chat_id,role,content) VALUES($1,$2,$3) RETURNING id,role,content,created_at", id, role, content).Scan(&m.ID, &m.Role, &m.Content, &m.CreatedAt)
	if err != nil {
		return m, err
	}
	if _, err = tx.Exec(ctx, "UPDATE chats SET updated_at=clock_timestamp() WHERE id=$1", id); err != nil {
		return m, err
	}
	return m, tx.Commit(ctx)
}
