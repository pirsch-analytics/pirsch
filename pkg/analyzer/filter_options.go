package analyzer

import (
	"fmt"
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"strings"
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
func (options *FilterOptions) Hostnames(filter *Filter, search string) ([]string, error) {
	return options.selectFilterOptions(filter, "hostname", "session", search)
}

// Pages returns all paths.
// This can also be used for the entry and exit pages.
func (options *FilterOptions) Pages(filter *Filter, search string) ([]string, error) {
	return options.selectFilterOptions(filter, "path", "page_view", search)
}

// Referrer returns all referrers.
func (options *FilterOptions) Referrer(filter *Filter, search string) ([]string, error) {
	return options.selectFilterOptions(filter, "referrer", "session", search)
}

// ReferrerName returns all referrer names.
func (options *FilterOptions) ReferrerName(filter *Filter, search string) ([]string, error) {
	return options.selectFilterOptions(filter, "referrer_name", "session", search)
}

// UTMSource returns all UTM sources.
func (options *FilterOptions) UTMSource(filter *Filter, search string) ([]string, error) {
	return options.selectFilterOptions(filter, "utm_source", "session", search)
}

// UTMMedium returns all UTM media.
func (options *FilterOptions) UTMMedium(filter *Filter, search string) ([]string, error) {
	return options.selectFilterOptions(filter, "utm_medium", "session", search)
}

// UTMCampaign returns all UTM campaigns.
func (options *FilterOptions) UTMCampaign(filter *Filter, search string) ([]string, error) {
	return options.selectFilterOptions(filter, "utm_campaign", "session", search)
}

// UTMContent returns all UTM contents.
func (options *FilterOptions) UTMContent(filter *Filter, search string) ([]string, error) {
	return options.selectFilterOptions(filter, "utm_content", "session", search)
}

// UTMTerm returns all UTM terms.
func (options *FilterOptions) UTMTerm(filter *Filter, search string) ([]string, error) {
	return options.selectFilterOptions(filter, "utm_term", "session", search)
}

// Channel returns all channels.
func (options *FilterOptions) Channel(filter *Filter, search string) ([]string, error) {
	return options.selectFilterOptions(filter, "channel", "session", search)
}

// Events return all event names.
func (options *FilterOptions) Events(filter *Filter, search string) ([]string, error) {
	return options.selectFilterOptions(filter, "event_name", "event", search)
}

// Countries returns all countries.
func (options *FilterOptions) Countries(filter *Filter, search string) ([]string, error) {
	return options.selectFilterOptions(filter, "country_code", "session", search)
}

// Regions returns all regions.
func (options *FilterOptions) Regions(filter *Filter, search string) ([]string, error) {
	return options.selectFilterOptions(filter, "region", "session", search)
}

// Cities returns all cities.
func (options *FilterOptions) Cities(filter *Filter, search string) ([]string, error) {
	return options.selectFilterOptions(filter, "city", "session", search)
}

// Languages returns all languages.
func (options *FilterOptions) Languages(filter *Filter, search string) ([]string, error) {
	return options.selectFilterOptions(filter, "language", "session", search)
}

// EventMetadataValues returns all metadata values.
func (options *FilterOptions) EventMetadataValues(filter *Filter, search string) ([]string, error) {
	if filter == nil || len(filter.EventName) == 0 {
		return []string{}, nil
	}

	filter = options.analyzer.getFilter(filter)
	search = strings.TrimSpace(search)
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
	searchQuery := ""

	if search != "" {
		searchQuery = `AND "values" ILIKE ? `
		args = append(args, fmt.Sprintf("%%%s%%", search))
	}

	q := fmt.Sprintf(`SELECT DISTINCT arrayJoin(event_meta_values) AS "values"
		FROM event
		%s
		AND length(event_meta_values) > 0
		%s %s
		ORDER BY "values" ASC
		LIMIT %d`, timeQuery, builder.q.String(), searchQuery, maxOptions)
	return options.store.SelectOptions(filter.Ctx, q, args...)
}

// TagKeys returns all tag keys.
func (options *FilterOptions) TagKeys(filter *Filter, search string) ([]string, error) {
	if filter == nil {
		return []string{}, nil
	}

	filter = options.analyzer.getFilter(filter)
	search = strings.TrimSpace(search)
	timeQuery, args := filter.buildTimeQuery()
	searchQuery := ""

	if search != "" {
		searchQuery = `AND "values" ILIKE ? `
		args = append(args, fmt.Sprintf("%%%s%%", search))
	}

	q := fmt.Sprintf(`SELECT DISTINCT arrayJoin(tag_keys) AS "values"
		FROM page_view
		%s %s
		AND length(tag_values) > 0
		ORDER BY "values" ASC
		LIMIT %d`, timeQuery, searchQuery, maxOptions)
	return options.store.SelectOptions(filter.Ctx, q, args...)
}

// TagValues returns all tag values.
// The Filter.Tag must be set to exactly one tag, or otherwise the result set will be empty.
func (options *FilterOptions) TagValues(filter *Filter, search string) ([]string, error) {
	if filter == nil || len(filter.Tag) != 1 {
		return []string{}, nil
	}

	filter = options.analyzer.getFilter(filter)
	search = strings.TrimSpace(search)
	args := make([]any, 0)
	args = append(args, filter.Tag[0])
	timeQuery, timeArgs := filter.buildTimeQuery()
	args = append(args, timeArgs...)
	args = append(args, filter.Tag[0])
	searchQuery := ""

	if search != "" {
		searchQuery = `AND "keys" ILIKE ? `
		args = append(args, fmt.Sprintf("%%%s%%", search))
	}

	q := fmt.Sprintf(`SELECT DISTINCT tag_values[indexOf(tag_keys, ?)] AS "keys"
		FROM page_view
		%s
		AND length(tag_values) > 0
		AND has(tag_keys, ?)
		%s
		ORDER BY "keys" ASC
		LIMIT %d`, timeQuery, searchQuery, maxOptions)
	return options.store.SelectOptions(filter.Ctx, q, args...)
}

func (options *FilterOptions) selectFilterOptions(filter *Filter, field, table, search string) ([]string, error) {
	filter = options.analyzer.getFilter(filter)
	search = strings.TrimSpace(search)
	timeQuery, args := filter.buildTimeQuery()
	query := fmt.Sprintf(`SELECT DISTINCT %s FROM %s %s `, field, table, timeQuery)

	if search != "" {
		query += fmt.Sprintf(`AND %s ILIKE ? `, field)
		args = append(args, fmt.Sprintf("%%%s%%", search))
	}

	query += fmt.Sprintf(`ORDER BY %s ASC LIMIT %d`, field, maxOptions)
	return options.store.SelectOptions(filter.Ctx, query, args...)
}
