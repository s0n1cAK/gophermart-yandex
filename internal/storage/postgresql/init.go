package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	Database *sql.DB
}

func NewPostgresStorage(DSN *url.URL) (*PostgresStorage, error) {
	var db *sql.DB
	var err error

	err = retryWrapper(context.Background(), func() error {
		db, err = sql.Open("postgres", DSN.String())
		err := db.Ping()
		return err
	})
	if err != nil {
		return &PostgresStorage{}, err
	}

	err = migration(DSN.String())
	if err != nil {
		return &PostgresStorage{}, err
	}

	return &PostgresStorage{
		db,
	}, nil
}

func migration(DSN string) error {
	m, err := migrate.New(
		"file://migrations",
		DSN,
	)
	if err != nil {
		return fmt.Errorf("can't make migration: %v", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}
