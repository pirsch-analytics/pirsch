package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v4"
	"github.com/pirsch-analytics/pirsch/v4/db"
	"github.com/pirsch-analytics/pirsch/v4/model"
	"github.com/pirsch-analytics/pirsch/v4/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFilter_Validate(t *testing.T) {
	filter := NewFilter(pirsch.NullClient)
	filter.validate()
	assert.NotNil(t, filter)
	assert.NotNil(t, filter.Timezone)
	assert.Equal(t, time.UTC, filter.Timezone)
	assert.Zero(t, filter.From)
	assert.Zero(t, filter.To)
	filter = &Filter{From: util.PastDay(2), To: util.PastDay(5), Limit: 42}
	filter.validate()
	assert.Equal(t, util.PastDay(5), filter.From)
	assert.Equal(t, util.PastDay(2), filter.To)
	assert.Equal(t, 42, filter.Limit)
	filter = &Filter{From: util.PastDay(2), To: util.Today().Add(time.Hour * 24 * 5)}
	filter.validate()
	assert.Equal(t, util.PastDay(2), filter.From)
	assert.Equal(t, util.Today().Add(time.Hour*24), filter.To)
	filter = &Filter{Limit: -42, Path: []string{"/path"}, PathPattern: []string{"pattern"}}
	filter.validate()
	assert.Zero(t, filter.Limit)
	assert.Len(t, filter.Path, 1)
	assert.Equal(t, "/path", filter.Path[0])
	assert.Empty(t, filter.PathPattern)
	filter = &Filter{Limit: -42, PathPattern: []string{"pattern", "pattern"}}
	filter.validate()
	assert.Empty(t, filter.Path)
	assert.Len(t, filter.PathPattern, 1)
	assert.Equal(t, "pattern", filter.PathPattern[0])
}

func TestFilter_RemoveDuplicates(t *testing.T) {
	filter := NewFilter(pirsch.NullClient)
	filter.Path = []string{
		"/",
		"/",
		"/foo",
		"/Foo",
		"/bar",
		"/foo",
	}
	filter.validate()
	assert.Len(t, filter.Path, 4)
	assert.Equal(t, "/", filter.Path[0])
	assert.Equal(t, "/foo", filter.Path[1])
	assert.Equal(t, "/Foo", filter.Path[2])
	assert.Equal(t, "/bar", filter.Path[3])
}

