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

// Keys return the visitor count grouped by tag keys.
func (tags *Tags) Keys(filter *Filter) ([]model.TagStats, error) {
	filter = tags.analyzer.getFilter(filter)
	q, args := filter.buildQuery([]Field{
		FieldTagKey,
		FieldVisitors,
		FieldViews,
		FieldRelativeVisitors,
		FieldRelativeViews,
	}, []Field{
		FieldTagKey,
	}, []Field{
		FieldVisitors,
		FieldTagKey,
	}, nil, "")
	stats, err := tags.store.SelectTagStats(filter.Ctx, false, q, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// Breakdown returns the visitor count for tags grouping them by a given key.
// The Filter.Tag must be set, or otherwise the result set will be empty.
func (tags *Tags) Breakdown(filter *Filter) ([]model.TagStats, error) {
	filter = tags.analyzer.getFilter(filter)

	if len(filter.Tag) == 0 {
		return []model.TagStats{}, nil
	}

	q, args := filter.buildQuery([]Field{
		FieldTagValue,
		FieldVisitors,
		FieldViews,
		FieldRelativeVisitors,
		FieldRelativeViews,
	}, []Field{
		FieldTagValue,
	}, []Field{
		FieldVisitors,
		FieldTagValue,
	}, nil, "")
	stats, err := tags.store.SelectTagStats(filter.Ctx, true, q, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}
