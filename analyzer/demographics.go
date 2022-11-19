package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v5/db"
	"github.com/pirsch-analytics/pirsch/v5/model"
)

// Demographics aggregates metadata statistics like the referrer, browser, and OS.
type Demographics struct {
	analyzer *Analyzer
	store    db.Store
}

// Languages returns the visitor count grouped by language.
func (demographics *Demographics) Languages(filter *Filter) ([]model.LanguageStats, error) {
	q, args := demographics.analyzer.selectByAttribute(filter, FieldLanguage)
	return demographics.store.SelectLanguageStats(q, args...)
}

// Countries returns the visitor count grouped by country.
func (demographics *Demographics) Countries(filter *Filter) ([]model.CountryStats, error) {
	q, args := demographics.analyzer.selectByAttribute(filter, FieldCountry)
	return demographics.store.SelectCountryStats(q, args...)
}

// Cities returns the visitor count grouped by city.
func (demographics *Demographics) Cities(filter *Filter) ([]model.CityStats, error) {
	q, args := demographics.analyzer.selectByAttribute(filter, FieldCity, FieldCountryCity)
	return demographics.store.SelectCityStats(q, args...)
}
