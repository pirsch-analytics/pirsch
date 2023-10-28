package db

import (
	"context"
	"github.com/pirsch-analytics/pirsch/v6/pkg"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"time"
)

// Store is the database storage interface.
type Store interface {
	// SavePageViews saves given hits.
	SavePageViews(context.Context, []model.PageView) error

	// SaveSessions saves given sessions.
	SaveSessions(context.Context, []model.Session) error

	// SaveEvents saves given events.
	SaveEvents(context.Context, []model.Event) error

	// SaveUserAgents saves given UserAgent headers.
	SaveUserAgents(context.Context, []model.UserAgent) error

	// SaveBots saves given bots.
	SaveBots(context.Context, []model.Bot) error

	// Session returns the last hit for given client, fingerprint, and maximum age.
	Session(context.Context, uint64, uint64, time.Time) (*model.Session, error)

	// Count returns the number of results for given query.
	Count(context.Context, string, ...any) (int, error)

	// SelectActiveVisitorStats selects ActiveVisitorStats.
	SelectActiveVisitorStats(context.Context, bool, string, ...any) ([]model.ActiveVisitorStats, error)

	// GetTotalVisitorStats returns the TotalVisitorStats.
	GetTotalVisitorStats(context.Context, string, bool, bool, ...any) (*model.TotalVisitorStats, error)

	// GetTotalUniqueVisitorStats returns the total number of unique visitors.
	GetTotalUniqueVisitorStats(context.Context, string, ...any) (int, error)

	// GetTotalPageViewStats returns the total number of page views.
	GetTotalPageViewStats(context.Context, string, ...any) (int, error)

	// GetTotalSessionStats returns the total number of sessions.
	GetTotalSessionStats(context.Context, string, ...any) (int, error)

	// GetTotalVisitorsPageViewsStats returns the TotalVisitorsPageViewsStats.
	GetTotalVisitorsPageViewsStats(context.Context, string, ...any) (*model.TotalVisitorsPageViewsStats, error)

	// SelectVisitorStats selects VisitorStats.
	SelectVisitorStats(context.Context, pkg.Period, string, bool, bool, ...any) ([]model.VisitorStats, error)

	// SelectTimeSpentStats selects TimeSpentStats.
	SelectTimeSpentStats(context.Context, pkg.Period, string, ...any) ([]model.TimeSpentStats, error)

	// GetGrowthStats returns the GrowthStats.
	GetGrowthStats(context.Context, string, bool, bool, ...any) (*model.GrowthStats, error)

	// SelectVisitorHourStats selects VisitorHourStats.
	SelectVisitorHourStats(context.Context, string, bool, bool, ...any) ([]model.VisitorHourStats, error)

	// SelectPageStats selects PageStats.
	SelectPageStats(context.Context, bool, bool, string, ...any) ([]model.PageStats, error)

	// SelectAvgTimeSpentStats selects AvgTimeSpentStats.
	SelectAvgTimeSpentStats(context.Context, string, ...any) ([]model.AvgTimeSpentStats, error)

	// SelectEntryStats selects EntryStats.
	SelectEntryStats(context.Context, bool, string, ...any) ([]model.EntryStats, error)

	// SelectExitStats selects ExitStats.
	SelectExitStats(context.Context, bool, string, ...any) ([]model.ExitStats, error)

	// SelectTotalSessions returns the total number of unique sessions.
	SelectTotalSessions(context.Context, string, ...any) (int, error)

	// SelectTotalVisitorSessionStats selects TotalVisitorSessionStats.
	SelectTotalVisitorSessionStats(context.Context, string, ...any) ([]model.TotalVisitorSessionStats, error)

	// GetConversionsStats returns the ConversionsStats.
	GetConversionsStats(context.Context, string, bool, ...any) (*model.ConversionsStats, error)

	// SelectEventStats selects EventStats.
	SelectEventStats(context.Context, bool, string, ...any) ([]model.EventStats, error)

	// SelectEventListStats selects EventListStats.
	SelectEventListStats(context.Context, string, ...any) ([]model.EventListStats, error)

	// SelectReferrerStats selects ReferrerStats.
	SelectReferrerStats(context.Context, string, ...any) ([]model.ReferrerStats, error)

	// GetPlatformStats returns the PlatformStats.
	GetPlatformStats(context.Context, string, ...any) (*model.PlatformStats, error)

	// SelectLanguageStats selects LanguageStats.
	SelectLanguageStats(context.Context, string, ...any) ([]model.LanguageStats, error)

	// SelectCountryStats selects CountryStats.
	SelectCountryStats(context.Context, string, ...any) ([]model.CountryStats, error)

	// SelectCityStats selects CityStats.
	SelectCityStats(context.Context, string, ...any) ([]model.CityStats, error)

	// SelectBrowserStats selects BrowserStats.
	SelectBrowserStats(context.Context, string, ...any) ([]model.BrowserStats, error)

	// SelectOSStats selects OSStats.
	SelectOSStats(context.Context, string, ...any) ([]model.OSStats, error)

	// SelectScreenClassStats selects ScreenClassStats.
	SelectScreenClassStats(context.Context, string, ...any) ([]model.ScreenClassStats, error)

	// SelectUTMSourceStats selects UTMSourceStats.
	SelectUTMSourceStats(context.Context, string, ...any) ([]model.UTMSourceStats, error)

	// SelectUTMMediumStats selects UTMMediumStats.
	SelectUTMMediumStats(context.Context, string, ...any) ([]model.UTMMediumStats, error)

	// SelectUTMCampaignStats selects UTMCampaignStats.
	SelectUTMCampaignStats(context.Context, string, ...any) ([]model.UTMCampaignStats, error)

	// SelectUTMContentStats selects UTMContentStats.
	SelectUTMContentStats(context.Context, string, ...any) ([]model.UTMContentStats, error)

	// SelectUTMTermStats selects UTMTermStats.
	SelectUTMTermStats(context.Context, string, ...any) ([]model.UTMTermStats, error)

	// SelectOSVersionStats selects OSVersionStats.
	SelectOSVersionStats(context.Context, string, ...any) ([]model.OSVersionStats, error)

	// SelectBrowserVersionStats selects BrowserVersionStats.
	SelectBrowserVersionStats(context.Context, string, ...any) ([]model.BrowserVersionStats, error)

	// SelectOptions selects a list of filter options.
	SelectOptions(context.Context, string, ...any) ([]string, error)
}
