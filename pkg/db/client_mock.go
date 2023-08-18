package db

import (
	"github.com/pirsch-analytics/pirsch/v6"
	model2 "github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"sort"
	"sync"
	"time"
)

// ClientMock is a mock Store implementation.
type ClientMock struct {
	pageViews     []model2.PageView
	sessions      []model2.Session
	events        []model2.Event
	userAgents    []model2.UserAgent
	bots          []model2.Bot
	ReturnSession *model2.Session
	m             sync.Mutex
}

// NewClientMock returns a new mock client.
func NewClientMock() *ClientMock {
	return &ClientMock{
		pageViews:  make([]model2.PageView, 0),
		sessions:   make([]model2.Session, 0),
		events:     make([]model2.Event, 0),
		userAgents: make([]model2.UserAgent, 0),
	}
}

// GetPageViews returns a copy of the page views slice.
func (client *ClientMock) GetPageViews() []model2.PageView {
	client.m.Lock()
	defer client.m.Unlock()
	data := make([]model2.PageView, len(client.pageViews))
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
func (client *ClientMock) GetSessions() []model2.Session {
	client.m.Lock()
	defer client.m.Unlock()
	data := make([]model2.Session, len(client.sessions))
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
func (client *ClientMock) GetEvents() []model2.Event {
	client.m.Lock()
	defer client.m.Unlock()
	data := make([]model2.Event, len(client.events))
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
func (client *ClientMock) GetUserAgents() []model2.UserAgent {
	client.m.Lock()
	defer client.m.Unlock()
	data := make([]model2.UserAgent, len(client.userAgents))
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
func (client *ClientMock) GetBots() []model2.Bot {
	client.m.Lock()
	defer client.m.Unlock()
	data := make([]model2.Bot, len(client.bots))
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
func (client *ClientMock) SavePageViews(pageViews []model2.PageView) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.pageViews = append(client.pageViews, pageViews...)
	return nil
}

// SaveSessions implements the Store interface.
func (client *ClientMock) SaveSessions(sessions []model2.Session) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.sessions = append(client.sessions, sessions...)
	return nil
}

// SaveEvents implements the Store interface.
func (client *ClientMock) SaveEvents(events []model2.Event) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.events = append(client.events, events...)
	return nil
}

// SaveUserAgents implements the Store interface.
func (client *ClientMock) SaveUserAgents(userAgents []model2.UserAgent) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.userAgents = append(client.userAgents, userAgents...)
	return nil
}

func (client *ClientMock) SaveBots(bots []model2.Bot) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.bots = append(client.bots, bots...)
	return nil
}

// Session implements the Store interface.
func (client *ClientMock) Session(uint64, uint64, time.Time) (*model2.Session, error) {
	if client.ReturnSession != nil {
		return client.ReturnSession, nil
	}

	return nil, nil
}

// Count implements the Store interface.
func (client *ClientMock) Count(string, ...any) (int, error) {
	return 0, nil
}

// SelectActiveVisitorStats implements the Store interface.
func (client *ClientMock) SelectActiveVisitorStats(bool, string, ...any) ([]model2.ActiveVisitorStats, error) {
	return nil, nil
}

// GetTotalVisitorStats implements the Store interface.
func (client *ClientMock) GetTotalVisitorStats(string, ...any) (*model2.TotalVisitorStats, error) {
	return nil, nil
}

// GetTotalVisitorsPageViewsStats implements the Store interface.
func (client *ClientMock) GetTotalVisitorsPageViewsStats(string, ...any) (*model2.TotalVisitorsPageViewsStats, error) {
	return nil, nil
}

// SelectVisitorStats implements the Store interface.
func (client *ClientMock) SelectVisitorStats(pirsch.Period, string, ...any) ([]model2.VisitorStats, error) {
	return nil, nil
}

// SelectTimeSpentStats implements the Store interface.
func (client *ClientMock) SelectTimeSpentStats(pirsch.Period, string, ...any) ([]model2.TimeSpentStats, error) {
	return nil, nil
}

// GetGrowthStats implements the Store interface.
func (client *ClientMock) GetGrowthStats(string, ...any) (*model2.GrowthStats, error) {
	return nil, nil
}

// SelectVisitorHourStats implements the Store interface.
func (client *ClientMock) SelectVisitorHourStats(string, ...any) ([]model2.VisitorHourStats, error) {
	return nil, nil
}

// SelectPageStats implements the Store interface.
func (client *ClientMock) SelectPageStats(bool, bool, string, ...any) ([]model2.PageStats, error) {
	return nil, nil
}

