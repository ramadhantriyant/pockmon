package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// setupGoogleCredentials writes GOOGLE_CREDENTIALS_JSON to a temp file and
// points GOOGLE_APPLICATION_CREDENTIALS at it so ADC picks it up without any
// deprecated credential-parsing APIs. The file must outlive the process because
// ADC re-reads it on every token refresh; the OS reclaims it on container exit.
func setupGoogleCredentials() error {
	credsJSON := os.Getenv("GOOGLE_CREDENTIALS_JSON")
	if credsJSON == "" {
		return fmt.Errorf("GOOGLE_CREDENTIALS_JSON environment variable is required")
	}

	f, err := os.CreateTemp("", "google-creds-*.json")
	if err != nil {
		return fmt.Errorf("create temp credentials file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(credsJSON); err != nil {
		os.Remove(f.Name())
		return fmt.Errorf("write credentials: %w", err)
	}

	if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", f.Name()); err != nil {
		os.Remove(f.Name())
		return fmt.Errorf("set GOOGLE_APPLICATION_CREDENTIALS: %w", err)
	}

	return nil
}

func connectDatabase(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}

func runMigration(dbURL string) error {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "sql/schema"); err != nil {
		return err
	}

	log.Println("database migrated successfully")
	return nil
}
