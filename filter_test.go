package pirsch

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFilter_Validate(t *testing.T) {
	filter := NewFilter(NullClient)
	filter.validate()
	assert.NotNil(t, filter)
	assert.NotNil(t, filter.Timezone)
	assert.Equal(t, time.UTC, filter.Timezone)
	assert.Zero(t, filter.From)
	assert.Zero(t, filter.To)
	filter = &Filter{From: pastDay(2), To: pastDay(5), Limit: 42}
	filter.validate()
	assert.Equal(t, pastDay(5), filter.From)
	assert.Equal(t, pastDay(2), filter.To)
	assert.Equal(t, 42, filter.Limit)
	filter = &Filter{From: pastDay(2), To: Today().Add(time.Hour * 24 * 5)}
	filter.validate()
	assert.Equal(t, pastDay(2), filter.From)
	assert.Equal(t, Today().Add(time.Hour*24), filter.To)
	filter = &Filter{Limit: -42, Path: "/path", PathPattern: "pattern"}
	filter.validate()
	assert.Zero(t, filter.Limit)
	assert.Equal(t, "/path", filter.Path)
	assert.Empty(t, filter.PathPattern)
	filter = &Filter{Limit: -42, PathPattern: "pattern"}
	filter.validate()
	assert.Empty(t, filter.Path)
	assert.Equal(t, "pattern", filter.PathPattern)
}

