package pirsch

import (
	"database/sql"
	"sync"
	"time"
)

// MockClient is a mock Store implementation.
type MockClient struct {
	Hits []Hit
	m    sync.Mutex
}

// NewMockClient returns a new mock client.
func NewMockClient() *MockClient {
	return &MockClient{
		Hits: make([]Hit, 0),
	}
}

// SaveHits implements the Store interface.
func (client *MockClient) SaveHits(hits []Hit) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.Hits = append(client.Hits, hits...)
	return nil
}

// Session implements the Store interface.
func (client *MockClient) Session(tenantID sql.NullInt64, fingerprint string, maxAge time.Time) (time.Time, error) {
	return time.Now(), nil
}

// Count implements the Store interface.
func (client *MockClient) Count(query *Query) (int, error) {
	return 0, nil
}

// Select implements the Store interface.
func (client *MockClient) Select(query *Query) ([]Stats, error) {
	return nil, nil
}
