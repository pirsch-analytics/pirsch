package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"sort"
	"time"
)

// Sessions aggregates statistics regarding single sessions.
type Sessions struct {
	analyzer *Analyzer
	store    db.Store
}

// List returns a list of sessions for given filter.
func (sessions *Sessions) List() ([]model.Session, error) {
	// TODO
	return nil, nil
}

// Breakdown returns the page views and events for a single session in chronological order.
func (sessions *Sessions) Breakdown(filter *Filter) ([]model.SessionStep, error) {
	filter = sessions.analyzer.getFilter(filter)

	if filter.VisitorID == 0 || filter.SessionID == 0 {
		return nil, nil
	}

	q, args := filter.buildQuery([]Field{FieldPageViewsAll}, nil, nil)
	pageViews, err := sessions.store.SelectPageViews(filter.Ctx, q, args...)

	if err != nil {
		return nil, err
	}

	q, args = filter.buildQuery([]Field{FieldEventsAll}, nil, nil)
	events, err := sessions.store.SelectEvents(filter.Ctx, q, args...)

	if err != nil {
		return nil, err
	}

	stats := make([]model.SessionStep, 0, len(pageViews)+len(events))

	for i := range pageViews {
		stats = append(stats, model.SessionStep{
			PageView: &pageViews[i],
		})
	}

	for i := range events {
		stats = append(stats, model.SessionStep{
			Event: &events[i],
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		var a, b time.Time

		if stats[i].PageView != nil {
			a = stats[i].PageView.Time
		} else {
			a = stats[i].Event.Time
		}

		if stats[j].PageView != nil {
			b = stats[j].PageView.Time
		} else {
			b = stats[j].Event.Time
		}

		return a.Before(b)
	})
	return stats, nil
}
