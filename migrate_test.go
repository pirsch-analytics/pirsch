package pirsch

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMigrate(t *testing.T) {
	assert.NoError(t, Migrate("tcp://127.0.0.1:9000?database=pirschtest"))
}
