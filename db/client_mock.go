package db

import (
	"github.com/pirsch-analytics/pirsch/v4"
	"github.com/pirsch-analytics/pirsch/v4/model"
	"sync"
	"time"
)

// ClientMock is a mock Store implementation.
type ClientMock struct {
	pageViews     []model.PageView
	sessions      []model.Session
	events        []model.Event
	userAgents    []model.UserAgent
	ReturnSession *model.Session
	m             sync.Mutex
}

// NewMockClient returns a new mock client.
func NewMockClient() *ClientMock {
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
	return data
}

// GetSessions returns a copy of the sessions slice.
func (client *ClientMock) GetSessions() []model.Session {
	client.m.Lock()
	defer client.m.Unlock()
	data := make([]model.Session, len(client.sessions))
	copy(data, client.sessions)
	return data
}

// GetEvents returns a copy of the events slice.
func (client *ClientMock) GetEvents() []model.Event {
	client.m.Lock()
	defer client.m.Unlock()
	data := make([]model.Event, len(client.events))
	copy(data, client.events)
	return data
}

// GetUserAgents returns a copy of the user agents slice.
func (client *ClientMock) GetUserAgents() []model.UserAgent {
	client.m.Lock()
	defer client.m.Unlock()
	data := make([]model.UserAgent, len(client.userAgents))
	copy(data, client.userAgents)
	return data
}

// SavePageViews implements the Store interface.
func (client *ClientMock) SavePageViews(pageViews []model.PageView) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.pageViews = append(client.pageViews, pageViews...)
	return nil
}

// SaveSessions implements the Store interface.
func (client *ClientMock) SaveSessions(sessions []model.Session) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.sessions = append(client.sessions, sessions...)
	return nil
}

// SaveEvents implements the Store interface.
func (client *ClientMock) SaveEvents(events []model.Event) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.events = append(client.events, events...)
	return nil
}

// SaveUserAgents implements the Store interface.
func (client *ClientMock) SaveUserAgents(userAgents []model.UserAgent) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.userAgents = append(client.userAgents, userAgents...)
	return nil
}

