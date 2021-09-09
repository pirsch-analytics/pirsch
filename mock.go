package pirsch

import (
	"sync"
	"time"
)

// MockClient is a mock Store implementation.
type MockClient struct {
	Hits          []Hit
	Events        []Event
	ReturnSession *Session
	m             sync.Mutex
}

// NewMockClient returns a new mock client.
func NewMockClient() *MockClient {
	return &MockClient{
		Hits:   make([]Hit, 0),
		Events: make([]Event, 0),
	}
}

// SaveHits implements the Store interface.
func (client *MockClient) SaveHits(hits []Hit) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.Hits = append(client.Hits, hits...)
	return nil
}

// SaveEvents implements the Store interface.
func (client *MockClient) SaveEvents(events []Event) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.Events = append(client.Events, events...)
	return nil
}

// Session implements the Store interface.
func (client *MockClient) Session(clientID int64, fingerprint string, maxAge time.Time) (Session, error) {
	if client.ReturnSession != nil {
		return *client.ReturnSession, nil
	}

	return Session{
		Path:    "",
		Time:    time.Now().UTC(),
		Session: time.Now().UTC(),
	}, nil
}

// Count implements the Store interface.
func (client *MockClient) Count(query string, args ...interface{}) (int, error) {
	return 0, nil
}

// Get implements the Store interface.
func (client *MockClient) Get(result interface{}, query string, args ...interface{}) error {
	return nil
}

// Select implements the Store interface.
func (client *MockClient) Select(results interface{}, query string, args ...interface{}) error {
	return nil
}
