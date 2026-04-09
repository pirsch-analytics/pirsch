package db

import (
	"context"

	"github.com/pirsch-analytics/pirsch/v7/pkg/model"
)

// Storage is the database storage interface.
type Storage interface {
	// SavePageViews saves given hits.
	SavePageViews(context.Context, []model.PageView) error

	// SaveSessions saves given sessions.
	SaveSessions(context.Context, []model.Session) error

	// SaveEvents saves given events.
	SaveEvents(context.Context, []model.Event) error

	// SaveRequests saves given requests.
	SaveRequests(context.Context, []model.Request) error
}
