package db

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg/model"
)

// Storage is the database storage interface.
type Storage interface {
	// SavePageViews saves given hits.
	SavePageViews([]model.PageView) error

	// SaveSessions saves given sessions.
	SaveSessions([]model.Session) error

	// SaveEvents saves given events.
	SaveEvents([]model.Event) error

	// SaveRequests saves given requests.
	SaveRequests([]model.Request) error
}
