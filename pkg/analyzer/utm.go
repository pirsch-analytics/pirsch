package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
)

// UTM aggregates UTM campaign statistics.
type UTM struct {
	analyzer *Analyzer
	store    db.Store
}

// Source returns the visitor count grouped by utm source.
func (utm *UTM) Source(filter *Filter) ([]model.UTMSourceStats, error) {
	ctx, q, args := utm.analyzer.selectByAttribute(filter, "imported_utm_source", FieldUTMSource)
	return utm.store.SelectUTMSourceStats(ctx, q, args...)
}

// Medium returns the visitor count grouped by utm medium.
func (utm *UTM) Medium(filter *Filter) ([]model.UTMMediumStats, error) {
	ctx, q, args := utm.analyzer.selectByAttribute(filter, "imported_utm_medium", FieldUTMMedium)
	return utm.store.SelectUTMMediumStats(ctx, q, args...)
}

// Campaign returns the visitor count grouped by utm source.
func (utm *UTM) Campaign(filter *Filter) ([]model.UTMCampaignStats, error) {
	ctx, q, args := utm.analyzer.selectByAttribute(filter, "imported_utm_campaign", FieldUTMCampaign)
	return utm.store.SelectUTMCampaignStats(ctx, q, args...)
}

// Content returns the visitor count grouped by utm source.
func (utm *UTM) Content(filter *Filter) ([]model.UTMContentStats, error) {
	ctx, q, args := utm.analyzer.selectByAttribute(filter, "", FieldUTMContent)
	return utm.store.SelectUTMContentStats(ctx, q, args...)
}

// Term returns the visitor count grouped by utm source.
func (utm *UTM) Term(filter *Filter) ([]model.UTMTermStats, error) {
	ctx, q, args := utm.analyzer.selectByAttribute(filter, "", FieldUTMTerm)
	return utm.store.SelectUTMTermStats(ctx, q, args...)
}