func TestFilter_BuildQuery(t *testing.T) {
	db.CleanupDB(t, dbClient)
	assert.NoError(t, dbClient.SavePageViews([]model.PageView{
		{VisitorID: 1, Time: util.Today(), Path: "/"},
		{VisitorID: 1, Time: util.Today().Add(time.Minute * 2), Path: "/foo"},
		{VisitorID: 1, Time: util.Today().Add(time.Minute*2 + time.Second*2), Path: "/foo"},
		{VisitorID: 1, Time: util.Today().Add(time.Minute*2 + time.Second*23), Path: "/bar"},

		{VisitorID: 2, Time: util.Today(), Path: "/bar"},
		{VisitorID: 2, Time: util.Today().Add(time.Second * 16), Path: "/foo"},
		{VisitorID: 2, Time: util.Today().Add(time.Second*16 + time.Second*8), Path: "/"},
	}))
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: util.Today(), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 2, Time: util.Today(), Start: time.Now(), EntryPath: "/bar", ExitPath: "/bar", PageViews: 1},
		},
		{
			{Sign: -1, VisitorID: 1, Time: util.Today(), Start: time.Now(), EntryPath: "/", ExitPath: "/", PageViews: 1},
			{Sign: 1, VisitorID: 1, Time: util.Today().Add(time.Minute * 2), Start: time.Now(), EntryPath: "/", ExitPath: "/foo", PageViews: 2},
			{Sign: -1, VisitorID: 2, Time: util.Today(), Start: time.Now(), EntryPath: "/bar", ExitPath: "/bar", PageViews: 1},
			{Sign: 1, VisitorID: 2, Time: util.Today().Add(time.Second * 16), Start: time.Now(), EntryPath: "/bar", ExitPath: "/foo", PageViews: 2},
		},
		{
			{Sign: -1, VisitorID: 1, Time: util.Today().Add(time.Minute * 2), Start: time.Now(), EntryPath: "/", ExitPath: "/foo", PageViews: 2},
			{Sign: 1, VisitorID: 1, Time: util.Today().Add(time.Minute*2 + time.Second*23), Start: time.Now(), EntryPath: "/", ExitPath: "/bar", PageViews: 3},
			{Sign: -1, VisitorID: 2, Time: util.Today().Add(time.Second * 16), Start: time.Now(), EntryPath: "/bar", ExitPath: "/foo", PageViews: 2},
			{Sign: 1, VisitorID: 2, Time: util.Today().Add(time.Second*16 + time.Second*8), Start: time.Now(), EntryPath: "/bar", ExitPath: "/", PageViews: 3},
		},
	})

	// no filter (from page views)
	analyzer := NewAnalyzer(dbClient, nil)
	args, query := analyzer.getFilter(nil).buildQuery([]Field{FieldPath, FieldVisitors}, []Field{FieldPath}, []Field{FieldVisitors, FieldPath})
	var stats []model.PageStats
	rows, err := dbClient.Query(query, args...)
	assert.NoError(t, err)

	for rows.Next() {
		var stat model.PageStats
		assert.NoError(t, rows.Scan(&stat.Path, &stat.Visitors))
		stats = append(stats, stat)
	}

	assert.Len(t, stats, 3)
	assert.Equal(t, 2, stats[0].Visitors)
	assert.Equal(t, 2, stats[1].Visitors)
	assert.Equal(t, 2, stats[2].Visitors)
	assert.Equal(t, "/", stats[0].Path)
	assert.Equal(t, "/bar", stats[1].Path)
	assert.Equal(t, "/foo", stats[2].Path)

	// join (from page views)
	args, query = analyzer.getFilter(&Filter{EntryPath: []string{"/"}}).buildQuery([]Field{FieldPath, FieldVisitors}, []Field{FieldPath}, []Field{FieldPath})
	stats = stats[:0]
	rows, err = dbClient.Query(query, args...)
	assert.NoError(t, err)

	for rows.Next() {
		var stat model.PageStats
		assert.NoError(t, rows.Scan(&stat.Path, &stat.Visitors))
		stats = append(stats, stat)
	}

	assert.Len(t, stats, 3)
	assert.Equal(t, 1, stats[0].Visitors)
	assert.Equal(t, 1, stats[1].Visitors)
	assert.Equal(t, 1, stats[2].Visitors)
	assert.Equal(t, "/", stats[0].Path)
	assert.Equal(t, "/bar", stats[1].Path)
	assert.Equal(t, "/foo", stats[2].Path)

	// join and filter (from page views)
	args, query = analyzer.getFilter(&Filter{EntryPath: []string{"/"}, Path: []string{"/foo"}}).buildQuery([]Field{FieldPath, FieldVisitors}, []Field{FieldPath}, []Field{FieldPath})
	stats = stats[:0]
	rows, err = dbClient.Query(query, args...)
	assert.NoError(t, err)

	for rows.Next() {
		var stat model.PageStats
		assert.NoError(t, rows.Scan(&stat.Path, &stat.Visitors))
		stats = append(stats, stat)
	}

	assert.Len(t, stats, 1)
	assert.Equal(t, "/foo", stats[0].Path)
	assert.Equal(t, 1, stats[0].Visitors)

	// filter (from page views)
	args, query = analyzer.getFilter(&Filter{Path: []string{"/foo"}}).buildQuery([]Field{FieldPath, FieldVisitors}, []Field{FieldPath}, []Field{FieldPath})
	stats = stats[:0]
	rows, err = dbClient.Query(query, args...)
	assert.NoError(t, err)

	for rows.Next() {
		var stat model.PageStats
		assert.NoError(t, rows.Scan(&stat.Path, &stat.Visitors))
		stats = append(stats, stat)
	}

	assert.Len(t, stats, 1)
	assert.Equal(t, "/foo", stats[0].Path)
	assert.Equal(t, 2, stats[0].Visitors)

	// no filter (from sessions)
	args, query = analyzer.getFilter(nil).buildQuery([]Field{FieldVisitors, FieldSessions, FieldViews, FieldBounces, FieldBounceRate}, nil, nil)
	var vstats model.PageStats
	assert.NoError(t, dbClient.QueryRow(query, args...).Scan(&vstats.Visitors, &vstats.Sessions, &vstats.Views, &vstats.Bounces, &vstats.BounceRate))
	assert.Equal(t, 2, vstats.Visitors)
	assert.Equal(t, 2, vstats.Sessions)
	assert.Equal(t, 6, vstats.Views)
	assert.Equal(t, 0, vstats.Bounces)
	assert.InDelta(t, 0, vstats.BounceRate, 0.01)

	// filter (from page views)
	args, query = analyzer.getFilter(&Filter{Path: []string{"/foo"}, EntryPath: []string{"/"}}).buildQuery([]Field{FieldVisitors, FieldRelativeVisitors, FieldSessions, FieldViews, FieldRelativeViews, FieldBounces, FieldBounceRate}, nil, nil)
	assert.NoError(t, dbClient.QueryRow(query, args...).Scan(&vstats.Visitors, &vstats.RelativeVisitors, &vstats.Sessions, &vstats.Views, &vstats.RelativeViews, &vstats.Bounces, &vstats.BounceRate))
	assert.Equal(t, 1, vstats.Visitors)
	assert.Equal(t, 1, vstats.Sessions)
	assert.Equal(t, 2, vstats.Views)
	assert.Equal(t, 0, vstats.Bounces)
	assert.InDelta(t, 0, vstats.BounceRate, 0.01)
	assert.InDelta(t, 0.5, vstats.RelativeVisitors, 0.01)
	assert.InDelta(t, 0.3333, vstats.RelativeViews, 0.01)

	// filter period
	args, query = analyzer.getFilter(&Filter{Period: pirsch.PeriodWeek}).buildQuery([]Field{FieldDay, FieldVisitors}, []Field{FieldDay}, []Field{FieldDay})
	var visitors []model.VisitorStats
	rows, err = dbClient.Query(query, args...)
	assert.NoError(t, err)

	for rows.Next() {
		var stat model.VisitorStats
		assert.NoError(t, rows.Scan(&stat.Day, &stat.Visitors))
		visitors = append(visitors, stat)
	}

	assert.Len(t, visitors, 1)
}

