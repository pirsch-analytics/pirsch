package pirsch

import "time"

// Analyzer provides an interface to analyze statistics.
type Analyzer struct {
	store Store
}

// NewAnalyzer returns a new Analyzer for given Store.
func NewAnalyzer(store Store) *Analyzer {
	return &Analyzer{
		store,
	}
}

// ActiveVisitors returns the active visitors per path and the total number of active visitors for given duration.
// Use time.Minute*5 for example to see the active visitors for the past 5 minutes.
// The correct date/time is not included.
func (analyzer *Analyzer) ActiveVisitors(filter *Filter, duration time.Duration) ([]Stats, int, error) {
	filter = analyzer.getFilter(filter)
	filter.From = time.Now().UTC().Add(-duration)
	visitors, err := analyzer.store.Select(NewQuery(filter).
		Fields(`count(DISTINCT fingerprint) "visitors"`).
		Group(FieldPath).
		Order(QueryOrder{Field: FieldVisitors, Direction: DESC}, QueryOrder{Field: FieldPath, Direction: ASC}))

	if err != nil {
		return nil, 0, err
	}

	count, err := analyzer.store.Count(NewQuery(filter).Fields(`count(DISTINCT fingerprint) "visitors"`))

	if err != nil {
		return nil, 0, err
	}

	return visitors, count, nil
}

// Languages returns the visitor count per language.
func (analyzer *Analyzer) Languages(filter *Filter) ([]Stats, error) {
	return analyzer.store.Select(NewQuery(analyzer.getFilter(filter)).
		Fields(`count(DISTINCT fingerprint) "visitors"`).
		Group(FieldLanguage).
		Order(QueryOrder{Field: FieldVisitors, Direction: DESC}, QueryOrder{Field: FieldLanguage, Direction: ASC}))
}

// Countries returns the visitor count per country.
func (analyzer *Analyzer) Countries(filter *Filter) ([]Stats, error) {
	return analyzer.store.Select(NewQuery(analyzer.getFilter(filter)).
		Fields(`count(DISTINCT fingerprint) "visitors"`).
		Group(FieldCountryCode).
		Order(QueryOrder{Field: FieldVisitors, Direction: DESC}, QueryOrder{Field: FieldCountryCode, Direction: ASC}))
}

// Browser returns the visitor count per browser.
func (analyzer *Analyzer) Browser(filter *Filter) ([]Stats, error) {
	return analyzer.store.Select(NewQuery(analyzer.getFilter(filter)).
		Fields(`count(DISTINCT fingerprint) "visitors"`).
		Group(FieldBrowser).
		Order(QueryOrder{Field: FieldVisitors, Direction: DESC}, QueryOrder{Field: FieldBrowser, Direction: ASC}))
}

// OS returns the visitor count per operating system.
func (analyzer *Analyzer) OS(filter *Filter) ([]Stats, error) {
	return analyzer.store.Select(NewQuery(analyzer.getFilter(filter)).
		Fields(`count(DISTINCT fingerprint) "visitors"`).
		Group(FieldOS).
		Order(QueryOrder{Field: FieldVisitors, Direction: DESC}, QueryOrder{Field: FieldOS, Direction: ASC}))
}

func (analyzer *Analyzer) getFilter(filter *Filter) *Filter {
	if filter == nil {
		return NewFilter(NullTenant)
	}

	filter.validate()
	return filter
}
