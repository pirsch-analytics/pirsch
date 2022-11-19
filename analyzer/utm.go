package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v5/db"
	"github.com/pirsch-analytics/pirsch/v5/model"
)

// UTM aggregates UTM campaign statistics.
type UTM struct {
	analyzer *Analyzer
	store    db.Store
}

// Source returns the visitor count grouped by utm source.
func (utm *UTM) Source(filter *Filter) ([]model.UTMSourceStats, error) {
	q, args := utm.analyzer.selectByAttribute(filter, FieldUTMSource)
	return utm.store.SelectUTMSourceStats(q, args...)
}

// Medium returns the visitor count grouped by utm medium.
func (utm *UTM) Medium(filter *Filter) ([]model.UTMMediumStats, error) {
	q, args := utm.analyzer.selectByAttribute(filter, FieldUTMMedium)
	return utm.store.SelectUTMMediumStats(q, args...)
}

// Campaign returns the visitor count grouped by utm source.
func (utm *UTM) Campaign(filter *Filter) ([]model.UTMCampaignStats, error) {
	q, args := utm.analyzer.selectByAttribute(filter, FieldUTMCampaign)
	return utm.store.SelectUTMCampaignStats(q, args...)
}

// Content returns the visitor count grouped by utm source.
func (utm *UTM) Content(filter *Filter) ([]model.UTMContentStats, error) {
	q, args := utm.analyzer.selectByAttribute(filter, FieldUTMContent)
	return utm.store.SelectUTMContentStats(q, args...)
}

// Term returns the visitor count grouped by utm source.
func (utm *UTM) Term(filter *Filter) ([]model.UTMTermStats, error) {
	q, args := utm.analyzer.selectByAttribute(filter, FieldUTMTerm)
	return utm.store.SelectUTMTermStats(q, args...)
}
