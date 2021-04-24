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

// CountActiveVisitors implements the Store interface.
/*func (client *MockClient) CountActiveVisitors(filter *Run) int {
	return 0
}

// ActiveVisitors implements the Store interface.
func (client *MockClient) ActiveVisitors(filter *Run) ([]Stats, error) {
	return nil, nil
}

// VisitorLanguages implements the Store interface.
func (client *MockClient) VisitorLanguages(filter *Run) ([]Stats, error) {
	return nil, nil
}*/

// Run implements the Store interface.
func (client *MockClient) Run(query *Query) ([]Stats, error) {
	return nil, nil
}
