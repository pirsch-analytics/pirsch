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
	}))
	assert.NoError(t, dbClient.SaveEvents(context.Background(), []model.Event{
		{VisitorID: 1, Time: util.Today(), Path: "/product", Name: "Add to Cart", MetaKeys: []string{"product"}, MetaValues: []string{"42"}},
		{VisitorID: 1, Time: util.Today(), Path: "/checkout", Name: "Purchase", MetaKeys: []string{"amount", "currency"}, MetaValues: []string{"89.90", "USD"}},
		{VisitorID: 2, Time: util.Today(), Path: "/product", Name: "Add to Cart", MetaKeys: []string{"product"}, MetaValues: []string{"24"}},
	}))
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	analyzer.Funnel.Steps([]Filter{
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
}
