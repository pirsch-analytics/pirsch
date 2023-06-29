package db

import (
	"github.com/pirsch-analytics/pirsch/v5/model"
	"github.com/pirsch-analytics/pirsch/v5/util"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient(&ClientConfig{
		Hostname:      "127.0.0.1",
		Port:          9000,
		Database:      "pirschtest",
		SSLSkipVerify: true,
		Debug:         true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NoError(t, client.DB.Ping())
}

func TestClient_SavePageViews(t *testing.T) {
	CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{
			ClientID:        1,
			VisitorID:       1,
			Time:            time.Now(),
			SessionID:       util.RandUint32(),
			DurationSeconds: 42,
			Path:            "/path",
			Title:           "title",
			Language:        "en",
			Referrer:        "ref",
			ReferrerName:    "ref_name",
			ReferrerIcon:    "ref_icon",
			OS:              "os",
			OSVersion:       "10",
			Browser:         "browser",
			BrowserVersion:  "89",
			CountryCode:     "en",
			City:            "London",
			Desktop:         true,
			Mobile:          false,
			ScreenClass:     "XL",
		},
		{
			VisitorID: 1,
			Time:      time.Now().UTC(),
			Path:      "/path",
		},
	}))
}

func TestClient_SaveSessions(t *testing.T) {
	CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SaveSessions([]model.Session{
		{
			Sign:            1,
			ClientID:        1,
			VisitorID:       1,
			Time:            time.Now(),
			Start:           time.Now(),
			SessionID:       util.RandUint32(),
			DurationSeconds: 42,
			EntryPath:       "/entry-path",
			ExitPath:        "/exit-path",
			PageViews:       7,
			IsBounce:        true,
			EntryTitle:      "entry-title",
			ExitTitle:       "exit_title",
			Language:        "en",
			Referrer:        "ref",
			ReferrerName:    "ref_name",
			ReferrerIcon:    "ref_icon",
			OS:              "os",
			OSVersion:       "10",
			Browser:         "browser",
			BrowserVersion:  "89",
			CountryCode:     "en",
			City:            "London",
			Desktop:         true,
			Mobile:          false,
			ScreenClass:     "XL",
			Extended:        123,
		},
		{
			Sign:      -1,
			VisitorID: 1,
			Time:      time.Now().UTC(),
			Start:     time.Now(),
			ExitPath:  "/path",
		},
	}))
}

func TestClient_SaveSessionsBatch(t *testing.T) {
	CleanupDB(t, dbClient)
	sessions := make([]model.Session, 0, 101)

	for i := 0; i < 50; i++ {
		now := time.Now().Add(time.Millisecond * time.Duration(i))
		sessions = append(sessions, model.Session{
			Sign:      1,
			ClientID:  2,
			VisitorID: 3,
			Time:      now,
			Start:     now,
			SessionID: 4,
			PageViews: uint16(i + 1),
		})
		sessions = append(sessions, model.Session{
			Sign:      -1,
			ClientID:  2,
			VisitorID: 3,
			Time:      now,
			Start:     now,
			SessionID: 4,
			PageViews: uint16(i + 1),
		})
	}

	sessions = append(sessions, model.Session{
		Sign:      1,
		ClientID:  2,
		VisitorID: 3,
		Time:      time.Now().Add(time.Millisecond * 10),
		Start:     time.Now(),
		SessionID: 4,
		PageViews: 101,
	})

	// randomize insertion order
	rand.Shuffle(len(sessions), func(i, j int) {
		sessions[i], sessions[j] = sessions[j], sessions[i]
	})

	assert.NoError(t, dbClient.SaveSessions(sessions))
	count := 0
	assert.NoError(t, dbClient.QueryRow(`SELECT count(*) FROM "session" GROUP BY page_views HAVING sum(sign) > 0`).Scan(&count))
	assert.Equal(t, 1, count)
}

func TestClient_SaveEvents(t *testing.T) {
	CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{
			ClientID:        1,
			VisitorID:       1,
			Time:            time.Now(),
			SessionID:       util.RandUint32(),
			Name:            "event_name",
			MetaKeys:        []string{"meta", "keys"},
			MetaValues:      []string{"some", "values"},
			DurationSeconds: 21,
			Path:            "/path",
			Title:           "title",
			Language:        "en",
			Referrer:        "ref",
			ReferrerName:    "ref_name",
			ReferrerIcon:    "ref_icon",
			OS:              "os",
			OSVersion:       "10",
			Browser:         "browser",
			BrowserVersion:  "89",
			CountryCode:     "en",
			City:            "London",
			Desktop:         true,
			Mobile:          false,
			ScreenClass:     "XL",
		},
		{
			VisitorID: 1,
			Time:      time.Now().UTC(),
			Name:      "different_event",
			Path:      "/path",
		},
	}))
}

func TestClient_SaveUserAgents(t *testing.T) {
	CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SaveUserAgents([]model.UserAgent{
		{
			Time:      time.Now(),
			UserAgent: "ua1",
		},
		{
			Time:      time.Now().Add(time.Second),
			UserAgent: "ua2",
		},
	}))
}

func TestClient_SaveBots(t *testing.T) {
	CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SaveBots([]model.Bot{
		{
			ClientID:  1,
			VisitorID: 1,
			Time:      time.Now(),
			UserAgent: "ua1",
			Path:      "/foo",
			Event:     "event",
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

func TestClient_Session(t *testing.T) {
	CleanupDB(t, dbClient)
	now := time.Now().UTC().Add(-time.Second * 20)
	assert.NoError(t, dbClient.SaveSessions([]model.Session{
		{
			Sign:      1,
			ClientID:  1,
			VisitorID: 1,
			Time:      now.Add(-time.Second * 20),
			Start:     time.Now(),
			SessionID: util.RandUint32(),
			ExitPath:  "/path1",
			EntryPath: "/entry1",
			PageViews: 2,
		},
		{
			Sign:      -1,
			ClientID:  1,
			VisitorID: 1,
			Time:      now,
			Start:     time.Now(),
			SessionID: 123456,
			ExitPath:  "/path2",
			EntryPath: "/entry2",
			PageViews: 3,
		},
		{
			Sign:      -1,
			ClientID:  1,
			VisitorID: 1,
			Time:      now.Add(-time.Second * 10),
			Start:     time.Now(),
			SessionID: util.RandUint32(),
			ExitPath:  "/path3",
			EntryPath: "/entry3",
			PageViews: 4,
		},
	}))
	session, err := dbClient.Session(1, 1, time.Now().UTC().Add(-time.Minute))
	assert.NoError(t, err)
	assert.Equal(t, now.Unix(), session.Time.Unix())
	assert.Equal(t, uint32(123456), session.SessionID)
	assert.Equal(t, "/path2", session.ExitPath)
	assert.Equal(t, "/entry2", session.EntryPath)
	assert.Equal(t, uint16(3), session.PageViews)
}

func TestClient_GetNoError(t *testing.T) {
	CleanupDB(t, dbClient)
	var sessions int
	assert.NoError(t, dbClient.QueryRow(`SELECT count(*) FROM "session" LIMIT 1`).Scan(&sessions))
	assert.Equal(t, 0, sessions)
}
