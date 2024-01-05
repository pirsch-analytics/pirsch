package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
)

// Tags aggregates statistics regarding tags.
type Tags struct {
	analyzer *Analyzer
	store    db.Store
}

// Keys returns the visitor count grouped by tag keys.
func (events *Tags) Keys(filter *Filter) ([]model.TagStats, error) {
	filter = events.analyzer.getFilter(filter)
	q, args := filter.buildQuery([]Field{
		FieldTagKey,
		FieldVisitors,
	}, []Field{
		FieldTagKey,
	}, []Field{
		FieldVisitors,
		FieldTagKey,
	})
	stats, err := events.store.SelectTagStats(filter.Ctx, false, q, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// Breakdown returns the visitor count for tags grouping them by given key.
// The Filter.Tag must be set, or otherwise the result set will be empty.
func (events *Tags) Breakdown(filter *Filter) ([]model.TagStats, error) {
	filter = events.analyzer.getFilter(filter)

	if len(filter.Tag) == 0 {
		return []model.TagStats{}, nil
	}

	q, args := filter.buildQuery([]Field{
		FieldTagKey,
		FieldVisitors,
		FieldTagValue,
	}, []Field{
		FieldTagKey,
		FieldTagValue,
	}, []Field{
		FieldVisitors,
		FieldTagKey,
		FieldTagValue,
	})
	stats, err := events.store.SelectTagStats(filter.Ctx, true, q, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}
