package tracker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_validate(t *testing.T) {
	cfg := Config{}
	cfg.validate()
	assert.Len(t, cfg.Salt, 20)
	assert.NotZero(t, cfg.FingerprintKey0)
	assert.NotZero(t, cfg.FingerprintKey1)
	assert.Greater(t, cfg.Worker, 1)
	assert.Equal(t, defaultWorkerBufferSize, cfg.WorkerBufferSize)
	assert.Equal(t, defaultWorkerTimeout, cfg.WorkerTimeout)
	assert.NotNil(t, cfg.SessionCache)
	assert.NotNil(t, cfg.Logger)
	cfg.WorkerTimeout = time.Second * 999
	cfg.validate()
	assert.Equal(t, maxWorkerTimeout, cfg.WorkerTimeout)
}
