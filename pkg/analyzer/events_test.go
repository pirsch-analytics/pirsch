package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v6/pkg"
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"github.com/pirsch-analytics/pirsch/v6/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAnalyzer_Events(t *testing.T) {
	db.CleanupDB(t, dbClient)

	// create sessions for the conversion rate
	for i := 0; i < 10; i++ {
		saveSessions(t, [][]model.Session{
			{
				{Sign: 1, VisitorID: uint64(i), Time: util.Today(), Start: time.Now(), EntryPath: "/", ExitPath: "/exit", PageViews: 1},
			},
			{
				{Sign: -1, VisitorID: uint64(i), Time: util.Today(), Start: time.Now(), EntryPath: "/", ExitPath: "/exit", PageViews: 1},
				{Sign: 1, VisitorID: uint64(i), Time: util.Today(), Start: time.Now(), EntryPath: "/", ExitPath: "/exit", PageViews: 1},
			},
		})
	}

	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.Today(), Path: "/"},
		{VisitorID: 2, Time: util.Today().Add(time.Second), Path: "/simple/page"},
		{VisitorID: 3, Time: util.Today().Add(time.Second * 2), Path: "/simple/page/1"},
		{VisitorID: 3, Time: util.Today().Add(time.Minute), Path: "/simple/page/2"},
		{VisitorID: 4, Time: util.Today(), Path: "/"},
		{VisitorID: 1, Time: util.Today().Add(time.Second * 4), Path: "/"},
		{VisitorID: 2, Time: util.Today().Add(time.Second * 5), Path: "/"},
		{VisitorID: 2, Time: util.Today().Add(time.Minute), Path: "/simple/page"},
		{VisitorID: 3, Time: util.Today().Add(time.Second * 6), Path: "/simple/page"},
		{VisitorID: 4, Time: util.Today().Add(time.Second * 7), Path: "/"},
		{VisitorID: 5, Time: util.Today().Add(time.Second * 8), Path: "/"},
	}))
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{Name: "event1", DurationSeconds: 5, MetaKeys: []string{"status", "price"}, MetaValues: []string{"in", "34.56"}, VisitorID: 1, Time: util.Today(), Path: "/"},
		{Name: "event1", DurationSeconds: 8, MetaKeys: []string{"status", "price"}, MetaValues: []string{"out", "34.56"}, VisitorID: 2, Time: util.Today().Add(time.Second), Path: "/simple/page"},
		{Name: "event1", DurationSeconds: 3, VisitorID: 3, Time: util.Today().Add(time.Second * 2), Path: "/simple/page/1"},
		{Name: "event1", DurationSeconds: 8, VisitorID: 3, Time: util.Today().Add(time.Minute), Path: "/simple/page/2"},
		{Name: "event1", DurationSeconds: 2, MetaKeys: []string{"status"}, MetaValues: []string{"in"}, VisitorID: 4, Time: util.Today().Add(time.Second * 3), Path: "/"},
		{Name: "event2", DurationSeconds: 1, VisitorID: 1, Time: util.Today().Add(time.Second * 4), Path: "/"},
		{Name: "event2", DurationSeconds: 5, VisitorID: 2, Time: util.Today().Add(time.Second * 5), Path: "/"},
		{Name: "event2", DurationSeconds: 7, MetaKeys: []string{"status", "price"}, MetaValues: []string{"in", "34.56"}, VisitorID: 2, Time: util.Today().Add(time.Minute), Path: "/simple/page"},
		{Name: "event2", DurationSeconds: 9, MetaKeys: []string{"status", "price", "third"}, MetaValues: []string{"in", "13.74", "param"}, VisitorID: 3, Time: util.Today().Add(time.Second * 6), Path: "/simple/page"},
		{Name: "event2", DurationSeconds: 3, MetaKeys: []string{"price"}, MetaValues: []string{"34.56"}, VisitorID: 4, Time: util.Today().Add(time.Second * 7), Path: "/"},
		{Name: "event2", DurationSeconds: 4, VisitorID: 5, Time: util.Today().Add(time.Second * 8), Path: "/"},
	}))
	time.Sleep(time.Millisecond * 100)
	analyzer := NewAnalyzer(dbClient)
	stats, err := analyzer.Events.Events(nil)
	assert.NoError(t, err)
	assert.Len(t, stats, 2)
	assert.Equal(t, "event2", stats[0].Name)
	assert.Equal(t, "event1", stats[1].Name)
	assert.Equal(t, 6, stats[0].Count)
	assert.Equal(t, 5, stats[1].Count)
	assert.Equal(t, 5, stats[0].Visitors)
	assert.Equal(t, 4, stats[1].Visitors)
	assert.Equal(t, 6, stats[0].Views)
	assert.Equal(t, 5, stats[1].Views)
	assert.InDelta(t, 0.5, stats[0].CR, 0.001)
	assert.InDelta(t, 0.4, stats[1].CR, 0.001)
	assert.InDelta(t, 4, stats[0].AverageDurationSeconds, 0.001)
	assert.InDelta(t, 5, stats[1].AverageDurationSeconds, 0.001)
	assert.Len(t, stats[0].MetaKeys, 3)
	assert.Len(t, stats[1].MetaKeys, 2)
	stats, err = analyzer.Events.Events(&Filter{Sort: []Sort{{Field: FieldCount, Direction: pkg.DirectionASC}}})
	assert.NoError(t, err)
	assert.Len(t, stats, 2)
	assert.Equal(t, "event1", stats[0].Name)
	assert.Equal(t, "event2", stats[1].Name)
	assert.Equal(t, 5, stats[0].Count)
	assert.Equal(t, 6, stats[1].Count)
	stats, err = analyzer.Events.Events(&Filter{EntryPath: []string{"/exit"}})
	assert.NoError(t, err)
	assert.Len(t, stats, 0)
	stats, err = analyzer.Events.Events(&Filter{EntryPath: []string{"/"}, ExitPath: []string{"/exit"}})
	assert.NoError(t, err)
	assert.Len(t, stats, 2)
	assert.Equal(t, "event2", stats[0].Name)
	assert.Equal(t, "event1", stats[1].Name)
	assert.Equal(t, 6, stats[0].Count)
	assert.Equal(t, 5, stats[1].Count)
	assert.Equal(t, 5, stats[0].Visitors)
	assert.Equal(t, 4, stats[1].Visitors)
	assert.Equal(t, 6, stats[0].Views)
	assert.Equal(t, 5, stats[1].Views)
	assert.InDelta(t, 0.5, stats[0].CR, 0.001)
	assert.InDelta(t, 0.4, stats[1].CR, 0.001)
	assert.InDelta(t, 4, stats[0].AverageDurationSeconds, 0.001)
	assert.InDelta(t, 5, stats[1].AverageDurationSeconds, 0.001)
	assert.Len(t, stats[0].MetaKeys, 3)
	assert.Len(t, stats[1].MetaKeys, 2)
	stats, err = analyzer.Events.Events(&Filter{EventName: []string{"event2"}})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, "event2", stats[0].Name)
	assert.Equal(t, 6, stats[0].Count)
	assert.Equal(t, 5, stats[0].Visitors)
	assert.Equal(t, 6, stats[0].Views)
	assert.InDelta(t, 0.5, stats[0].CR, 0.001)
	assert.InDelta(t, 4, stats[0].AverageDurationSeconds, 0.001)
	assert.Len(t, stats[0].MetaKeys, 3)
	stats, err = analyzer.Events.Events(&Filter{EventName: []string{"does-not-exist"}})
	assert.NoError(t, err)
	assert.Empty(t, stats)
	_, err = analyzer.Events.Events(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Events.Events(&Filter{Offset: 1, Limit: 10, Sort: []Sort{
		{
			Field:     FieldEventName,
			Direction: pkg.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldEventName,
			Input: "event",
		},
	}})
	assert.NoError(t, err)
	stats, err = analyzer.Events.Breakdown(&Filter{EventName: []string{"event1"}, EventMetaKey: []string{"status"}})
	assert.NoError(t, err)
	assert.Len(t, stats, 2)
	assert.Equal(t, "event1", stats[0].Name)
	assert.Equal(t, "event1", stats[1].Name)
	assert.Equal(t, 2, stats[0].Count)
	assert.Equal(t, 1, stats[1].Count)
	assert.Equal(t, 2, stats[0].Visitors)
	assert.Equal(t, 1, stats[1].Visitors)
	assert.Equal(t, 2, stats[0].Views)
	assert.Equal(t, 1, stats[1].Views)
	assert.InDelta(t, 0.2, stats[0].CR, 0.001)
	assert.InDelta(t, 0.1, stats[1].CR, 0.001)
	assert.InDelta(t, 3, stats[0].AverageDurationSeconds, 0.001)
	assert.InDelta(t, 8, stats[1].AverageDurationSeconds, 0.001)
	assert.Equal(t, "in", stats[0].MetaValue)
	assert.Equal(t, "out", stats[1].MetaValue)
	stats, err = analyzer.Events.Breakdown(&Filter{EventName: []string{"event1"}, EventMetaKey: []string{"status"}, Sort: []Sort{{Field: FieldCount, Direction: pkg.DirectionASC}}})
	assert.NoError(t, err)
	assert.Len(t, stats, 2)
	assert.Equal(t, "event1", stats[0].Name)
	assert.Equal(t, "event1", stats[1].Name)
	assert.Equal(t, 1, stats[0].Count)
	assert.Equal(t, 2, stats[1].Count)
	stats, err = analyzer.Events.Breakdown(&Filter{EventName: []string{"event2"}, EventMetaKey: []string{"status"}})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, "event2", stats[0].Name)
	assert.Equal(t, 2, stats[0].Count)
	assert.Equal(t, 2, stats[0].Visitors)
	assert.Equal(t, 2, stats[0].Views)
	assert.InDelta(t, 0.2, stats[0].CR, 0.001)
	assert.InDelta(t, 8, stats[0].AverageDurationSeconds, 0.001)
	assert.Equal(t, "in", stats[0].MetaValue)
	stats, err = analyzer.Events.Breakdown(&Filter{EventName: []string{"event2"}, EventMetaKey: []string{"price"}})
	assert.NoError(t, err)
	assert.Len(t, stats, 2)
	assert.Equal(t, "event2", stats[0].Name)
	assert.Equal(t, "event2", stats[1].Name)
	assert.Equal(t, 2, stats[0].Count)
	assert.Equal(t, 1, stats[1].Count)
	assert.Equal(t, 2, stats[0].Visitors)
	assert.Equal(t, 1, stats[1].Visitors)
	assert.Equal(t, 2, stats[0].Views)
	assert.Equal(t, 1, stats[1].Views)
	assert.InDelta(t, 0.2, stats[0].CR, 0.001)
	assert.InDelta(t, 0.1, stats[1].CR, 0.001)
	assert.InDelta(t, 5, stats[0].AverageDurationSeconds, 0.001)
	assert.InDelta(t, 9, stats[1].AverageDurationSeconds, 0.001)
	assert.Equal(t, "34.56", stats[0].MetaValue)
	assert.Equal(t, "13.74", stats[1].MetaValue)
	stats, err = analyzer.Events.Breakdown(&Filter{EventName: []string{"event2"}, EventMetaKey: []string{"third"}})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, "event2", stats[0].Name)
	assert.Equal(t, 1, stats[0].Count)
	assert.Equal(t, 1, stats[0].Visitors)
	assert.Equal(t, 1, stats[0].Views)
	assert.InDelta(t, 0.1, stats[0].CR, 0.001)
	assert.InDelta(t, 9, stats[0].AverageDurationSeconds, 0.001)
	assert.Equal(t, "param", stats[0].MetaValue)
	stats, err = analyzer.Events.Breakdown(&Filter{
		EventName:    []string{"event1"},
		EventMetaKey: []string{"status"},
		PathPattern:  []string{"(?i)/simple/.*$"},
	})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, "event1", stats[0].Name)
	assert.Equal(t, 2, stats[0].Count)
	assert.Equal(t, 1, stats[0].Visitors)
	assert.Equal(t, 2, stats[0].Views)
	assert.InDelta(t, 0.1, stats[0].CR, 0.001)
	assert.Equal(t, "out", stats[0].MetaValue)
	stats, err = analyzer.Events.Breakdown(&Filter{EventName: []string{"does-not-exist"}, EventMetaKey: []string{"status"}})
	assert.NoError(t, err)
	assert.Empty(t, stats)
	stats, err = analyzer.Events.Breakdown(&Filter{EventName: []string{"event1"}, EventMetaKey: []string{"does-not-exist"}})
	assert.NoError(t, err)
	assert.Empty(t, stats)
	_, err = analyzer.Events.Breakdown(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Events.Breakdown(getMaxFilter("event"))
	assert.NoError(t, err)
}

