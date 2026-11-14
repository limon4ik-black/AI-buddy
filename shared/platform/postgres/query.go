package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Query[T any](ctx context.Context, db *pgxpool.Pool, sql string, args ...any) ([]T, error) {
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[T])
}

func One[T any](ctx context.Context, db *pgxpool.Pool, sql string, args ...any) (T, error) {
	var zero T
	rows, err := Query[T](ctx, db, sql, args...)
	if err != nil {
		return zero, err
	}
	if len(rows) == 0 {
		return zero, pgx.ErrNoRows
	}
	return rows[0], nil
}
