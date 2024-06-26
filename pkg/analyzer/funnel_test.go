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

func TestFunnel_Steps(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{
				Sign:      1,
				VisitorID: 1,
				Time:      util.Today(),
				Start:     time.Now(),
				EntryPath: "/",
				ExitPath:  "/thank-you",
				IsBounce:  false,
				PageViews: 5,
			},
			{
				Sign:      1,
				VisitorID: 2,
				Time:      util.Today(),
				Start:     time.Now(),
				EntryPath: "/product",
				ExitPath:  "/cart",
				IsBounce:  false,
				PageViews: 2,
			},
			{
				Sign:      1,
				VisitorID: 3,
				Time:      util.Today(),
				Start:     time.Now(),
				EntryPath: "/",
				ExitPath:  "/product",
				IsBounce:  false,
				PageViews: 2,
			},
			{
				Sign:      1,
				VisitorID: 4,
				Time:      util.Today(),
				Start:     time.Now(),
				EntryPath: "/checkout",
				ExitPath:  "/checkout",
				IsBounce:  true,
				PageViews: 1,
			},
		},
	})
	assert.NoError(t, dbClient.SavePageViews(context.Background(), []model.PageView{
		{VisitorID: 1, Time: util.Today(), Path: "/"},
		{VisitorID: 1, Time: util.Today().Add(time.Second * 15), Path: "/product"},
		{VisitorID: 1, Time: util.Today().Add(time.Second * 131), Path: "/cart"},
		{VisitorID: 1, Time: util.Today().Add(time.Second * 140), Path: "/checkout"},
		{VisitorID: 1, Time: util.Today().Add(time.Second * 298), Path: "/thank-you"},

		{VisitorID: 2, Time: util.Today(), Path: "/product"},
		{VisitorID: 2, Time: util.Today().Add(time.Second * 5), Path: "/cart"},

		{VisitorID: 3, Time: util.Today(), Path: "/"},
		{VisitorID: 3, Time: util.Today().Add(time.Second * 12), Path: "/product"},

		{VisitorID: 4, Time: util.Today(), Path: "/checkout"},
	}))
	assert.NoError(t, dbClient.SaveEvents(context.Background(), []model.Event{
		{VisitorID: 1, Time: util.Today(), Path: "/product", Name: "Add to Cart", MetaKeys: []string{"product"}, MetaValues: []string{"42"}},
		{VisitorID: 1, Time: util.Today(), Path: "/checkout", Name: "Purchase", MetaKeys: []string{"amount", "currency"}, MetaValues: []string{"89.90", "USD"}},
		{VisitorID: 2, Time: util.Today(), Path: "/product", Name: "Add to Cart", MetaKeys: []string{"product"}, MetaValues: []string{"24"}},
		{VisitorID: 4, Time: util.Today(), Path: "/checkout", Name: "Purchase", MetaKeys: []string{"amount", "currency"}, MetaValues: []string{"29.95", "USD"}},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	funnel, err := analyzer.Funnel.Steps(context.Background(), []Filter{
		{
			Ctx:  context.Background(),
			From: util.Today(),
			To:   util.Today(),
			Path: []string{"/product"},
		},
		{
			Ctx:       context.Background(),
			From:      util.Today(),
			To:        util.Today(),
			EventName: []string{"Add to Cart"},
		},
		{
			Ctx:       context.Background(),
			From:      util.Today(),
			To:        util.Today(),
			EventName: []string{"Purchase"},
		},
	})
	assert.NoError(t, err)
	assert.Len(t, funnel, 3)
	assert.Equal(t, 0, funnel[0].Step)
	assert.Equal(t, 1, funnel[1].Step)
	assert.Equal(t, 2, funnel[2].Step)
	assert.Equal(t, 3, funnel[0].Visitors)
	assert.Equal(t, 2, funnel[1].Visitors)
	assert.Equal(t, 1, funnel[2].Visitors)
	assert.InDelta(t, 1, funnel[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.6666, funnel[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.3333, funnel[2].RelativeVisitors, 0.01)
	assert.Equal(t, 0, funnel[0].PreviousVisitors)
	assert.Equal(t, 3, funnel[1].PreviousVisitors)
	assert.Equal(t, 2, funnel[2].PreviousVisitors)
	assert.InDelta(t, 0, funnel[0].RelativePreviousVisitors, 0.01)
	assert.InDelta(t, 1, funnel[1].RelativePreviousVisitors, 0.01)
	assert.InDelta(t, 0.6666, funnel[2].RelativePreviousVisitors, 0.01)
	assert.Equal(t, 0, funnel[0].Dropped)
	assert.Equal(t, 1, funnel[1].Dropped)
	assert.Equal(t, 1, funnel[2].Dropped)
	assert.InDelta(t, 0, funnel[0].DropOff, 0.01)
	assert.InDelta(t, 0.3333, funnel[1].DropOff, 0.01)
	assert.InDelta(t, 0.5, funnel[2].DropOff, 0.01)
}