// Session implements the Store interface.
func (client *ClientMock) Session(uint64, uint64, time.Time) (*model.Session, error) {
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
func (client *ClientMock) SelectActiveVisitorStats(bool, string, ...any) ([]model.ActiveVisitorStats, error) {
	return nil, nil
}

// GetTotalVisitorStats implements the Store interface.
func (client *ClientMock) GetTotalVisitorStats(string, ...any) (*model.TotalVisitorStats, error) {
	return nil, nil
}

// SelectVisitorStats implements the Store interface.
func (client *ClientMock) SelectVisitorStats(pirsch.Period, string, ...any) ([]model.VisitorStats, error) {
	return nil, nil
}

// SelectTimeSpentStats implements the Store interface.
func (client *ClientMock) SelectTimeSpentStats(pirsch.Period, string, ...any) ([]model.TimeSpentStats, error) {
	return nil, nil
}

// GetGrowthStats implements the Store interface.
func (client *ClientMock) GetGrowthStats(string, ...any) (*model.GrowthStats, error) {
	return nil, nil
}

// SelectVisitorHourStats implements the Store interface.
func (client *ClientMock) SelectVisitorHourStats(string, ...any) ([]model.VisitorHourStats, error) {
	return nil, nil
}

// SelectPageStats implements the Store interface.
func (client *ClientMock) SelectPageStats(bool, bool, string, ...any) ([]model.PageStats, error) {
	return nil, nil
}

// SelectAvgTimeSpentStats implements the Store interface.
func (client *ClientMock) SelectAvgTimeSpentStats(string, ...any) ([]model.AvgTimeSpentStats, error) {
	return nil, nil
}

// SelectEntryStats implements the Store interface.
func (client *ClientMock) SelectEntryStats(bool, string, ...any) ([]model.EntryStats, error) {
	return nil, nil
}

// SelectExitStats implements the Store interface.
func (client *ClientMock) SelectExitStats(bool, string, ...any) ([]model.ExitStats, error) {
	return nil, nil
}

// SelectTotalSessions implements the Store interface.
func (client *ClientMock) SelectTotalSessions(string, ...any) (int, error) {
	return 0, nil
}

// SelectTotalVisitorSessionStats implements the Store interface.
func (client *ClientMock) SelectTotalVisitorSessionStats(string, ...any) ([]model.TotalVisitorSessionStats, error) {
	return nil, nil
}

// GetPageConversionsStats implements the Store interface.
func (client *ClientMock) GetPageConversionsStats(string, ...any) (*model.PageConversionsStats, error) {
	return nil, nil
}

// SelectEventStats implements the Store interface.
func (client *ClientMock) SelectEventStats(bool, string, ...any) ([]model.EventStats, error) {
	return nil, nil
}

// SelectEventListStats implements the Store interface.
func (client *ClientMock) SelectEventListStats(string, ...any) ([]model.EventListStats, error) {
	return nil, nil
}

// SelectReferrerStats implements the Store interface.
func (client *ClientMock) SelectReferrerStats(string, ...any) ([]model.ReferrerStats, error) {
	return nil, nil
}

// GetPlatformStats implements the Store interface.
func (client *ClientMock) GetPlatformStats(string, ...any) (*model.PlatformStats, error) {
	return nil, nil
}

// SelectLanguageStats implements the Store interface.
func (client *ClientMock) SelectLanguageStats(string, ...any) ([]model.LanguageStats, error) {
	return nil, nil
}

// SelectCountryStats implements the Store interface.
func (client *ClientMock) SelectCountryStats(string, ...any) ([]model.CountryStats, error) {
	return nil, nil
}

// SelectCityStats implements the Store interface.
func (client *ClientMock) SelectCityStats(string, ...any) ([]model.CityStats, error) {
	return nil, nil
}

// SelectBrowserStats implements the Store interface.
func (client *ClientMock) SelectBrowserStats(string, ...any) ([]model.BrowserStats, error) {
	return nil, nil
}

// SelectOSStats implements the Store interface.
func (client *ClientMock) SelectOSStats(string, ...any) ([]model.OSStats, error) {
	return nil, nil
}

// SelectScreenClassStats implements the Store interface.
func (client *ClientMock) SelectScreenClassStats(string, ...any) ([]model.ScreenClassStats, error) {
	return nil, nil
}

// SelectUTMSourceStats implements the Store interface.
func (client *ClientMock) SelectUTMSourceStats(string, ...any) ([]model.UTMSourceStats, error) {
	return nil, nil
}

// SelectUTMMediumStats implements the Store interface.
func (client *ClientMock) SelectUTMMediumStats(string, ...any) ([]model.UTMMediumStats, error) {
	return nil, nil
}

// SelectUTMCampaignStats implements the Store interface.
func (client *ClientMock) SelectUTMCampaignStats(string, ...any) ([]model.UTMCampaignStats, error) {
	return nil, nil
}

// SelectUTMContentStats implements the Store interface.
func (client *ClientMock) SelectUTMContentStats(string, ...any) ([]model.UTMContentStats, error) {
	return nil, nil
}

// SelectUTMTermStats implements the Store interface.
func (client *ClientMock) SelectUTMTermStats(string, ...any) ([]model.UTMTermStats, error) {
	return nil, nil
}

// SelectOSVersionStats implements the Store interface.
func (client *ClientMock) SelectOSVersionStats(string, ...any) ([]model.OSVersionStats, error) {
	return nil, nil
}

// SelectBrowserVersionStats implements the Store interface.
func (client *ClientMock) SelectBrowserVersionStats(string, ...any) ([]model.BrowserVersionStats, error) {
	return nil, nil
}

// SelectOptions implements the Store interface.
func (client *ClientMock) SelectOptions(string, ...any) ([]string, error) {
	return nil, nil
}
