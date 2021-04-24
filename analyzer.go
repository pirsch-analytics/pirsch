package pirsch

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
/*func (analyzer *Analyzer) ActiveVisitors(filter *Run, duration time.Duration) ([]Stats, int, error) {
	filter = analyzer.getFilter(filter)
	filter.From = time.Now().UTC().Add(-duration)
	visitors, err := analyzer.store.ActiveVisitors(filter)

	if err != nil {
		return nil, 0, err
	}

	return visitors, analyzer.store.CountActiveVisitors(filter), nil
}*/

// Languages returns the visitor count per language.
func (analyzer *Analyzer) Languages(filter *Filter) ([]Stats, error) {
	return analyzer.store.Run(NewQuery(analyzer.getFilter(filter)).
		Group(FieldLanguage).
		Order(QueryOrder{Field: FieldVisitors, Direction: DESC}, QueryOrder{Field: FieldLanguage, Direction: ASC}))
}

func (analyzer *Analyzer) getFilter(filter *Filter) *Filter {
	if filter == nil {
		return NewFilter(NullTenant)
	}

	filter.validate()
	return filter
}