func TestFilter_Table(t *testing.T) {
	filter := NewFilter(pirsch.NullClient)
	assert.Equal(t, "session", filter.table())
	filter.eventFilter = true
	assert.Equal(t, "event", filter.table())
	filter.eventFilter = false
	filter.EventName = []string{"event"}
	assert.Equal(t, "event", filter.table())
}

func TestFilter_QueryTime(t *testing.T) {
	filter := NewFilter(pirsch.NullClient)
	filter.From = util.PastDay(5)
	filter.To = util.PastDay(2)
	args, query := filter.queryTime(false)
	assert.Len(t, args, 3)
	assert.Equal(t, pirsch.NullClient, args[0])
	assert.Equal(t, filter.From.Format(dateFormat), args[1])
	assert.Equal(t, filter.To.Format(dateFormat), args[2])
	assert.Equal(t, "client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) ", query)
	filter.From = util.PastDay(2)
	filter.validate()
	args, query = filter.queryTime(false)
	assert.Len(t, args, 2)
	assert.Equal(t, pirsch.NullClient, args[0])
	assert.Equal(t, filter.From.Format(dateFormat), args[1])
	assert.Equal(t, "client_id = ? AND toDate(time, 'UTC') = toDate(?) ", query)
	filter.IncludeTime = true
	filter.From = filter.From.Add(time.Hour * 5)
	filter.To = filter.To.Add(time.Hour * 19)
	filter.validate()
	args, query = filter.queryTime(false)
	assert.Len(t, args, 3)
	assert.Equal(t, pirsch.NullClient, args[0])
	assert.Equal(t, filter.From, args[1])
	assert.Equal(t, filter.To, args[2])
	assert.Equal(t, "client_id = ? AND toDateTime(time, 'UTC') >= toDateTime(?, 'UTC') AND toDateTime(time, 'UTC') <= toDateTime(?, 'UTC') ", query)
}

