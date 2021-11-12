package pirsch

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient("tcp://127.0.0.1:9000", nil)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NoError(t, client.DB.Ping())
}

func TestClient_SavePageViews(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SavePageViews([]PageView{
		{
			ClientID:        1,
			VisitorID:       1,
			Time:            time.Now(),
			SessionID:       rand.Uint32(),
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
			ScreenWidth:     1920,
			ScreenHeight:    1080,
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
	cleanupDB()
	assert.NoError(t, dbClient.SaveSessions([]Session{
		{
			Sign:            1,
			ClientID:        1,
			VisitorID:       1,
			Time:            time.Now(),
			Start:           time.Now(),
			SessionID:       rand.Uint32(),
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
			ScreenWidth:     1920,
			ScreenHeight:    1080,
			ScreenClass:     "XL",
		},
		{
			Sign:      -1,
			VisitorID: 1,
			Time:      time.Now().UTC(),
			ExitPath:  "/path",
		},
	}))
}

func TestClient_SaveEvents(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveEvents([]Event{
		{
			ClientID:        1,
			VisitorID:       1,
			Time:            time.Now(),
			SessionID:       rand.Uint32(),
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
			ScreenWidth:     1920,
			ScreenHeight:    1080,
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
	cleanupDB()
	assert.NoError(t, dbClient.SaveUserAgents([]UserAgent{
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

func TestClient_Session(t *testing.T) {
	cleanupDB()
	now := time.Now().UTC().Add(-time.Second * 20)
	assert.NoError(t, dbClient.SaveSessions([]Session{
		{
			Sign:      1,
			ClientID:  1,
			VisitorID: 1,
			Time:      now.Add(-time.Second * 20),
			SessionID: rand.Uint32(),
			ExitPath:  "/path1",
			EntryPath: "/entry1",
			PageViews: 2,
		},
		{
			Sign:      -1,
			ClientID:  1,
			VisitorID: 1,
			Time:      now,
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
			SessionID: rand.Uint32(),
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
	cleanupDB()
	var session Session
	assert.NoError(t, dbClient.Get(&session, `SELECT * FROM "session" LIMIT 1`))
}
