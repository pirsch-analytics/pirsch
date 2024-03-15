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

func TestSessions_List(t *testing.T) {
	// TODO
}

func TestSessions_Breakdown(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.Today(), Start: util.Today(), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
		},
		{
			{Sign: -1, VisitorID: 1, SessionID: 1, Time: util.Today(), Start: util.Today(), EntryPath: "/", ExitPath: "/", PageViews: 1, IsBounce: true},
			{Sign: 1, VisitorID: 1, SessionID: 1, Time: util.Today().Add(time.Second * 10), Start: util.Today(), EntryPath: "/", ExitPath: "/pricing", PageViews: 2, IsBounce: false},
		},
	})
	assert.NoError(t, dbClient.SavePageViews(context.Background(), []model.PageView{
		{VisitorID: 1, SessionID: 1, Time: util.Today(), Path: "/"},
		{VisitorID: 1, SessionID: 1, Time: util.Today().Add(time.Second * 10), Path: "/pricing"},
	}))
	assert.NoError(t, dbClient.SaveEvents(context.Background(), []model.Event{
		{VisitorID: 1, SessionID: 1, Time: util.Today().Add(time.Second * 5), Name: "event", MetaKeys: []string{"key"}, MetaValues: []string{"value"}},
	}))
	analyzer := NewAnalyzer(dbClient)
	steps, err := analyzer.Sessions.Breakdown(nil)
	assert.NoError(t, err)
	assert.Nil(t, steps)
	steps, err = analyzer.Sessions.Breakdown(&Filter{
		VisitorID: 1,
		SessionID: 1,
		From:      util.Today(),
		To:        util.Today(),
	})
	assert.NoError(t, err)
	assert.Len(t, steps, 3)
	assert.Equal(t, util.Today(), steps[0].PageView.Time)
	assert.Equal(t, util.Today().Add(time.Second*5), steps[1].Event.Time)
	assert.Equal(t, util.Today().Add(time.Second*10), steps[2].PageView.Time)
	assert.Equal(t, "/", steps[0].PageView.Path)
	assert.Equal(t, "/pricing", steps[2].PageView.Path)
	assert.Equal(t, "event", steps[1].Event.Name)
	assert.Equal(t, "key", steps[1].Event.MetaKeys[0])
	assert.Equal(t, "value", steps[1].Event.MetaValues[0])
}
