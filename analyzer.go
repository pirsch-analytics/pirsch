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

// Stats returns the visitors per page per day for given time frame.
func (analyzer *Analyzer) PageVisits(filter *Filter) ([]Stats, error) {
	// clean up filter and select all paths
	filter = analyzer.validateFilter(filter)
	paths, err := analyzer.store.Paths(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	// select all processed path views and store them
	pageVisits := make([]Stats, len(paths))

	for i, path := range paths {
		visitors, err := analyzer.store.PageVisits(filter.TenantID, path, filter.From, filter.To)

		if err != nil {
			return nil, err
		}

		pageVisits[i].Path = sql.NullString{String: path, Valid: true}
		pageVisits[i].VisitorsPerDay = visitors
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
				if visitToday.Path == visit.Path.String {
					visit.VisitorsPerDay[len(visit.VisitorsPerDay)-1].Visitors = visitToday.Visitors
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
				pageVisits = append(pageVisits, Stats{
					Path:           sql.NullString{String: visitToday.Path, Valid: true},
					VisitorsPerDay: visits,
				})
			}
		}

		// sort paths by length and alphabetically
		sort.Slice(pageVisits, func(i, j int) bool {
			return len(pageVisits[i].Path.String) < len(pageVisits[j].Path.String) ||
				pageVisits[i].Path.String < pageVisits[j].Path.String
		})
	}

	return pageVisits, nil
}

// ReferrerVisits returns the visitors per referrer per day for given time frame.
func (analyzer *Analyzer) ReferrerVisits(filter *Filter) ([]Stats, error) {
	// clean up filter and select all referrer
	filter = analyzer.validateFilter(filter)
	referrer, err := analyzer.store.Referrer(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	// select all processed referrer views and store them
	referrerVisits := make([]Stats, len(referrer))

	for i, ref := range referrer {
		visitors, err := analyzer.store.ReferrerVisits(filter.TenantID, ref, filter.From, filter.To)

		if err != nil {
			return nil, err
		}

		referrerVisits[i].Referrer = sql.NullString{String: ref, Valid: true}
		referrerVisits[i].VisitorsPerReferrer = visitors
	}

	// add hits to list of referrer views in case the filter includes today
	today := analyzer.today()

	if today.Equal(filter.To) {
		referrerVisitsToday, err := analyzer.store.CountVisitorsPerReferrer(filter.TenantID, today)

		if err != nil {
			return nil, err
		}

		for _, visitToday := range referrerVisitsToday {
			// find the referrer we can set the visitor count for, ...
			found := false

			for _, visit := range referrerVisits {
				if visitToday.Ref == visit.Referrer.String {
					visit.VisitorsPerReferrer[len(visit.VisitorsPerReferrer)-1].Visitors = visitToday.Visitors
					found = true
					break
				}
			}

			// ... or else add the referrer
			if !found {
				visits := make([]VisitorsPerReferrer, filter.Days()+1)

				for i := range visits {
					visits[i].Day = filter.From.Add(time.Hour * 24 * time.Duration(i))
				}

				visits[len(visits)-1].Visitors = visitToday.Visitors
				referrerVisits = append(referrerVisits, Stats{
					Referrer:            sql.NullString{String: visitToday.Ref, Valid: true},
					VisitorsPerReferrer: visits,
				})
			}
		}

		// sort referrer by length and alphabetically
		sort.Slice(referrerVisits, func(i, j int) bool {
			return len(referrerVisits[i].Referrer.String) < len(referrerVisits[j].Referrer.String) ||
				referrerVisits[i].Referrer.String < referrerVisits[j].Referrer.String
		})
	}

	return referrerVisits, nil
}

// Pages returns the absolute visitor count per page for given time frame.
func (analyzer *Analyzer) Pages(filter *Filter) ([]Stats, error) {
	filter = analyzer.validateFilter(filter)
	pages, err := analyzer.store.VisitorPages(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	return pages, nil
}

// Languages returns the absolute and relative visitor count per language for given time frame.
func (analyzer *Analyzer) Languages(filter *Filter) ([]Stats, int, error) {
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

// Referrer returns the absolute visitor count per referrer for given time frame.
func (analyzer *Analyzer) Referrer(filter *Filter) ([]Stats, error) {
	filter = analyzer.validateFilter(filter)
	referrer, err := analyzer.store.VisitorReferrer(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	return referrer, nil
}

// OS returns the absolute visitor count per operating system for given time frame.
func (analyzer *Analyzer) OS(filter *Filter) ([]Stats, error) {
	filter = analyzer.validateFilter(filter)
	os, err := analyzer.store.VisitorOS(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	return os, nil
}

// Browser returns the absolute visitor count per browser for given time frame.
func (analyzer *Analyzer) Browser(filter *Filter) ([]Stats, error) {
	filter = analyzer.validateFilter(filter)
	browser, err := analyzer.store.VisitorBrowser(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	return browser, nil
}

// HourlyVisitors returns the absolute and relative visitor count per language for given time frame.
func (analyzer *Analyzer) HourlyVisitors(filter *Filter) ([]Stats, error) {
	filter = analyzer.validateFilter(filter)
	visitors, err := analyzer.store.HourlyVisitors(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	visitorsPerHour := make([]Stats, 24)

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
func (analyzer *Analyzer) ActiveVisitorsPages(tenantID sql.NullInt64, d time.Duration) ([]Stats, error) {
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
