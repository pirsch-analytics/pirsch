package pirsch

import (
	"sync"
	"time"
)

// ClientMock is a mock Store implementation.
type ClientMock struct {
	PageViews     []PageView
	Sessions      []Session
	Events        []Event
	UserAgents    []UserAgent
	ReturnSession *Session
	m             sync.Mutex
}

// NewMockClient returns a new mock client.
func NewMockClient() *ClientMock {
	return &ClientMock{
		PageViews:  make([]PageView, 0),
		Sessions:   make([]Session, 0),
		Events:     make([]Event, 0),
		UserAgents: make([]UserAgent, 0),
	}
}

// SavePageViews implements the Store interface.
func (client *ClientMock) SavePageViews(pageViews []PageView) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.PageViews = append(client.PageViews, pageViews...)
	return nil
}

// SaveSessions implements the Store interface.
func (client *ClientMock) SaveSessions(sessions []Session) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.Sessions = append(client.Sessions, sessions...)
	return nil
}

// SaveEvents implements the Store interface.
func (client *ClientMock) SaveEvents(events []Event) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.Events = append(client.Events, events...)
	return nil
}

// SaveUserAgents implements the Store interface.
func (client *ClientMock) SaveUserAgents(userAgents []UserAgent) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.UserAgents = append(client.UserAgents, userAgents...)
	return nil
}

// Session implements the Store interface.
func (client *ClientMock) Session(uint64, uint64, time.Time) (*Session, error) {
	if client.ReturnSession != nil {
		return client.ReturnSession, nil
	}

	return nil, nil
}

// Count implements the Store interface.
func (client *ClientMock) Count(string, ...any) (int, error) {
	return 0, nil
}
