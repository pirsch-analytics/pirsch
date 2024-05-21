package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
)

// Demographics aggregates metadata statistics like the referrer, browser, and OS.
type Demographics struct {
	analyzer *Analyzer
	store    db.Store
}

// Languages returns the visitor count grouped by language.
func (demographics *Demographics) Languages(filter *Filter) ([]model.LanguageStats, error) {
	ctx, q, args := demographics.analyzer.selectByAttribute(filter, FieldLanguage)
	return demographics.store.SelectLanguageStats(ctx, q, args...)
}

// Countries returns the visitor count grouped by country.
func (demographics *Demographics) Countries(filter *Filter) ([]model.CountryStats, error) {
	ctx, q, args := demographics.analyzer.selectByAttribute(filter, FieldCountry)
	return demographics.store.SelectCountryStats(ctx, q, args...)
}

// Regions returns the visitor count grouped by region.
func (demographics *Demographics) Regions(filter *Filter) ([]model.RegionStats, error) {
	ctx, q, args := demographics.analyzer.selectByAttribute(filter, FieldRegion, FieldCountryRegion)
	return demographics.store.SelectRegionStats(ctx, q, args...)
}

// Cities returns the visitor count grouped by city.
func (demographics *Demographics) Cities(filter *Filter) ([]model.CityStats, error) {
	ctx, q, args := demographics.analyzer.selectByAttribute(filter, FieldCity, FieldRegionCity, FieldCountryCity)
	return demographics.store.SelectCityStats(ctx, q, args...)
}