func TestFilter_QueryFields(t *testing.T) {
	filter := NewFilter(pirsch.NullClient)
	filter.Path = []string{"/", "/foo"}
	filter.EntryPath = []string{"/entry"}
	filter.ExitPath = []string{"/exit"}
	filter.PathPattern = []string{"pattern"}
	filter.Language = []string{"en"}
	filter.Country = []string{"jp"}
	filter.City = []string{"Tokyo"}
	filter.Referrer = []string{"ref"}
	filter.ReferrerName = []string{"refname"}
	filter.OS = []string{pirsch.OSWindows}
	filter.OSVersion = []string{"10"}
	filter.Browser = []string{pirsch.BrowserEdge}
	filter.BrowserVersion = []string{"89"}
	filter.Platform = pirsch.PlatformUnknown
	filter.ScreenClass = []string{"XXL"}
	filter.ScreenWidth = []string{"1920"}
	filter.ScreenHeight = []string{"1080"}
	filter.UTMSource = []string{"source"}
	filter.UTMMedium = []string{"medium"}
	filter.UTMCampaign = []string{"campaign"}
	filter.UTMContent = []string{"content"}
	filter.UTMTerm = []string{"term"}
	filter.EventMeta = map[string]string{
		"foo": "bar",
	}
	filter.validate()
	args, query := filter.queryFields()
	assert.Len(t, args, 21)
	assert.Equal(t, "/", args[0])
	assert.Equal(t, "/foo", args[1])
	assert.Equal(t, "/entry", args[2])
	assert.Equal(t, "/exit", args[3])
	assert.Equal(t, "en", args[4])
	assert.Equal(t, "jp", args[5])
	assert.Equal(t, "Tokyo", args[6])
	assert.Equal(t, "ref", args[7])
	assert.Equal(t, "refname", args[8])
	assert.Equal(t, pirsch.OSWindows, args[9])
	assert.Equal(t, "10", args[10])
	assert.Equal(t, pirsch.BrowserEdge, args[11])
	assert.Equal(t, "89", args[12])
	assert.Equal(t, "XXL", args[13])
	assert.Equal(t, uint16(1920), args[14])
	assert.Equal(t, uint16(1080), args[15])
	assert.Equal(t, "source", args[16])
	assert.Equal(t, "medium", args[17])
	assert.Equal(t, "campaign", args[18])
	assert.Equal(t, "content", args[19])
	assert.Equal(t, "term", args[20])
	assert.Equal(t, "(path = ? OR path = ? ) AND entry_path = ? AND exit_path = ? AND language = ? AND country_code = ? AND city = ? AND referrer = ? AND referrer_name = ? AND os = ? AND os_version = ? AND browser = ? AND browser_version = ? AND screen_class = ? AND screen_width = ? AND screen_height = ? AND utm_source = ? AND utm_medium = ? AND utm_campaign = ? AND utm_content = ? AND utm_term = ? AND desktop = 0 AND mobile = 0 ", query)
	filter.EventName = []string{"event"}
	filter.EventMeta = map[string]string{
		"foo": "bar",
	}
	args, query = filter.queryFields()
	assert.Len(t, args, 23)
	assert.Equal(t, "event", args[21])
	assert.Equal(t, "(path = ? OR path = ? ) AND entry_path = ? AND exit_path = ? AND language = ? AND country_code = ? AND city = ? AND referrer = ? AND referrer_name = ? AND os = ? AND os_version = ? AND browser = ? AND browser_version = ? AND screen_class = ? AND screen_width = ? AND screen_height = ? AND utm_source = ? AND utm_medium = ? AND utm_campaign = ? AND utm_content = ? AND utm_term = ? AND event_name = ? AND desktop = 0 AND mobile = 0 AND event_meta_values[indexOf(event_meta_keys, 'foo')] = ? ", query)
}

