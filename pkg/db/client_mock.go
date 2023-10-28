package db

import (
	"context"
	"github.com/pirsch-analytics/pirsch/v6/pkg"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"sort"
	"sync"
	"time"
)

// ClientMock is a mock Store implementation.
type ClientMock struct {
	pageViews     []model.PageView
	sessions      []model.Session
	events        []model.Event
	userAgents    []model.UserAgent
	bots          []model.Bot
	ReturnSession *model.Session
	m             sync.Mutex
}

// NewClientMock returns a new mock client.
func NewClientMock() *ClientMock {
	return &ClientMock{
		pageViews:  make([]model.PageView, 0),
		sessions:   make([]model.Session, 0),
		events:     make([]model.Event, 0),
		userAgents: make([]model.UserAgent, 0),
	}
}

// GetPageViews returns a copy of the page views slice.
func (client *ClientMock) GetPageViews() []model.PageView {
	client.m.Lock()
	defer client.m.Unlock()
	data := make([]model.PageView, len(client.pageViews))
	copy(data, client.pageViews)
	sort.Slice(data, func(i, j int) bool {
		if data[i].Time.Before(data[j].Time) {
			return true
		}

		return false
	})
	return data
}

// GetSessions returns a copy of the sessions slice.
func (client *ClientMock) GetSessions() []model.Session {
	client.m.Lock()
	defer client.m.Unlock()
	data := make([]model.Session, len(client.sessions))
	copy(data, client.sessions)
	sort.Slice(data, func(i, j int) bool {
		if data[i].Time.Before(data[j].Time) {
			return true
		}

		return false
	})
	return data
}

// GetEvents returns a copy of the events slice.
func (client *ClientMock) GetEvents() []model.Event {
	client.m.Lock()
	defer client.m.Unlock()
	data := make([]model.Event, len(client.events))
	copy(data, client.events)
	sort.Slice(data, func(i, j int) bool {
		if data[i].Time.Before(data[j].Time) {
			return true
		}

		return false
	})
	return data
}

// GetUserAgents returns a copy of the user agents slice.
func (client *ClientMock) GetUserAgents() []model.UserAgent {
	client.m.Lock()
	defer client.m.Unlock()
	data := make([]model.UserAgent, len(client.userAgents))
	copy(data, client.userAgents)
	sort.Slice(data, func(i, j int) bool {
		if data[i].Time.Before(data[j].Time) {
			return true
		}

		return false
	})
	return data
}

// GetBots returns a copy of the bots slice.
func (client *ClientMock) GetBots() []model.Bot {
	client.m.Lock()
	defer client.m.Unlock()
	data := make([]model.Bot, len(client.bots))
	copy(data, client.bots)
	sort.Slice(data, func(i, j int) bool {
		if data[i].Time.Before(data[j].Time) {
			return true
		}

		return false
	})
	return data
}

// SavePageViews implements the Store interface.
func (client *ClientMock) SavePageViews(_ context.Context, pageViews []model.PageView) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.pageViews = append(client.pageViews, pageViews...)
	return nil
}

// SaveSessions implements the Store interface.
func (client *ClientMock) SaveSessions(_ context.Context, sessions []model.Session) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.sessions = append(client.sessions, sessions...)
	return nil
}

// SaveEvents implements the Store interface.
func (client *ClientMock) SaveEvents(_ context.Context, events []model.Event) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.events = append(client.events, events...)
	return nil
}

// SaveUserAgents implements the Store interface.
func (client *ClientMock) SaveUserAgents(_ context.Context, userAgents []model.UserAgent) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.userAgents = append(client.userAgents, userAgents...)
	return nil
}

func (client *ClientMock) SaveBots(_ context.Context, bots []model.Bot) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.bots = append(client.bots, bots...)
	return nil
}

// Session implements the Store interface.
func (client *ClientMock) Session(context.Context, uint64, uint64, time.Time) (*model.Session, error) {
	if client.ReturnSession != nil {
		return client.ReturnSession, nil
	}

	return nil, nil
}

// Count implements the Store interface.
func (client *ClientMock) Count(context.Context, string, ...any) (int, error) {
	return 0, nil
}

// SelectActiveVisitorStats implements the Store interface.
func (client *ClientMock) SelectActiveVisitorStats(context.Context, bool, string, ...any) ([]model.ActiveVisitorStats, error) {
	return nil, nil
}

// GetTotalVisitorStats implements the Store interface.
func (client *ClientMock) GetTotalVisitorStats(context.Context, string, bool, bool, ...any) (*model.TotalVisitorStats, error) {
	return nil, nil
}

// GetTotalVisitorsPageViewsStats implements the Store interface.
func (client *ClientMock) GetTotalVisitorsPageViewsStats(context.Context, string, ...any) (*model.TotalVisitorsPageViewsStats, error) {
	return nil, nil
}

// SelectVisitorStats implements the Store interface.
func (client *ClientMock) SelectVisitorStats(context.Context, pkg.Period, string, bool, bool, ...any) ([]model.VisitorStats, error) {
	return nil, nil
}

// GetTotalUniqueVisitorStats implements the Store interface.
func (client *ClientMock) GetTotalUniqueVisitorStats(context.Context, string, ...any) (int, error) {
	return 0, nil
}

// GetTotalPageViewStats implements the Store interface.
func (client *ClientMock) GetTotalPageViewStats(context.Context, string, ...any) (int, error) {
	return 0, nil
}

// GetTotalSessionStats implements the Store interface.
func (client *ClientMock) GetTotalSessionStats(context.Context, string, ...any) (int, error) {
	return 0, nil
}

