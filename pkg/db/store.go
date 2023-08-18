package db

import (
	"github.com/pirsch-analytics/pirsch/v6/pkg"
	model2 "github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"time"
)

// Store is the database storage interface.
type Store interface {
	// SavePageViews saves given hits.
	SavePageViews([]model2.PageView) error

	// SaveSessions saves given sessions.
	SaveSessions([]model2.Session) error

	// SaveEvents saves given events.
	SaveEvents([]model2.Event) error

	// SaveUserAgents saves given UserAgent headers.
	SaveUserAgents([]model2.UserAgent) error

	// SaveBots saves given bots.
	SaveBots([]model2.Bot) error

	// Session returns the last hit for given client, fingerprint, and maximum age.
	Session(uint64, uint64, time.Time) (*model2.Session, error)

	// Count returns the number of results for given query.
	Count(string, ...any) (int, error)

	// SelectActiveVisitorStats selects ActiveVisitorStats.
	SelectActiveVisitorStats(bool, string, ...any) ([]model2.ActiveVisitorStats, error)

	// GetTotalVisitorStats returns the TotalVisitorStats.
	GetTotalVisitorStats(string, ...any) (*model2.TotalVisitorStats, error)

	// GetTotalVisitorsPageViewsStats returns the TotalVisitorsPageViewsStats.
	GetTotalVisitorsPageViewsStats(string, ...any) (*model2.TotalVisitorsPageViewsStats, error)

	// SelectVisitorStats selects VisitorStats.
	SelectVisitorStats(pkg.Period, string, ...any) ([]model2.VisitorStats, error)

	// SelectTimeSpentStats selects TimeSpentStats.
	SelectTimeSpentStats(pkg.Period, string, ...any) ([]model2.TimeSpentStats, error)

	// GetGrowthStats returns the GrowthStats.
	GetGrowthStats(string, ...any) (*model2.GrowthStats, error)

	// SelectVisitorHourStats selects VisitorHourStats.
	SelectVisitorHourStats(string, ...any) ([]model2.VisitorHourStats, error)

	// SelectPageStats selects PageStats.
	SelectPageStats(bool, bool, string, ...any) ([]model2.PageStats, error)

	// SelectAvgTimeSpentStats selects AvgTimeSpentStats.
	SelectAvgTimeSpentStats(string, ...any) ([]model2.AvgTimeSpentStats, error)

	// SelectEntryStats selects EntryStats.
	SelectEntryStats(bool, string, ...any) ([]model2.EntryStats, error)

	// SelectExitStats selects ExitStats.
	SelectExitStats(bool, string, ...any) ([]model2.ExitStats, error)

	// SelectTotalSessions returns the total number of unique sessions.
	SelectTotalSessions(string, ...any) (int, error)

	// SelectTotalVisitorSessionStats selects TotalVisitorSessionStats.
	SelectTotalVisitorSessionStats(string, ...any) ([]model2.TotalVisitorSessionStats, error)

	// GetPageConversionsStats returns the PageConversionsStats.
	GetPageConversionsStats(string, ...any) (*model2.PageConversionsStats, error)

	// SelectEventStats selects EventStats.
	SelectEventStats(bool, string, ...any) ([]model2.EventStats, error)

	// SelectEventListStats selects EventListStats.
	SelectEventListStats(string, ...any) ([]model2.EventListStats, error)

	// SelectReferrerStats selects ReferrerStats.
	SelectReferrerStats(string, ...any) ([]model2.ReferrerStats, error)

	// GetPlatformStats returns the PlatformStats.
	GetPlatformStats(string, ...any) (*model2.PlatformStats, error)

	// SelectLanguageStats selects LanguageStats.
	SelectLanguageStats(string, ...any) ([]model2.LanguageStats, error)

	// SelectCountryStats selects CountryStats.
	SelectCountryStats(string, ...any) ([]model2.CountryStats, error)

	// SelectCityStats selects CityStats.
	SelectCityStats(string, ...any) ([]model2.CityStats, error)

	// SelectBrowserStats selects BrowserStats.
	SelectBrowserStats(string, ...any) ([]model2.BrowserStats, error)

	// SelectOSStats selects OSStats.
	SelectOSStats(string, ...any) ([]model2.OSStats, error)

	// SelectScreenClassStats selects ScreenClassStats.
	SelectScreenClassStats(string, ...any) ([]model2.ScreenClassStats, error)

	// SelectUTMSourceStats selects UTMSourceStats.
	SelectUTMSourceStats(string, ...any) ([]model2.UTMSourceStats, error)

	// SelectUTMMediumStats selects UTMMediumStats.
	SelectUTMMediumStats(string, ...any) ([]model2.UTMMediumStats, error)

	// SelectUTMCampaignStats selects UTMCampaignStats.
	SelectUTMCampaignStats(string, ...any) ([]model2.UTMCampaignStats, error)

	// SelectUTMContentStats selects UTMContentStats.
	SelectUTMContentStats(string, ...any) ([]model2.UTMContentStats, error)

	// SelectUTMTermStats selects UTMTermStats.
	SelectUTMTermStats(string, ...any) ([]model2.UTMTermStats, error)

	// SelectOSVersionStats selects OSVersionStats.
	SelectOSVersionStats(string, ...any) ([]model2.OSVersionStats, error)

	// SelectBrowserVersionStats selects BrowserVersionStats.
	SelectBrowserVersionStats(string, ...any) ([]model2.BrowserVersionStats, error)

	// SelectOptions selects a list of filter options.
	SelectOptions(string, ...any) ([]string, error)
}
