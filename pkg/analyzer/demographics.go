package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
)

// Demographics aggregate metadata statistics like the referrer, browser, and OS.
type Demographics struct {
	analyzer *Analyzer
	store    db.Store
}

// Languages return the visitor count grouped by language.
func (demographics *Demographics) Languages(filter *Filter) ([]model.LanguageStats, error) {
	ctx, q, args := demographics.analyzer.selectByAttribute(filter, "imported_language", FieldLanguage)
	return demographics.store.SelectLanguageStats(ctx, q, args...)
}

// Countries return the visitor count grouped by country.
func (demographics *Demographics) Countries(filter *Filter) ([]model.CountryStats, error) {
	ctx, q, args := demographics.analyzer.selectByAttribute(filter, "imported_country", FieldCountry)
	return demographics.store.SelectCountryStats(ctx, q, args...)
}

// Regions return the visitor count grouped by region.
func (demographics *Demographics) Regions(filter *Filter) ([]model.RegionStats, error) {
	ctx, q, args := demographics.analyzer.selectByAttribute(filter, "imported_region", FieldRegion, FieldCountryRegion)
	return demographics.store.SelectRegionStats(ctx, q, args...)
}

// Cities return the visitor count grouped by city.
func (demographics *Demographics) Cities(filter *Filter) ([]model.CityStats, error) {
	ctx, q, args := demographics.analyzer.selectByAttribute(filter, "imported_city", FieldCity, FieldRegionCity, FieldCountryCity)
	return demographics.store.SelectCityStats(ctx, q, args...)
}
