package db

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMigrate(t *testing.T) {
	assert.NoError(t, Migrate("clickhouse://127.0.0.1:9000?x-multi-statement=true"))
}
