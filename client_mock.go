package pirsch

import (
	"sync"
	"time"
)

// ClientMock is a mock Store implementation.
type ClientMock struct {
	PageViews     []PageView
	Sessions      []Session
	Events        []Event
	UserAgents    []UserAgent
	ReturnSession *Session
	m             sync.Mutex
}

// NewMockClient returns a new mock client.
func NewMockClient() *ClientMock {
	return &ClientMock{
		PageViews:  make([]PageView, 0),
		Sessions:   make([]Session, 0),
		Events:     make([]Event, 0),
		UserAgents: make([]UserAgent, 0),
	}
}

// SavePageViews implements the Store interface.
func (client *ClientMock) SavePageViews(pageViews []PageView) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.PageViews = append(client.PageViews, pageViews...)
	return nil
}

// SaveSessions implements the Store interface.
func (client *ClientMock) SaveSessions(sessions []Session) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.Sessions = append(client.Sessions, sessions...)
	return nil
}

// SaveEvents implements the Store interface.
func (client *ClientMock) SaveEvents(events []Event) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.Events = append(client.Events, events...)
	return nil
}

// SaveUserAgents implements the Store interface.
func (client *ClientMock) SaveUserAgents(userAgents []UserAgent) error {
	client.m.Lock()
	defer client.m.Unlock()
	client.UserAgents = append(client.UserAgents, userAgents...)
	return nil
}

// Session implements the Store interface.
func (client *ClientMock) Session(uint64, uint64, time.Time) (*Session, error) {
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
func (client *ClientMock) SelectActiveVisitorStats(bool, string, ...any) ([]ActiveVisitorStats, error) {
	return nil, nil
}

// GetTotalVisitorStats implements the Store interface.
func (client *ClientMock) GetTotalVisitorStats(string, ...any) (*TotalVisitorStats, error) {
	return nil, nil
}

// SelectVisitorStats implements the Store interface.
func (client *ClientMock) SelectVisitorStats(Period, string, ...any) ([]VisitorStats, error) {
	return nil, nil
}

// SelectTimeSpentStats implements the Store interface.
func (client *ClientMock) SelectTimeSpentStats(Period, string, ...any) ([]TimeSpentStats, error) {
	return nil, nil
}

// GetGrowthStats implements the Store interface.
func (client *ClientMock) GetGrowthStats(string, ...any) (*GrowthStats, error) {
	return nil, nil
}

// SelectVisitorHourStats implements the Store interface.
func (client *ClientMock) SelectVisitorHourStats(string, ...any) ([]VisitorHourStats, error) {
	return nil, nil
}

// SelectPageStats implements the Store interface.
func (client *ClientMock) SelectPageStats(bool, bool, string, ...any) ([]PageStats, error) {
	return nil, nil
}

// SelectAvgTimeSpentStats implements the Store interface.
func (client *ClientMock) SelectAvgTimeSpentStats(string, ...any) ([]AvgTimeSpentStats, error) {
	return nil, nil
}

// SelectEntryStats implements the Store interface.
func (client *ClientMock) SelectEntryStats(bool, string, ...any) ([]EntryStats, error) {
	return nil, nil
}

// SelectExitStats implements the Store interface.
func (client *ClientMock) SelectExitStats(bool, string, ...any) ([]ExitStats, error) {
	return nil, nil
}

// SelectTotalVisitorSessionStats implements the Store interface.
func (client *ClientMock) SelectTotalVisitorSessionStats(string, ...any) ([]TotalVisitorSessionStats, error) {
	return nil, nil
}

// GetPageConversionsStats implements the Store interface.
func (client *ClientMock) GetPageConversionsStats(string, ...any) (*PageConversionsStats, error) {
	return nil, nil
}

// SelectEventStats implements the Store interface.
func (client *ClientMock) SelectEventStats(bool, string, ...any) ([]EventStats, error) {
	return nil, nil
}

// SelectEventListStats implements the Store interface.
func (client *ClientMock) SelectEventListStats(string, ...any) ([]EventListStats, error) {
	return nil, nil
}

// SelectReferrerStats implements the Store interface.
func (client *ClientMock) SelectReferrerStats(string, ...any) ([]ReferrerStats, error) {
	return nil, nil
}

// GetPlatformStats implements the Store interface.
func (client *ClientMock) GetPlatformStats(string, ...any) (*PlatformStats, error) {
	return nil, nil
}

// SelectLanguageStats implements the Store interface.
func (client *ClientMock) SelectLanguageStats(string, ...any) ([]LanguageStats, error) {
	return nil, nil
}

// SelectCountryStats implements the Store interface.
func (client *ClientMock) SelectCountryStats(string, ...any) ([]CountryStats, error) {
	return nil, nil
}

// SelectCityStats implements the Store interface.
func (client *ClientMock) SelectCityStats(string, ...any) ([]CityStats, error) {
	return nil, nil
}

// SelectBrowserStats implements the Store interface.
func (client *ClientMock) SelectBrowserStats(string, ...any) ([]BrowserStats, error) {
	return nil, nil
}

// SelectOSStats implements the Store interface.
func (client *ClientMock) SelectOSStats(string, ...any) ([]OSStats, error) {
	return nil, nil
}

// SelectScreenClassStats implements the Store interface.
func (client *ClientMock) SelectScreenClassStats(string, ...any) ([]ScreenClassStats, error) {
	return nil, nil
}

// SelectUTMSourceStats implements the Store interface.
func (client *ClientMock) SelectUTMSourceStats(string, ...any) ([]UTMSourceStats, error) {
	return nil, nil
}

// SelectUTMMediumStats implements the Store interface.
func (client *ClientMock) SelectUTMMediumStats(string, ...any) ([]UTMMediumStats, error) {
	return nil, nil
}

// SelectUTMCampaignStats implements the Store interface.
func (client *ClientMock) SelectUTMCampaignStats(string, ...any) ([]UTMCampaignStats, error) {
	return nil, nil
}

// SelectUTMContentStats implements the Store interface.
func (client *ClientMock) SelectUTMContentStats(string, ...any) ([]UTMContentStats, error) {
	return nil, nil
}

// SelectUTMTermStats implements the Store interface.
func (client *ClientMock) SelectUTMTermStats(string, ...any) ([]UTMTermStats, error) {
	return nil, nil
}

// SelectOSVersionStats implements the Store interface.
func (client *ClientMock) SelectOSVersionStats(string, ...any) ([]OSVersionStats, error) {
	return nil, nil
}

// SelectBrowserVersionStats implements the Store interface.
func (client *ClientMock) SelectBrowserVersionStats(string, ...any) ([]BrowserVersionStats, error) {
	return nil, nil
}
