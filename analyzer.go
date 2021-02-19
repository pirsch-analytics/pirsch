package pirsch

import (
	"sort"
	"time"
)

// PathVisitors represents visitor statistics per day for a path, including the total visitor count, relative visitor count and bounce rate.
type PathVisitors struct {
	Path             string  `json:"path"`
	Stats            []Stats `json:"stats"`
	Visitors         int     `json:"visitors"`
	Bounces          int     `json:"bounces"`
	Views            int     `json:"views"`
	RelativeVisitors float64 `json:"relative_visitors"`
	BounceRate       float64 `json:"bounce_rate"`
	RelativeViews    float64 `json:"relative_views"`
}

// TimeOfDayVisitors represents the visitor count per day and hour for a path.
type TimeOfDayVisitors struct {
	Day   time.Time          `json:"day"`
	Stats []VisitorTimeStats `json:"stats"`
}

// Growth represents the visitors, sessions, and bounces growth between two time periods.
type Growth struct {
	Current        *Stats  `json:"current"`
	Previous       *Stats  `json:"previous"`
	VisitorsGrowth float64 `json:"visitors_growth"`
	SessionsGrowth float64 `json:"sessions_growth"`
	BouncesGrowth  float64 `json:"bounces_growth"`
	ViewsGrowth    float64 `json:"views_growth"`
}

// AnalyzerConfig is the (optional) configuration for the Analyzer.
type AnalyzerConfig struct {
	// Timezone sets the time zone for the result set.
	// If not set, UTC will be used.
	Timezone *time.Location
}

func (config *AnalyzerConfig) validate() {
	if config.Timezone == nil {
		config.Timezone = time.UTC
	}
}

// Analyzer provides an interface to analyze processed data and hits.
type Analyzer struct {
	store    Store
	timezone *time.Location
}

// NewAnalyzer returns a new Analyzer for given Store.
func NewAnalyzer(store Store, config *AnalyzerConfig) *Analyzer {
	if config == nil {
		config = new(AnalyzerConfig)
	}

	config.validate()
	return &Analyzer{
		store:    store,
		timezone: config.Timezone,
	}
}

// ActiveVisitors returns the active visitors per path and the total number of active visitors for given duration.
// Use time.Minute*5 for example to see the active visitors for the past 5 minutes.
// The correct date/time is not included.
func (analyzer *Analyzer) ActiveVisitors(filter *Filter, duration time.Duration) ([]Stats, int, error) {
	filter = analyzer.getFilter(filter)
	from := time.Now().UTC().Add(-duration)
	stats, err := analyzer.store.ActivePageVisitors(filter.TenantID, from)

	if err != nil {
		return nil, 0, err
	}

	return stats, analyzer.store.ActiveVisitors(filter.TenantID, from), nil
}