func TestFilter_QueryFieldsInvert(t *testing.T) {
	filter := NewFilter(pirsch.NullClient)
	filter.Path = []string{"!/"}
	filter.EntryPath = []string{"!/entry"}
	filter.ExitPath = []string{"!/exit"}
	filter.PathPattern = []string{"!pattern"}
	filter.Language = []string{"!en"}
	filter.Country = []string{"!jp"}
	filter.City = []string{"!Tokyo"}
	filter.Referrer = []string{"!ref"}
	filter.ReferrerName = []string{"!refname"}
	filter.OS = []string{"!" + pirsch.OSWindows}
	filter.OSVersion = []string{"!10"}
	filter.Browser = []string{"!" + pirsch.BrowserEdge}
	filter.BrowserVersion = []string{"!89"}
	filter.Platform = "!" + pirsch.PlatformUnknown
	filter.ScreenClass = []string{"!XXL"}
	filter.ScreenWidth = []string{"!1920"}
	filter.ScreenHeight = []string{"!1080"}
	filter.UTMSource = []string{"!source"}
	filter.UTMMedium = []string{"!medium"}
	filter.UTMCampaign = []string{"!campaign"}
	filter.UTMContent = []string{"!content"}
	filter.UTMTerm = []string{"!term"}
	filter.EventMeta = map[string]string{
		"foo": "!bar",
	}
	filter.validate()
	args, query := filter.queryFields()
	assert.Len(t, args, 20)
	assert.Equal(t, "/", args[0])
	assert.Equal(t, "/entry", args[1])
	assert.Equal(t, "/exit", args[2])
	assert.Equal(t, "en", args[3])
	assert.Equal(t, "jp", args[4])
	assert.Equal(t, "Tokyo", args[5])
	assert.Equal(t, "ref", args[6])
	assert.Equal(t, "refname", args[7])
	assert.Equal(t, pirsch.OSWindows, args[8])
	assert.Equal(t, "10", args[9])
	assert.Equal(t, pirsch.BrowserEdge, args[10])
	assert.Equal(t, "89", args[11])
	assert.Equal(t, "XXL", args[12])
	assert.Equal(t, uint16(1920), args[13])
	assert.Equal(t, uint16(1080), args[14])
	assert.Equal(t, "source", args[15])
	assert.Equal(t, "medium", args[16])
	assert.Equal(t, "campaign", args[17])
	assert.Equal(t, "content", args[18])
	assert.Equal(t, "term", args[19])
	assert.Equal(t, "path != ? AND entry_path != ? AND exit_path != ? AND language != ? AND country_code != ? AND city != ? AND referrer != ? AND referrer_name != ? AND os != ? AND os_version != ? AND browser != ? AND browser_version != ? AND screen_class != ? AND screen_width != ? AND screen_height != ? AND utm_source != ? AND utm_medium != ? AND utm_campaign != ? AND utm_content != ? AND utm_term != ? AND (desktop = 1 OR mobile = 1) ", query)
	filter.EventName = []string{"!event"}
	filter.EventMeta = map[string]string{
		"foo": "!bar",
	}
	args, query = filter.queryFields()
	assert.Len(t, args, 22)
	assert.Equal(t, "event", args[20])
	assert.Equal(t, "path != ? AND entry_path != ? AND exit_path != ? AND language != ? AND country_code != ? AND city != ? AND referrer != ? AND referrer_name != ? AND os != ? AND os_version != ? AND browser != ? AND browser_version != ? AND screen_class != ? AND screen_width != ? AND screen_height != ? AND utm_source != ? AND utm_medium != ? AND utm_campaign != ? AND utm_content != ? AND utm_term != ? AND event_name != ? AND (desktop = 1 OR mobile = 1) AND event_meta_values[indexOf(event_meta_keys, 'foo')] != ? ", query)
}

