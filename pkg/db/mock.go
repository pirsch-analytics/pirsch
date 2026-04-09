package db

import (
	"sort"
	"sync"

	"github.com/pirsch-analytics/pirsch/v7/pkg/model"
)

// Mock implements the Storage interface.
type Mock struct {
	pageViews     []model.PageView
	sessions      []model.Session
	events        []model.Event
	requests      []model.Request
	ReturnSession *model.Session
	m             sync.Mutex
}

// NewMock returns a new mock client.
func NewMock() *Mock {
	return &Mock{
		pageViews: make([]model.PageView, 0),
		sessions:  make([]model.Session, 0),
		events:    make([]model.Event, 0),
		requests:  make([]model.Request, 0),
	}
}

// SavePageViews implements the Storage interface.
func (client *Mock) SavePageViews(pageViews []model.PageView) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.pageViews = append(client.pageViews, pageViews...)
	return nil
}

// SaveSessions implements the Storage interface.
func (client *Mock) SaveSessions(sessions []model.Session) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.sessions = append(client.sessions, sessions...)
	return nil
}

// SaveEvents implements the Storage interface.
func (client *Mock) SaveEvents(events []model.Event) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.events = append(client.events, events...)
	return nil
}

// SaveRequests implements the Storage interface.
func (client *Mock) SaveRequests(bots []model.Request) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.requests = append(client.requests, bots...)
	return nil
}

// GetPageViews returns a sorted copy of the page views slice.
func (client *Mock) GetPageViews() []model.PageView {
	client.m.Lock()
	defer client.m.Unlock()
	data := make([]model.PageView, len(client.pageViews))
	copy(data, client.pageViews)
	sort.Slice(data, func(i, j int) bool {
		if data[i].Time.Before(data[j].Time) {
			return true
		}

		return false
	})
	return data
}

// GetSessions returns a sorted copy of the session slice.
func (client *Mock) GetSessions() []model.Session {
	client.m.Lock()
	defer client.m.Unlock()
	data := make([]model.Session, len(client.sessions))
	copy(data, client.sessions)
	sort.Slice(data, func(i, j int) bool {
		if data[i].Time.Before(data[j].Time) {
			return true
		}

		return false
	})
	return data
}

// GetEvents returns a sorted copy of the events slice.
func (client *Mock) GetEvents() []model.Event {
	client.m.Lock()
	defer client.m.Unlock()
	data := make([]model.Event, len(client.events))
	copy(data, client.events)
	sort.Slice(data, func(i, j int) bool {
		if data[i].Time.Before(data[j].Time) {
			return true
		}

		return false
	})
	return data
}

// GetRequests returns a sorted copy of the request slice.
func (client *Mock) GetRequests() []model.Request {
	client.m.Lock()
	defer client.m.Unlock()
	data := make([]model.Request, len(client.requests))
	copy(data, client.requests)
	sort.Slice(data, func(i, j int) bool {
		if data[i].Time.Before(data[j].Time) {
			return true
		}

		return false
	})
	return data
}
