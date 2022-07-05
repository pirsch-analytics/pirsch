package pirsch

import (
	"time"
)

// Store is the database storage interface.
type Store interface {
	// SavePageViews saves given hits.
	SavePageViews([]PageView) error

	// SaveSessions saves given sessions.
	SaveSessions([]Session) error

	// SaveEvents saves given events.
	SaveEvents([]Event) error

	// SaveUserAgents saves given UserAgent headers.
	SaveUserAgents([]UserAgent) error

	// Session returns the last hit for given client, fingerprint, and maximum age.
	Session(uint64, uint64, time.Time) (*Session, error)

	// Count returns the number of results for given query.
	Count(string, ...any) (int, error)

	// SelectActiveVisitorStats selects ActiveVisitorStats.
	SelectActiveVisitorStats(bool, string, ...any) ([]ActiveVisitorStats, error)

	// GetTotalVisitorStats returns the TotalVisitorStats.
	GetTotalVisitorStats(string, ...any) (*TotalVisitorStats, error)

	// SelectVisitorStats selects VisitorStats.
	SelectVisitorStats(Period, string, ...any) ([]VisitorStats, error)

	// SelectTimeSpentStats selects TimeSpentStats.
	SelectTimeSpentStats(Period, string, ...any) ([]TimeSpentStats, error)

	// GetGrowthStats returns the GrowthStats.
	GetGrowthStats(string, ...any) (*GrowthStats, error)

	// SelectVisitorHourStats selects VisitorHourStats.
	SelectVisitorHourStats(string, ...any) ([]VisitorHourStats, error)

	// SelectPageStats selects PageStats.
	SelectPageStats(bool, bool, string, ...any) ([]PageStats, error)

	// SelectAvgTimeSpentStats selects AvgTimeSpentStats.
	SelectAvgTimeSpentStats(string, ...any) ([]AvgTimeSpentStats, error)

	// SelectEntryStats selects EntryStats.
	SelectEntryStats(bool, string, ...any) ([]EntryStats, error)

	// SelectExitStats selects ExitStats.
	SelectExitStats(bool, string, ...any) ([]ExitStats, error)

	// SelectTotalVisitorSessionStats selects TotalVisitorSessionStats.
	SelectTotalVisitorSessionStats(string, ...any) ([]TotalVisitorSessionStats, error)

	// GetPageConversionsStats returns the PageConversionsStats.
	GetPageConversionsStats(string, ...any) (*PageConversionsStats, error)

	// SelectEventStats selects EventStats.
	SelectEventStats(bool, string, ...any) ([]EventStats, error)

	// SelectEventListStats selects EventListStats.
	SelectEventListStats(string, ...any) ([]EventListStats, error)

	// SelectReferrerStats selects ReferrerStats.
	SelectReferrerStats(string, ...any) ([]ReferrerStats, error)

	// GetPlatformStats returns the PlatformStats.
	GetPlatformStats(string, ...any) (*PlatformStats, error)

	// SelectLanguageStats selects LanguageStats.
	SelectLanguageStats(string, ...any) ([]LanguageStats, error)

	// SelectCountryStats selects CountryStats.
	SelectCountryStats(string, ...any) ([]CountryStats, error)

	// SelectCityStats selects CityStats.
	SelectCityStats(string, ...any) ([]CityStats, error)

	// SelectBrowserStats selects BrowserStats.
	SelectBrowserStats(string, ...any) ([]BrowserStats, error)

	// SelectOSStats selects OSStats.
	SelectOSStats(string, ...any) ([]OSStats, error)

	// SelectScreenClassStats selects ScreenClassStats.
	SelectScreenClassStats(string, ...any) ([]ScreenClassStats, error)

	// SelectUTMSourceStats selects UTMSourceStats.
	SelectUTMSourceStats(string, ...any) ([]UTMSourceStats, error)

	// SelectUTMMediumStats selects UTMMediumStats.
	SelectUTMMediumStats(string, ...any) ([]UTMMediumStats, error)

	// SelectUTMCampaignStats selects UTMCampaignStats.
	SelectUTMCampaignStats(string, ...any) ([]UTMCampaignStats, error)

	// SelectUTMContentStats selects UTMContentStats.
	SelectUTMContentStats(string, ...any) ([]UTMContentStats, error)

	// SelectUTMTermStats selects UTMTermStats.
	SelectUTMTermStats(string, ...any) ([]UTMTermStats, error)

	// SelectOSVersionStats selects OSVersionStats.
	SelectOSVersionStats(string, ...any) ([]OSVersionStats, error)

	// SelectBrowserVersionStats selects BrowserVersionStats.
	SelectBrowserVersionStats(string, ...any) ([]BrowserVersionStats, error)
}
