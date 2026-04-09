package db

import (
	"bufio"
	"bytes"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

//go:embed schema
var migrationFiles embed.FS

type migration struct {
	version    int
	statements []string
}

// Migrate runs the database migration for a given connection string.
// This will use the embedded schema migration scripts.
func Migrate(config *ClientConfig) error {
	client, err := NewClient(config)

	if err != nil {
		return err
	}

	if err := createMigrationsTable(client, config.Cluster); err != nil {
		return err
	}

	version, err := getMigrationVersion(client)

	if err != nil {
		return err
	}

	if err := runMigrations(client, version, config.Cluster); err != nil {
		return err
	}

	return client.Close()
}

func createMigrationsTable(client *Client, cluster string) error {
	table := ""
	err := client.QueryRow("SHOW TABLES LIKE 'schema_migrations'").Scan(&table)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if table == "" {
		query := ""

		if cluster != "" {
			query = fmt.Sprintf("CREATE TABLE schema_migrations ON CLUSTER '%s' (version Int64, dirty UInt8, sequence UInt64) Engine=ReplicatedMergeTree ORDER BY (version, dirty, sequence)", cluster)
		} else {
			query = "CREATE TABLE schema_migrations (version Int64, dirty UInt8, sequence UInt64) Engine=MergeTree ORDER BY (version, dirty, sequence)"
		}

		if _, err := client.DB.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func getMigrationVersion(client *Client) (int, error) {
	migration := struct {
		Version int
		Dirty   bool
	}{}
	err := client.QueryRow("SELECT version, dirty FROM schema_migrations ORDER BY sequence DESC LIMIT 1").Scan(&migration.Version, &migration.Dirty)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	if migration.Dirty {
		return 0, errors.New("database dirty")
	}

	return migration.Version, nil
}

func setMigrationVersion(client *Client, version int, dirty bool) error {
	_, err := client.DB.Exec("INSERT INTO schema_migrations (version, dirty, sequence) VALUES (?, ?, ?)", version, dirty, time.Now().UnixNano())
	return err
}

func runMigrations(client *Client, version int, cluster string) error {
	files, err := migrationFiles.ReadDir("schema")

	if err != nil {
		return err
	}

	migrations, err := loadMigrations(files, version, cluster)

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

func loadMigrations(files []fs.DirEntry, version int, cluster string) ([]migration, error) {
	migrations := make([]migration, 0)

	for _, f := range files {
		if filepath.Ext(f.Name()) == ".sql" {
			v, err := parseVersion(f.Name())

			if err != nil {
				return nil, err
			}

			statements, err := parseStatements(f.Name(), cluster)

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

func parseStatements(name, cluster string) ([]string, error) {
	content, err := fs.ReadFile(migrationFiles, filepath.Join("schema", name))

	if err != nil {
		return nil, fmt.Errorf("error reading migration file: %s", err)
	}

	tpl, err := template.New("").Parse(string(content))

	if err != nil {
		return nil, fmt.Errorf("error parsing migration template: %s", err)
	}

	var out bytes.Buffer

	if err := tpl.Execute(&out, struct {
		Cluster string
	}{
		Cluster: cluster,
	}); err != nil {
		return nil, fmt.Errorf("error executing migration template: %s", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(out.String()))
	var buffer strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "--") {
			buffer.WriteString(fmt.Sprintf("%s\n", line))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error parsing migration file: %s", err)
	}

	statements := strings.Split(buffer.String(), ";")
	statementsClean := make([]string, 0, len(statements))

	for _, statement := range statements {
		statement = strings.TrimSpace(statement)

		if statement != "" {
			statementsClean = append(statementsClean, statement)
		}
	}

	return statementsClean, nil
}
