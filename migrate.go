package pirsch

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"time"
)

//go:embed schema
var migrationFiles embed.FS

// Migrate runs the database migration for given connection string.
// This will use the embedded schema migration scripts.
func Migrate(connection string) error {
	client, err := NewClient(connection, nil)

	if err != nil {
		return err
	}

	// TODO
	schema := "pirschtest"

	if err := createMigrationsTable(client, schema); err != nil {
		return err
	}

	version, err := getMigrationVersion(client, schema)

	if err != nil {
		return err
	}

	if err := runMigrations(client, schema, version); err != nil {
		return err
	}

	return client.Close()
}

func createMigrationsTable(client *Client, schema string) error {
	table := ""
	err := client.DB.Get(&table, fmt.Sprintf("SHOW TABLES FROM %s LIKE 'schema_migrations'", schema))
	log.Println(table, err)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err != nil {
		if _, err := client.DB.Exec(fmt.Sprintf("CREATE TABLE %s.schema_migrations (version Int64, dirty UInt8, sequence UInt64) Engine=TinyLog", schema)); err != nil {
			return err
		}
	}

	return nil
}

func getMigrationVersion(client *Client, schema string) (int, error) {
	version := struct {
		Version int
		Dirty   bool
	}{}

	if err := client.DB.Get(&version, fmt.Sprintf("SELECT version, dirty FROM %s.schema_migrations ORDER BY sequence DESC LIMIT 1", schema)); err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if version.Dirty {
		return 0, errors.New("database dirty")
	}

	return version.Version, nil
}

func setMigrationVersion(client *Client, schema string, version int, dirty bool) error {
	_, err := client.DB.Exec(fmt.Sprintf("INSERT INTO %s.schema_migrations (version, dirty, sequence) VALUES (?, ?, ?)", schema), version, dirty, time.Now().UnixNano())
	return err
}

func runMigrations(client *Client, schema string, version int) error {
	files, err := migrationFiles.ReadDir("schema")

	if err != nil {
		return err
	}

	type migration struct {
		version    int
		statements []string
	}

	migrations := make([]migration, 0)

	for _, f := range files {
		if filepath.Ext(f.Name()) == ".sql" {
			v, err := parseVersion(f.Name())

			if err != nil {
				return err
			}

			if v > version {
				migrations = append(migrations, migration{
					version:    v,
					statements: parseStatements(f.Name()),
				})
			}
		}
	}

	/*if err := setMigrationVersion(client, schema, version, true); err != nil {
		return err
	}

	if err := setMigrationVersion(client, schema, version, false); err != nil {
		return err
	}*/

	return nil
}

func parseVersion(name string) (int, error) {
	return 0, nil
}

func parseStatements(name string) []string {
	return nil
}
