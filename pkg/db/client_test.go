package db

import (
	"context"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"github.com/pirsch-analytics/pirsch/v6/pkg/util"
	"github.com/stretchr/testify/assert"
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
	assert.NoError(t, dbClient.SavePageViews(context.Background(), []model.PageView{
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
			Region:          "England",
			City:            "London",
			Desktop:         true,
			Mobile:          false,
			ScreenClass:     "XL",
			TagKeys:         []string{"key0", "key1"},
			TagValues:       []string{"value0", "value1"},
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
	assert.NoError(t, dbClient.SaveSessions(context.Background(), []model.Session{
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
			Region:          "England",
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
		PageViews: 51,
	})
	assert.NoError(t, dbClient.SaveSessions(context.Background(), sessions))
	count := 0
	assert.NoError(t, dbClient.QueryRow(`SELECT page_views FROM "session" FINAL`).Scan(&count))
	assert.Equal(t, 51, count)
}

func TestClient_SaveEvents(t *testing.T) {
	CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SaveEvents(context.Background(), []model.Event{
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
			Region:          "England",
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

func TestClient_SaveRequests(t *testing.T) {
	CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SaveRequests(context.Background(), []model.Request{
		{
			ClientID:    1,
			VisitorID:   1,
			Time:        time.Now(),
			IP:          "123.456.789.9",
			UserAgent:   "ua1",
			Path:        "/foo",
			Event:       "event",
			Referrer:    "ref",
			UTMSource:   "source",
			UTMMedium:   "medium",
			UTMCampaign: "campaign",
			Bot:         true,
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
	assert.NoError(t, dbClient.SaveSessions(context.Background(), []model.Session{
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
	session, err := dbClient.Session(context.Background(), 1, 1, time.Now().UTC().Add(-time.Minute))
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

func TestClient_SelectImported(t *testing.T) {
	CleanupDB(t, dbClient)
	queries := []string{
		"INSERT INTO imported_browser (client_id, date, browser, visitors) VALUES (42, '2024-07-23', 'Firefox', 1)",
		"INSERT INTO imported_campaign (client_id, date, campaign, visitors) VALUES (42, '2024-07-23', 'Campaign', 2)",
		"INSERT INTO imported_city (client_id, date, city, visitors) VALUES (42, '2024-07-23', 'Tokyo', 3)",
		"INSERT INTO imported_country (client_id, date, country_code, visitors) VALUES (42, '2024-07-23', 'ja', 4)",
		"INSERT INTO imported_device (client_id, date, category, visitors) VALUES (42, '2024-07-23', 'XXL', 5)",
		"INSERT INTO imported_entry_page (client_id, date, entry_path, visitors, sessions) VALUES (42, '2024-07-23', '/entry', 6, 1)",
		"INSERT INTO imported_exit_page (client_id, date, exit_path, visitors, sessions) VALUES (42, '2024-07-23', '/exit', 7, 2)",
		"INSERT INTO imported_language (client_id, date, language, visitors) VALUES (42, '2024-07-23', 'ja', 8)",
		"INSERT INTO imported_medium (client_id, date, medium, visitors) VALUES (42, '2024-07-23', 'Medium', 9)",
		"INSERT INTO imported_os (client_id, date, os, visitors) VALUES (42, '2024-07-23', 'Windows', 10)",
		"INSERT INTO imported_page (client_id, date, path, visitors, page_views, sessions, bounces) VALUES (42, '2024-07-23', '/', 11, 1, 3, 1)",
		"INSERT INTO imported_referrer (client_id, date, referrer, visitors, sessions, bounces) VALUES (42, '2024-07-23', 'Referrer', 12, 4, 2)",
		"INSERT INTO imported_region (client_id, date, region, visitors) VALUES (42, '2024-07-23', 'Kantō', 13)",
		"INSERT INTO imported_source (client_id, date, source, visitors) VALUES (42, '2024-07-23', 'Source', 14)",
		"INSERT INTO imported_visitors (client_id, date, visitors, page_views, sessions, bounces, session_duration) VALUES (42, '2024-07-23', 15, 2, 5, 3, 1)",
	}

	for _, query := range queries {
		_, err := dbClient.Exec(query)
		assert.NoError(t, err)
	}

	date := time.Date(2024, 7, 23, 0, 0, 0, 0, time.UTC)
	browser, err := dbClient.SelectImportedBrowser(context.Background(), 42, date, date)
	assert.NoError(t, err)
	assert.Len(t, browser, 1)
	assert.Equal(t, uint64(42), browser[0].ClientID)
	assert.Equal(t, date, browser[0].Date)
	assert.Equal(t, "Firefox", browser[0].Browser)
	assert.Equal(t, 1, browser[0].Visitors)
	campaign, err := dbClient.SelectImportedCampaign(context.Background(), 42, date, date)
	assert.NoError(t, err)
	assert.Len(t, campaign, 1)
	assert.Equal(t, uint64(42), campaign[0].ClientID)
	assert.Equal(t, date, campaign[0].Date)
	assert.Equal(t, "Campaign", campaign[0].Campaign)
	assert.Equal(t, 2, campaign[0].Visitors)
	city, err := dbClient.SelectImportedCity(context.Background(), 42, date, date)
	assert.NoError(t, err)
	assert.Len(t, city, 1)
	assert.Equal(t, uint64(42), city[0].ClientID)
	assert.Equal(t, date, city[0].Date)
	assert.Equal(t, "Tokyo", city[0].City)
	assert.Equal(t, 3, city[0].Visitors)
	country, err := dbClient.SelectImportedCountry(context.Background(), 42, date, date)
	assert.NoError(t, err)
	assert.Len(t, country, 1)
	assert.Equal(t, uint64(42), country[0].ClientID)
	assert.Equal(t, date, country[0].Date)
	assert.Equal(t, "ja", country[0].CountryCode)
	assert.Equal(t, 4, country[0].Visitors)
	device, err := dbClient.SelectImportedDevice(context.Background(), 42, date, date)
	assert.NoError(t, err)
	assert.Len(t, device, 1)
	assert.Equal(t, uint64(42), device[0].ClientID)
	assert.Equal(t, date, device[0].Date)
	assert.Equal(t, "XXL", device[0].Category)
	assert.Equal(t, 5, device[0].Visitors)
	entryPage, err := dbClient.SelectImportedEntryPage(context.Background(), 42, date, date)
	assert.NoError(t, err)
	assert.Len(t, entryPage, 1)
	assert.Equal(t, uint64(42), entryPage[0].ClientID)
	assert.Equal(t, date, entryPage[0].Date)
	assert.Equal(t, "/entry", entryPage[0].EntryPath)
	assert.Equal(t, 6, entryPage[0].Visitors)
	assert.Equal(t, 1, entryPage[0].Sessions)
	exitPage, err := dbClient.SelectImportedExitPage(context.Background(), 42, date, date)
	assert.NoError(t, err)
	assert.Len(t, exitPage, 1)
	assert.Equal(t, uint64(42), exitPage[0].ClientID)
	assert.Equal(t, date, exitPage[0].Date)
	assert.Equal(t, "/exit", exitPage[0].ExitPath)
	assert.Equal(t, 7, exitPage[0].Visitors)
	assert.Equal(t, 2, exitPage[0].Sessions)
	language, err := dbClient.SelectImportedLanguage(context.Background(), 42, date, date)
	assert.NoError(t, err)
	assert.Len(t, language, 1)
	assert.Equal(t, uint64(42), language[0].ClientID)
	assert.Equal(t, date, language[0].Date)
	assert.Equal(t, "ja", language[0].Language)
	assert.Equal(t, 8, language[0].Visitors)
	medium, err := dbClient.SelectImportedMedium(context.Background(), 42, date, date)
	assert.NoError(t, err)
	assert.Len(t, medium, 1)
	assert.Equal(t, uint64(42), medium[0].ClientID)
	assert.Equal(t, date, medium[0].Date)
	assert.Equal(t, "Medium", medium[0].Medium)
	assert.Equal(t, 9, medium[0].Visitors)
	os, err := dbClient.SelectImportedOS(context.Background(), 42, date, date)
	assert.NoError(t, err)
	assert.Len(t, os, 1)
	assert.Equal(t, uint64(42), os[0].ClientID)
	assert.Equal(t, date, os[0].Date)
	assert.Equal(t, "Windows", os[0].OS)
	assert.Equal(t, 10, os[0].Visitors)
	page, err := dbClient.SelectImportedPage(context.Background(), 42, date, date)
	assert.NoError(t, err)
	assert.Len(t, page, 1)
	assert.Equal(t, uint64(42), page[0].ClientID)
	assert.Equal(t, date, page[0].Date)
	assert.Equal(t, "/", page[0].Path)
	assert.Equal(t, 11, page[0].Visitors)
	assert.Equal(t, 1, page[0].PageViews)
	assert.Equal(t, 3, page[0].Sessions)
	assert.Equal(t, 1, page[0].Bounces)
	referrer, err := dbClient.SelectImportedReferrer(context.Background(), 42, date, date)
	assert.NoError(t, err)
	assert.Len(t, referrer, 1)
	assert.Equal(t, uint64(42), referrer[0].ClientID)
	assert.Equal(t, date, referrer[0].Date)
	assert.Equal(t, "Referrer", referrer[0].Referrer)
	assert.Equal(t, 12, referrer[0].Visitors)
	assert.Equal(t, 4, referrer[0].Sessions)
	assert.Equal(t, 2, referrer[0].Bounces)
	region, err := dbClient.SelectImportedRegion(context.Background(), 42, date, date)
	assert.NoError(t, err)
	assert.Len(t, region, 1)
	assert.Equal(t, uint64(42), region[0].ClientID)
	assert.Equal(t, date, region[0].Date)
	assert.Equal(t, "Kantō", region[0].Region)
	assert.Equal(t, 13, region[0].Visitors)
	source, err := dbClient.SelectImportedSource(context.Background(), 42, date, date)
	assert.NoError(t, err)
	assert.Len(t, source, 1)
	assert.Equal(t, uint64(42), source[0].ClientID)
	assert.Equal(t, date, source[0].Date)
	assert.Equal(t, "Source", source[0].Source)
	assert.Equal(t, 14, source[0].Visitors)
	visitors, err := dbClient.SelectImportedVisitors(context.Background(), 42, date, date)
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, uint64(42), visitors[0].ClientID)
	assert.Equal(t, date, visitors[0].Date)
	assert.Equal(t, 15, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[0].PageViews)
	assert.Equal(t, 5, visitors[0].Sessions)
	assert.Equal(t, 3, visitors[0].Bounces)
	assert.Equal(t, 1, visitors[0].SessionDuration)
}

func TestClient_SelectImportedFill(t *testing.T) {
	CleanupDB(t, dbClient)
	queries := []string{
		"INSERT INTO imported_browser (client_id, date, browser, visitors) VALUES (42, '2024-07-21', 'Firefox', 67)",
		"INSERT INTO imported_browser (client_id, date, browser, visitors) VALUES (42, '2024-07-22', 'Firefox', 43)",
		"INSERT INTO imported_browser (client_id, date, browser, visitors) VALUES (42, '2024-07-23', 'Firefox', 86)",
	}

	for _, query := range queries {
		_, err := dbClient.Exec(query)
		assert.NoError(t, err)
	}

	from := time.Date(2024, 7, 20, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 7, 23, 0, 0, 0, 0, time.UTC)
	browser, err := dbClient.SelectImportedBrowser(context.Background(), 42, from, to)
	assert.NoError(t, err)
	assert.Len(t, browser, 4)
	assert.Equal(t, from, browser[0].Date)
	assert.Equal(t, to, browser[3].Date)
	assert.Equal(t, 0, browser[0].Visitors)
	assert.Equal(t, 67, browser[1].Visitors)
	assert.Equal(t, 43, browser[2].Visitors)
	assert.Equal(t, 86, browser[3].Visitors)
}
