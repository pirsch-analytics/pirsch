package db

import (
	"context"
	"math/rand/v2"
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
	assert.NoError(t, client.Ping(context.Background()))
}

func TestClient_SaveSessions(t *testing.T) {
	CleanupDB(t, client)
	assert.NoError(t, client.SaveSessions(context.Background(), []model.Session{
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
			Sign:            1,
			Version:         1,
			Start:           time.Now(),
			DurationSeconds: 42,
			PageViews:       7,
			IsBounce:        true,
			EntryPath:       "/entry-path",
			ExitPath:        "/exit-path",
			EntryTitle:      "entry-title",
			ExitTitle:       "exit_title",
			Extended:        123,
		},
		{
			Data: model.Data{
				VisitorID: 1,
				Time:      time.Now().UTC(),
			},
			Sign:     -1,
			Version:  1,
			Start:    time.Now(),
			ExitPath: "/path",
		},
	}))
	var sessions uint64
	assert.NoError(t, client.QueryRow(context.Background(), `SELECT count(*) FROM "session_v7" LIMIT 1`).Scan(&sessions))
	assert.Equal(t, uint64(2), sessions)
}

func TestClient_SaveSessionsBatch(t *testing.T) {
	CleanupDB(t, client)
	sessions := make([]model.Session, 0, 101)

	for i := range 50 {
		now := time.Now().Add(time.Millisecond * time.Duration(i))
		sessions = append(sessions, model.Session{
			Data: model.Data{
				ClientID:  2,
				VisitorID: 3,
				Time:      now,
				SessionID: 4,
			},
			Sign:      1,
			Version:   1,
			Start:     now,
			PageViews: uint16(i + 1),
		})
		sessions = append(sessions, model.Session{
			Data: model.Data{
				ClientID:  2,
				VisitorID: 3,
				Time:      now,
				SessionID: 4,
			},
			Sign:      -1,
			Version:   1,
			Start:     now,
			PageViews: uint16(i + 1),
		})
	}

	sessions = append(sessions, model.Session{
		Data: model.Data{
			ClientID:  2,
			VisitorID: 3,
			Time:      time.Now().Add(time.Millisecond * 10),
			SessionID: 4,
		},
		Sign:      1,
		Version:   1,
		Start:     time.Now(),
		PageViews: 51,
	})
	assert.NoError(t, client.SaveSessions(context.Background(), sessions))
	count := uint16(0)
	assert.NoError(t, client.QueryRow(context.Background(), `SELECT page_views FROM "session_v7" FINAL`).Scan(&count))
	assert.Equal(t, uint16(51), count)
}

func TestClient_SavePageViews(t *testing.T) {
	CleanupDB(t, client)
	assert.NoError(t, client.SavePageViews(context.Background(), []model.PageView{
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
	var pageViews uint64
	assert.NoError(t, client.QueryRow(context.Background(), `SELECT count(*) FROM "page_view_v7" LIMIT 1`).Scan(&pageViews))
	assert.Equal(t, uint64(2), pageViews)
}

func TestClient_SaveEvents(t *testing.T) {
	CleanupDB(t, client)
	assert.NoError(t, client.SaveEvents(context.Background(), []model.Event{
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
	var events uint64
	assert.NoError(t, client.QueryRow(context.Background(), `SELECT count(*) FROM "event_v7" LIMIT 1`).Scan(&events))
	assert.Equal(t, uint64(2), events)
}

func TestClient_SaveRequests(t *testing.T) {
	CleanupDB(t, client)
	assert.NoError(t, client.SaveRequests(context.Background(), []model.Request{
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
	var requests uint64
	assert.NoError(t, client.QueryRow(context.Background(), `SELECT count(*) FROM "request" LIMIT 1`).Scan(&requests))
	assert.Equal(t, uint64(2), requests)
}

func TestClient_Session(t *testing.T) {
	CleanupDB(t, client)
	now := time.Now().UTC().Add(-time.Second * 20)
	assert.NoError(t, client.SaveSessions(context.Background(), []model.Session{
		{
			Data: model.Data{
				ClientID:  1,
				VisitorID: 1,
				Time:      now.Add(-time.Second * 20),
				SessionID: rand.Uint32(),
				Hostname:  "example.com",
			},
			Sign:      1,
			Start:     time.Now(),
			PageViews: 2,
			ExitPath:  "/path1",
			EntryPath: "/entry1",
		},
		{
			Data: model.Data{
				ClientID:  1,
				VisitorID: 1,
				Time:      now,
				SessionID: 123456,
				Hostname:  "example.com",
			},
			Sign:      -1,
			Start:     time.Now(),
			PageViews: 3,
			ExitPath:  "/path2",
			EntryPath: "/entry2",
		},
		{
			Data: model.Data{
				ClientID:  1,
				VisitorID: 1,
				Time:      now.Add(-time.Second * 10),
				SessionID: rand.Uint32(),
				Hostname:  "example.com",
			},
			Sign:      -1,
			Start:     time.Now(),
			PageViews: 4,
			ExitPath:  "/path3",
			EntryPath: "/entry3",
		},
	}))
	session, err := client.Session(context.Background(), 1, 1, time.Now().UTC().Add(-time.Minute))
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, now.Unix(), session.Time.Unix())
	assert.Equal(t, uint32(123456), session.SessionID)
	assert.Equal(t, "/path2", session.ExitPath)
	assert.Equal(t, "/entry2", session.EntryPath)
	assert.Equal(t, uint16(3), session.PageViews)
}

func TestClient_GetNoError(t *testing.T) {
	CleanupDB(t, client)
	var sessions uint64
	assert.NoError(t, client.QueryRow(context.Background(), `SELECT count(*) FROM "session_v7" LIMIT 1`).Scan(&sessions))
	assert.Equal(t, uint64(0), sessions)
}
