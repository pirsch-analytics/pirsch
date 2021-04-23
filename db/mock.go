package db

import (
	"database/sql"
	"github.com/pirsch-analytics/pirsch/model"
	"sync"
	"time"
)

// MockClient is a mock Store implementation.
type MockClient struct {
	Hits []model.Hit
	m    sync.Mutex
}

// NewMockClient returns a new mock client.
func NewMockClient() *MockClient {
	return &MockClient{
		Hits: make([]model.Hit, 0),
	}
}

// SaveHits implements the Store interface.
func (client *MockClient) SaveHits(hits []model.Hit) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.Hits = append(client.Hits, hits...)
	return nil
}

// Session implements the Store interface.
func (client *MockClient) Session(tenantID sql.NullInt64, fingerprint string, maxAge time.Time) (time.Time, error) {
	return time.Now(), nil
}
