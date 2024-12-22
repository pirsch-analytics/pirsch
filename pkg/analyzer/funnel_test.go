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
				ClientID:  1,
				VisitorID: 1,
				Time:      util.Today(),
				Start:     time.Now(),
				EntryPath: "/",
				ExitPath:  "/thank-you",
				IsBounce:  false,
				PageViews: 5,
				Language:  "en",
			},
			{
				Sign:      1,
				ClientID:  1,
				VisitorID: 2,
				Time:      util.Today(),
				Start:     time.Now(),
				EntryPath: "/product",
				ExitPath:  "/cart",
				IsBounce:  false,
				PageViews: 2,
				Language:  "en",
			},
			{
				Sign:      1,
				ClientID:  1,
				VisitorID: 3,
				Time:      util.Today(),
				Start:     time.Now(),
				EntryPath: "/",
				ExitPath:  "/product",
				IsBounce:  false,
				PageViews: 2,
				Language:  "de",
			},
			{
				Sign:      1,
				ClientID:  1,
				VisitorID: 4,
				Time:      util.Today(),
				Start:     time.Now(),
				EntryPath: "/checkout",
				ExitPath:  "/checkout",
				IsBounce:  true,
				PageViews: 1,
				Language:  "en",
			},
		},
	})
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{ClientID: 1, VisitorID: 1, Time: util.Today(), Path: "/", Language: "en", TagKeys: []string{"currency"}, TagValues: []string{"USD"}},
		{ClientID: 1, VisitorID: 1, Time: util.Today().Add(time.Second * 15), Path: "/product", Language: "en", TagKeys: []string{"currency"}, TagValues: []string{"USD"}},
		{ClientID: 1, VisitorID: 1, Time: util.Today().Add(time.Second * 131), Path: "/cart", Language: "en", TagKeys: []string{"currency"}, TagValues: []string{"USD"}},
		{ClientID: 1, VisitorID: 1, Time: util.Today().Add(time.Second * 140), Path: "/checkout", Language: "en", TagKeys: []string{"currency"}, TagValues: []string{"USD"}},
		{ClientID: 1, VisitorID: 1, Time: util.Today().Add(time.Second * 298), Path: "/thank-you", Language: "en", TagKeys: []string{"currency"}, TagValues: []string{"USD"}},

		{ClientID: 1, VisitorID: 2, Time: util.Today(), Path: "/product", Language: "en", TagKeys: []string{"currency"}, TagValues: []string{"USD"}},
		{ClientID: 1, VisitorID: 2, Time: util.Today().Add(time.Second * 5), Path: "/cart", Language: "en", TagKeys: []string{"currency"}, TagValues: []string{"USD"}},

		{ClientID: 1, VisitorID: 3, Time: util.Today(), Path: "/", Language: "de", TagKeys: []string{"currency"}, TagValues: []string{"EUR"}},
		{ClientID: 1, VisitorID: 3, Time: util.Today().Add(time.Second * 12), Path: "/product", Language: "de", TagKeys: []string{"currency"}, TagValues: []string{"EUR"}},

		{ClientID: 1, VisitorID: 4, Time: util.Today(), Path: "/checkout", Language: "en", TagKeys: []string{"currency"}, TagValues: []string{"USD"}},
	}))
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{ClientID: 1, VisitorID: 1, Time: util.Today(), Path: "/product", Name: "Add to Cart", MetaKeys: []string{"product"}, MetaValues: []string{"42"}, Language: "en"},
		{ClientID: 1, VisitorID: 1, Time: util.Today(), Path: "/checkout", Name: "Purchase", MetaKeys: []string{"amount", "currency"}, MetaValues: []string{"89.90", "USD"}, Language: "en"},
		{ClientID: 1, VisitorID: 2, Time: util.Today(), Path: "/product", Name: "Add to Cart", MetaKeys: []string{"product"}, MetaValues: []string{"24"}, Language: "en"},
		{ClientID: 1, VisitorID: 4, Time: util.Today(), Path: "/checkout", Name: "Purchase", MetaKeys: []string{"amount", "currency"}, MetaValues: []string{"29.95", "USD"}, Language: "en"},
	}))
	time.Sleep(time.Millisecond * 100)
	analyzer := NewAnalyzer(dbClient)
	_, err := analyzer.Funnel.Steps(context.Background(), []Filter{})
	assert.Equal(t, "not enough steps", err.Error())
	_, err = analyzer.Funnel.Steps(context.Background(), []Filter{{}})
	assert.Equal(t, "not enough steps", err.Error())

	// regular three-step funnel
	funnel, err := analyzer.Funnel.Steps(context.Background(), []Filter{
		{
			ClientID:    1,
			From:        util.Today(),
			To:          util.Today(),
			PathPattern: []string{"(?i)^/product.*$"},
			Language:    []string{"en", "de"},
			Sample:      1000,
		},
		{
			ClientID:  1,
			From:      util.Today(),
			To:        util.Today(),
			EntryPath: []string{"/", "/product"},
			Path:      []string{"/product"},
			EventName: []string{"Add to Cart"},
			Sample:    1000,
		},
		{
			ClientID:  1,
			From:      util.Today(),
			To:        util.Today(),
			EventName: []string{"Purchase"},
			EventMeta: map[string]string{"amount": "89.90", "currency": "USD"},
			Tags:      map[string]string{"currency": "USD"},
			Sample:    1000,
		},
	})
	assert.NoError(t, err)
	assert.Len(t, funnel, 3)
	assert.Equal(t, 1, funnel[0].Step)
	assert.Equal(t, 2, funnel[1].Step)
	assert.Equal(t, 3, funnel[2].Step)
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

	// minimum two-step funnel
	funnel, err = analyzer.Funnel.Steps(context.Background(), []Filter{
		{
			ClientID:  1,
			From:      util.Today(),
			To:        util.Today(),
			EntryPath: []string{"/"},
			Sample:    1000,
		},
		{
			ClientID: 1,
			From:     util.Today(),
			To:       util.Today(),
			Path:     []string{"/pricing"},
			Sample:   1000,
		},
	})
	assert.NoError(t, err)
	assert.Len(t, funnel, 2)

	// start with no visitors
	funnel, err = analyzer.Funnel.Steps(context.Background(), []Filter{
		{
			ClientID:  1,
			From:      util.Today(),
			To:        util.Today(),
			EntryPath: []string{"/does-not-exist"},
			Sample:    1000,
		},
		{
			ClientID: 1,
			From:     util.Today(),
			To:       util.Today(),
			Path:     []string{"/pricing"},
			Sample:   1000,
		},
	})
	assert.NoError(t, err)
	assert.Len(t, funnel, 2)
	assert.Equal(t, 1, funnel[0].Step)
	assert.Equal(t, 2, funnel[1].Step)
	assert.Equal(t, 0, funnel[0].Visitors)
	assert.Equal(t, 0, funnel[1].Visitors)
	assert.InDelta(t, 0, funnel[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0, funnel[1].RelativeVisitors, 0.01)
	assert.Equal(t, 0, funnel[0].PreviousVisitors)
	assert.Equal(t, 0, funnel[1].PreviousVisitors)
	assert.InDelta(t, 0, funnel[0].RelativePreviousVisitors, 0.01)
	assert.InDelta(t, 0, funnel[1].RelativePreviousVisitors, 0.01)
	assert.Equal(t, 0, funnel[0].Dropped)
	assert.Equal(t, 0, funnel[1].Dropped)
	assert.InDelta(t, 0, funnel[0].DropOff, 0.01)
	assert.InDelta(t, 0, funnel[1].DropOff, 0.01)
}
