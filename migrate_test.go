package pirsch

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMigrate(t *testing.T) {
	dropDB()
	assert.NotNil(t, Migrate(nil))
	assert.NoError(t, Migrate(&ClientConfig{
		Hostname:      "127.0.0.1",
		Port:          9000,
		Database:      "pirschtest",
		SSLSkipVerify: true,
		Debug:         true,
	}))
}

func TestParseVersion(t *testing.T) {
	version, err := parseVersion("0001_baseline.up.sql")
	assert.NoError(t, err)
	assert.Equal(t, 1, version)
	version, err = parseVersion("0015_is_bounce.up.sql")
	assert.NoError(t, err)
	assert.Equal(t, 15, version)
	version, err = parseVersion("baseline.up.sql")
	assert.Equal(t, "migration filename needs to start with the version number", err.Error())
	assert.Equal(t, 0, version)
}

func TestParseStatements(t *testing.T) {
	statements, err := parseStatements("0013_remove_backup.up.sql")
	assert.NoError(t, err)
	assert.Len(t, statements, 3)
	assert.Equal(t, "DROP TABLE IF EXISTS session_backup", statements[0])
	assert.Equal(t, "DROP TABLE IF EXISTS page_view_backup", statements[1])
	assert.Equal(t, "DROP TABLE IF EXISTS event_backup", statements[2])
}
