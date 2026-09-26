package query

import (
	"context"
	"encoding/json/v2"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/pirsch-analytics/pirsch/v7/pkg/db"
	"github.com/pirsch-analytics/pirsch/v7/pkg/model"
	"github.com/stretchr/testify/assert"
)

type sessionData struct {
	model.Session
	Scenario string `csv:"scenario"`
}

type pageViewData struct {
	model.PageView
	Scenario string `csv:"scenario"`
	Tags     string `csv:"tags"`
}

type eventData struct {
	model.Event
	Scenario string `csv:"scenario"`
	MetaData string `csv:"meta_data"`
}

type ImportedData struct {
	Scenario string    `csv:"scenario"`
	SiteID   uint64    `csv:"site_id"`
	Date     time.Time `csv:"date"`
}

type importedReferrerData struct {
	ImportedData
	Referrer string `csv:"referrer"`
	Visitors uint64 `csv:"visitors"`
	Sessions uint64 `csv:"sessions"`
	Bounces  uint64 `csv:"bounces"`
}

func loadTestData(t *testing.T, scenarios []string) {
	db.CleanupDB(t, client)

	// load and store sessions
	sessionsFile, err := os.ReadFile("../../../test/sessions.csv")
	assert.NoError(t, err)
	var sessionData []sessionData
	assert.NoError(t, gocsv.UnmarshalBytes(sessionsFile, &sessionData))
	sessions := make([]model.Session, 0, len(sessionData))

	for _, s := range sessionData {
		if len(scenarios) == 0 || slices.Contains(scenarios, s.Scenario) {
			sessions = append(sessions, s.Session)
		}
	}

	assert.NoError(t, client.SaveSessions(context.Background(), sessions))

	// load and store page views
	pageViewsFile, err := os.ReadFile("../../../test/page_views.csv")
	assert.NoError(t, err)
	var pageViewData []pageViewData
	assert.NoError(t, gocsv.UnmarshalBytes(pageViewsFile, &pageViewData))
	pageViews := make([]model.PageView, 0, len(pageViewData))

	for _, pv := range pageViewData {
		if len(scenarios) == 0 || slices.Contains(scenarios, pv.Scenario) {
			pageViews = append(pageViews, pv.PageView)

			if pv.Tags != "" {
				var tags map[string]string

				if err := json.Unmarshal([]byte(pv.Tags), &tags); err != nil {
					t.Fatal(err)
				}

				pageViews[len(pageViews)-1].Tags = tags
			}
		}
	}

	assert.NoError(t, client.SavePageViews(context.Background(), pageViews))

	// load and store events
	eventsFile, err := os.ReadFile("../../../test/events.csv")
	assert.NoError(t, err)
	var eventData []eventData
	assert.NoError(t, gocsv.UnmarshalBytes(eventsFile, &eventData))
	events := make([]model.Event, 0, len(eventData))

	for _, e := range eventData {
		if len(scenarios) == 0 || slices.Contains(scenarios, e.Scenario) {
			events = append(events, e.Event)

			if e.MetaData != "" {
				var metaData map[string]any

				if err := json.Unmarshal([]byte(e.MetaData), &metaData); err != nil {
					t.Fatal(err)
				}

				events[len(events)-1].MetaData = metaData
			}
		}
	}

	assert.NoError(t, client.SaveEvents(context.Background(), events))

	// load and store imported statistics
	importedReferrerFile, err := os.ReadFile("../../../test/imported_referrer.csv")
	assert.NoError(t, err)
	var referrerData []importedReferrerData
	assert.NoError(t, gocsv.UnmarshalBytes(importedReferrerFile, &referrerData))
	referrer := make([]importedReferrerData, 0, len(referrerData))

	for _, r := range referrerData {
		if len(scenarios) == 0 || slices.Contains(scenarios, r.Scenario) {
			referrer = append(referrer, r)
		}
	}

	assert.NoError(t, saveImportedReferrer(referrer))
}

func saveImportedReferrer(referrer []importedReferrerData) error {
	stmt, err := client.Conn.PrepareBatch(context.Background(), `INSERT INTO "imported_referrer" (client_id, date, referrer, visitors, sessions, bounces)`)

	if err != nil {
		return err
	}

	for _, ref := range referrer {
		if err := stmt.Append(ref.SiteID,
			ref.Date,
			ref.Referrer,
			ref.Visitors,
			ref.Sessions,
			ref.Bounces); err != nil {
			return err
		}
	}

	return stmt.Send()
}