// Visitors returns the visitor count, session count, bounce rate, and views per day.
func (analyzer *Analyzer) Visitors(filter *Filter) ([]Stats, error) {
	filter = analyzer.getFilter(filter)
	today := today()
	addToday := today.Equal(filter.To)
	stats, err := analyzer.store.Visitors(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	if addToday {
		visitorsToday := analyzer.store.CountVisitors(nil, filter.TenantID, today)
		bouncesToday := analyzer.store.CountVisitorsByPathAndMaxOneHit(nil, filter.TenantID, today, "")

		if len(stats) > 0 {
			if visitorsToday != nil {
				stats[len(stats)-1].Visitors += visitorsToday.Visitors
				stats[len(stats)-1].Sessions += visitorsToday.Sessions
				stats[len(stats)-1].Bounces += bouncesToday
				stats[len(stats)-1].Views += visitorsToday.Views
			}
		} else {
			stats = append(stats, Stats{
				Visitors: visitorsToday.Visitors,
				Sessions: visitorsToday.Sessions,
				Bounces:  visitorsToday.Bounces,
				Views:    visitorsToday.Views,
			})
		}
	}

	for i := range stats {
		if stats[i].Visitors > 0 {
			stats[i].BounceRate = float64(stats[i].Bounces) / float64(stats[i].Visitors)
		}
	}

	return stats, nil
}

// VisitorHours returns the visitor count grouped by hour of day for given time frame.
// Note that the sum of them is not the number of unique visitors for the day, as visitors can re-appear at different times on the same day!
func (analyzer *Analyzer) VisitorHours(filter *Filter) ([]VisitorTimeStats, error) {
	filter = analyzer.getFilter(filter)
	stats, err := analyzer.store.VisitorHours(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	for i := range stats {
		stats[i].Hour = hourInTimezone(stats[i].Hour, analyzer.timezone)
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Hour < stats[j].Hour
	})

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

	if addToday {
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

	if addToday {
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
		stats[i].BounceRate = float64(stats[i].Bounces) / float64(stats[i].Visitors)
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

	if addToday {
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

	if addToday {
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

	if sum != 0 {
		stats.RelativePlatformDesktop = float64(stats.PlatformDesktop) / sum
		stats.RelativePlatformMobile = float64(stats.PlatformMobile) / sum
		stats.RelativePlatformUnknown = float64(stats.PlatformUnknown) / sum
	}

	return stats
}

// Screen returns the visitor count per screen size (width and height).
func (analyzer *Analyzer) Screen(filter *Filter) ([]ScreenStats, error) {
	filter = analyzer.getFilter(filter)
	today := today()
	addToday := today.Equal(filter.To)
	stats, err := analyzer.store.VisitorScreenSize(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	if addToday {
		visitorsToday, err := analyzer.store.CountVisitorsByScreenSize(nil, filter.TenantID, today)

		if err != nil {
			return nil, err
		}

		for _, v := range visitorsToday {
			found := false

			for i, s := range stats {
				if s.Width == v.Width && s.Height == v.Height {
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

// ScreenClass returns the visitor count per screen class.
func (analyzer *Analyzer) ScreenClass(filter *Filter) ([]ScreenStats, error) {
	filter = analyzer.getFilter(filter)
	today := today()
	addToday := today.Equal(filter.To)
	stats, err := analyzer.store.VisitorScreenClass(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	if addToday {
		visitorsToday, err := analyzer.store.CountVisitorsByScreenClass(nil, filter.TenantID, today)

		if err != nil {
			return nil, err
		}

		for _, v := range visitorsToday {
			found := false

			for i, s := range stats {
				if s.Class.String == v.Class.String {
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

// Country returns the visitor count per country.
func (analyzer *Analyzer) Country(filter *Filter) ([]CountryStats, error) {
	filter = analyzer.getFilter(filter)
	today := today()
	addToday := today.Equal(filter.To)
	stats, err := analyzer.store.VisitorCountry(filter.TenantID, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	if addToday {
		visitorsToday, err := analyzer.store.CountVisitorsByCountryCode(nil, filter.TenantID, today)

		if err != nil {
			return nil, err
		}

		for _, v := range visitorsToday {
			found := false

			for i, s := range stats {
				if s.CountryCode == v.CountryCode {
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

// TimeOfDay returns the visitor count per day and hour for given time frame.
func (analyzer *Analyzer) TimeOfDay(filter *Filter) ([]TimeOfDayVisitors, error) {
	filter = analyzer.getFilter(filter)
	from := filter.From
	stats := make([]TimeOfDayVisitors, 0)

	for !from.After(filter.To) {
		s, err := analyzer.VisitorHours(&Filter{TenantID: filter.TenantID, From: from, To: from})

		if err != nil {
			return nil, err
		}

		// no need to set timezone for hours and sort stats, as this is done by VisitorHours
		stats = append(stats, TimeOfDayVisitors{
			Day:   from,
			Stats: s,
		})
		from = from.Add(time.Hour * 24)
	}

	return stats, nil
}

// PageVisitors returns the visitor count, session count, bounce rate, and views per day for the given time frame grouped by path.
func (analyzer *Analyzer) PageVisitors(filter *Filter) ([]PathVisitors, error) {
	filter = analyzer.getFilter(filter)
	paths := analyzer.getPaths(filter)
	today := today()
	addToday := today.Equal(filter.To)
	stats := make([]PathVisitors, 0, len(paths))
	var totalVisitors, totalViews int

	for _, path := range paths {
		visitors, err := analyzer.store.PageVisitors(filter.TenantID, path, filter.From, filter.To)

		if err != nil {
			return nil, err
		}

		if addToday {
			visitorsToday, err := analyzer.store.CountVisitorsByPath(nil, filter.TenantID, today, path, false)

			if err != nil {
				return nil, err
			}

			bouncesToday := analyzer.store.CountVisitorsByPathAndMaxOneHit(nil, filter.TenantID, today, path)

			if len(visitorsToday) > 0 {
				if len(visitors) > 0 {
					visitors[len(visitors)-1].Visitors += visitorsToday[0].Visitors
					visitors[len(visitors)-1].Sessions += visitorsToday[0].Sessions
					visitors[len(visitors)-1].Bounces += bouncesToday
					visitors[len(visitors)-1].Views += visitorsToday[0].Views
				} else {
					visitors = append(visitors, Stats{
						Visitors: visitorsToday[0].Visitors,
						Sessions: visitorsToday[0].Sessions,
						Bounces:  bouncesToday,
						Views:    visitorsToday[0].Views,
					})
				}
			}
		}

		var visitorSum, bouncesSum, viewsSum int
		var bounceRate float64

		for i := range visitors {
			visitorSum += visitors[i].Visitors
			bouncesSum += visitors[i].Bounces
			viewsSum += visitors[i].Views
		}

		// we don't need to check for viewsSum, as it will be > 0 if there are unique visitors
		if visitorSum > 0 {
			for i := range visitors {
				visitors[i].RelativeVisitors = float64(visitors[i].Visitors) / float64(visitorSum)

				if visitors[i].Visitors > 0 {
					visitors[i].BounceRate = float64(visitors[i].Bounces) / float64(visitors[i].Visitors)
				}

				if visitorSum > 0 {
					visitors[i].RelativeViews = float64(visitors[i].Views) / float64(viewsSum)
				}
			}

			bounceRate = float64(bouncesSum) / float64(visitorSum)
		}

		stats = append(stats, PathVisitors{
			Path:       path,
			Stats:      visitors,
			Visitors:   visitorSum,
			Bounces:    bouncesSum,
			Views:      viewsSum,
			BounceRate: bounceRate,
		})
		totalVisitors += visitorSum
		totalViews += viewsSum
	}

	// same as above, don't need to check totalViews
	if totalVisitors > 0 {
		for i := range stats {
			stats[i].RelativeVisitors = float64(stats[i].Visitors) / float64(totalVisitors)
			stats[i].RelativeViews = float64(stats[i].Views) / float64(totalViews)
		}
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Visitors > stats[j].Visitors
	})

	return stats, nil
}

// PageLanguages returns the visitor count per language, day, path, and for the given time frame.
// The path is mandatory.
func (analyzer *Analyzer) PageLanguages(filter *Filter) ([]LanguageStats, error) {
	filter = analyzer.getFilter(filter)

	if filter.Path == "" {
		return []LanguageStats{}, nil
	}

	stats, err := analyzer.store.PageLanguages(filter.TenantID, filter.Path, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	var sum float64

	for i := range stats {
		sum += float64(stats[i].Visitors)
	}

	for i := range stats {
		stats[i].RelativeVisitors = float64(stats[i].Visitors) / sum
	}

	return stats, nil
}

// PageReferrer returns the visitor count per referrer, day, path, and for the given time frame.
// The path is mandatory. Bounces for today are not included, as they cannot be calculated for path filtered results.
func (analyzer *Analyzer) PageReferrer(filter *Filter) ([]ReferrerStats, error) {
	filter = analyzer.getFilter(filter)

	if filter.Path == "" {
		return []ReferrerStats{}, nil
	}

	stats, err := analyzer.store.PageReferrer(filter.TenantID, filter.Path, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	var sum float64

	for i := range stats {
		sum += float64(stats[i].Visitors)
	}

	for i := range stats {
		stats[i].RelativeVisitors = float64(stats[i].Visitors) / sum
		stats[i].BounceRate = float64(stats[i].Bounces) / float64(stats[i].Visitors)
	}

	return stats, nil
}

// PageOS returns the visitor count per operating system, day, path, and for the given time frame.
// The path is mandatory.
func (analyzer *Analyzer) PageOS(filter *Filter) ([]OSStats, error) {
	filter = analyzer.getFilter(filter)

	if filter.Path == "" {
		return []OSStats{}, nil
	}

	stats, err := analyzer.store.PageOS(filter.TenantID, filter.Path, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	var sum float64

	for i := range stats {
		sum += float64(stats[i].Visitors)
	}

	for i := range stats {
		stats[i].RelativeVisitors = float64(stats[i].Visitors) / sum
	}

	return stats, nil
}

// PageBrowser returns the visitor count per brower, day, path, and for the given time frame.
// The path is mandatory.
func (analyzer *Analyzer) PageBrowser(filter *Filter) ([]BrowserStats, error) {
	filter = analyzer.getFilter(filter)

	if filter.Path == "" {
		return []BrowserStats{}, nil
	}

	stats, err := analyzer.store.PageBrowser(filter.TenantID, filter.Path, filter.From, filter.To)

	if err != nil {
		return nil, err
	}

	var sum float64

	for i := range stats {
		sum += float64(stats[i].Visitors)
	}

	for i := range stats {
		stats[i].RelativeVisitors = float64(stats[i].Visitors) / sum
	}

	return stats, nil
}

// PagePlatform returns the visitor count per platform, day, path, and for the given time frame.
// The path is mandatory.
func (analyzer *Analyzer) PagePlatform(filter *Filter) *VisitorStats {
	filter = analyzer.getFilter(filter)

	if filter.Path == "" {
		return &VisitorStats{}
	}

	stats := analyzer.store.PagePlatform(filter.TenantID, filter.Path, filter.From, filter.To)

	if stats == nil {
		return &VisitorStats{}
	}

	sum := float64(stats.PlatformDesktop + stats.PlatformMobile + stats.PlatformUnknown)

	if sum > 0 {
		stats.RelativePlatformDesktop = float64(stats.PlatformDesktop) / sum
		stats.RelativePlatformMobile = float64(stats.PlatformMobile) / sum
		stats.RelativePlatformUnknown = float64(stats.PlatformUnknown) / sum
	}

	return stats
}

// Growth returns the total number of visitors, sessions, and bounces for given time frame and path
// and calculates the growth of each metric relative to the previous time frame. The path is optional.
// It does not include today, as that won't be accurate (the day needs to be over to be comparable).
func (analyzer *Analyzer) Growth(filter *Filter) (*Growth, error) {
	filter = analyzer.getFilter(filter)
	current, err := analyzer.store.VisitorsSum(filter.TenantID, filter.From, filter.To, filter.Path)

	if err != nil {
		return nil, err
	}

	days := filter.To.Sub(filter.From)
	filter.To = filter.From.Add(-time.Hour * 24)
	filter.From = filter.To.Add(-days)
	previous, err := analyzer.store.VisitorsSum(filter.TenantID, filter.From, filter.To, filter.Path)

	if err != nil {
		return nil, err
	}

	return &Growth{
		Current:        current,
		Previous:       previous,
		VisitorsGrowth: analyzer.calculateGrowth(current.Visitors, previous.Visitors),
		SessionsGrowth: analyzer.calculateGrowth(current.Sessions, previous.Sessions),
		BouncesGrowth:  analyzer.calculateBouncesGrowth(current, previous),
		ViewsGrowth:    analyzer.calculateGrowth(current.Views, previous.Views),
	}, nil
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

// calculateGrowth calculates the relative growth between two numbers.
func (analyzer *Analyzer) calculateGrowth(current, previous int) float64 {
	if current == 0 && previous == 0 {
		return 0
	} else if previous == 0 {
		return 1
	}

	c := float64(current)
	p := float64(previous)
	return (c - p) / p
}

func (analyzer *Analyzer) calculateBouncesGrowth(current, previous *Stats) float64 {
	var currentBounceRate float64
	var previousBounceRate float64

	if current.Visitors > 0 {
		currentBounceRate = float64(current.Bounces) / float64(current.Visitors)
	}

	if previous.Visitors > 0 {
		previousBounceRate = float64(previous.Bounces) / float64(previous.Visitors)
	}

	var bounceGrowth float64

	// use visitors instead of previousBounceRate, as that's an integer and more reliable for this type of comparison
	if previous.Visitors > 0 {
		bounceGrowth = (currentBounceRate - previousBounceRate) / previousBounceRate
	}

	return bounceGrowth
}
