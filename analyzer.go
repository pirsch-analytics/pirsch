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
func (analyzer *Analyzer) Languages(filter *Query) ([]Stats, error) {
	filter = analyzer.getFilter(filter)
	filter.GroupBy = append(filter.GroupBy, Language)
	filter.OrderBy = append(filter.OrderBy, FilterOrder{Field: Visitors, Direction: DESC})
	filter.OrderBy = append(filter.OrderBy, FilterOrder{Field: Language, Direction: ASC})
	return analyzer.store.Run(filter)
}

func (analyzer *Analyzer) getFilter(filter *Query) *Query {
	if filter == nil {
		return NewQuery(NullTenant)
	}

	filter.validate()
	return filter
}
