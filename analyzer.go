package pirsch

import (
	"time"
)

// PathVisitors assigns a path to visitor statistics per day.
type PathVisitors struct {
	Path  string
	Stats []Stats
}

// Analyzer provides an interface to analyze processed data and hits.
type Analyzer struct {
	store Store
}

// NewAnalyzer returns a new Analyzer for given Store.
func NewAnalyzer(store Store) *Analyzer {
	return &Analyzer{store}
}

// ActiveVisitors returns the active visitors per path and the total number of active visitors for given duration.
// Use time.Minute*5 for example to see the active visitors for the past 5 minutes.
func (analyzer *Analyzer) ActiveVisitors(filter *Filter, duration time.Duration) ([]Stats, int, error) {
	filter = analyzer.getFilter(filter)
	visitors, err := analyzer.store.ActiveVisitors(filter.TenantID, filter.Path, time.Now().UTC().Add(-duration))

	if err != nil {
		return nil, 0, err
	}

	sum := 0

	for _, v := range visitors {
		sum += v.Visitors
	}

	return visitors, sum, nil
}

// Visitors returns the visitors per day for the given time frame and path.
func (analyzer *Analyzer) Visitors(filter *Filter) ([]PathVisitors, error) {
	filter = analyzer.getFilter(filter)
	paths := analyzer.getPaths(filter)
	today := today()
	addToday := today.Equal(filter.To)
	stats := make([]PathVisitors, 0, len(paths))

	for _, path := range paths {
		visitors, err := analyzer.store.Visitors(filter.TenantID, path, filter.From, filter.To)

		if err != nil {
			return nil, err
		}

		if addToday {
			visitorsToday, err := analyzer.store.CountVisitorsByPath(nil, filter.TenantID, today, path, false)

			if err != nil {
				return nil, err
			}

			if len(visitorsToday) > 0 {
				visitors[len(visitors)-1].Visitors += visitorsToday[0].Visitors
			}
		}

		stats = append(stats, PathVisitors{
			Path:  path,
			Stats: visitors,
		})
	}

	return stats, nil
}

// getFilter validates and returns the given filter or a default filter if it is nil.
func (analyzer *Analyzer) getFilter(filter *Filter) *Filter {
	if filter == nil {
		return NewFilter(NullTenant)
	}

	filter.validate()
	return filter
}

// getPaths returns the paths to filter for. This can either be the one passed in,
// or all relevant paths for the given time frame otherwise.
func (analyzer *Analyzer) getPaths(filter *Filter) []string {
	if filter.Path != "" {
		return []string{filter.Path}
	}

	paths, err := analyzer.store.Paths(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return []string{}
	}

	return paths
}

/*
// Visitors returns the visitors per day for the given time frame.
func (analyzer *Analyzer) Visitors(filter *Filter) ([]VisitorsPerDay, error) {
	filter = analyzer.getFilter(filter)
	visitors, err := analyzer.store.Visitors(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	today := today()

	if today.Equal(filter.To) {
		visitorsToday, err := analyzer.store.CountVisitorsPerDay(nil, filter.TenantID, today)

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
func (analyzer *Analyzer) PageVisits(filter *Filter) ([]Stats, error) {
	// clean up filter and select all paths
	filter = analyzer.getFilter(filter)
	paths, err := analyzer.store.HitPaths(filter.TenantID, filter.From, filter.To)

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
	today := today()

	if today.Equal(filter.To) {
		pageVisitsToday, err := analyzer.store.CountVisitorsPerPage(nil, filter.TenantID, today)

		if err != nil {
			return nil, err
		}

		for _, visitToday := range pageVisitsToday {
			// find the path we can set the visitor count for, ...
			found := false

			for _, visit := range pageVisits {
				if visitToday.Path.String == visit.Path.String {
					visit.VisitorsPerDay[len(visit.VisitorsPerDay)-1].Visitors = visitToday.Visitors
					found = true
					break
				}
			}

			// ... or else add the path
			if !found {
				visits := make([]VisitorsPerDay, filter.HitDays()+1)

				for i := range visits {
					visits[i].Day = filter.From.Add(time.Hour * 24 * time.Duration(i))
				}

				visits[len(visits)-1].Visitors = visitToday.Visitors
				pageVisits = append(pageVisits, Stats{
					Path:           visitToday.Path,
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
	filter = analyzer.getFilter(filter)
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
	today := today()

	if today.Equal(filter.To) {
		referrerVisitsToday, err := analyzer.store.CountVisitorsPerReferrer(nil, filter.TenantID, today)

		if err != nil {
			return nil, err
		}

		for _, visitToday := range referrerVisitsToday {
			// find the referrer we can set the visitor count for, ...
			found := false

			for _, visit := range referrerVisits {
				if visitToday.Referrer.String == visit.Referrer.String {
					visit.VisitorsPerReferrer[len(visit.VisitorsPerReferrer)-1].Visitors = visitToday.Visitors
					found = true
					break
				}
			}

			// ... or else add the referrer
			if !found {
				visits := make([]VisitorsPerReferrer, filter.HitDays()+1)

				for i := range visits {
					visits[i].Day = filter.From.Add(time.Hour * 24 * time.Duration(i))
				}

				visits[len(visits)-1].Visitors = visitToday.Visitors
				referrerVisits = append(referrerVisits, Stats{
					Referrer:            visitToday.Referrer,
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
	filter = analyzer.getFilter(filter)
	pages, err := analyzer.store.VisitorPages(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	return pages, nil
}

// Languages returns the absolute and relative visitor count per language for given time frame.
func (analyzer *Analyzer) Languages(filter *Filter) ([]Stats, int, error) {
	filter = analyzer.getFilter(filter)
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
	filter = analyzer.getFilter(filter)
	referrer, err := analyzer.store.VisitorReferrer(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	return referrer, nil
}

// OS returns the absolute visitor count per operating system for given time frame.
func (analyzer *Analyzer) OS(filter *Filter) ([]Stats, error) {
	filter = analyzer.getFilter(filter)
	os, err := analyzer.store.VisitorOS(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	sum := 0

	for _, o := range os {
		sum += o.Visitors
	}

	for i := range os {
		os[i].RelativeVisitors = float64(os[i].Visitors) / float64(sum)
	}

	return os, nil
}

// Browser returns the absolute visitor count per browser for given time frame.
func (analyzer *Analyzer) Browser(filter *Filter) ([]Stats, error) {
	filter = analyzer.getFilter(filter)
	browser, err := analyzer.store.VisitorBrowser(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	sum := 0

	for _, b := range browser {
		sum += b.Visitors
	}

	for i := range browser {
		browser[i].RelativeVisitors = float64(browser[i].Visitors) / float64(sum)
	}

	return browser, nil
}

// Platform returns the relative platform usage for given time frame.
func (analyzer *Analyzer) Platform(filter *Filter) (*Stats, error) {
	filter = analyzer.getFilter(filter)
	platform, err := analyzer.store.VisitorPlatform(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	sum := float64(platform.PlatformDesktopVisitors + platform.PlatformMobileVisitors + platform.PlatformUnknownVisitors)
	platform.PlatformDesktopRelative = float64(platform.PlatformDesktopVisitors) / sum
	platform.PlatformMobileRelative = float64(platform.PlatformMobileVisitors) / sum
	platform.PlatformUnknownRelative = float64(platform.PlatformUnknownVisitors) / sum
	return platform, nil
}

// HourlyVisitors returns the absolute and relative visitor count per language for given time frame.
func (analyzer *Analyzer) HourlyVisitors(filter *Filter) ([]Stats, error) {
	filter = analyzer.getFilter(filter)
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
*/
