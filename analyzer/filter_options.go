package analyzer

import (
	"fmt"
	"github.com/pirsch-analytics/pirsch/v5/db"
)

// FilterOptions returns options that can be used to filter results.
// This includes distinct pages, referrers, ... for a given period.
// Common options, like the operating system or browser, are not read from the database.
type FilterOptions struct {
	analyzer *Analyzer
	store    db.Store
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

// Cities returns all cities.
func (options *FilterOptions) Cities(filter *Filter) ([]string, error) {
	return options.selectFilterOptions(filter, "city", "session")
}

// Languages returns all languages.
func (options *FilterOptions) Languages(filter *Filter) ([]string, error) {
	return options.selectFilterOptions(filter, "language", "session")
}

func (options *FilterOptions) selectFilterOptions(filter *Filter, field, table string) ([]string, error) {
	filter = options.analyzer.getFilter(filter)
	timeQuery, args := filter.buildTimeQuery()
	q := fmt.Sprintf(`SELECT DISTINCT %s FROM %s %s ORDER BY %s ASC`, field, table, timeQuery, field)
	return options.store.SelectOptions(q, args...)
}
