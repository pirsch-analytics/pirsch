package db

import (
	"testing"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/_pkg/util"
	"github.com/pirsch-analytics/pirsch/v7/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	client, err := NewClickHouse(&ClickHouseConfig{
		Hostnames:     []string{"127.0.0.1"},
		Port:          9000,
		Database:      "pirschtest",
		Password:      "default",
		SSLSkipVerify: true,
		Debug:         true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NoError(t, client.DB.Ping())
}

func TestClient_SaveSessions(t *testing.T) {
	CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SaveSessions([]model.Session{
		{
			Data: model.Data{
				ClientID:       1,
				VisitorID:      1,
				Time:           time.Now(),
				Start:          time.Now(),
				SessionID:      util.RandUint32(),
				Hostname:       "example.com",
				PageViews:      7,
				IsBounce:       true,
				Language:       "en",
				Referrer:       "ref",
				ReferrerName:   "ref_name",
				ReferrerIcon:   "ref_icon",
				OS:             "os",
				OSVersion:      "10",
				Browser:        "browser",
				BrowserVersion: "89",
				CountryCode:    "en",
				Region:         "England",
				City:           "London",
				Desktop:        true,
				Mobile:         false,
				ScreenClass:    "XL",
				UTMCampaign:    "Campaign",
				UTMMedium:      "Medium",
				UTMSource:      "Source",
				UTMContent:     "Content",
				UTMTerm:        "Term",
				Channel:        "Channel",
			},
			Sign:       1,
			Version:    1,
			EntryPath:  "/entry-path",
			ExitPath:   "/exit-path",
			EntryTitle: "entry-title",
			ExitTitle:  "exit_title",
			Extended:   123,
		},
		{
			Data: model.Data{
				VisitorID: 1,
				Time:      time.Now().UTC(),
				Start:     time.Now(),
			},
			Sign:     -1,
			Version:  1,
			ExitPath: "/path",
		},
	}))
}

func TestClient_SaveSessionsBatch(t *testing.T) {
	CleanupDB(t, dbClient)
	sessions := make([]model.Session, 0, 101)

	for i := range 50 {
		now := time.Now().Add(time.Millisecond * time.Duration(i))
		sessions = append(sessions, model.Session{
			Data: model.Data{
				ClientID:  2,
				VisitorID: 3,
				Time:      now,
				Start:     now,
				SessionID: 4,
				PageViews: uint16(i + 1),
			},
			Sign:    1,
			Version: 1,
		})
		sessions = append(sessions, model.Session{
			Data: model.Data{
				ClientID:  2,
				VisitorID: 3,
				Time:      now,
				Start:     now,
				SessionID: 4,
				PageViews: uint16(i + 1),
			},
			Sign:    -1,
			Version: 1,
		})
	}

	sessions = append(sessions, model.Session{
		Data: model.Data{
			ClientID:  2,
			VisitorID: 3,
			Time:      time.Now().Add(time.Millisecond * 10),
			Start:     time.Now(),
			SessionID: 4,
			PageViews: 51,
		},
		Sign:    1,
		Version: 1,
	})
	assert.NoError(t, dbClient.SaveSessions(sessions))
	count := 0
	assert.NoError(t, dbClient.QueryRow(`SELECT page_views FROM "session_v7" FINAL`).Scan(&count))
	assert.Equal(t, 51, count)
}

func TestClient_SavePageViews(t *testing.T) {
	CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{
			Data: model.Data{
				ClientID:       1,
				VisitorID:      1,
				Time:           time.Now(),
				SessionID:      util.RandUint32(),
				Hostname:       "example.com",
				Language:       "en",
				Referrer:       "ref",
				ReferrerName:   "ref_name",
				ReferrerIcon:   "ref_icon",
				OS:             "os",
				OSVersion:      "10",
				Browser:        "browser",
				BrowserVersion: "89",
				CountryCode:    "en",
				Region:         "England",
				City:           "London",
				Desktop:        true,
				Mobile:         false,
				ScreenClass:    "XL",
				UTMCampaign:    "Campaign",
				UTMMedium:      "Medium",
				UTMSource:      "Source",
				UTMContent:     "Content",
				UTMTerm:        "Term",
				Channel:        "Channel",
			},
			Path:  "/path",
			Title: "title",
			Tags:  map[string]string{"key0": "value0", "key1": "value1"},
		},
		{
			Data: model.Data{
				VisitorID: 1,
				Time:      time.Now().UTC(),
			},
			Path: "/path",
		},
	}))
}

func TestClient_SaveEvents(t *testing.T) {
	CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{
			Data: model.Data{
				ClientID:       1,
				VisitorID:      1,
				Time:           time.Now(),
				SessionID:      util.RandUint32(),
				Hostname:       "example.com",
				Language:       "en",
				Referrer:       "ref",
				ReferrerName:   "ref_name",
				ReferrerIcon:   "ref_icon",
				OS:             "os",
				OSVersion:      "10",
				Browser:        "browser",
				BrowserVersion: "89",
				CountryCode:    "en",
				Region:         "England",
				City:           "London",
				Desktop:        true,
				Mobile:         false,
				ScreenClass:    "XL",
				UTMCampaign:    "Campaign",
				UTMMedium:      "Medium",
				UTMSource:      "Source",
				UTMContent:     "Content",
				UTMTerm:        "Term",
				Channel:        "Channel",
			},
			Name:     "event_name",
			Path:     "/path",
			Title:    "title",
			MetaData: map[string]any{"meta": "some", "data": []int{1, 2, 3}},
		},
		{
			Data: model.Data{
				VisitorID: 1,
				Time:      time.Now().UTC(),
			},
			Name: "different_event",
			Path: "/path",
		},
	}))
}

func TestClient_SaveRequests(t *testing.T) {
	CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SaveRequests([]model.Request{
		{
			ClientID:    1,
			VisitorID:   1,
			Time:        time.Now(),
			IP:          "123.456.789.9",
			UserAgent:   "ua1",
			Hostname:    "example.com",
			Path:        "/foo",
			Event:       "event",
			Referrer:    "ref",
			UTMSource:   "source",
			UTMMedium:   "medium",
			UTMCampaign: "campaign",
			Bot:         true,
			BotReason:   "ip",
		},
		{
			ClientID:  2,
			VisitorID: 2,
			Time:      time.Now(),
			UserAgent: "ua2",
			Path:      "/bar",
		},
	}))
}

func TestClient_GetNoError(t *testing.T) {
	CleanupDB(t, dbClient)
	var sessions int
	assert.NoError(t, dbClient.QueryRow(`SELECT count(*) FROM "session_v7" LIMIT 1`).Scan(&sessions))
	assert.Equal(t, 0, sessions)
}