func TestAnalyzer_EventsSortCR(t *testing.T) {
	db.CleanupDB(t, dbClient)

	// create sessions for the conversion rate
	for i := 0; i < 10; i++ {
		saveSessions(t, [][]model.Session{
			{
				{Sign: 1, VisitorID: uint64(i), Time: util.Today(), Start: time.Now(), EntryPath: "/", ExitPath: "/exit", PageViews: 1},
			},
		})
	}

	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{Name: "event1", VisitorID: 1, Time: util.Today(), Path: "/"},
		{Name: "event1", VisitorID: 2, Time: util.Today().Add(time.Second), Path: "/"},
		{Name: "event1", VisitorID: 3, Time: util.Today().Add(time.Second * 2), Path: "/"},
		{Name: "event2", VisitorID: 3, Time: util.Today().Add(time.Minute), Path: "/"},
	}))
	time.Sleep(time.Millisecond * 100)
	analyzer := NewAnalyzer(dbClient)
	stats, err := analyzer.Events.Events(&Filter{Sort: []Sort{
		{
			Field:     FieldCR,
			Direction: pkg.DirectionASC,
		},
	}})
	assert.NoError(t, err)
	assert.Len(t, stats, 2)
	assert.Equal(t, "event2", stats[0].Name)
	assert.Equal(t, "event1", stats[1].Name)
	assert.InDelta(t, 0.1, stats[0].CR, 0.001)
	assert.InDelta(t, 0.3, stats[1].CR, 0.001)
	stats, err = analyzer.Events.Events(&Filter{Sort: []Sort{
		{
			Field:     FieldCR,
			Direction: pkg.DirectionDESC,
		},
	}})
	assert.NoError(t, err)
	assert.Len(t, stats, 2)
	assert.Equal(t, "event1", stats[0].Name)
	assert.Equal(t, "event2", stats[1].Name)
	assert.InDelta(t, 0.3, stats[0].CR, 0.001)
	assert.InDelta(t, 0.1, stats[1].CR, 0.001)
}

