package pirsch

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//go:embed schema
var migrationFiles embed.FS

type migration struct {
	version    int
	statements []string
}

// Migrate runs the database migration for given connection string.
// This will use the embedded schema migration scripts.
func Migrate(connection string) error {
	client, err := NewClient(connection, nil)

	if err != nil {
		return err
	}

	database, err := getDatabaseName(connection)

	if err != nil {
		return err
	}

	if _, err := client.DB.Exec(fmt.Sprintf("USE %s", database)); err != nil {
		return err
	}

	if err := createMigrationsTable(client); err != nil {
		return err
	}

	version, err := getMigrationVersion(client)

	if err != nil {
		return err
	}

	if err := runMigrations(client, version); err != nil {
		return err
	}

	return client.Close()
}

func getDatabaseName(connection string) (string, error) {
	u, err := url.Parse(connection)

	if err != nil {
		return "", err
	}

	schema := strings.TrimLeft(u.Path, "/")

	if schema == "" {
		return "", errors.New("the database name must be set")
	}

	return schema, nil
}

func createMigrationsTable(client *Client) error {
	table := ""
	err := client.DB.Get(&table, "SHOW TABLES LIKE 'schema_migrations'")

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if err != nil {
		if _, err := client.DB.Exec("CREATE TABLE schema_migrations (version Int64, dirty UInt8, sequence UInt64) Engine=TinyLog"); err != nil {
			return err
		}
	}

	return nil
}

func getMigrationVersion(client *Client) (int, error) {
	version := struct {
		Version int
		Dirty   bool
	}{}

	if err := client.DB.Get(&version, "SELECT version, dirty FROM schema_migrations ORDER BY sequence DESC LIMIT 1"); err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if version.Dirty {
		return 0, errors.New("database dirty")
	}

	return version.Version, nil
}

func setMigrationVersion(client *Client, version int, dirty bool) error {
	_, err := client.DB.Exec("INSERT INTO schema_migrations (version, dirty, sequence) VALUES (?, ?, ?)", version, dirty, time.Now().UnixNano())
	return err
}

func runMigrations(client *Client, version int) error {
	files, err := migrationFiles.ReadDir("schema")

	if err != nil {
		return err
	}

	migrations, err := loadMigrations(files, version)

	if err != nil {
		return err
	}

	for _, m := range migrations {
		if err := setMigrationVersion(client, m.version, true); err != nil {
			return err
		}

		for _, s := range m.statements {
			if _, err := client.DB.Exec(s); err != nil {
				return err
			}
		}

		if err := setMigrationVersion(client, m.version, false); err != nil {
			return err
		}
	}

	return nil
}

func loadMigrations(files []fs.DirEntry, version int) ([]migration, error) {
	migrations := make([]migration, 0)

	for _, f := range files {
		if filepath.Ext(f.Name()) == ".sql" {
			v, err := parseVersion(f.Name())

			if err != nil {
				return nil, err
			}

			statements, err := parseStatements(f.Name())

			if err != nil {
				return nil, err
			}

			if v > version {
				migrations = append(migrations, migration{
					version:    v,
					statements: statements,
				})
			}
		}
	}

	return migrations, nil
}

func parseVersion(name string) (int, error) {
	left, _, found := strings.Cut(name, "_")

	if !found {
		return 0, errors.New("migration filename needs to start with the version number")
	}

	version, err := strconv.ParseInt(left, 10, 64)
	return int(version), err
}

func parseStatements(name string) ([]string, error) {
	content, err := fs.ReadFile(migrationFiles, filepath.Join("schema", name))

	if err != nil {
		return nil, errors.New(fmt.Sprintf("error reading migrationi file: %s", err))
	}

	statements := strings.Split(string(content), ";")
	statementsClean := make([]string, 0, len(statements))

	for _, statement := range statements {
		statement = strings.TrimSpace(statement)

		if statement != "" {
			statementsClean = append(statementsClean, statement)
		}
	}

	return statementsClean, nil
}
