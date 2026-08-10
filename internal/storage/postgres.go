package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/luma434/audittrail/internal/chain"
)

// Postgres implements chain.Repository against the insert-only events
// table (see migrations/000001_create_events_table.up.sql).
type Postgres struct {
	pool *pgxpool.Pool
}

var _ chain.Repository = (*Postgres)(nil)

func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}

func (p *Postgres) Latest(ctx context.Context) (*chain.Entry, error) {
	row := p.pool.QueryRow(ctx, `
		SELECT payload, "timestamp", prev_hash, hash
		FROM events
		ORDER BY id DESC
		LIMIT 1`)

	var e chain.Entry
	if err := row.Scan(&e.Payload, &e.Timestamp, &e.PrevHash, &e.Hash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query latest event: %w", err)
	}
	return &e, nil
}

func (p *Postgres) Insert(ctx context.Context, e chain.Entry) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO events (payload, "timestamp", prev_hash, hash)
		VALUES ($1, $2, $3, $4)`,
		e.Payload, e.Timestamp, e.PrevHash, e.Hash)
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}
	return nil
}

func (p *Postgres) List(ctx context.Context, limit, offset int) ([]chain.Entry, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT payload, "timestamp", prev_hash, hash
		FROM events
		ORDER BY id ASC
		LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()
	return scanEntries(rows)
}

func (p *Postgres) All(ctx context.Context) ([]chain.Entry, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT payload, "timestamp", prev_hash, hash
		FROM events
		ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list all events: %w", err)
	}
	defer rows.Close()
	return scanEntries(rows)
}

func scanEntries(rows pgx.Rows) ([]chain.Entry, error) {
	var entries []chain.Entry
	for rows.Next() {
		var e chain.Entry
		if err := rows.Scan(&e.Payload, &e.Timestamp, &e.PrevHash, &e.Hash); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read events: %w", err)
	}
	return entries, nil
}
