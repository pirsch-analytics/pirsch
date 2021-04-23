package db

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient("tcp://127.0.0.1:9000?debug=true")
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NoError(t, client.DB.Ping())
}
