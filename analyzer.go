package pirsch

import "time"

// Filter is used to specify the time frame for the Analyzer.
type Filter struct {
	From time.Time
	To   time.Time
}

type PageVisits struct {
	Path   string
	Visits []VisitorsPerDay
}

// Analyzer provides an interface to analyze processed data and hits.
type Analyzer struct {
	store Store
}

// NewAnalyzer returns a new Analyzer for given Store.
func NewAnalyzer(store Store) *Analyzer {
	return &Analyzer{store}
}

// Visitors returns the visitors per day for the given time frame.
func (analyzer *Analyzer) Visitors(filter *Filter) ([]VisitorsPerDay, error) {
	filter = analyzer.validateFilter(filter)
	visitors, err := analyzer.store.Visitors(filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	today := analyzer.today()

	if today.Equal(filter.To) {
		visitorsToday, err := analyzer.store.VisitorsPerDay(today)

		if err != nil {
			return nil, err
		}

		if len(visitors) > 0 {
			visitors[len(visitors)-1].Visitors = visitorsToday
		} else {
			visitors = append(visitors, VisitorsPerDay{Day: today, Visitors: visitorsToday})
		}
	}

	return visitors, nil
}

func (analyzer *Analyzer) PageVisits(filter *Filter) ([]PageVisits, error) {
	filter = analyzer.validateFilter(filter)
	paths, err := analyzer.store.Paths(filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	pageVisits := make([]PageVisits, len(paths))

	for i, path := range paths {
		visitors, err := analyzer.store.PageVisits(path, filter.From, filter.To)

		if err != nil {
			return nil, err
		}

		pageVisits[i].Path = path
		pageVisits[i].Visits = visitors
	}

	return pageVisits, nil
}

func (analyzer *Analyzer) validateFilter(filter *Filter) *Filter {
	today := analyzer.today()

	if filter == nil {
		return &Filter{
			From: today.Add(-time.Hour * 24 * 6), // 7 including today
			To:   today,
		}
	}

	filter.From = time.Date(filter.From.Year(), filter.From.Month(), filter.From.Day(), 0, 0, 0, 0, time.UTC)
	filter.To = time.Date(filter.To.Year(), filter.To.Month(), filter.To.Day(), 0, 0, 0, 0, time.UTC)
	return filter
}

func (analyzer *Analyzer) today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}