func TestFilter_QueryFieldsContains(t *testing.T) {
	filter := NewFilter(pirsch.NullClient)
	filter.Path = []string{"~/"}
	filter.EntryPath = []string{"~/entry"}
	filter.ExitPath = []string{"~/exit"}
	filter.PathPattern = []string{"~pattern"}
	filter.Language = []string{"~en"}
	filter.Country = []string{"~jp"}
	filter.City = []string{"~Tokyo"}
	filter.Referrer = []string{"~ref"}
	filter.ReferrerName = []string{"~refname"}
	filter.OS = []string{"~" + pirsch.OSWindows}
	filter.OSVersion = []string{"~10"}
	filter.Browser = []string{"~" + pirsch.BrowserEdge}
	filter.BrowserVersion = []string{"~89"}
	filter.Platform = "~" + pirsch.PlatformUnknown
	filter.ScreenClass = []string{"~XXL"}
	filter.ScreenWidth = []string{"1920"}
	filter.ScreenHeight = []string{"1080"}
	filter.UTMSource = []string{"~source"}
	filter.UTMMedium = []string{"~medium"}
	filter.UTMCampaign = []string{"~campaign"}
	filter.UTMContent = []string{"~content"}
	filter.UTMTerm = []string{"~term"}
	filter.EventMeta = map[string]string{
		"foo": "~bar",
	}
	filter.validate()
	args, query := filter.queryFields()
	assert.Len(t, args, 20)
	assert.Equal(t, "%/%", args[0])
	assert.Equal(t, "%/entry%", args[1])
	assert.Equal(t, "%/exit%", args[2])
	assert.Equal(t, "en", args[3])
	assert.Equal(t, "jp", args[4])
	assert.Equal(t, "%Tokyo%", args[5])
	assert.Equal(t, "%ref%", args[6])
	assert.Equal(t, "%refname%", args[7])
	assert.Equal(t, "%"+pirsch.OSWindows+"%", args[8])
	assert.Equal(t, "%10%", args[9])
	assert.Equal(t, "%"+pirsch.BrowserEdge+"%", args[10])
	assert.Equal(t, "%89%", args[11])
	assert.Equal(t, "%XXL%", args[12])
	assert.Equal(t, uint16(1920), args[13])
	assert.Equal(t, uint16(1080), args[14])
	assert.Equal(t, "%source%", args[15])
	assert.Equal(t, "%medium%", args[16])
	assert.Equal(t, "%campaign%", args[17])
	assert.Equal(t, "%content%", args[18])
	assert.Equal(t, "%term%", args[19])
	assert.Equal(t, "ilike(path, ?) = 1 AND ilike(entry_path, ?) = 1 AND ilike(exit_path, ?) = 1 AND has(splitByChar(',', ?), language) = 1 AND has(splitByChar(',', ?), country_code) = 1 AND ilike(city, ?) = 1 AND ilike(referrer, ?) = 1 AND ilike(referrer_name, ?) = 1 AND ilike(os, ?) = 1 AND ilike(os_version, ?) = 1 AND ilike(browser, ?) = 1 AND ilike(browser_version, ?) = 1 AND ilike(screen_class, ?) = 1 AND screen_width = ? AND screen_height = ? AND ilike(utm_source, ?) = 1 AND ilike(utm_medium, ?) = 1 AND ilike(utm_campaign, ?) = 1 AND ilike(utm_content, ?) = 1 AND ilike(utm_term, ?) = 1 AND desktop = 0 AND mobile = 0 ", query)
	filter.EventName = []string{"~event"}
	filter.EventMeta = map[string]string{
		"foo": "~bar",
	}
	args, query = filter.queryFields()
	assert.Len(t, args, 22)
	assert.Equal(t, "%event%", args[20])
	assert.Equal(t, "ilike(path, ?) = 1 AND ilike(entry_path, ?) = 1 AND ilike(exit_path, ?) = 1 AND has(splitByChar(',', ?), language) = 1 AND has(splitByChar(',', ?), country_code) = 1 AND ilike(city, ?) = 1 AND ilike(referrer, ?) = 1 AND ilike(referrer_name, ?) = 1 AND ilike(os, ?) = 1 AND ilike(os_version, ?) = 1 AND ilike(browser, ?) = 1 AND ilike(browser_version, ?) = 1 AND ilike(screen_class, ?) = 1 AND screen_width = ? AND screen_height = ? AND ilike(utm_source, ?) = 1 AND ilike(utm_medium, ?) = 1 AND ilike(utm_campaign, ?) = 1 AND ilike(utm_content, ?) = 1 AND ilike(utm_term, ?) = 1 AND ilike(event_name, ?) = 1 AND desktop = 0 AND mobile = 0 AND ilike(event_meta_values[indexOf(event_meta_keys, 'foo')], ?) = 1 ", query)
}

func TestFilter_QueryFieldsNull(t *testing.T) {
	filter := NewFilter(pirsch.NullClient)
	filter.Path = []string{"null"}
	filter.EntryPath = []string{"null"}
	filter.ExitPath = []string{"null"}
	// not for path pattern
	filter.Language = []string{"Null"}
	filter.Country = []string{"NULL"}
	filter.City = []string{"null"}
	filter.Referrer = []string{"null"}
	filter.ReferrerName = []string{"null"}
	filter.OS = []string{"null"}
	filter.OSVersion = []string{"null"}
	filter.Browser = []string{"null"}
	filter.BrowserVersion = []string{"null"}
	filter.Platform = "null"
	filter.ScreenClass = []string{"null"}
	filter.ScreenWidth = []string{"null"}
	filter.ScreenHeight = []string{"Null"}
	filter.UTMSource = []string{"null"}
	filter.UTMMedium = []string{"null"}
	filter.UTMCampaign = []string{"null"}
	filter.UTMContent = []string{"null"}
	filter.UTMTerm = []string{"null"}
	filter.EventMeta = map[string]string{
		"foo": "null",
	}
	filter.validate()
	args, query := filter.queryFields()
	assert.Len(t, args, 20)

	for i := 0; i < len(args); i++ {
		assert.Empty(t, args[i])
	}

	assert.Equal(t, "path = ? AND entry_path = ? AND exit_path = ? AND language = ? AND country_code = ? AND city = ? AND referrer = ? AND referrer_name = ? AND os = ? AND os_version = ? AND browser = ? AND browser_version = ? AND screen_class = ? AND screen_width = ? AND screen_height = ? AND utm_source = ? AND utm_medium = ? AND utm_campaign = ? AND utm_content = ? AND utm_term = ? AND desktop = 0 AND mobile = 0 ", query)
	filter.EventName = []string{"null"}
	filter.EventMeta = map[string]string{
		"foo": "null",
	}
	filter.validate()
	args, query = filter.queryFields()
	assert.Len(t, args, 22)

	for i := 0; i < len(args); i++ {
		assert.Empty(t, args[i])
	}

	assert.Equal(t, "path = ? AND entry_path = ? AND exit_path = ? AND language = ? AND country_code = ? AND city = ? AND referrer = ? AND referrer_name = ? AND os = ? AND os_version = ? AND browser = ? AND browser_version = ? AND screen_class = ? AND screen_width = ? AND screen_height = ? AND utm_source = ? AND utm_medium = ? AND utm_campaign = ? AND utm_content = ? AND utm_term = ? AND event_name = ? AND desktop = 0 AND mobile = 0 AND event_meta_values[indexOf(event_meta_keys, 'foo')] = ? ", query)
}

