package db

import (
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func NewDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)

	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(time.Minute * 3)
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		return nil, err
	}

	slog.Info("Database ready", "Path", path)
	return db, nil
}

// creates tables and indexes if not exists
func migrate(db *sql.DB) error {

	statements := []string{
		`PRAGMA journal_mode=WAL`, // WAL gives better concurrent read performance
		`PRAGMA foreign_keys=ON`,  // enforce FK constraints

		//user Table
		`CREATE TABLE IF NOT EXISTS users (
			id            INTEGER  PRIMARY KEY AUTOINCREMENT,
			email         TEXT     NOT NULL UNIQUE,
			username      TEXT     NOT NULL UNIQUE,
			password_hash TEXT     NOT NULL,
			is_verified   INTEGER  NOT NULL DEFAULT 0,
			created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			password_changed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP

		)`,

		//OTP table
		`CREATE TABLE IF NOT EXISTS otps (
			id         INTEGER  PRIMARY KEY AUTOINCREMENT,
			user_id    INTEGER  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			code       TEXT     NOT NULL,
			type       TEXT     NOT NULL CHECK(type IN ('signup','reset_password')),
			expires_at DATETIME NOT NULL,
			used       INTEGER  NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,

		//TABEL INDEX
		`CREATE INDEX IF NOT EXISTS idx_otps_user_type ON otps(user_id, type)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email    ON users(email)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return err

		}
	}
	return nil
}
