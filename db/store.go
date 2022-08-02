package db

import (
	"github.com/pirsch-analytics/pirsch/v4"
	"github.com/pirsch-analytics/pirsch/v4/model"
	"time"
)

// Store is the database storage interface.
type Store interface {
	// SavePageViews saves given hits.
	SavePageViews([]model.PageView) error

	// SaveSessions saves given sessions.
	SaveSessions([]model.Session) error

	// SaveEvents saves given events.
	SaveEvents([]model.Event) error

	// SaveUserAgents saves given UserAgent headers.
	SaveUserAgents([]model.UserAgent) error

	// Session returns the last hit for given client, fingerprint, and maximum age.
	Session(uint64, uint64, time.Time) (*model.Session, error)

	// Count returns the number of results for given query.
	Count(string, ...any) (int, error)

	// SelectActiveVisitorStats selects ActiveVisitorStats.
	SelectActiveVisitorStats(bool, string, ...any) ([]model.ActiveVisitorStats, error)

	// GetTotalVisitorStats returns the TotalVisitorStats.
	GetTotalVisitorStats(string, ...any) (*model.TotalVisitorStats, error)

	// SelectVisitorStats selects VisitorStats.
	SelectVisitorStats(pirsch.Period, string, ...any) ([]model.VisitorStats, error)

	// SelectTimeSpentStats selects TimeSpentStats.
	SelectTimeSpentStats(pirsch.Period, string, ...any) ([]model.TimeSpentStats, error)

	// GetGrowthStats returns the GrowthStats.
	GetGrowthStats(string, ...any) (*model.GrowthStats, error)

	// SelectVisitorHourStats selects VisitorHourStats.
	SelectVisitorHourStats(string, ...any) ([]model.VisitorHourStats, error)

	// SelectPageStats selects PageStats.
	SelectPageStats(bool, bool, string, ...any) ([]model.PageStats, error)

	// SelectAvgTimeSpentStats selects AvgTimeSpentStats.
	SelectAvgTimeSpentStats(string, ...any) ([]model.AvgTimeSpentStats, error)

	// SelectEntryStats selects EntryStats.
	SelectEntryStats(bool, string, ...any) ([]model.EntryStats, error)

	// SelectExitStats selects ExitStats.
	SelectExitStats(bool, string, ...any) ([]model.ExitStats, error)

	// SelectTotalSessions returns the total number of unique sessions.
	SelectTotalSessions(string, ...any) (int, error)

	// SelectTotalVisitorSessionStats selects TotalVisitorSessionStats.
	SelectTotalVisitorSessionStats(string, ...any) ([]model.TotalVisitorSessionStats, error)

	// GetPageConversionsStats returns the PageConversionsStats.
	GetPageConversionsStats(string, ...any) (*model.PageConversionsStats, error)

	// SelectEventStats selects EventStats.
	SelectEventStats(bool, string, ...any) ([]model.EventStats, error)

	// SelectEventListStats selects EventListStats.
	SelectEventListStats(string, ...any) ([]model.EventListStats, error)

	// SelectReferrerStats selects ReferrerStats.
	SelectReferrerStats(string, ...any) ([]model.ReferrerStats, error)

	// GetPlatformStats returns the PlatformStats.
	GetPlatformStats(string, ...any) (*model.PlatformStats, error)

	// SelectLanguageStats selects LanguageStats.
	SelectLanguageStats(string, ...any) ([]model.LanguageStats, error)

	// SelectCountryStats selects CountryStats.
	SelectCountryStats(string, ...any) ([]model.CountryStats, error)

	// SelectCityStats selects CityStats.
	SelectCityStats(string, ...any) ([]model.CityStats, error)

	// SelectBrowserStats selects BrowserStats.
	SelectBrowserStats(string, ...any) ([]model.BrowserStats, error)

	// SelectOSStats selects OSStats.
	SelectOSStats(string, ...any) ([]model.OSStats, error)

	// SelectScreenClassStats selects ScreenClassStats.
	SelectScreenClassStats(string, ...any) ([]model.ScreenClassStats, error)

	// SelectUTMSourceStats selects UTMSourceStats.
	SelectUTMSourceStats(string, ...any) ([]model.UTMSourceStats, error)

	// SelectUTMMediumStats selects UTMMediumStats.
	SelectUTMMediumStats(string, ...any) ([]model.UTMMediumStats, error)

	// SelectUTMCampaignStats selects UTMCampaignStats.
	SelectUTMCampaignStats(string, ...any) ([]model.UTMCampaignStats, error)

	// SelectUTMContentStats selects UTMContentStats.
	SelectUTMContentStats(string, ...any) ([]model.UTMContentStats, error)

	// SelectUTMTermStats selects UTMTermStats.
	SelectUTMTermStats(string, ...any) ([]model.UTMTermStats, error)

	// SelectOSVersionStats selects OSVersionStats.
	SelectOSVersionStats(string, ...any) ([]model.OSVersionStats, error)

	// SelectBrowserVersionStats selects BrowserVersionStats.
	SelectBrowserVersionStats(string, ...any) ([]model.BrowserVersionStats, error)

	// SelectOptions selects a list of filter options.
	SelectOptions(string, ...any) ([]string, error)
}