func TestFilter_QueryFieldsLogicalOperators(t *testing.T) {
	filter := NewFilter(pirsch.NullClient)
	filter.Path = []string{"/foo", "/bar", "!/exclude/path"}
	filter.Country = []string{"ja", "!de", "!gb"}
	filter.Language = []string{"~ja"}
	filter.UTMSource = []string{"!Twitter"}
	filter.validate()
	args, query := filter.queryFields()
	assert.Len(t, args, 8)
	assert.Equal(t, "(path = ? OR path = ? ) AND path != ? AND has(splitByChar(',', ?), language) = 1 AND country_code = ? AND country_code != ?  AND country_code != ? AND utm_source != ? ", query)
}

func TestFilter_QueryFieldsPlatform(t *testing.T) {
	filter := NewFilter(pirsch.NullClient)
	filter.Platform = pirsch.PlatformDesktop
	args, query := filter.queryFields()
	assert.Len(t, args, 0)
	assert.Equal(t, "desktop = 1 ", query)
	filter = NewFilter(pirsch.NullClient)
	filter.Platform = pirsch.PlatformMobile
	args, query = filter.queryFields()
	assert.Len(t, args, 0)
	assert.Equal(t, "mobile = 1 ", query)
	filter = NewFilter(pirsch.NullClient)
	filter.Platform = pirsch.PlatformUnknown
	args, query = filter.queryFields()
	assert.Len(t, args, 0)
	assert.Equal(t, "desktop = 0 AND mobile = 0 ", query)
	_, query = filter.query(false)
	assert.Contains(t, query, "desktop = 0 AND mobile = 0")
}

func TestFilter_QueryFieldsPlatformInvert(t *testing.T) {
	filter := NewFilter(pirsch.NullClient)
	filter.Platform = "!" + pirsch.PlatformDesktop
	args, query := filter.queryFields()
	assert.Len(t, args, 0)
	assert.Equal(t, "desktop != 1 ", query)
	filter = NewFilter(pirsch.NullClient)
	filter.Platform = "!" + pirsch.PlatformMobile
	args, query = filter.queryFields()
	assert.Len(t, args, 0)
	assert.Equal(t, "mobile != 1 ", query)
	filter = NewFilter(pirsch.NullClient)
	filter.Platform = "!" + pirsch.PlatformUnknown
	args, query = filter.queryFields()
	assert.Len(t, args, 0)
	assert.Equal(t, "(desktop = 1 OR mobile = 1) ", query)
	_, query = filter.query(false)
	assert.Contains(t, query, "(desktop = 1 OR mobile = 1)")
}

func TestFilter_QueryIsBot(t *testing.T) {
	filter := NewFilter(pirsch.NullClient)
	filter.Path = []string{"/path"}
	filter.minIsBot = 5
	args, query := filter.query(true)
	assert.Len(t, args, 3)
	assert.Equal(t, uint8(5), args[2])
	assert.Equal(t, "client_id = ? AND path = ?  AND is_bot < ? ", query)
}

func TestFilter_QueryFieldsPathPattern(t *testing.T) {
	filter := NewFilter(pirsch.NullClient)
	filter.PathPattern = []string{"/some/pattern"}
	args, query := filter.queryFields()
	assert.Len(t, args, 1)
	assert.Equal(t, "/some/pattern", args[0])
	assert.Equal(t, `match("path", ?) = 1 `, query)
}

func TestFilter_QueryFieldsPathPatternInvert(t *testing.T) {
	filter := NewFilter(pirsch.NullClient)
	filter.PathPattern = []string{"!/some/pattern"}
	args, query := filter.queryFields()
	assert.Len(t, args, 1)
	assert.Equal(t, "/some/pattern", args[0])
	assert.Equal(t, `match("path", ?) = 0 `, query)
}

