package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ashwinkg/monitoring-service/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDB struct {
	pool *pgxpool.Pool
}

type Event struct {
	ID        string          `json:"id"` // UUID
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"` // arbitrary JSON
	Status    string          `json:"status"`  // "pending" | "processed" | "failed"
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type EventRepository interface {
	Insert(ctx context.Context, event *Event) error
	GetByID(ctx context.Context, id string) (*Event, error)
	List(ctx context.Context, limit, offset int) ([]*Event, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	Delete(ctx context.Context, id string) error
}

const (
	StatusPending   = "pending"
	StatusProcessed = "processed"
	StatusFailed    = "failed"
)

func NewPostgres(cfg *config.Config) *PostgresDB {
	poolCfg, err := pgxpool.ParseConfig(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to parse Postgres URL: %v", err)
	}

	// Pool tuning
	poolCfg.MaxConns = int32(cfg.PostgresMaxConns)
	poolCfg.MinConns = int32(cfg.PostgresMinConns)
	poolCfg.MaxConnLifetime = 1 * time.Hour
	poolCfg.MaxConnIdleTime = 30 * time.Minute
	poolCfg.HealthCheckPeriod = 1 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), cfg.PostgresConnTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping Postgres: %v", err)
	}

	db := &PostgresDB{pool: pool}

	// Auto-migrate
	if err := db.migrate(context.Background()); err != nil {
		log.Fatalf("Failed to auto-migrate Postgres: %v", err)
	}

	log.Println("Postgres connection established")
	return db
}

func (db *PostgresDB) Close() {
	db.pool.Close()
	log.Println("Postgres connection closed")
}

func (db *PostgresDB) migrate(ctx context.Context) error {
	query := `
		CREATE EXTENSION IF NOT EXISTS "pgcrypto";

		CREATE TABLE IF NOT EXISTS events (
			id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			type       TEXT        NOT NULL,
			payload    JSONB       NOT NULL DEFAULT '{}',
			status     TEXT        NOT NULL DEFAULT 'pending',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_events_type      ON events(type);
		CREATE INDEX IF NOT EXISTS idx_events_status    ON events(status);
		CREATE INDEX IF NOT EXISTS idx_events_created_at ON events(created_at DESC);
	`
	_, err := db.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	log.Println("Postgres schema migrated")
	return nil
}

// Insert adds a new event and returns its generated ID and timestamps
func (db *PostgresDB) Insert(ctx context.Context, e *Event) error {
	query := `
		INSERT INTO events (type, payload, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	return db.pool.QueryRow(ctx, query, e.Type, e.Payload, e.Status).
		Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
}

// GetByID fetches a single event by UUID
func (db *PostgresDB) GetByID(ctx context.Context, id string) (*Event, error) {
	query := `
		SELECT id, type, payload, status, created_at, updated_at
		FROM events
		WHERE id = $1
	`
	e := &Event{}
	err := db.pool.QueryRow(ctx, query, id).
		Scan(&e.ID, &e.Type, &e.Payload, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("GetByID %s: %w", id, err)
	}
	return e, nil
}

// List returns paginated events ordered by newest first
func (db *PostgresDB) List(ctx context.Context, limit, offset int) ([]*Event, error) {
	query := `
		SELECT id, type, payload, status, created_at, updated_at
		FROM events
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := db.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("List: %w", err)
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		e := &Event{}
		if err := rows.Scan(&e.ID, &e.Type, &e.Payload, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("List scan: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// UpdateStatus changes the status of an event and bumps updated_at
func (db *PostgresDB) UpdateStatus(ctx context.Context, id, status string) error {
	query := `
		UPDATE events
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	result, err := db.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("UpdateStatus %s: %w", id, err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("UpdateStatus: event %s not found", id)
	}
	return nil
}

// Delete removes an event by UUID
func (db *PostgresDB) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM events WHERE id = $1`
	result, err := db.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("Delete %s: %w", id, err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("Delete: event %s not found", id)
	}
	return nil
}
