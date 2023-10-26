package analyzer

import (
	"context"
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"github.com/pirsch-analytics/pirsch/v6/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFilterOptions_Pages(t *testing.T) {
	db.CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SavePageViews(context.Background(), []model.PageView{
		{VisitorID: 1, SessionID: 1, Time: util.PastDay(4), Path: "/"},
		{VisitorID: 1, SessionID: 1, Time: util.PastDay(2), Path: "/foo"},
		{VisitorID: 1, SessionID: 2, Time: util.PastDay(2), Path: "/foo"},
		{VisitorID: 1, SessionID: 1, Time: util.PastDay(1), Path: "/bar"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	options, err := analyzer.Options.Pages(nil)
	assert.NoError(t, err)
	assert.Len(t, options, 3)
	assert.Equal(t, "/", options[0])
	assert.Equal(t, "/bar", options[1])
	assert.Equal(t, "/foo", options[2])
	options, err = analyzer.Options.Pages(&Filter{From: util.PastDay(3), To: util.Today()})
	assert.NoError(t, err)
	assert.Len(t, options, 2)
	assert.Equal(t, "/bar", options[0])
	assert.Equal(t, "/foo", options[1])
}

func TestFilterOptions_Referrer(t *testing.T) {
	db.CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SaveSessions(context.Background(), []model.Session{
		{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.PastDay(4), Start: util.PastDay(4), Referrer: "https://google.com", ReferrerName: "Google"},
		{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.PastDay(2), Start: util.PastDay(2), Referrer: "https://pirsch.io", ReferrerName: "Pirsch"},
		{Sign: 1, VisitorID: 1, SessionID: 2, Time: util.PastDay(2), Start: util.PastDay(2), Referrer: "https://pirsch.io", ReferrerName: "Pirsch"},
		{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.PastDay(1), Start: util.PastDay(1), Referrer: "https://twitter.com", ReferrerName: "Twitter"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	options, err := analyzer.Options.Referrer(nil)
	assert.NoError(t, err)
	assert.Len(t, options, 3)
	assert.Equal(t, "https://google.com", options[0])
	assert.Equal(t, "https://pirsch.io", options[1])
	assert.Equal(t, "https://twitter.com", options[2])
	options, err = analyzer.Options.Referrer(&Filter{From: util.PastDay(3), To: util.Today()})
	assert.NoError(t, err)
	assert.Len(t, options, 2)
	assert.Equal(t, "https://pirsch.io", options[0])
	assert.Equal(t, "https://twitter.com", options[1])
	options, err = analyzer.Options.ReferrerName(nil)
	assert.NoError(t, err)
	assert.Len(t, options, 3)
	assert.Equal(t, "Google", options[0])
	assert.Equal(t, "Pirsch", options[1])
	assert.Equal(t, "Twitter", options[2])
	options, err = analyzer.Options.ReferrerName(&Filter{From: util.PastDay(3), To: util.Today()})
	assert.NoError(t, err)
	assert.Len(t, options, 2)
	assert.Equal(t, "Pirsch", options[0])
	assert.Equal(t, "Twitter", options[1])
}

func TestFilterOptions_UTM(t *testing.T) {
	db.CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SaveSessions(context.Background(), []model.Session{
		{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.PastDay(4), Start: util.PastDay(4), UTMSource: "source", UTMMedium: "medium", UTMCampaign: "campaign", UTMContent: "content", UTMTerm: "term"},
		{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.PastDay(2), Start: util.PastDay(2), UTMSource: "foo", UTMMedium: "foo", UTMCampaign: "foo", UTMContent: "foo", UTMTerm: "foo"},
		{Sign: 1, VisitorID: 1, SessionID: 2, Time: util.PastDay(2), Start: util.PastDay(2), UTMSource: "foo", UTMMedium: "foo", UTMCampaign: "foo", UTMContent: "foo", UTMTerm: "foo"},
		{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.PastDay(1), Start: util.PastDay(1), UTMSource: "bar", UTMMedium: "bar", UTMCampaign: "bar", UTMContent: "bar", UTMTerm: "bar"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	utmSource, err := analyzer.Options.UTMSource(nil)
	assert.NoError(t, err)
	assert.Len(t, utmSource, 3)
	assert.Equal(t, "bar", utmSource[0])
	assert.Equal(t, "foo", utmSource[1])
	assert.Equal(t, "source", utmSource[2])
	utmSource, err = analyzer.Options.UTMSource(&Filter{From: util.PastDay(3), To: util.Today()})
	assert.NoError(t, err)
	assert.Len(t, utmSource, 2)
	assert.Equal(t, "bar", utmSource[0])
	assert.Equal(t, "foo", utmSource[1])
	utmMedium, err := analyzer.Options.UTMMedium(nil)
	assert.NoError(t, err)
	assert.Len(t, utmMedium, 3)
	assert.Equal(t, "bar", utmMedium[0])
	assert.Equal(t, "foo", utmMedium[1])
	assert.Equal(t, "medium", utmMedium[2])
	utmMedium, err = analyzer.Options.UTMMedium(&Filter{From: util.PastDay(3), To: util.Today()})
	assert.NoError(t, err)
	assert.Len(t, utmMedium, 2)
	assert.Equal(t, "bar", utmMedium[0])
	assert.Equal(t, "foo", utmMedium[1])
	utmCampaign, err := analyzer.Options.UTMCampaign(nil)
	assert.NoError(t, err)
	assert.Len(t, utmCampaign, 3)
	assert.Equal(t, "bar", utmCampaign[0])
	assert.Equal(t, "campaign", utmCampaign[1])
	assert.Equal(t, "foo", utmCampaign[2])
	utmCampaign, err = analyzer.Options.UTMCampaign(&Filter{From: util.PastDay(3), To: util.Today()})
	assert.NoError(t, err)
	assert.Len(t, utmCampaign, 2)
	assert.Equal(t, "bar", utmCampaign[0])
	assert.Equal(t, "foo", utmCampaign[1])
	utmContent, err := analyzer.Options.UTMContent(nil)
	assert.NoError(t, err)
	assert.Len(t, utmContent, 3)
	assert.Equal(t, "bar", utmContent[0])
	assert.Equal(t, "content", utmContent[1])
	assert.Equal(t, "foo", utmContent[2])
	utmContent, err = analyzer.Options.UTMContent(&Filter{From: util.PastDay(3), To: util.Today()})
	assert.NoError(t, err)
	assert.Len(t, utmContent, 2)
	assert.Equal(t, "bar", utmContent[0])
	assert.Equal(t, "foo", utmContent[1])
	utmTerm, err := analyzer.Options.UTMTerm(nil)
	assert.NoError(t, err)
	assert.Len(t, utmTerm, 3)
	assert.Equal(t, "bar", utmTerm[0])
	assert.Equal(t, "foo", utmTerm[1])
	assert.Equal(t, "term", utmTerm[2])
	utmTerm, err = analyzer.Options.UTMTerm(&Filter{From: util.PastDay(3), To: util.Today()})
	assert.NoError(t, err)
	assert.Len(t, utmTerm, 2)
	assert.Equal(t, "bar", utmTerm[0])
	assert.Equal(t, "foo", utmTerm[1])
}

func TestFilterOptions_Events(t *testing.T) {
	db.CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SaveEvents(context.Background(), []model.Event{
		{VisitorID: 1, SessionID: 1, Time: util.PastDay(4), Name: "event"},
		{VisitorID: 1, SessionID: 1, Time: util.PastDay(2), Name: "foo"},
		{VisitorID: 1, SessionID: 2, Time: util.PastDay(2), Name: "foo"},
		{VisitorID: 1, SessionID: 1, Time: util.PastDay(1), Name: "bar"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	options, err := analyzer.Options.Events(nil)
	assert.NoError(t, err)
	assert.Len(t, options, 3)
	assert.Equal(t, "bar", options[0])
	assert.Equal(t, "event", options[1])
	assert.Equal(t, "foo", options[2])
	options, err = analyzer.Options.Events(&Filter{From: util.PastDay(3), To: util.Today()})
	assert.NoError(t, err)
	assert.Len(t, options, 2)
	assert.Equal(t, "bar", options[0])
	assert.Equal(t, "foo", options[1])
}

func TestFilterOptions_Countries(t *testing.T) {
	db.CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SaveSessions(context.Background(), []model.Session{
		{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.PastDay(4), Start: util.PastDay(4), CountryCode: "us"},
		{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.PastDay(2), Start: util.PastDay(2), CountryCode: "ja"},
		{Sign: 1, VisitorID: 1, SessionID: 2, Time: util.PastDay(2), Start: util.PastDay(2), CountryCode: "ja"},
		{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.PastDay(1), Start: util.PastDay(1), CountryCode: "de"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	options, err := analyzer.Options.Countries(nil)
	assert.NoError(t, err)
	assert.Len(t, options, 3)
	assert.Equal(t, "de", options[0])
	assert.Equal(t, "ja", options[1])
	assert.Equal(t, "us", options[2])
	options, err = analyzer.Options.Countries(&Filter{From: util.PastDay(3), To: util.Today()})
	assert.NoError(t, err)
	assert.Len(t, options, 2)
	assert.Equal(t, "de", options[0])
	assert.Equal(t, "ja", options[1])
}

func TestFilterOptions_Cities(t *testing.T) {
	db.CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SaveSessions(context.Background(), []model.Session{
		{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.PastDay(4), Start: util.PastDay(4), City: "Boston"},
		{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.PastDay(2), Start: util.PastDay(2), City: "Tokio"},
		{Sign: 1, VisitorID: 1, SessionID: 2, Time: util.PastDay(2), Start: util.PastDay(2), City: "Tokio"},
		{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.PastDay(1), Start: util.PastDay(1), City: "Berlin"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	options, err := analyzer.Options.Cities(nil)
	assert.NoError(t, err)
	assert.Len(t, options, 3)
	assert.Equal(t, "Berlin", options[0])
	assert.Equal(t, "Boston", options[1])
	assert.Equal(t, "Tokio", options[2])
	options, err = analyzer.Options.Cities(&Filter{From: util.PastDay(3), To: util.Today()})
	assert.NoError(t, err)
	assert.Len(t, options, 2)
	assert.Equal(t, "Berlin", options[0])
	assert.Equal(t, "Tokio", options[1])
}

func TestFilterOptions_Languages(t *testing.T) {
	db.CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SaveSessions(context.Background(), []model.Session{
		{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.PastDay(4), Start: util.PastDay(4), Language: "en"},
		{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.PastDay(2), Start: util.PastDay(2), Language: "ja"},
		{Sign: 1, VisitorID: 1, SessionID: 2, Time: util.PastDay(2), Start: util.PastDay(2), Language: "ja"},
		{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.PastDay(1), Start: util.PastDay(1), Language: "de"},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	utmSource, err := analyzer.Options.Languages(nil)
	assert.NoError(t, err)
	assert.Len(t, utmSource, 3)
	assert.Equal(t, "de", utmSource[0])
	assert.Equal(t, "en", utmSource[1])
	assert.Equal(t, "ja", utmSource[2])
	utmSource, err = analyzer.Options.Languages(&Filter{From: util.PastDay(3), To: util.Today()})
	assert.NoError(t, err)
	assert.Len(t, utmSource, 2)
	assert.Equal(t, "de", utmSource[0])
	assert.Equal(t, "ja", utmSource[1])
}

func TestFilterOptions_EventMetadataValues(t *testing.T) {
	db.CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SaveEvents(context.Background(), []model.Event{
		{VisitorID: 1, SessionID: 1, Time: util.PastDay(4), Name: "event0", MetaKeys: []string{"key0", "key1"}, MetaValues: []string{"val0", "val1"}},
		{VisitorID: 1, SessionID: 1, Time: util.PastDay(2), Name: "event1"},
		{VisitorID: 1, SessionID: 2, Time: util.PastDay(2), Name: "event2", MetaKeys: []string{"key0", "key1"}, MetaValues: []string{"val1", "val2"}},
		{VisitorID: 1, SessionID: 1, Time: util.PastDay(1), Name: "event3", MetaKeys: []string{"key1", "key2"}, MetaValues: []string{"val1", "val3"}},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	options, err := analyzer.Options.EventMetadataValues(nil)
	assert.NoError(t, err)
	assert.Len(t, options, 0)
	options, err = analyzer.Options.EventMetadataValues(&Filter{From: util.PastDay(3), To: util.Today(), EventName: []string{"event0", "event1", "event2", "event3"}})
	assert.NoError(t, err)
	assert.Len(t, options, 3)
	assert.Equal(t, "val1", options[0])
	assert.Equal(t, "val2", options[1])
	assert.Equal(t, "val3", options[2])
	options, err = analyzer.Options.EventMetadataValues(&Filter{
		EventName: []string{"event3"},
		Path:      []string{"/unknown"},
	})
	assert.NoError(t, err)
	assert.Len(t, options, 2)
	assert.Equal(t, "val1", options[0])
	assert.Equal(t, "val3", options[1])
}
