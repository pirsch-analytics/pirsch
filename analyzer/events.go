package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v4"
	"github.com/pirsch-analytics/pirsch/v4/db"
	"github.com/pirsch-analytics/pirsch/v4/model"
)

// Events aggregates statistics regarding events.
type Events struct {
	analyzer *Analyzer
	store    db.Store
}

// Events returns the visitor count, views, and conversion rate for custom events.
func (events *Events) Events(filter *Filter) ([]model.EventStats, error) {
	filter = events.analyzer.getFilter(filter)
	q, args := filter.buildQuery([]Field{
		FieldEventName,
		FieldVisitors,
		FieldViews,
		FieldCR,
		FieldEventTimeSpent,
		FieldEventMetaKeys,
	}, []Field{
		FieldEventName,
	}, []Field{
		FieldVisitors,
		FieldEventName,
	})
	stats, err := events.store.SelectEventStats(false, q, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// Breakdown returns the visitor count, views, and conversion rate for a custom event grouping them by a meta value for given key.
// The Filter.EventName and Filter.EventMetaKey must be set, or otherwise the result set will be empty.
func (events *Events) Breakdown(filter *Filter) ([]model.EventStats, error) {
	filter = events.analyzer.getFilter(filter)

	if len(filter.EventName) == 0 || len(filter.EventMetaKey) == 0 {
		return []model.EventStats{}, nil
	}

	q, args := filter.buildQuery([]Field{
		FieldEventName,
		FieldVisitors,
		FieldViews,
		FieldCR,
		FieldEventTimeSpent,
		FieldEventMetaValues,
	}, []Field{
		FieldEventName,
		FieldEventMetaValues,
	}, []Field{
		FieldVisitors,
		FieldEventMetaValues,
	})
	stats, err := events.store.SelectEventStats(true, q, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// List returns events as a list. The metadata is grouped as key-value pairs.
func (events *Events) List(filter *Filter) ([]model.EventListStats, error) {
	if filter == nil {
		filter = NewFilter(pirsch.NullClient)
	}

	filter = events.analyzer.getFilter(filter)
	q, args := filter.buildQuery([]Field{
		FieldEventName,
		FieldEventMeta,
		FieldVisitors,
		FieldCount,
	}, []Field{
		FieldEventName,
		FieldEventMeta,
	}, []Field{
		FieldCount,
		FieldEventName,
	})
	stats, err := events.store.SelectEventListStats(q, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}