func TestFilter_QueryFieldsSearch(t *testing.T) {
	filter := NewFilter(pirsch.NullClient)
	filter.Search = []Search{
		{
			Field: FieldPath,
			Input: "/blog",
		},
		{
			Field: FieldBrowser,
			Input: "!fox",
		},
		{
			Field: FieldOS,
			Input: " ",
		},
	}
	filter.validate()
	args, query := filter.queryFields()
	assert.Len(t, args, 2)
	assert.Equal(t, "%/blog%", args[0])
	assert.Equal(t, "%fox%", args[1])
	assert.Equal(t, "ilike(path, ?) = 1 AND ilike(browser, ?) = 0 ", query)
}

func TestFilter_WithFill(t *testing.T) {
	filter := NewFilter(pirsch.NullClient)
	args, query := filter.withFill()
	assert.Len(t, args, 0)
	assert.Empty(t, query)
	filter.From = util.PastDay(10)
	filter.To = util.PastDay(5)
	args, query = filter.withFill()
	assert.Len(t, args, 2)
	assert.Equal(t, filter.From.Format(dateFormat), args[0])
	assert.Equal(t, filter.To.Format(dateFormat), args[1])
	assert.Equal(t, "WITH FILL FROM toDate(?) TO toDate(?)+1 STEP INTERVAL 1 DAY ", query)
}

func TestFilter_WithLimit(t *testing.T) {
	filter := NewFilter(pirsch.NullClient)
	assert.Empty(t, filter.withLimit())
	filter.Limit = 42
	assert.Equal(t, "LIMIT 42 ", filter.withLimit())
	filter.Offset = 9
	assert.Equal(t, "LIMIT 42 OFFSET 9 ", filter.withLimit())
}

func TestFilter_Fields(t *testing.T) {
	filter := NewFilter(pirsch.NullClient)
	filter.Path = []string{"/"}
	filter.EntryPath = []string{"/entry"}
	filter.ExitPath = []string{"/exit"}
	filter.PathPattern = []string{"pattern"}
	filter.Language = []string{"en"}
	filter.Country = []string{"jp"}
	filter.City = []string{"Tokyo"}
	filter.Referrer = []string{"ref"}
	filter.ReferrerName = []string{"refname"}
	filter.OS = []string{pirsch.OSWindows}
	filter.OSVersion = []string{"10"}
	filter.Browser = []string{pirsch.BrowserEdge}
	filter.BrowserVersion = []string{"89"}
	filter.Platform = pirsch.PlatformUnknown
	filter.ScreenClass = []string{"XXL"}
	filter.UTMSource = []string{"source"}
	filter.UTMMedium = []string{"medium"}
	filter.UTMCampaign = []string{"campaign"}
	filter.UTMContent = []string{"content"}
	filter.UTMTerm = []string{"term"}
	filter.validate()

	// exit_path not included
	assert.Equal(t, "path,entry_path,exit_path,language,country_code,city,referrer,referrer_name,os,os_version,browser,browser_version,screen_class,utm_source,utm_medium,utm_campaign,utm_content,utm_term,desktop,mobile", filter.fields())

	filter.validate()
	filter.EventName = []string{"event"}
	filter.EventMeta = map[string]string{
		"foo": "bar",
	}
	assert.Equal(t, "path,entry_path,exit_path,language,country_code,city,referrer,referrer_name,os,os_version,browser,browser_version,screen_class,utm_source,utm_medium,utm_campaign,utm_content,utm_term,event_name,desktop,mobile,event_meta_keys,event_meta_values", filter.fields())
}

func TestFilter_JoinOrderBy(t *testing.T) {
	filter := NewFilter(pirsch.NullClient)
	filter.From = util.PastDay(1)
	filter.To = util.Today()
	args := make([]any, 0)
	query := filter.joinOrderBy(&args, []Field{
		FieldDay,
		FieldVisitors,
	})
	assert.Len(t, args, 2)
	assert.Equal(t, "day ASC WITH FILL FROM toDate(?) TO toDate(?)+1 STEP INTERVAL 1 DAY ,visitors DESC", query)
	args = make([]any, 0)
	filter.Sort = []Sort{
		{
			Field:     FieldPath,
			Direction: pirsch.DirectionDESC,
		},
		{
			Field:     FieldVisitors,
			Direction: pirsch.DirectionASC,
		},
	}
	query = filter.joinOrderBy(&args, []Field{
		FieldDay,
		FieldVisitors,
	})
	assert.Len(t, args, 0)
	assert.Equal(t, "path DESC,visitors ASC", query)
}
