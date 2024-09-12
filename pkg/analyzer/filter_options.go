package analyzer

import (
	"fmt"
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
)

const (
	maxOptions = 200
)

// FilterOptions returns options that can be used to filter results.
// This includes distinct pages, referrers, ... for a given period.
// Common options, like the operating system or browser, are not read from the database.
type FilterOptions struct {
	analyzer *Analyzer
	store    db.Store
}

// Hostnames returns all hostnames.
func (options *FilterOptions) Hostnames(filter *Filter) ([]string, error) {
	return options.selectFilterOptions(filter, "hostname", "session")
}

// Pages returns all paths.
// This can also be used for the entry and exit pages.
func (options *FilterOptions) Pages(filter *Filter) ([]string, error) {
	return options.selectFilterOptions(filter, "path", "page_view")
}

// Referrer returns all referrers.
func (options *FilterOptions) Referrer(filter *Filter) ([]string, error) {
	return options.selectFilterOptions(filter, "referrer", "session")
}

// ReferrerName returns all referrer names.
func (options *FilterOptions) ReferrerName(filter *Filter) ([]string, error) {
	return options.selectFilterOptions(filter, "referrer_name", "session")
}

// UTMSource returns all UTM sources.
func (options *FilterOptions) UTMSource(filter *Filter) ([]string, error) {
	return options.selectFilterOptions(filter, "utm_source", "session")
}

// UTMMedium returns all UTM media.
func (options *FilterOptions) UTMMedium(filter *Filter) ([]string, error) {
	return options.selectFilterOptions(filter, "utm_medium", "session")
}

// UTMCampaign returns all UTM campaigns.
func (options *FilterOptions) UTMCampaign(filter *Filter) ([]string, error) {
	return options.selectFilterOptions(filter, "utm_campaign", "session")
}

// UTMContent returns all UTM contents.
func (options *FilterOptions) UTMContent(filter *Filter) ([]string, error) {
	return options.selectFilterOptions(filter, "utm_content", "session")
}

// UTMTerm returns all UTM terms.
func (options *FilterOptions) UTMTerm(filter *Filter) ([]string, error) {
	return options.selectFilterOptions(filter, "utm_term", "session")
}

// Events returns all event names.
func (options *FilterOptions) Events(filter *Filter) ([]string, error) {
	return options.selectFilterOptions(filter, "event_name", "event")
}

// Countries returns all countries.
func (options *FilterOptions) Countries(filter *Filter) ([]string, error) {
	return options.selectFilterOptions(filter, "country_code", "session")
}

// Regions returns all regions.
func (options *FilterOptions) Regions(filter *Filter) ([]string, error) {
	return options.selectFilterOptions(filter, "region", "session")
}

// Cities returns all cities.
func (options *FilterOptions) Cities(filter *Filter) ([]string, error) {
	return options.selectFilterOptions(filter, "city", "session")
}

// Languages returns all languages.
func (options *FilterOptions) Languages(filter *Filter) ([]string, error) {
	return options.selectFilterOptions(filter, "language", "session")
}

// EventMetadataValues returns all metadata values.
func (options *FilterOptions) EventMetadataValues(filter *Filter) ([]string, error) {
	if filter == nil || len(filter.EventName) == 0 {
		return []string{}, nil
	}

	filter = options.analyzer.getFilter(filter)
	timeQuery, args := filter.buildTimeQuery()
	builder := queryBuilder{
		filter: &Filter{
			EventName: filter.EventName,
		},
		from:   events,
		fields: []Field{FieldEventName},
	}
	builder.whereFields()
	args = append(args, builder.args...)
	q := fmt.Sprintf(`SELECT DISTINCT arrayJoin(event_meta_values) AS "values"
		FROM event
		%s
		AND length(event_meta_values) > 0
		%s
		ORDER BY "values" ASC
		LIMIT %d`, timeQuery, builder.q.String(), maxOptions)
	return options.store.SelectOptions(filter.Ctx, q, args...)
}

// TagKeys returns all tag keys.
func (options *FilterOptions) TagKeys(filter *Filter) ([]string, error) {
	if filter == nil {
		return []string{}, nil
	}

	filter = options.analyzer.getFilter(filter)
	timeQuery, args := filter.buildTimeQuery()
	q := fmt.Sprintf(`SELECT DISTINCT arrayJoin(tag_keys) AS "values"
		FROM page_view
		%s
		AND length(tag_values) > 0
		ORDER BY "values" ASC
		LIMIT %d`, timeQuery, maxOptions)
	return options.store.SelectOptions(filter.Ctx, q, args...)
}

// TagValues returns all tag values.
// The Filter.Tag must be set to exactly one tag, or otherwise the result set will be empty.
func (options *FilterOptions) TagValues(filter *Filter) ([]string, error) {
	if filter == nil || len(filter.Tag) != 1 {
		return []string{}, nil
	}

	filter = options.analyzer.getFilter(filter)
	args := make([]any, 0)
	args = append(args, filter.Tag[0])
	timeQuery, timeArgs := filter.buildTimeQuery()
	args = append(args, timeArgs...)
	args = append(args, filter.Tag[0])
	q := fmt.Sprintf(`SELECT DISTINCT tag_values[indexOf(tag_keys, ?)] AS "keys"
		FROM page_view
		%s
		AND length(tag_values) > 0
		AND has(tag_keys, ?)
		ORDER BY "keys" ASC
		LIMIT %d`, timeQuery, maxOptions)
	return options.store.SelectOptions(filter.Ctx, q, args...)
}

func (options *FilterOptions) selectFilterOptions(filter *Filter, field, table string) ([]string, error) {
	filter = options.analyzer.getFilter(filter)
	timeQuery, args := filter.buildTimeQuery()
	q := fmt.Sprintf(`SELECT DISTINCT %s FROM %s %s ORDER BY %s ASC LIMIT %d`, field, table, timeQuery, field, maxOptions)
	return options.store.SelectOptions(filter.Ctx, q, args...)
}