// SelectAvgTimeSpentStats implements the Store interface.
func (client *ClientMock) SelectAvgTimeSpentStats(string, ...any) ([]model2.AvgTimeSpentStats, error) {
	return nil, nil
}

// SelectEntryStats implements the Store interface.
func (client *ClientMock) SelectEntryStats(bool, string, ...any) ([]model2.EntryStats, error) {
	return nil, nil
}

// SelectExitStats implements the Store interface.
func (client *ClientMock) SelectExitStats(bool, string, ...any) ([]model2.ExitStats, error) {
	return nil, nil
}

// SelectTotalSessions implements the Store interface.
func (client *ClientMock) SelectTotalSessions(string, ...any) (int, error) {
	return 0, nil
}

// SelectTotalVisitorSessionStats implements the Store interface.
func (client *ClientMock) SelectTotalVisitorSessionStats(string, ...any) ([]model2.TotalVisitorSessionStats, error) {
	return nil, nil
}

// GetPageConversionsStats implements the Store interface.
func (client *ClientMock) GetPageConversionsStats(string, ...any) (*model2.PageConversionsStats, error) {
	return nil, nil
}

// SelectEventStats implements the Store interface.
func (client *ClientMock) SelectEventStats(bool, string, ...any) ([]model2.EventStats, error) {
	return nil, nil
}

// SelectEventListStats implements the Store interface.
func (client *ClientMock) SelectEventListStats(string, ...any) ([]model2.EventListStats, error) {
	return nil, nil
}

// SelectReferrerStats implements the Store interface.
func (client *ClientMock) SelectReferrerStats(string, ...any) ([]model2.ReferrerStats, error) {
	return nil, nil
}

// GetPlatformStats implements the Store interface.
func (client *ClientMock) GetPlatformStats(string, ...any) (*model2.PlatformStats, error) {
	return nil, nil
}

// SelectLanguageStats implements the Store interface.
func (client *ClientMock) SelectLanguageStats(string, ...any) ([]model2.LanguageStats, error) {
	return nil, nil
}

// SelectCountryStats implements the Store interface.
func (client *ClientMock) SelectCountryStats(string, ...any) ([]model2.CountryStats, error) {
	return nil, nil
}

// SelectCityStats implements the Store interface.
func (client *ClientMock) SelectCityStats(string, ...any) ([]model2.CityStats, error) {
	return nil, nil
}

// SelectBrowserStats implements the Store interface.
func (client *ClientMock) SelectBrowserStats(string, ...any) ([]model2.BrowserStats, error) {
	return nil, nil
}

// SelectOSStats implements the Store interface.
func (client *ClientMock) SelectOSStats(string, ...any) ([]model2.OSStats, error) {
	return nil, nil
}

// SelectScreenClassStats implements the Store interface.
func (client *ClientMock) SelectScreenClassStats(string, ...any) ([]model2.ScreenClassStats, error) {
	return nil, nil
}

// SelectUTMSourceStats implements the Store interface.
func (client *ClientMock) SelectUTMSourceStats(string, ...any) ([]model2.UTMSourceStats, error) {
	return nil, nil
}

// SelectUTMMediumStats implements the Store interface.
func (client *ClientMock) SelectUTMMediumStats(string, ...any) ([]model2.UTMMediumStats, error) {
	return nil, nil
}

// SelectUTMCampaignStats implements the Store interface.
func (client *ClientMock) SelectUTMCampaignStats(string, ...any) ([]model2.UTMCampaignStats, error) {
	return nil, nil
}

// SelectUTMContentStats implements the Store interface.
func (client *ClientMock) SelectUTMContentStats(string, ...any) ([]model2.UTMContentStats, error) {
	return nil, nil
}

// SelectUTMTermStats implements the Store interface.
func (client *ClientMock) SelectUTMTermStats(string, ...any) ([]model2.UTMTermStats, error) {
	return nil, nil
}

// SelectOSVersionStats implements the Store interface.
func (client *ClientMock) SelectOSVersionStats(string, ...any) ([]model2.OSVersionStats, error) {
	return nil, nil
}

// SelectBrowserVersionStats implements the Store interface.
func (client *ClientMock) SelectBrowserVersionStats(string, ...any) ([]model2.BrowserVersionStats, error) {
	return nil, nil
}

// SelectOptions implements the Store interface.
func (client *ClientMock) SelectOptions(string, ...any) ([]string, error) {
	return nil, nil
}
