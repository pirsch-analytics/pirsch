package integration

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
	"github.com/stretchr/testify/assert"
)

func TestIntegrationRequests(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// set up a simple pipeline with bot filters
		p, s, _ := newPipe(t)

		// entry request
		assert.NoError(t, p.Process(&ingest.Request{
			ClientID:     1,
			Request:      newRequest(requestOptions{}),
			ScreenWidth:  1920,
			ScreenHeight: 1080,
			Title:        "Title",
			Tags:         map[string]string{"foo": "bar"},
		}))
		time.Sleep(time.Second * 10)

		// check that there is one pageview with a session bound to it
		sessions := s.Sessions()
		pageViews := s.PageViews()
		events := s.Events()
		assert.Len(t, sessions, 1)
		assert.Len(t, pageViews, 1)
		assert.Len(t, events, 0)

		// check the session
		assert.Equal(t, int8(1), sessions[0].Sign)
		assert.Equal(t, uint16(1), sessions[0].Version)
		assert.Equal(t, uint64(1), sessions[0].ClientID)
		assert.NotZero(t, sessions[0].VisitorID)
		assert.NotZero(t, sessions[0].SessionID)
		assert.Equal(t, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), sessions[0].Start)
		assert.Equal(t, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), sessions[0].Time)
		assert.Equal(t, uint32(0), sessions[0].DurationSeconds)
		assert.Equal(t, uint16(1), sessions[0].PageViews)
		assert.True(t, sessions[0].IsBounce)
		assert.Equal(t, "/", sessions[0].EntryPath)
		assert.Equal(t, "/", sessions[0].ExitPath)
		assert.Equal(t, "Title", sessions[0].EntryTitle)
		assert.Equal(t, "Title", sessions[0].ExitTitle)
		assert.Equal(t, uint16(0), sessions[0].Extended)
		assert.Equal(t, "example.com", sessions[0].Hostname)
		assert.Equal(t, "en", sessions[0].Language)
		assert.Equal(t, "gb", sessions[0].CountryCode)
		assert.Equal(t, "England", sessions[0].Region)
		assert.Equal(t, "London", sessions[0].City)
		assert.Equal(t, "https://google.com", sessions[0].Referrer)
		assert.Equal(t, "Google", sessions[0].ReferrerName)
		assert.Empty(t, sessions[0].ReferrerIcon)
		assert.Equal(t, pkg.OSWindows, sessions[0].OS)
		assert.Equal(t, "10", sessions[0].OSVersion)
		assert.Equal(t, pkg.BrowserChrome, sessions[0].Browser)
		assert.Equal(t, "147.0", sessions[0].BrowserVersion)
		assert.True(t, sessions[0].Desktop)
		assert.False(t, sessions[0].Mobile)
		assert.Equal(t, "Full HD", sessions[0].ScreenClass)
		assert.Equal(t, "Source", sessions[0].UTMSource)
		assert.Equal(t, "Medium", sessions[0].UTMMedium)
		assert.Equal(t, "Campaign", sessions[0].UTMCampaign)
		assert.Equal(t, "Content", sessions[0].UTMContent)
		assert.Equal(t, "Term", sessions[0].UTMTerm)
		assert.Equal(t, "Organic Search", sessions[0].Channel)

		// check the page view
		assert.Equal(t, uint64(1), pageViews[0].ClientID)
		assert.NotZero(t, pageViews[0].VisitorID)
		assert.NotZero(t, pageViews[0].SessionID)
		assert.Equal(t, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), pageViews[0].Time)
		assert.Equal(t, uint32(0), pageViews[0].DurationSeconds)
		assert.Equal(t, "example.com", pageViews[0].Hostname)
		assert.Equal(t, "/", pageViews[0].Path)
		assert.Equal(t, "Title", pageViews[0].Title)
		assert.Equal(t, "en", pageViews[0].Language)
		assert.Equal(t, "gb", pageViews[0].CountryCode)
		assert.Equal(t, "England", pageViews[0].Region)
		assert.Equal(t, "London", pageViews[0].City)
		assert.Equal(t, "https://google.com", pageViews[0].Referrer)
		assert.Equal(t, "Google", pageViews[0].ReferrerName)
		assert.Empty(t, pageViews[0].ReferrerIcon)
		assert.Equal(t, pkg.OSWindows, pageViews[0].OS)
		assert.Equal(t, "10", pageViews[0].OSVersion)
		assert.Equal(t, pkg.BrowserChrome, pageViews[0].Browser)
		assert.Equal(t, "147.0", pageViews[0].BrowserVersion)
		assert.True(t, pageViews[0].Desktop)
		assert.False(t, pageViews[0].Mobile)
		assert.Equal(t, "Full HD", pageViews[0].ScreenClass)
		assert.Equal(t, "Source", pageViews[0].UTMSource)
		assert.Equal(t, "Medium", pageViews[0].UTMMedium)
		assert.Equal(t, "Campaign", pageViews[0].UTMCampaign)
		assert.Equal(t, "Content", pageViews[0].UTMContent)
		assert.Equal(t, "Term", pageViews[0].UTMTerm)
		assert.Equal(t, "Organic Search", pageViews[0].Channel)
		assert.Equal(t, "bar", pageViews[0].Tags["foo"])

		// second request
		assert.NoError(t, p.Process(&ingest.Request{
			ClientID: 1,
			Request: newRequest(requestOptions{
				URL:      "https://example.com/pricing",
				Referrer: new(""),
			}),
			ScreenWidth:  1920,
			ScreenHeight: 1080,
			Title:        "Pricing",
		}))
		time.Sleep(time.Second * 20)

		// check that there are two pageviews with a session bound to them
		sessions = s.Sessions()
		pageViews = s.PageViews()
		events = s.Events()
		assert.Len(t, sessions, 3)
		assert.Len(t, pageViews, 2)
		assert.Len(t, events, 0)

		// check the session cancellation
		assert.Equal(t, int8(-1), sessions[1].Sign)
		assert.Equal(t, uint16(1), sessions[1].Version)
		assert.Equal(t, uint64(1), sessions[1].ClientID)
		assert.Equal(t, sessions[0].VisitorID, sessions[1].VisitorID)
		assert.Equal(t, sessions[0].SessionID, sessions[1].SessionID)
		assert.Equal(t, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), sessions[1].Start)
		assert.Equal(t, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), sessions[1].Time)
		assert.Equal(t, uint32(0), sessions[1].DurationSeconds)
		assert.Equal(t, uint16(1), sessions[1].PageViews)
		assert.True(t, sessions[1].IsBounce)
		assert.Equal(t, "/", sessions[1].EntryPath)
		assert.Equal(t, "/", sessions[1].ExitPath)
		assert.Equal(t, "Title", sessions[1].EntryTitle)
		assert.Equal(t, "Title", sessions[1].ExitTitle)
		assert.Equal(t, uint16(0), sessions[1].Extended)
		assert.Equal(t, "example.com", sessions[1].Hostname)
		assert.Equal(t, "en", sessions[1].Language)
		assert.Equal(t, "gb", sessions[1].CountryCode)
		assert.Equal(t, "England", sessions[1].Region)
		assert.Equal(t, "London", sessions[1].City)
		assert.Equal(t, "https://google.com", sessions[1].Referrer)
		assert.Equal(t, "Google", sessions[1].ReferrerName)
		assert.Empty(t, sessions[1].ReferrerIcon)
		assert.Equal(t, pkg.OSWindows, sessions[1].OS)
		assert.Equal(t, "10", sessions[1].OSVersion)
		assert.Equal(t, pkg.BrowserChrome, sessions[1].Browser)
		assert.Equal(t, "147.0", sessions[1].BrowserVersion)
		assert.True(t, sessions[1].Desktop)
		assert.False(t, sessions[1].Mobile)
		assert.Equal(t, "Full HD", sessions[1].ScreenClass)
		assert.Equal(t, "Source", sessions[1].UTMSource)
		assert.Equal(t, "Medium", sessions[1].UTMMedium)
		assert.Equal(t, "Campaign", sessions[1].UTMCampaign)
		assert.Equal(t, "Content", sessions[1].UTMContent)
		assert.Equal(t, "Term", sessions[1].UTMTerm)
		assert.Equal(t, "Organic Search", sessions[1].Channel)

		// check the session update
		assert.Equal(t, int8(1), sessions[2].Sign)
		assert.Equal(t, uint16(2), sessions[2].Version)
		assert.Equal(t, uint64(1), sessions[2].ClientID)
		assert.Equal(t, sessions[0].VisitorID, sessions[1].VisitorID)
		assert.Equal(t, sessions[0].SessionID, sessions[1].SessionID)
		assert.Equal(t, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), sessions[2].Start)
		assert.Equal(t, time.Date(2000, 1, 1, 0, 0, 10, 0, time.UTC), sessions[2].Time)
		assert.Equal(t, uint32(10), sessions[2].DurationSeconds)
		assert.Equal(t, uint16(2), sessions[2].PageViews)
		assert.False(t, sessions[2].IsBounce)
		assert.Equal(t, "/", sessions[2].EntryPath)
		assert.Equal(t, "/pricing", sessions[2].ExitPath)
		assert.Equal(t, "Title", sessions[2].EntryTitle)
		assert.Equal(t, "Pricing", sessions[2].ExitTitle)
		assert.Equal(t, uint16(0), sessions[2].Extended)
		assert.Equal(t, "example.com", sessions[2].Hostname)
		assert.Equal(t, "en", sessions[2].Language)
		assert.Equal(t, "gb", sessions[2].CountryCode)
		assert.Equal(t, "England", sessions[2].Region)
		assert.Equal(t, "London", sessions[2].City)
		assert.Equal(t, "https://google.com", sessions[2].Referrer)
		assert.Equal(t, "Google", sessions[2].ReferrerName)
		assert.Empty(t, sessions[2].ReferrerIcon)
		assert.Equal(t, pkg.OSWindows, sessions[2].OS)
		assert.Equal(t, "10", sessions[2].OSVersion)
		assert.Equal(t, pkg.BrowserChrome, sessions[2].Browser)
		assert.Equal(t, "147.0", sessions[2].BrowserVersion)
		assert.True(t, sessions[2].Desktop)
		assert.False(t, sessions[2].Mobile)
		assert.Equal(t, "Full HD", sessions[2].ScreenClass)
		assert.Equal(t, "Source", sessions[2].UTMSource)
		assert.Equal(t, "Medium", sessions[2].UTMMedium)
		assert.Equal(t, "Campaign", sessions[2].UTMCampaign)
		assert.Equal(t, "Content", sessions[2].UTMContent)
		assert.Equal(t, "Term", sessions[2].UTMTerm)
		assert.Equal(t, "Organic Search", sessions[2].Channel)

		// check the page view
		assert.Equal(t, uint64(1), pageViews[1].ClientID)
		assert.Equal(t, sessions[2].VisitorID, pageViews[1].VisitorID)
		assert.Equal(t, sessions[2].SessionID, pageViews[1].SessionID)
		assert.Equal(t, time.Date(2000, 1, 1, 0, 0, 10, 0, time.UTC), pageViews[1].Time)
		assert.Equal(t, uint32(10), pageViews[1].DurationSeconds)
		assert.Equal(t, "example.com", pageViews[1].Hostname)
		assert.Equal(t, "/pricing", pageViews[1].Path)
		assert.Equal(t, "Pricing", pageViews[1].Title)
		assert.Equal(t, "en", pageViews[1].Language)
		assert.Equal(t, "gb", pageViews[1].CountryCode)
		assert.Equal(t, "England", pageViews[1].Region)
		assert.Equal(t, "London", pageViews[1].City)
		assert.Equal(t, "https://google.com", pageViews[1].Referrer)
		assert.Equal(t, "Google", pageViews[1].ReferrerName)
		assert.Empty(t, pageViews[1].ReferrerIcon)
		assert.Equal(t, pkg.OSWindows, pageViews[1].OS)
		assert.Equal(t, "10", pageViews[1].OSVersion)
		assert.Equal(t, pkg.BrowserChrome, pageViews[1].Browser)
		assert.Equal(t, "147.0", pageViews[1].BrowserVersion)
		assert.True(t, pageViews[1].Desktop)
		assert.False(t, pageViews[1].Mobile)
		assert.Equal(t, "Full HD", pageViews[1].ScreenClass)
		assert.Equal(t, "Source", pageViews[1].UTMSource)
		assert.Equal(t, "Medium", pageViews[1].UTMMedium)
		assert.Equal(t, "Campaign", pageViews[1].UTMCampaign)
		assert.Equal(t, "Content", pageViews[1].UTMContent)
		assert.Equal(t, "Term", pageViews[1].UTMTerm)
		assert.Equal(t, "Organic Search", pageViews[1].Channel)
		assert.Empty(t, pageViews[1].Tags)

		// trigger an event on the second page
		assert.NoError(t, p.Process(&ingest.Request{
			ClientID: 1,
			Request: newRequest(requestOptions{
				URL:      "https://example.com/pricing",
				Referrer: new(""),
			}),
			ScreenWidth:  1920,
			ScreenHeight: 1080,
			Title:        "Pricing",
			EventName:    "Form Submission",
			EventMetaData: map[string]interface{}{
				"options_selected": []string{"module_a", "module_f"},
				"total_amount": struct {
					Currency    string
					AmountCents int `json:"amount_cents"`
				}{
					"USD",
					9998,
				},
			},
		}))
		time.Sleep(time.Second * 17)

		// check that there are two pageviews with a session bound to them
		sessions = s.Sessions()
		pageViews = s.PageViews()
		events = s.Events()
		assert.Len(t, sessions, 5)
		assert.Len(t, pageViews, 2)
		assert.Len(t, events, 1)

		// check the session time
		assert.Equal(t, time.Date(2000, 1, 1, 0, 0, 30, 0, time.UTC), sessions[4].Time)
		assert.Equal(t, uint32(30), sessions[4].DurationSeconds)

		// check the page view
		assert.Equal(t, uint64(1), events[0].ClientID)
		assert.Equal(t, sessions[2].VisitorID, events[0].VisitorID)
		assert.Equal(t, sessions[2].SessionID, events[0].SessionID)
		assert.Equal(t, time.Date(2000, 1, 1, 0, 0, 30, 0, time.UTC), events[0].Time)
		assert.Equal(t, "Form Submission", events[0].Name)
		assert.Equal(t, []string{"module_a", "module_f"}, events[0].MetaData["options_selected"])
		assert.Equal(t, struct {
			Currency    string
			AmountCents int `json:"amount_cents"`
		}{
			"USD",
			9998,
		}, events[0].MetaData["total_amount"])
		assert.Equal(t, "example.com", events[0].Hostname)
		assert.Equal(t, "/pricing", events[0].Path)
		assert.Equal(t, "Pricing", events[0].Title)
		assert.Equal(t, "en", events[0].Language)
		assert.Equal(t, "gb", events[0].CountryCode)
		assert.Equal(t, "England", events[0].Region)
		assert.Equal(t, "London", events[0].City)
		assert.Equal(t, "https://google.com", events[0].Referrer)
		assert.Equal(t, "Google", events[0].ReferrerName)
		assert.Empty(t, events[0].ReferrerIcon)
		assert.Equal(t, pkg.OSWindows, events[0].OS)
		assert.Equal(t, "10", events[0].OSVersion)
		assert.Equal(t, pkg.BrowserChrome, events[0].Browser)
		assert.Equal(t, "147.0", events[0].BrowserVersion)
		assert.True(t, events[0].Desktop)
		assert.False(t, events[0].Mobile)
		assert.Equal(t, "Full HD", events[0].ScreenClass)
		assert.Equal(t, "Source", events[0].UTMSource)
		assert.Equal(t, "Medium", events[0].UTMMedium)
		assert.Equal(t, "Campaign", events[0].UTMCampaign)
		assert.Equal(t, "Content", events[0].UTMContent)
		assert.Equal(t, "Term", events[0].UTMTerm)
		assert.Equal(t, "Organic Search", events[0].Channel)

		// third request
		assert.NoError(t, p.Process(&ingest.Request{
			ClientID: 1,
			Request: newRequest(requestOptions{
				URL:      "https://example.com/about",
				Referrer: new(""),
			}),
			ScreenWidth:  1920,
			ScreenHeight: 1080,
			Title:        "About",
		}))
		time.Sleep(time.Second * 20)

		// check that there are three pageviews with a session bound to them
		sessions = s.Sessions()
		pageViews = s.PageViews()
		events = s.Events()
		assert.Len(t, sessions, 7)
		assert.Len(t, pageViews, 3)
		assert.Len(t, events, 1)

		// check the time on page and some other important fields
		assert.Equal(t, uint64(1), pageViews[2].ClientID)
		assert.Equal(t, sessions[2].VisitorID, pageViews[2].VisitorID)
		assert.Equal(t, sessions[2].SessionID, pageViews[2].SessionID)
		assert.Equal(t, time.Date(2000, 1, 1, 0, 0, 47, 0, time.UTC), pageViews[2].Time)
		assert.Equal(t, uint32(17), pageViews[2].DurationSeconds)
		assert.Equal(t, "/about", pageViews[2].Path)
		assert.Equal(t, "About", pageViews[2].Title)

		// stop the pipeline
		p.Stop()
	})
}

func TestIntegrationConcurrency(t *testing.T) {
	// TODO
}

func TestIntegrationOverwriteTimeAndOrder(t *testing.T) {
	// TODO
}

func TestIntegrationExtendSession(t *testing.T) {
	// TODO
}

func TestIntegrationResetSession(t *testing.T) {
	// TODO
}

func TestIntegrationEventNonInteractive(t *testing.T) {
	// TODO
}