func TestFilter_BuildQuery(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SavePageViews([]PageView{
		{VisitorID: 1, Time: Today(), Path: "/"},
		{VisitorID: 1, Time: Today().Add(time.Minute * 2), Path: "/foo"},
		{VisitorID: 1, Time: Today().Add(time.Minute*2 + time.Second*2), Path: "/foo"},
		{VisitorID: 1, Time: Today().Add(time.Minute*2 + time.Second*23), Path: "/bar"},

		{VisitorID: 2, Time: Today(), Path: "/bar"},
		{VisitorID: 2, Time: Today().Add(time.Second * 16), Path: "/foo"},
		{VisitorID: 2, Time: Today().Add(time.Second*16 + time.Second*8), Path: "/"},
	}))
	saveSessions(t, [][]Session{
		{
			{Sign: 1, VisitorID: 1, Time: Today(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 2, Time: Today(), EntryPath: "/bar", ExitPath: "/bar", PageViews: 1},
		},
		{
			{Sign: -1, VisitorID: 1, Time: Today(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 1, Time: Today().Add(time.Minute * 2), EntryPath: "/", ExitPath: "/foo", PageViews: 2},
			{Sign: -1, VisitorID: 2, Time: Today(), EntryPath: "/bar", ExitPath: "/bar", PageViews: 1},
			{Sign: 1, VisitorID: 2, Time: Today().Add(time.Second * 16), EntryPath: "/bar", ExitPath: "/foo", PageViews: 2},
		},
		{
			{Sign: -1, VisitorID: 1, Time: Today().Add(time.Minute * 2), EntryPath: "/", ExitPath: "/foo", PageViews: 2},
			{Sign: 1, VisitorID: 1, Time: Today().Add(time.Minute*2 + time.Second*23), EntryPath: "/", ExitPath: "/bar", PageViews: 3},
			{Sign: -1, VisitorID: 2, Time: Today().Add(time.Second * 16), EntryPath: "/bar", ExitPath: "/foo", PageViews: 2},
			{Sign: 1, VisitorID: 2, Time: Today().Add(time.Second*16 + time.Second*8), EntryPath: "/bar", ExitPath: "/", PageViews: 3},
		},
	})

	// no filter (from page views)
	analyzer := NewAnalyzer(dbClient, nil)
	args, query := analyzer.getFilter(nil).buildQuery([]Field{FieldPath, FieldVisitors}, []Field{FieldPath}, []Field{FieldVisitors, FieldPath})
	var stats []PageStats
	assert.NoError(t, dbClient.Select(&stats, query, args...))
	assert.Len(t, stats, 3)
	assert.Equal(t, 2, stats[0].Visitors)
	assert.Equal(t, 2, stats[1].Visitors)
	assert.Equal(t, 2, stats[2].Visitors)
	assert.Equal(t, "/", stats[0].Path)
	assert.Equal(t, "/bar", stats[1].Path)
	assert.Equal(t, "/foo", stats[2].Path)

	// join (from page views)
	args, query = analyzer.getFilter(&Filter{EntryPath: "/"}).buildQuery([]Field{FieldPath, FieldVisitors}, []Field{FieldPath}, []Field{FieldPath})
	stats = stats[:0]
	assert.NoError(t, dbClient.Select(&stats, query, args...))
	assert.Len(t, stats, 3)
	assert.Equal(t, 1, stats[0].Visitors)
	assert.Equal(t, 1, stats[1].Visitors)
	assert.Equal(t, 1, stats[2].Visitors)
	assert.Equal(t, "/", stats[0].Path)
	assert.Equal(t, "/bar", stats[1].Path)
	assert.Equal(t, "/foo", stats[2].Path)

	// join and filter (from page views)
	args, query = analyzer.getFilter(&Filter{EntryPath: "/", Path: "/foo"}).buildQuery([]Field{FieldPath, FieldVisitors}, []Field{FieldPath}, []Field{FieldPath})
	stats = stats[:0]
	assert.NoError(t, dbClient.Select(&stats, query, args...))
	assert.Len(t, stats, 1)
	assert.Equal(t, "/foo", stats[0].Path)
	assert.Equal(t, 1, stats[0].Visitors)

	// filter (from page views)
	args, query = analyzer.getFilter(&Filter{Path: "/foo"}).buildQuery([]Field{FieldPath, FieldVisitors}, []Field{FieldPath}, []Field{FieldPath})
	stats = stats[:0]
	assert.NoError(t, dbClient.Select(&stats, query, args...))
	assert.Len(t, stats, 1)
	assert.Equal(t, "/foo", stats[0].Path)
	assert.Equal(t, 2, stats[0].Visitors)

	// no filter (from sessions)
	args, query = analyzer.getFilter(nil).buildQuery([]Field{FieldVisitors, FieldSessions, FieldViews, FieldBounces, FieldBounceRate}, nil, nil)
	var vstats PageStats
	assert.NoError(t, dbClient.Get(&vstats, query, args...))
	assert.Equal(t, 2, vstats.Visitors)
	assert.Equal(t, 2, vstats.Sessions)
	assert.Equal(t, 6, vstats.Views)
	assert.Equal(t, 0, vstats.Bounces)
	assert.InDelta(t, 0, vstats.BounceRate, 0.01)

	// filter (from page views)
	args, query = analyzer.getFilter(&Filter{Path: "/foo", EntryPath: "/"}).buildQuery([]Field{FieldVisitors, FieldRelativeVisitors, FieldSessions, FieldViews, FieldRelativeViews, FieldBounces, FieldBounceRate}, nil, nil)
	assert.NoError(t, dbClient.Get(&vstats, query, args...))
	assert.Equal(t, 1, vstats.Visitors)
	assert.Equal(t, 1, vstats.Sessions)
	assert.Equal(t, 2, vstats.Views)
	assert.Equal(t, 0, vstats.Bounces)
	assert.InDelta(t, 0, vstats.BounceRate, 0.01)
	assert.InDelta(t, 0.5, vstats.RelativeVisitors, 0.01)
	assert.InDelta(t, 0.3333, vstats.RelativeViews, 0.01)

	// filter period
	args, query = analyzer.getFilter(&Filter{Period: PeriodWeek}).buildQuery([]Field{FieldDay, FieldVisitors}, []Field{FieldDay}, []Field{FieldDay})
	var visitors []VisitorStats
	assert.NoError(t, dbClient.Select(&visitors, query, args...))
	assert.Len(t, visitors, 1)
}

func TestFilter_Table(t *testing.T) {
	filter := NewFilter(NullClient)
	assert.Equal(t, "session", filter.table())
	filter.eventFilter = true
	assert.Equal(t, "event", filter.table())
	filter.eventFilter = false
	filter.EventName = "event"
	assert.Equal(t, "event", filter.table())
}

func TestFilter_QueryTime(t *testing.T) {
	filter := NewFilter(NullClient)
	filter.From = pastDay(5)
	filter.To = pastDay(2)
	args, query := filter.queryTime(false)
	assert.Len(t, args, 3)
	assert.Equal(t, NullClient, args[0])
	assert.Equal(t, filter.From, args[1])
	assert.Equal(t, filter.To, args[2])
	assert.Equal(t, "client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) ", query)
	filter.From = pastDay(2)
	filter.validate()
	args, query = filter.queryTime(false)
	assert.Len(t, args, 2)
	assert.Equal(t, NullClient, args[0])
	assert.Equal(t, filter.From, args[1])
	assert.Equal(t, "client_id = ? AND toDate(time, 'UTC') = toDate(?) ", query)
	filter.IncludeTime = true
	filter.From = filter.From.Add(time.Hour * 5)
	filter.To = filter.To.Add(time.Hour * 19)
	filter.validate()
	args, query = filter.queryTime(false)
	assert.Len(t, args, 3)
	assert.Equal(t, NullClient, args[0])
	assert.Equal(t, filter.From, args[1])
	assert.Equal(t, filter.To, args[2])
	assert.Equal(t, "client_id = ? AND toDateTime(time, 'UTC') >= toDateTime(?, 'UTC') AND toDateTime(time, 'UTC') <= toDateTime(?, 'UTC') ", query)
}

func TestFilter_QueryFields(t *testing.T) {
	filter := NewFilter(NullClient)
	filter.Path = "/"
	filter.EntryPath = "/entry"
	filter.ExitPath = "/exit"
	filter.PathPattern = "pattern"
	filter.Language = "en"
	filter.Country = "jp"
	filter.City = "Tokyo"
	filter.Referrer = "ref"
	filter.ReferrerName = "refname"
	filter.OS = OSWindows
	filter.OSVersion = "10"
	filter.Browser = BrowserEdge
	filter.BrowserVersion = "89"
	filter.Platform = PlatformUnknown
	filter.ScreenClass = "XXL"
	filter.ScreenWidth = "1920"
	filter.ScreenHeight = "1080"
	filter.UTMSource = "source"
	filter.UTMMedium = "medium"
	filter.UTMCampaign = "campaign"
	filter.UTMContent = "content"
	filter.UTMTerm = "term"
	filter.EventMeta = map[string]string{
		"foo": "bar",
	}
	filter.validate()
	args, query := filter.queryFields()
	assert.Len(t, args, 21)
	assert.Equal(t, "/", args[0])
	assert.Equal(t, "/entry", args[1])
	assert.Equal(t, "/exit", args[2])
	assert.Equal(t, "en", args[3])
	assert.Equal(t, "jp", args[4])
	assert.Equal(t, "Tokyo", args[5])
	assert.Equal(t, "ref", args[6])
	assert.Equal(t, "refname", args[7])
	assert.Equal(t, OSWindows, args[8])
	assert.Equal(t, "10", args[9])
	assert.Equal(t, BrowserEdge, args[10])
	assert.Equal(t, "89", args[11])
	assert.Equal(t, "XXL", args[12])
	assert.Equal(t, uint16(1920), args[13])
	assert.Equal(t, uint16(1080), args[14])
	assert.Equal(t, "source", args[15])
	assert.Equal(t, "medium", args[16])
	assert.Equal(t, "campaign", args[17])
	assert.Equal(t, "content", args[18])
	assert.Equal(t, "term", args[19])
	assert.Equal(t, "bar", args[20])
	assert.Equal(t, "path = ? AND entry_path = ? AND exit_path = ? AND language = ? AND country_code = ? AND city = ? AND referrer = ? AND referrer_name = ? AND os = ? AND os_version = ? AND browser = ? AND browser_version = ? AND screen_class = ? AND screen_width = ? AND screen_height = ? AND utm_source = ? AND utm_medium = ? AND utm_campaign = ? AND utm_content = ? AND utm_term = ? AND desktop = 0 AND mobile = 0 AND event_meta_values[indexOf(event_meta_keys, 'foo')] = ? ", query)
	filter.EventName = "event"
	args, query = filter.queryFields()
	assert.Len(t, args, 22)
	assert.Equal(t, "event", args[20])
	assert.Equal(t, "path = ? AND entry_path = ? AND exit_path = ? AND language = ? AND country_code = ? AND city = ? AND referrer = ? AND referrer_name = ? AND os = ? AND os_version = ? AND browser = ? AND browser_version = ? AND screen_class = ? AND screen_width = ? AND screen_height = ? AND utm_source = ? AND utm_medium = ? AND utm_campaign = ? AND utm_content = ? AND utm_term = ? AND event_name = ? AND desktop = 0 AND mobile = 0 AND event_meta_values[indexOf(event_meta_keys, 'foo')] = ? ", query)
}

func TestFilter_QueryFieldsInvert(t *testing.T) {
	filter := NewFilter(NullClient)
	filter.Path = "!/"
	filter.EntryPath = "!/entry"
	filter.ExitPath = "!/exit"
	filter.PathPattern = "!pattern"
	filter.Language = "!en"
	filter.Country = "!jp"
	filter.City = "!Tokyo"
	filter.Referrer = "!ref"
	filter.ReferrerName = "!refname"
	filter.OS = "!" + OSWindows
	filter.OSVersion = "!10"
	filter.Browser = "!" + BrowserEdge
	filter.BrowserVersion = "!89"
	filter.Platform = "!" + PlatformUnknown
	filter.ScreenClass = "!XXL"
	filter.ScreenWidth = "!1920"
	filter.ScreenHeight = "!1080"
	filter.UTMSource = "!source"
	filter.UTMMedium = "!medium"
	filter.UTMCampaign = "!campaign"
	filter.UTMContent = "!content"
	filter.UTMTerm = "!term"
	filter.EventMeta = map[string]string{
		"foo": "!bar",
	}
	filter.validate()
	args, query := filter.queryFields()
	assert.Len(t, args, 21)
	assert.Equal(t, "/", args[0])
	assert.Equal(t, "/entry", args[1])
	assert.Equal(t, "/exit", args[2])
	assert.Equal(t, "en", args[3])
	assert.Equal(t, "jp", args[4])
	assert.Equal(t, "Tokyo", args[5])
	assert.Equal(t, "ref", args[6])
	assert.Equal(t, "refname", args[7])
	assert.Equal(t, OSWindows, args[8])
	assert.Equal(t, "10", args[9])
	assert.Equal(t, BrowserEdge, args[10])
	assert.Equal(t, "89", args[11])
	assert.Equal(t, "XXL", args[12])
	assert.Equal(t, uint16(1920), args[13])
	assert.Equal(t, uint16(1080), args[14])
	assert.Equal(t, "source", args[15])
	assert.Equal(t, "medium", args[16])
	assert.Equal(t, "campaign", args[17])
	assert.Equal(t, "content", args[18])
	assert.Equal(t, "term", args[19])
	assert.Equal(t, "bar", args[20])
	assert.Equal(t, "path != ? AND entry_path != ? AND exit_path != ? AND language != ? AND country_code != ? AND city != ? AND referrer != ? AND referrer_name != ? AND os != ? AND os_version != ? AND browser != ? AND browser_version != ? AND screen_class != ? AND screen_width != ? AND screen_height != ? AND utm_source != ? AND utm_medium != ? AND utm_campaign != ? AND utm_content != ? AND utm_term != ? AND (desktop = 1 OR mobile = 1) AND event_meta_values[indexOf(event_meta_keys, 'foo')] != ? ", query)
	filter.EventName = "!event"
	args, query = filter.queryFields()
	assert.Len(t, args, 22)
	assert.Equal(t, "event", args[20])
	assert.Equal(t, "path != ? AND entry_path != ? AND exit_path != ? AND language != ? AND country_code != ? AND city != ? AND referrer != ? AND referrer_name != ? AND os != ? AND os_version != ? AND browser != ? AND browser_version != ? AND screen_class != ? AND screen_width != ? AND screen_height != ? AND utm_source != ? AND utm_medium != ? AND utm_campaign != ? AND utm_content != ? AND utm_term != ? AND event_name != ? AND (desktop = 1 OR mobile = 1) AND event_meta_values[indexOf(event_meta_keys, 'foo')] != ? ", query)
}

func TestFilter_QueryFieldsNull(t *testing.T) {
	filter := NewFilter(NullClient)
	filter.Path = "null"
	filter.EntryPath = "null"
	filter.ExitPath = "null"
	// not for path pattern
	filter.Language = "Null"
	filter.Country = "NULL"
	filter.City = "null"
	filter.Referrer = "null"
	filter.ReferrerName = "null"
	filter.OS = "null"
	filter.OSVersion = "null"
	filter.Browser = "null"
	filter.BrowserVersion = "null"
	filter.Platform = "null"
	filter.ScreenClass = "null"
	filter.ScreenWidth = "null"
	filter.ScreenHeight = "Null"
	filter.UTMSource = "null"
	filter.UTMMedium = "null"
	filter.UTMCampaign = "null"
	filter.UTMContent = "null"
	filter.UTMTerm = "null"
	filter.EventMeta = map[string]string{
		"foo": "null",
	}
	filter.validate()
	args, query := filter.queryFields()
	assert.Len(t, args, 21)

	for i := 0; i < len(args); i++ {
		assert.Empty(t, args[i])
	}

	assert.Equal(t, "path = ? AND entry_path = ? AND exit_path = ? AND language = ? AND country_code = ? AND city = ? AND referrer = ? AND referrer_name = ? AND os = ? AND os_version = ? AND browser = ? AND browser_version = ? AND screen_class = ? AND screen_width = ? AND screen_height = ? AND utm_source = ? AND utm_medium = ? AND utm_campaign = ? AND utm_content = ? AND utm_term = ? AND desktop = 0 AND mobile = 0 AND event_meta_values[indexOf(event_meta_keys, 'foo')] = ? ", query)
	filter.EventName = "null"
	filter.validate()
	args, query = filter.queryFields()
	assert.Len(t, args, 22)

	for i := 0; i < len(args); i++ {
		assert.Empty(t, args[i])
	}

	assert.Equal(t, "path = ? AND entry_path = ? AND exit_path = ? AND language = ? AND country_code = ? AND city = ? AND referrer = ? AND referrer_name = ? AND os = ? AND os_version = ? AND browser = ? AND browser_version = ? AND screen_class = ? AND screen_width = ? AND screen_height = ? AND utm_source = ? AND utm_medium = ? AND utm_campaign = ? AND utm_content = ? AND utm_term = ? AND event_name = ? AND desktop = 0 AND mobile = 0 AND event_meta_values[indexOf(event_meta_keys, 'foo')] = ? ", query)
}

func TestFilter_QueryFieldsPlatform(t *testing.T) {
	filter := NewFilter(NullClient)
	filter.Platform = PlatformDesktop
	args, query := filter.queryFields()
	assert.Len(t, args, 0)
	assert.Equal(t, "desktop = 1 ", query)
	filter = NewFilter(NullClient)
	filter.Platform = PlatformMobile
	args, query = filter.queryFields()
	assert.Len(t, args, 0)
	assert.Equal(t, "mobile = 1 ", query)
	filter = NewFilter(NullClient)
	filter.Platform = PlatformUnknown
	args, query = filter.queryFields()
	assert.Len(t, args, 0)
	assert.Equal(t, "desktop = 0 AND mobile = 0 ", query)
	_, query = filter.query(false)
	assert.Contains(t, query, "desktop = 0 AND mobile = 0")
}

func TestFilter_QueryFieldsPlatformInvert(t *testing.T) {
	filter := NewFilter(NullClient)
	filter.Platform = "!" + PlatformDesktop
	args, query := filter.queryFields()
	assert.Len(t, args, 0)
	assert.Equal(t, "desktop != 1 ", query)
	filter = NewFilter(NullClient)
	filter.Platform = "!" + PlatformMobile
	args, query = filter.queryFields()
	assert.Len(t, args, 0)
	assert.Equal(t, "mobile != 1 ", query)
	filter = NewFilter(NullClient)
	filter.Platform = "!" + PlatformUnknown
	args, query = filter.queryFields()
	assert.Len(t, args, 0)
	assert.Equal(t, "(desktop = 1 OR mobile = 1) ", query)
	_, query = filter.query(false)
	assert.Contains(t, query, "(desktop = 1 OR mobile = 1)")
}

func TestFilter_QueryIsBot(t *testing.T) {
	filter := NewFilter(NullClient)
	filter.Path = "/path"
	filter.minIsBot = 5
	args, query := filter.query(true)
	assert.Len(t, args, 3)
	assert.Equal(t, uint8(5), args[2])
	assert.Equal(t, "client_id = ? AND path = ?  AND is_bot < ? ", query)
}

func TestFilter_QueryFieldsPathPattern(t *testing.T) {
	filter := NewFilter(NullClient)
	filter.PathPattern = "/some/pattern"
	args, query := filter.queryFields()
	assert.Len(t, args, 1)
	assert.Equal(t, "/some/pattern", args[0])
	assert.Equal(t, `match("path", ?) = 1`, query)
}

func TestFilter_QueryFieldsPathPatternInvert(t *testing.T) {
	filter := NewFilter(NullClient)
	filter.PathPattern = "!/some/pattern"
	args, query := filter.queryFields()
	assert.Len(t, args, 1)
	assert.Equal(t, "/some/pattern", args[0])
	assert.Equal(t, `match("path", ?) = 0`, query)
}

func TestFilter_WithFill(t *testing.T) {
	filter := NewFilter(NullClient)
	args, query := filter.withFill()
	assert.Len(t, args, 0)
	assert.Empty(t, query)
	filter.From = pastDay(10)
	filter.To = pastDay(5)
	args, query = filter.withFill()
	assert.Len(t, args, 2)
	assert.Equal(t, filter.From, args[0])
	assert.Equal(t, filter.To, args[1])
	assert.Equal(t, "WITH FILL FROM toDate(?, 'UTC') TO toDate(?, 'UTC')+1 ", query)
}

func TestFilter_WithLimit(t *testing.T) {
	filter := NewFilter(NullClient)
	assert.Empty(t, filter.withLimit())
	filter.Limit = 42
	assert.Equal(t, "LIMIT 42 ", filter.withLimit())
	filter.Offset = 9
	assert.Equal(t, "LIMIT 42 OFFSET 9 ", filter.withLimit())
}

func TestFilter_Fields(t *testing.T) {
	filter := NewFilter(NullClient)
	filter.Path = "/"
	filter.EntryPath = "/entry"
	filter.ExitPath = "/exit"
	filter.PathPattern = "pattern"
	filter.Language = "en"
	filter.Country = "jp"
	filter.City = "Tokyo"
	filter.Referrer = "ref"
	filter.ReferrerName = "refname"
	filter.OS = OSWindows
	filter.OSVersion = "10"
	filter.Browser = BrowserEdge
	filter.BrowserVersion = "89"
	filter.Platform = PlatformUnknown
	filter.ScreenClass = "XXL"
	filter.UTMSource = "source"
	filter.UTMMedium = "medium"
	filter.UTMCampaign = "campaign"
	filter.UTMContent = "content"
	filter.UTMTerm = "term"
	filter.validate()

	// exit_path not included
	assert.Equal(t, "path,entry_path,exit_path,language,country_code,city,referrer,referrer_name,os,os_version,browser,browser_version,screen_class,utm_source,utm_medium,utm_campaign,utm_content,utm_term,desktop,mobile", filter.fields())

	filter.validate()
	filter.EventName = "event"
	filter.EventMeta = map[string]string{
		"foo": "bar",
	}
	assert.Equal(t, "path,entry_path,exit_path,language,country_code,city,referrer,referrer_name,os,os_version,browser,browser_version,screen_class,utm_source,utm_medium,utm_campaign,utm_content,utm_term,event_name,desktop,mobile,event_meta_keys,event_meta_values", filter.fields())
}

func pastDay(n int) time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()-n, 0, 0, 0, 0, time.UTC)
}

func pastWeek(n int) int {
	date := pastDay(n * 7)
	_, week := date.ISOWeek()
	return week
}

func pastMonth(n int) time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month()-time.Month(n), 1, 0, 0, 0, 0, time.UTC)
}