func TestAnalyzer_EventList(t *testing.T) {
	db.CleanupDB(t, dbClient)

	// create sessions for the conversion rate
	for i := 0; i < 5; i++ {
		saveSessions(t, [][]model.Session{
			{
				{Sign: 1, VisitorID: uint64(i + 1), Time: util.Today(), Start: time.Now(), EntryPath: "/", ExitPath: "/exit"},
			},
			{
				{Sign: -1, VisitorID: uint64(i + 1), Time: util.Today(), Start: time.Now(), EntryPath: "/", ExitPath: "/exit"},
				{Sign: 1, VisitorID: uint64(i + 1), Time: util.Today(), Start: time.Now(), EntryPath: "/", ExitPath: "/exit"},
			},
		})
	}

	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.Today(), Path: "/"},
		{VisitorID: 2, Time: util.Today(), Path: "/foo"},
		{VisitorID: 1, Time: util.Today(), Path: "/bar"},
		{VisitorID: 3, Time: util.Today(), Path: "/"},
		{VisitorID: 4, Time: util.Today(), Path: "/"},
		{VisitorID: 5, Time: util.Today(), Path: "/foo"},
	}))
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{Name: "event1", MetaKeys: []string{"a", "b"}, MetaValues: []string{"foo", "42"}, VisitorID: 1, Time: util.Today(), Path: "/"},
		{Name: "event1", MetaKeys: []string{"b", "a"}, MetaValues: []string{"42", "foo"}, VisitorID: 2, Time: util.Today(), Path: "/foo"},
		{Name: "event1", MetaKeys: []string{"a", "b"}, MetaValues: []string{"bar", "42"}, VisitorID: 1, Time: util.Today(), Path: "/bar"},
		{Name: "event2", MetaKeys: []string{"b", "a"}, MetaValues: []string{"42", "foo"}, VisitorID: 3, Time: util.Today(), Path: "/"},
		{Name: "event2", MetaKeys: []string{"b", "a"}, MetaValues: []string{"56", "foo"}, VisitorID: 4, Time: util.Today(), Path: "/"},
		{Name: "event2", MetaKeys: []string{"a", "b"}, MetaValues: []string{"foo", "42"}, VisitorID: 5, Time: util.Today(), Path: "/foo"},
	}))
	analyzer := NewAnalyzer(dbClient)
	stats, err := analyzer.Events.List(nil)
	assert.NoError(t, err)
	assert.Len(t, stats, 4)
	assert.Equal(t, "event1", stats[0].Name)
	assert.Equal(t, "event2", stats[1].Name)
	assert.Equal(t, "event1", stats[2].Name)
	assert.Equal(t, "event2", stats[3].Name)
	assert.Equal(t, 2, stats[0].Count)
	assert.Equal(t, 2, stats[1].Count)
	assert.Equal(t, 1, stats[2].Count)
	assert.Equal(t, 1, stats[3].Count)
	assert.Equal(t, "foo", stats[0].Meta["a"])
	assert.Equal(t, "42", stats[0].Meta["b"])
	assert.Equal(t, "foo", stats[1].Meta["a"])
	assert.Equal(t, "42", stats[1].Meta["b"])
	assert.Equal(t, "bar", stats[2].Meta["a"])
	assert.Equal(t, "42", stats[2].Meta["b"])
	assert.Equal(t, "foo", stats[3].Meta["a"])
	assert.Equal(t, "56", stats[3].Meta["b"])
	stats, err = analyzer.Events.List(&Filter{
		EventName: []string{"event1"},
		Path:      []string{"/foo"},
	})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, "event1", stats[0].Name)
	assert.Equal(t, 1, stats[0].Count)
	assert.Equal(t, "foo", stats[0].Meta["a"])
	assert.Equal(t, "42", stats[0].Meta["b"])
	stats, err = analyzer.Events.List(&Filter{Path: []string{"/foo"}})
	assert.NoError(t, err)
	assert.Len(t, stats, 2)
	assert.Equal(t, "event1", stats[0].Name)
	assert.Equal(t, "event2", stats[1].Name)
	stats, err = analyzer.Events.List(&Filter{EventMeta: map[string]string{"a": "bar"}})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, "event1", stats[0].Name)
	assert.Equal(t, 1, stats[0].Count)
	assert.Equal(t, "bar", stats[0].Meta["a"])
	stats, err = analyzer.Events.List(&Filter{EventMeta: map[string]string{"a": "foo", "b": "56"}})
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, "event2", stats[0].Name)
	assert.Equal(t, 1, stats[0].Count)
	assert.Equal(t, "foo", stats[0].Meta["a"])
	assert.Equal(t, "56", stats[0].Meta["b"])
	stats, err = analyzer.Events.List(&Filter{EventMeta: map[string]string{"a": "no", "b": "result"}})
	assert.NoError(t, err)
	assert.Len(t, stats, 0)
	_, err = analyzer.Events.List(&Filter{Offset: 1, Limit: 10, Sort: []Sort{
		{
			Field:     FieldEventName,
			Direction: pkg.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldEventName,
			Input: "event",
		},
	}})
	assert.NoError(t, err)
}

