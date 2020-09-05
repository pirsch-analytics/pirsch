package pirsch

import (
	"sort"
	"time"
)

// PathVisitors assigns a path to visitor statistics per day.
type PathVisitors struct {
	Path  string
	Stats []Stats
}

// PathReferrer assigns a path to visitor statistics per day and referrer.
type PathReferrer struct {
	Path  string
	Stats []ReferrerStats
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
	stats, err := analyzer.store.ActiveVisitors(filter.TenantID, filter.Path, time.Now().UTC().Add(-duration))

	if err != nil {
		return nil, 0, err
	}

	sum := 0

	for _, v := range stats {
		sum += v.Visitors
	}

	return stats, sum, nil
}

// Visitors returns the visitor count per day.
func (analyzer *Analyzer) Visitors(filter *Filter) ([]Stats, error) {
	filter = analyzer.getFilter(filter)
	today := today()
	addToday := today.Equal(filter.To)
	stats, err := analyzer.store.Visitors(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	if addToday && len(stats) > 0 {
		visitorsToday := analyzer.store.CountVisitors(nil, filter.TenantID, today)

		if visitorsToday != nil {
			stats[len(stats)-1].Visitors += visitorsToday.Visitors
		}
	}

	return stats, nil
}

// Languages returns the visitor count per language.
func (analyzer *Analyzer) Languages(filter *Filter) ([]LanguageStats, error) {
	filter = analyzer.getFilter(filter)
	today := today()
	addToday := today.Equal(filter.To)
	stats, err := analyzer.store.VisitorLanguages(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	if addToday && len(stats) > 0 {
		visitorsToday, err := analyzer.store.CountVisitorsByLanguage(nil, filter.TenantID, today)

		if err != nil {
			return nil, err
		}

		for _, v := range visitorsToday {
			found := false

			for i, s := range stats {
				if s.Language.String == v.Language.String {
					stats[i].Visitors += v.Visitors
					found = true
					break
				}
			}

			if !found {
				stats = append(stats, v)
			}
		}
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Visitors > stats[j].Visitors
	})

	var sum float64

	for i := range stats {
		sum += float64(stats[i].Visitors)
	}

	for i := range stats {
		stats[i].RelativeVisitors = float64(stats[i].Visitors) / sum
	}

	return stats, nil
}

// Referrer returns the visitor count per referrer.
func (analyzer *Analyzer) Referrer(filter *Filter) ([]ReferrerStats, error) {
	filter = analyzer.getFilter(filter)
	today := today()
	addToday := today.Equal(filter.To)
	stats, err := analyzer.store.VisitorReferrer(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	if addToday && len(stats) > 0 {
		visitorsToday, err := analyzer.store.CountVisitorsByReferrer(nil, filter.TenantID, today)

		if err != nil {
			return nil, err
		}

		for _, v := range visitorsToday {
			found := false

			for i, s := range stats {
				if s.Referrer.String == v.Referrer.String {
					stats[i].Visitors += v.Visitors
					found = true
					break
				}
			}

			if !found {
				stats = append(stats, v)
			}
		}
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Visitors > stats[j].Visitors
	})

	var sum float64

	for i := range stats {
		sum += float64(stats[i].Visitors)
	}

	for i := range stats {
		stats[i].RelativeVisitors = float64(stats[i].Visitors) / sum
	}

	return stats, nil
}

// OS returns the visitor count per operating system.
func (analyzer *Analyzer) OS(filter *Filter) ([]OSStats, error) {
	filter = analyzer.getFilter(filter)
	today := today()
	addToday := today.Equal(filter.To)
	stats, err := analyzer.store.VisitorOS(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	if addToday && len(stats) > 0 {
		visitorsToday, err := analyzer.store.CountVisitorsByOS(nil, filter.TenantID, today)

		if err != nil {
			return nil, err
		}

		for _, v := range visitorsToday {
			found := false

			for i, s := range stats {
				if s.OS.String == v.OS.String {
					stats[i].Visitors += v.Visitors
					found = true
					break
				}
			}

			if !found {
				stats = append(stats, v)
			}
		}
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Visitors > stats[j].Visitors
	})

	var sum float64

	for i := range stats {
		sum += float64(stats[i].Visitors)
	}

	for i := range stats {
		stats[i].RelativeVisitors = float64(stats[i].Visitors) / sum
	}

	return stats, nil
}

// Browser returns the visitor count per browser.
func (analyzer *Analyzer) Browser(filter *Filter) ([]BrowserStats, error) {
	filter = analyzer.getFilter(filter)
	today := today()
	addToday := today.Equal(filter.To)
	stats, err := analyzer.store.VisitorBrowser(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	if addToday && len(stats) > 0 {
		visitorsToday, err := analyzer.store.CountVisitorsByBrowser(nil, filter.TenantID, today)

		if err != nil {
			return nil, err
		}

		for _, v := range visitorsToday {
			found := false

			for i, s := range stats {
				if s.Browser.String == v.Browser.String {
					stats[i].Visitors += v.Visitors
					found = true
					break
				}
			}

			if !found {
				stats = append(stats, v)
			}
		}
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Visitors > stats[j].Visitors
	})

	var sum float64

	for i := range stats {
		sum += float64(stats[i].Visitors)
	}

	for i := range stats {
		stats[i].RelativeVisitors = float64(stats[i].Visitors) / sum
	}

	return stats, nil
}

// Platform returns the visitor count per browser.
func (analyzer *Analyzer) Platform(filter *Filter) *VisitorStats {
	filter = analyzer.getFilter(filter)
	today := today()
	addToday := today.Equal(filter.To)
	stats := analyzer.store.VisitorPlatform(filter.TenantID, filter.From, filter.To)

	if stats == nil {
		stats = &VisitorStats{}
	}

	if addToday {
		visitorsToday := analyzer.store.CountVisitorsByPlatform(nil, filter.TenantID, today)

		if visitorsToday != nil {
			stats.PlatformDesktop += visitorsToday.PlatformDesktop
			stats.PlatformMobile += visitorsToday.PlatformMobile
			stats.PlatformUnknown += visitorsToday.PlatformUnknown
		}
	}

	sum := float64(stats.PlatformDesktop + stats.PlatformMobile + stats.PlatformUnknown)
	stats.RelativePlatformDesktop = float64(stats.PlatformDesktop) / sum
	stats.RelativePlatformMobile = float64(stats.PlatformMobile) / sum
	stats.RelativePlatformUnknown = float64(stats.PlatformUnknown) / sum
	return stats
}

// PageVisitors returns the visitors per day for the given time frame grouped by path.
func (analyzer *Analyzer) PageVisitors(filter *Filter) ([]PathVisitors, error) {
	filter = analyzer.getFilter(filter)
	paths := analyzer.getPaths(filter)
	today := today()
	addToday := today.Equal(filter.To)
	stats := make([]PathVisitors, 0, len(paths))

	for _, path := range paths {
		visitors, err := analyzer.store.PageVisitors(filter.TenantID, path, filter.From, filter.To)

		if err != nil {
			return nil, err
		}

		if addToday && len(visitors) > 0 {
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

// Referrer returns the visitor count per referrer, day, and for the given time frame and path.
/*func (analyzer *Analyzer) Referrer(filter *Filter) ([]PathReferrer, error) {
	filter = analyzer.getFilter(filter)
	paths := analyzer.getPaths(filter)
	today := today()
	addToday := today.Equal(filter.To)
	stats := make([]PathReferrer, 0, len(paths))

	for _, path := range paths {
		visitors, err := analyzer.store.Referrer(filter.TenantID, path, filter.From, filter.To)

		if err != nil {
			return nil, err
		}

		if addToday && len(visitors) > 0 {
			visitorsToday, err := analyzer.store.CountVisitorsByPathAndReferrer(nil, filter.TenantID, today, path)

			if err != nil {
				return nil, err
			}

			if len(visitorsToday) > 0 {
				visitors[len(visitors)-1].Visitors += visitorsToday[0].Visitors
			}
		}

		stats = append(stats, PathReferrer{
			Path:  path,
			Stats: visitors,
		})
	}

	return stats, nil
}*/

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
