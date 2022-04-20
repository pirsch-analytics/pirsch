package pirsch

import (
	"time"
)

// Store is the database storage interface.
type Store interface {
	// SavePageViews saves given hits.
	SavePageViews([]PageView) error

	// SaveSessions saves given sessions.
	SaveSessions([]Session) error

	// SaveEvents saves given events.
	SaveEvents([]Event) error

	// SaveUserAgents saves given UserAgent headers.
	SaveUserAgents([]UserAgent) error

	// Session returns the last hit for given client, fingerprint, and maximum age.
	Session(uint64, uint64, time.Time) (*Session, error)

	// Count returns the number of results for given query.
	Count(string, ...any) (int, error)

	// Get returns a single result for given query.
	// The result must be a pointer.
	Get(any, string, ...any) error

	// Select returns the results for given query.
	// The results must be a pointer to a slice.
	Select(any, string, ...any) error
}