func TestAnalyzer_EventFilter(t *testing.T) {
	db.CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.Today(), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 1, Time: util.Today(), Path: "/foo", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 1, Time: util.Today(), Path: "/bar", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
		{VisitorID: 2, Time: util.Today(), Path: "/foo", TagKeys: []string{"author"}, TagValues: []string{"John"}},
		{VisitorID: 3, Time: util.Today(), Path: "/", TagKeys: []string{"author"}, TagValues: []string{"Alice"}},
	}))
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.Today(), Start: time.Now(), EntryPath: "/", ExitPath: "/", IsBounce: true, PageViews: 1},
			{Sign: 1, VisitorID: 1, Time: util.Today(), Start: time.Now(), EntryPath: "/", ExitPath: "/foo", IsBounce: false, PageViews: 2},
			{Sign: 1, VisitorID: 1, Time: util.Today(), Start: time.Now(), EntryPath: "/", ExitPath: "/bar", IsBounce: false, PageViews: 3},
			{Sign: 1, VisitorID: 2, Time: util.Today(), Start: time.Now(), EntryPath: "/foo", ExitPath: "/foo", IsBounce: true, PageViews: 1},
			{Sign: 1, VisitorID: 3, Time: util.Today(), Start: time.Now(), EntryPath: "/", ExitPath: "/", IsBounce: true, PageViews: 1},
		},
	})
	assert.NoError(t, dbClient.SaveEvents([]model.Event{
		{VisitorID: 1, Time: util.Today(), Name: "event1", MetaKeys: []string{"k0", "k1", "author"}, MetaValues: []string{"v0", "v1", "John"}},
		{VisitorID: 3, Time: util.Today(), Name: "event2", MetaKeys: []string{"k2", "k3", "author"}, MetaValues: []string{"v2", "v3", "Alice"}},
	}))
	time.Sleep(time.Millisecond * 100)
	analyzer := NewAnalyzer(dbClient)
	list, err := analyzer.Events.Events(&Filter{
		EventName: []string{"event1"},
	})
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "event1", list[0].Name)
	list, err = analyzer.Events.Events(&Filter{
		EventName: []string{"!event1"},
	})
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "event2", list[0].Name)
	list, err = analyzer.Events.Events(&Filter{
		EventName: []string{"event1"},
		Tags:      map[string]string{"author": "John"},
	})
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "event1", list[0].Name)
	list, err = analyzer.Events.Events(&Filter{
		EventName: []string{"event1"},
		Tags:      map[string]string{"author": "!John"},
	})
	assert.NoError(t, err)
	assert.Empty(t, list)
	list, err = analyzer.Events.Events(&Filter{
		EventName: []string{"~event"},
		EventMeta: map[string]string{"author": "~ice"},
	})
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "event2", list[0].Name)
	assert.Contains(t, list[0].MetaKeys, "author")
	list, err = analyzer.Events.Events(&Filter{
		EventName: []string{"~event"},
		Tags:      map[string]string{"author": "~ice"},
	})
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "event2", list[0].Name)
	assert.Contains(t, list[0].MetaKeys, "author")

	breakdown, err := analyzer.Events.Breakdown(&Filter{EventName: []string{"event1"}, EventMetaKey: []string{"k0"}})
	assert.NoError(t, err)
	assert.Len(t, breakdown, 1)
	assert.Equal(t, "event1", breakdown[0].Name)
	breakdown, err = analyzer.Events.Breakdown(&Filter{EventName: []string{"!event1"}, EventMetaKey: []string{"k2"}})
	assert.NoError(t, err)
	assert.Len(t, breakdown, 1)
	assert.Equal(t, "event2", breakdown[0].Name)
	breakdown, err = analyzer.Events.Breakdown(&Filter{
		EventName:    []string{"!event1"},
		EventMetaKey: []string{"k2"},
		Tags:         map[string]string{"author": "Alice"},
	})
	assert.NoError(t, err)
	assert.Len(t, breakdown, 1)
	assert.Equal(t, "event2", breakdown[0].Name)
	breakdown, err = analyzer.Events.Breakdown(&Filter{
		EventName:    []string{"!event1"},
		EventMetaKey: []string{"k2"},
		Tags:         map[string]string{"author": "!Alice"},
	})
	assert.NoError(t, err)
	assert.Empty(t, breakdown)
	breakdown, err = analyzer.Events.Breakdown(&Filter{
		EventName:    []string{"~event"},
		EventMetaKey: []string{"author"},
		EventMeta:    map[string]string{"author": "~ice"},
	})
	assert.NoError(t, err)
	assert.Len(t, breakdown, 1)
	assert.Equal(t, "event2", breakdown[0].Name)
	assert.Equal(t, "Alice", breakdown[0].MetaValue)
	breakdown, err = analyzer.Events.Breakdown(&Filter{
		EventName:    []string{"~event"},
		EventMetaKey: []string{"author"},
		Tags:         map[string]string{"author": "~ice"},
	})
	assert.NoError(t, err)
	assert.Len(t, breakdown, 1)
	assert.Equal(t, "event2", breakdown[0].Name)
	assert.Equal(t, "Alice", breakdown[0].MetaValue)

	eventList, err := analyzer.Events.List(&Filter{EventName: []string{"event1"}})
	assert.NoError(t, err)
	assert.Len(t, eventList, 1)
	assert.Equal(t, "event1", eventList[0].Name)
	eventList, err = analyzer.Events.List(&Filter{EventName: []string{"!event1"}})
	assert.NoError(t, err)
	assert.Len(t, eventList, 1)
	assert.Equal(t, "event2", eventList[0].Name)
	eventList, err = analyzer.Events.List(&Filter{EventName: []string{"!event1"}, EventMeta: map[string]string{"k2": "v2"}})
	assert.NoError(t, err)
	assert.Len(t, eventList, 1)
	assert.Equal(t, "event2", eventList[0].Name)
	assert.Equal(t, "v2", eventList[0].Meta["k2"])
	eventList, err = analyzer.Events.List(&Filter{
		EventName: []string{"event2"},
		EventMeta: map[string]string{"k2": "v2"},
		Tags:      map[string]string{"author": "Alice"},
	})
	assert.NoError(t, err)
	assert.Len(t, eventList, 1)
	assert.Equal(t, "event2", eventList[0].Name)
	assert.Equal(t, "v2", eventList[0].Meta["k2"])
	eventList, err = analyzer.Events.List(&Filter{
		EventName: []string{"event2"},
		EventMeta: map[string]string{"k2": "v2"},
		Tags:      map[string]string{"author": "!Alice"},
	})
	assert.NoError(t, err)
	assert.Empty(t, eventList)
}
