package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute) // releases idle conns before Neon drops them

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging db: %w", err)
	}

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	slog.Info("Database ready")
	return db, nil
}

// migrate creates tables and indexes if they don't exist
func migrate(db *sql.DB) error {
	statements := []string{

		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id                  BIGSERIAL    PRIMARY KEY,
			email               TEXT         NOT NULL UNIQUE,
			username            TEXT         NOT NULL,
			password_hash       TEXT         NOT NULL,
			is_verified         BOOLEAN      NOT NULL DEFAULT FALSE,
			created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			password_changed_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		)`,

		// OTPs table
		`CREATE TABLE IF NOT EXISTS otps (
			id         BIGSERIAL    PRIMARY KEY,
			user_id    BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			code       TEXT         NOT NULL,
			type       TEXT         NOT NULL CHECK(type IN ('signup', 'reset_password')),
			expires_at TIMESTAMPTZ  NOT NULL,
			used       BOOLEAN      NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		)`,

		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_otps_user_type ON otps(user_id, type)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email    ON users(email)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("executing migration: %w\nstatement: %s", err, stmt)
		}
	}

	slog.Info("Migrations applied successfully")
	return nil
}