// SelectTimeSpentStats implements the Store interface.
func (client *ClientMock) SelectTimeSpentStats(context.Context, pkg.Period, string, ...any) ([]model.TimeSpentStats, error) {
	return nil, nil
}

// GetGrowthStats implements the Store interface.
func (client *ClientMock) GetGrowthStats(context.Context, string, bool, bool, ...any) (*model.GrowthStats, error) {
	return nil, nil
}

// SelectVisitorHourStats implements the Store interface.
func (client *ClientMock) SelectVisitorHourStats(context.Context, string, bool, bool, ...any) ([]model.VisitorHourStats, error) {
	return nil, nil
}

// SelectPageStats implements the Store interface.
func (client *ClientMock) SelectPageStats(context.Context, bool, bool, string, ...any) ([]model.PageStats, error) {
	return nil, nil
}

// SelectAvgTimeSpentStats implements the Store interface.
func (client *ClientMock) SelectAvgTimeSpentStats(context.Context, string, ...any) ([]model.AvgTimeSpentStats, error) {
	return nil, nil
}

// SelectEntryStats implements the Store interface.
func (client *ClientMock) SelectEntryStats(context.Context, bool, string, ...any) ([]model.EntryStats, error) {
	return nil, nil
}

// SelectExitStats implements the Store interface.
func (client *ClientMock) SelectExitStats(context.Context, bool, string, ...any) ([]model.ExitStats, error) {
	return nil, nil
}

// SelectTotalSessions implements the Store interface.
func (client *ClientMock) SelectTotalSessions(context.Context, string, ...any) (int, error) {
	return 0, nil
}

// SelectTotalVisitorSessionStats implements the Store interface.
func (client *ClientMock) SelectTotalVisitorSessionStats(context.Context, string, ...any) ([]model.TotalVisitorSessionStats, error) {
	return nil, nil
}

// GetConversionsStats implements the Store interface.
func (client *ClientMock) GetConversionsStats(context.Context, string, bool, ...any) (*model.ConversionsStats, error) {
	return nil, nil
}

// SelectEventStats implements the Store interface.
func (client *ClientMock) SelectEventStats(context.Context, bool, string, ...any) ([]model.EventStats, error) {
	return nil, nil
}

// SelectEventListStats implements the Store interface.
func (client *ClientMock) SelectEventListStats(context.Context, string, ...any) ([]model.EventListStats, error) {
	return nil, nil
}

// SelectReferrerStats implements the Store interface.
func (client *ClientMock) SelectReferrerStats(context.Context, string, ...any) ([]model.ReferrerStats, error) {
	return nil, nil
}

// GetPlatformStats implements the Store interface.
func (client *ClientMock) GetPlatformStats(context.Context, string, ...any) (*model.PlatformStats, error) {
	return nil, nil
}

// SelectLanguageStats implements the Store interface.
func (client *ClientMock) SelectLanguageStats(context.Context, string, ...any) ([]model.LanguageStats, error) {
	return nil, nil
}

// SelectCountryStats implements the Store interface.
func (client *ClientMock) SelectCountryStats(context.Context, string, ...any) ([]model.CountryStats, error) {
	return nil, nil
}

// SelectCityStats implements the Store interface.
func (client *ClientMock) SelectCityStats(context.Context, string, ...any) ([]model.CityStats, error) {
	return nil, nil
}

// SelectBrowserStats implements the Store interface.
func (client *ClientMock) SelectBrowserStats(context.Context, string, ...any) ([]model.BrowserStats, error) {
	return nil, nil
}

// SelectOSStats implements the Store interface.
func (client *ClientMock) SelectOSStats(context.Context, string, ...any) ([]model.OSStats, error) {
	return nil, nil
}

// SelectScreenClassStats implements the Store interface.
func (client *ClientMock) SelectScreenClassStats(context.Context, string, ...any) ([]model.ScreenClassStats, error) {
	return nil, nil
}

// SelectUTMSourceStats implements the Store interface.
func (client *ClientMock) SelectUTMSourceStats(context.Context, string, ...any) ([]model.UTMSourceStats, error) {
	return nil, nil
}

// SelectUTMMediumStats implements the Store interface.
func (client *ClientMock) SelectUTMMediumStats(context.Context, string, ...any) ([]model.UTMMediumStats, error) {
	return nil, nil
}

// SelectUTMCampaignStats implements the Store interface.
func (client *ClientMock) SelectUTMCampaignStats(context.Context, string, ...any) ([]model.UTMCampaignStats, error) {
	return nil, nil
}

// SelectUTMContentStats implements the Store interface.
func (client *ClientMock) SelectUTMContentStats(context.Context, string, ...any) ([]model.UTMContentStats, error) {
	return nil, nil
}

// SelectUTMTermStats implements the Store interface.
func (client *ClientMock) SelectUTMTermStats(context.Context, string, ...any) ([]model.UTMTermStats, error) {
	return nil, nil
}

// SelectOSVersionStats implements the Store interface.
func (client *ClientMock) SelectOSVersionStats(context.Context, string, ...any) ([]model.OSVersionStats, error) {
	return nil, nil
}

// SelectBrowserVersionStats implements the Store interface.
func (client *ClientMock) SelectBrowserVersionStats(context.Context, string, ...any) ([]model.BrowserVersionStats, error) {
	return nil, nil
}

// SelectOptions implements the Store interface.
func (client *ClientMock) SelectOptions(context.Context, string, ...any) ([]string, error) {
	return nil, nil
}
