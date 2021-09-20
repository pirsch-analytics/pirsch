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

func TestClient_SaveHit(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{
			ClientID:        1,
			Fingerprint:     "fp",
			Time:            time.Now(),
			SessionID:       rand.Uint32(),
			DurationSeconds: 42,
			Path:            "/path",
			EntryPath:       "/entry-path",
			PageViews:       7,
			IsBounce:        true,
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
			Desktop:         true,
			Mobile:          false,
			ScreenWidth:     1920,
			ScreenHeight:    1080,
			ScreenClass:     "XL",
		},
		{
			Fingerprint: "fp",
			Time:        time.Now().UTC(),
			Path:        "/path",
		},
	}))
}

func TestClient_SaveEvent(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveEvents([]Event{
		{
			Hit: Hit{
				ClientID:        1,
				Fingerprint:     "fp",
				Time:            time.Now(),
				SessionID:       rand.Uint32(),
				DurationSeconds: 42,
				Path:            "/path",
				EntryPath:       "/entry-path",
				PageViews:       7,
				IsBounce:        true,
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
				Desktop:         true,
				Mobile:          false,
				ScreenWidth:     1920,
				ScreenHeight:    1080,
				ScreenClass:     "XL",
			},
			Name:            "event_name",
			DurationSeconds: 21,
			MetaKeys:        []string{"meta", "keys"},
			MetaValues:      []string{"some", "values"},
		},
		{
			Hit: Hit{
				Fingerprint: "fp",
				Time:        time.Now().UTC(),
				Path:        "/path",
			},
			Name: "different_event",
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
	fp := "session_fp"
	now := time.Now().UTC().Add(-time.Second * 20)
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{
			ClientID:    1,
			Fingerprint: fp,
			Time:        now.Add(-time.Second * 20),
			SessionID:   rand.Uint32(),
			Path:        "/path1",
			EntryPath:   "/entry1",
			PageViews:   2,
		},
		{
			ClientID:    1,
			Fingerprint: fp,
			Time:        now,
			SessionID:   123456,
			Path:        "/path2",
			EntryPath:   "/entry2",
			PageViews:   3,
		},
		{
			ClientID:    1,
			Fingerprint: fp,
			Time:        now.Add(-time.Second * 10),
			SessionID:   rand.Uint32(),
			Path:        "/path3",
			EntryPath:   "/entry3",
			PageViews:   4,
		},
	}))
	session, err := dbClient.Session(1, fp, time.Now().UTC().Add(-time.Minute))
	assert.NoError(t, err)
	assert.Equal(t, now.Unix(), session.Time.Unix())
	assert.Equal(t, uint32(123456), session.SessionID)
	assert.Equal(t, "/path2", session.Path)
	assert.Equal(t, "/entry2", session.EntryPath)
	assert.Equal(t, uint16(3), session.PageViews)
}
