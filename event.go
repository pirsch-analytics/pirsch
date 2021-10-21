package pirsch

import (
	"net/http"
	"time"
)

// EventOptions are the options to save a new event.
// The name is required. All other fields are optional.
type EventOptions struct {
	// Name is the name of the event (required).
	Name string

	// Duration is an optional duration that is used to calculate an average time on the dashboard.
	Duration uint32

	// Meta are optional fields used to break down the events that were send for a name.
	Meta map[string]string
}

func (options *EventOptions) getMetaData() ([]string, []string) {
	keys, values := make([]string, 0, len(options.Meta)), make([]string, 0, len(options.Meta))

	for k, v := range options.Meta {
		keys = append(keys, k)
		values = append(values, v)
	}

	return keys, values
}

// EventFromRequest returns the session for given request if found.
// The salt must stay consistent to track visitors across multiple calls.
// The easiest way to track events is to use the Tracker.
// The options must be set!
func EventFromRequest(r *http.Request, salt string, options *HitOptions) *Session {
	if options == nil {
		return nil
	}

	// set default options in case they're nil
	if options.SessionMaxAge.Seconds() == 0 {
		options.SessionMaxAge = defaultSessionMaxAge
	}

	fingerprint := Fingerprint(r, salt+options.Salt)
	session := options.SessionCache.Get(options.ClientID, fingerprint, time.Now().UTC().Add(-options.SessionMaxAge))

	if session != nil {
		getRequestURI(r, options)
		session.ExitPath = getPath(options.Path)
	}

	return session
}
