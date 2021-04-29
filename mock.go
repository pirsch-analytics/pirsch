package pirsch

import (
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
func (client *MockClient) Session(clientID int64, fingerprint string, maxAge time.Time) (string, time.Time, time.Time, error) {
	return "", time.Now(), time.Now(), nil
}

// Count implements the Store interface.
func (client *MockClient) Count(query string, args ...interface{}) (int, error) {
	return 0, nil
}

func (client *MockClient) Get(query string, args ...interface{}) (*Stats, error) {
	return nil, nil
}

// Select implements the Store interface.
func (client *MockClient) Select(query string, args ...interface{}) ([]Stats, error) {
	return nil, nil
}
