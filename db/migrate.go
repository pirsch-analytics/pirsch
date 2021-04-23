package db

import (
	"embed"
	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"net/http"
)

//go:embed schema
var schema embed.FS

// Migrate runs the database migration for given connection string.
// This will use the embedded schema migration scripts.
func Migrate(connection string) error {
	source, err := httpfs.New(http.FS(schema), "schema")

	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("httpfs", source, connection)

	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	if sourceErr, dbErr := m.Close(); sourceErr != nil || dbErr != nil {
		if sourceErr != nil {
			return sourceErr
		}

		return dbErr
	}

	return nil
}
