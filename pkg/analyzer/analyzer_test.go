package analyzer

import (
	"context"
	"github.com/pirsch-analytics/pirsch/v6/pkg"
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"github.com/pirsch-analytics/pirsch/v6/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAnalyzer_NoData(t *testing.T) {
	db.CleanupDB(t, dbClient)
	analyzer := NewAnalyzer(dbClient)
	_, _, err := analyzer.Visitors.Active(nil, time.Minute*15)
	assert.NoError(t, err)
	_, err = analyzer.Visitors.Total(nil)
	assert.NoError(t, err)
	_, err = analyzer.Visitors.ByPeriod(nil)
	assert.NoError(t, err)
	_, err = analyzer.Visitors.Growth(&Filter{From: util.PastDay(7), To: util.Today()})
	assert.NoError(t, err)
	_, err = analyzer.Visitors.ByHour(nil)
	assert.NoError(t, err)
	_, err = analyzer.Visitors.Referrer(nil)
	assert.NoError(t, err)
	_, err = analyzer.Pages.ByPath(nil)
	assert.NoError(t, err)
	_, err = analyzer.Pages.Entry(nil)
	assert.NoError(t, err)
	_, err = analyzer.Pages.Exit(nil)
	assert.NoError(t, err)
	_, err = analyzer.Pages.Conversions(nil)
	assert.NoError(t, err)
	_, err = analyzer.Events.Events(nil)
	assert.NoError(t, err)
	_, err = analyzer.Events.Breakdown(&Filter{EventName: []string{"event"}})
	assert.NoError(t, err)
	_, err = analyzer.Events.List(nil)
	assert.NoError(t, err)
	_, err = analyzer.Device.Platform(nil)
	assert.NoError(t, err)
	_, err = analyzer.Device.Browser(nil)
	assert.NoError(t, err)
	_, err = analyzer.Device.OS(nil)
	assert.NoError(t, err)
	_, err = analyzer.Device.OSVersion(nil)
	assert.NoError(t, err)
	_, err = analyzer.Device.BrowserVersion(nil)
	assert.NoError(t, err)
	_, err = analyzer.Device.ScreenClass(nil)
	assert.NoError(t, err)
	_, err = analyzer.Demographics.Languages(nil)
	assert.NoError(t, err)
	_, err = analyzer.Demographics.Countries(nil)
	assert.NoError(t, err)
	_, err = analyzer.Demographics.Regions(nil)
	assert.NoError(t, err)
	_, err = analyzer.Demographics.Cities(nil)
	assert.NoError(t, err)
	_, err = analyzer.Time.AvgSessionDuration(nil)
	assert.NoError(t, err)
	_, err = analyzer.Time.AvgTimeOnPage(nil)
	assert.NoError(t, err)
	_, err = analyzer.Tags.Keys(nil)
	assert.NoError(t, err)
	_, err = analyzer.Tags.Breakdown(nil)
	assert.NoError(t, err)
}

func getMaxFilter(eventName string) *Filter {
	var events []string

	if eventName != "" {
		events = append(events, eventName)
	}

	return &Filter{
		Ctx:            context.Background(),
		ClientID:       42,
		From:           util.PastDay(5),
		To:             util.PastDay(2),
		Path:           []string{"/path"},
		EntryPath:      []string{"/entry"},
		ExitPath:       []string{"/exit"},
		Language:       []string{"en"},
		Country:        []string{"en"},
		Region:         []string{"England"},
		City:           []string{"London"},
		Referrer:       []string{"ref"},
		ReferrerName:   []string{"refname"},
		OS:             []string{pkg.OSWindows},
		OSVersion:      []string{"10"},
		Browser:        []string{pkg.BrowserChrome},
		BrowserVersion: []string{"90"},
		Platform:       pkg.PlatformDesktop,
		ScreenClass:    []string{"XL"},
		UTMSource:      []string{"source"},
		UTMMedium:      []string{"medium"},
		UTMCampaign:    []string{"campaign"},
		UTMContent:     []string{"content"},
		UTMTerm:        []string{"term"},
		Tags:           map[string]string{"key": "value"},
		EventName:      events,
		Limit:          42,
		IncludeCR:      true,
	}
}

func saveSessions(t *testing.T, sessions [][]model.Session) {
	for _, entries := range sessions {
		assert.NoError(t, dbClient.SaveSessions(entries))
		time.Sleep(time.Millisecond * 100)
	}
}
