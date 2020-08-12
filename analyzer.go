package pirsch

import (
	"database/sql"
	"sort"
	"time"
)

// Filter is used to specify the time frame and tenant for the Analyzer.
type Filter struct {
	// TenantID is the optional tenant ID used to filter results.
	TenantID sql.NullInt64

	// From is the start of the selection.
	From time.Time

	// To is the end of the selection.
	To time.Time
}

// Days returns the number of days covered by the filter.
func (filter *Filter) Days() int {
	return int(filter.To.Sub(filter.From).Hours()) / 24
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
	visitors, err := analyzer.store.Visitors(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	today := analyzer.today()

	if today.Equal(filter.To) {
		visitorsToday, err := analyzer.store.CountVisitorsPerDay(filter.TenantID, today)

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

// PageVisits returns the visitors per page per day for given time frame.
func (analyzer *Analyzer) PageVisits(filter *Filter) ([]PageVisits, error) {
	// clean up filter and select all paths
	filter = analyzer.validateFilter(filter)
	paths, err := analyzer.store.Paths(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	// select all processed path views and store them
	pageVisits := make([]PageVisits, len(paths))

	for i, path := range paths {
		visitors, err := analyzer.store.PageVisits(filter.TenantID, path, filter.From, filter.To)

		if err != nil {
			return nil, err
		}

		pageVisits[i].Path = path
		pageVisits[i].Visits = visitors
	}

	// add hits to list of path views in case the filter includes today
	today := analyzer.today()

	if today.Equal(filter.To) {
		pageVisitsToday, err := analyzer.store.CountVisitorsPerPage(filter.TenantID, today)

		if err != nil {
			return nil, err
		}

		for _, visitToday := range pageVisitsToday {
			// find the path we can set the visitor count for, ...
			found := false

			for _, visit := range pageVisits {
				if visitToday.Path == visit.Path {
					visit.Visits[len(visit.Visits)-1].Visitors = visitToday.Visitors
					found = true
					break
				}
			}

			// ... or else add the path
			if !found {
				visits := make([]VisitorsPerDay, filter.Days()+1)

				for i := range visits {
					visits[i].Day = filter.From.Add(time.Hour * 24 * time.Duration(i))
				}

				visits[len(visits)-1].Visitors = visitToday.Visitors
				pageVisits = append(pageVisits, PageVisits{
					Path:   visitToday.Path,
					Visits: visits,
				})
			}
		}

		// sort paths by length and alphabetically
		sort.Slice(pageVisits, func(i, j int) bool {
			return len(pageVisits[i].Path) < len(pageVisits[j].Path) ||
				pageVisits[i].Path < pageVisits[j].Path
		})
	}

	return pageVisits, nil
}

// RefererVisits returns the visitors per referer per day for given time frame.
func (analyzer *Analyzer) RefererVisits(filter *Filter) ([]RefererVisits, error) {
	// clean up filter and select all referer
	filter = analyzer.validateFilter(filter)
	referer, err := analyzer.store.Referer(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	// select all processed referer views and store them
	refererVisits := make([]RefererVisits, len(referer))

	for i, ref := range referer {
		visitors, err := analyzer.store.RefererVisits(filter.TenantID, ref, filter.From, filter.To)

		if err != nil {
			return nil, err
		}

		refererVisits[i].Referer = ref
		refererVisits[i].Visits = visitors
	}

	// add hits to list of referer views in case the filter includes today
	today := analyzer.today()

	if today.Equal(filter.To) {
		refererVisitsToday, err := analyzer.store.CountVisitorsPerReferer(filter.TenantID, today)

		if err != nil {
			return nil, err
		}

		for _, visitToday := range refererVisitsToday {
			// find the referer we can set the visitor count for, ...
			found := false

			for _, visit := range refererVisits {
				if visitToday.Ref == visit.Referer {
					visit.Visits[len(visit.Visits)-1].Visitors = visitToday.Visitors
					found = true
					break
				}
			}

			// ... or else add the referer
			if !found {
				visits := make([]VisitorsPerReferer, filter.Days()+1)

				for i := range visits {
					visits[i].Day = filter.From.Add(time.Hour * 24 * time.Duration(i))
				}

				visits[len(visits)-1].Visitors = visitToday.Visitors
				refererVisits = append(refererVisits, RefererVisits{
					Referer: visitToday.Ref,
					Visits:  visits,
				})
			}
		}

		// sort referer by length and alphabetically
		sort.Slice(refererVisits, func(i, j int) bool {
			return len(refererVisits[i].Referer) < len(refererVisits[j].Referer) ||
				refererVisits[i].Referer < refererVisits[j].Referer
		})
	}

	return refererVisits, nil
	/*filter = analyzer.validateFilter(filter)
	referer, err := analyzer.store.Referer(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	refererVisits := make([]RefererVisits, len(referer))

	for i, ref := range referer {
		visitors, err := analyzer.store.RefererVisits(filter.TenantID, ref, filter.From, filter.To)

		if err != nil {
			return nil, err
		}

		refererVisits[i].Referer = ref
		refererVisits[i].Visits = visitors
	}

	today := analyzer.today()

	if today.Equal(filter.To) {
		refererVisitsToday, err := analyzer.store.CountVisitorsPerReferer(filter.TenantID, today)

		if err != nil {
			return nil, err
		}

		for _, visitToday := range refererVisitsToday {
			found := false

			for _, visit := range refererVisits {
				if visitToday.Ref == visit.Referer {
					visit.Visits[len(visit.Visits)-1].Visitors = visitToday.Visitors
					found = true
					break
				}
			}

			if !found {
				visits := make([]VisitorsPerReferer, filter.Days()+1)

				for i := range visits {
					visits[i].Day = filter.From.Add(time.Hour * 24 * time.Duration(i))
				}

				visits[len(visits)-1].Visitors = visitToday.Visitors
				refererVisits = append(refererVisits, RefererVisits{
					Referer: visitToday.Ref,
					Visits:  visits,
				})
			}
		}
	}

	return refererVisits, nil*/
}

// Languages returns the absolute and relative visitor count per language for given time frame.
func (analyzer *Analyzer) Languages(filter *Filter) ([]VisitorLanguage, int, error) {
	filter = analyzer.validateFilter(filter)
	langs, err := analyzer.store.VisitorLanguages(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, 0, err
	}

	total := 0

	for _, lang := range langs {
		total += lang.Visitors
	}

	for i := range langs {
		langs[i].RelativeVisitors = float64(langs[i].Visitors) / float64(total)
	}

	return langs, total, nil
}

// HourlyVisitors returns the absolute and relative visitor count per language for given time frame.
func (analyzer *Analyzer) HourlyVisitors(filter *Filter) ([]HourlyVisitors, error) {
	filter = analyzer.validateFilter(filter)
	visitors, err := analyzer.store.HourlyVisitors(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	visitorsPerHour := make([]HourlyVisitors, 24)

	for i := range visitorsPerHour {
		visitorsPerHour[i].Hour = i
	}

	for _, visitor := range visitors {
		visitorsPerHour[visitor.Hour].Visitors = visitor.Visitors
	}

	return visitorsPerHour, nil
}

// ActiveVisitors returns the number of unique visitors active within the given duration.
func (analyzer *Analyzer) ActiveVisitors(tenantID sql.NullInt64, d time.Duration) (int, error) {
	visitors, err := analyzer.store.ActiveVisitors(tenantID, time.Now().UTC().Add(-d))

	if err != nil {
		return 0, err
	}

	return visitors, nil
}

// ActiveVisitorsPages returns the number of unique visitors active within the given duration and the corresponding pages.
func (analyzer *Analyzer) ActiveVisitorsPages(tenantID sql.NullInt64, d time.Duration) ([]PageVisitors, error) {
	visitors, err := analyzer.store.ActiveVisitorsPerPage(tenantID, time.Now().UTC().Add(-d))

	if err != nil {
		return nil, err
	}

	return visitors, nil
}

func (analyzer *Analyzer) validateFilter(filter *Filter) *Filter {
	today := analyzer.today()

	if filter == nil {
		return &Filter{
			From: today.Add(-time.Hour * 24 * 6), // 7 including today
			To:   today,
		}
	}

	if filter.From.IsZero() && filter.To.IsZero() {
		filter.From = today.Add(-time.Hour * 24 * 6) // 7 including today
		filter.To = today
	} else {
		filter.From = time.Date(filter.From.Year(), filter.From.Month(), filter.From.Day(), 0, 0, 0, 0, time.UTC)
		filter.To = time.Date(filter.To.Year(), filter.To.Month(), filter.To.Day(), 0, 0, 0, 0, time.UTC)
	}

	if filter.From.After(filter.To) {
		filter.From, filter.To = filter.To, filter.From
	}

	return filter
}

func (analyzer *Analyzer) today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}
