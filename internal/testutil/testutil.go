package testutil

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/atcdot/SharedPrep/internal/storage"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
)

const databaseURL = "postgres://sharedprep:sharedprep@localhost:5432/sharedprep?sslmode=disable"

func NewTestDB(t *testing.T) *storage.Postgres {
	t.Helper()
	ctx := context.Background()

	logger := slog.Default()
	db, err := storage.New(ctx, databaseURL, logger)
	if err != nil {
		t.Fatalf("connect to db: %v", err)
	}

	cleanDB(t, db)
	t.Cleanup(func() { cleanDB(t, db) })

	return db
}

func cleanDB(t *testing.T, db *storage.Postgres) {
	t.Helper()
	_, err := db.Pool.Exec(context.Background(), `
		TRUNCATE items, participants, events, users RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("clean db: %v", err)
	}
}

// RunMigrationsOnce ensures migrations are applied. Call from TestMain.
func RunMigrationsOnce() error {
	goose.SetDialect("postgres")
	db, err := goose.OpenDBWithDriver("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		DELETE FROM items; DELETE FROM participants; DELETE FROM events;
	`)
	if err != nil {
		return fmt.Errorf("clean: %w", err)
	}
	return nil
}

func GetPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}
