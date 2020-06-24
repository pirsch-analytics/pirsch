package pirsch

import "time"

const (
	week = time.Hour * 24 * 7
)

// Filter is used to specify the time frame for the Analyzer.
type Filter struct {
	From time.Time
	To   time.Time
}

// Analyzer provides an interface to analyze processed data and hits.
type Analyzer struct {
	store Store
}

// Visitors returns the visitors per day for the given time frame.
func (analyzer *Analyzer) Visitors(filter *Filter) ([]VisitorsPerDay, error) {
	filter = analyzer.validateFilter(filter)
	analyzer.removeTime(filter)
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

func (analyzer *Analyzer) validateFilter(filter *Filter) *Filter {
	now := time.Now()

	if filter == nil {
		return &Filter{
			From: now.Add(-week),
			To:   now,
		}
	}

	if filter.To.After(now) {
		filter.To = now
	}

	if filter.From.After(filter.To) {
		filter.From = filter.To.Add(-week)
	}

	return filter
}

func (analyzer *Analyzer) removeTime(filter *Filter) {
	filter.From = time.Date(filter.From.Year(), filter.From.Month(), filter.From.Day(), 0, 0, 0, 0, filter.From.Location())
	filter.To = time.Date(filter.To.Year(), filter.To.Month(), filter.To.Day(), 0, 0, 0, 0, filter.To.Location())
}

func (analyzer *Analyzer) today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
