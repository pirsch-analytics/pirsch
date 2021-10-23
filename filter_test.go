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
	filter = &Filter{Day: time.Now().UTC(), Limit: -42, Path: "/path", PathPattern: "pattern"}
	filter.validate()
	assert.Zero(t, filter.Day.Hour())
	assert.Zero(t, filter.Day.Minute())
	assert.Zero(t, filter.Day.Second())
	assert.Zero(t, filter.Day.Nanosecond())
	assert.Zero(t, filter.Limit)
	assert.Equal(t, "/path", filter.Path)
	assert.Empty(t, filter.PathPattern)
	filter = &Filter{Day: time.Now().UTC(), Limit: -42, PathPattern: "pattern"}
	filter.validate()
	assert.Empty(t, filter.Path)
	assert.Equal(t, "pattern", filter.PathPattern)
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
	filter.Day = pastDay(1)
	filter.Start = time.Now().UTC()
	args, query := filter.queryTime()
	assert.Len(t, args, 5)
	assert.Equal(t, NullClient, args[0])
	assert.Equal(t, filter.From, args[1])
	assert.Equal(t, filter.To, args[2])
	assert.Equal(t, filter.Day, args[3])
	assert.Equal(t, filter.Start, args[4])
	assert.Equal(t, "client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND toDate(time, 'UTC') = toDate(?) AND toDateTime(time, 'UTC') >= toDateTime(?, 'UTC') ", query)
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
	filter.UTMSource = "source"
	filter.UTMMedium = "medium"
	filter.UTMCampaign = "campaign"
	filter.UTMContent = "content"
	filter.UTMTerm = "term"
	filter.validate()
	args, query := filter.queryFields()
	assert.Len(t, args, 18)
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
	assert.Equal(t, "source", args[13])
	assert.Equal(t, "medium", args[14])
	assert.Equal(t, "campaign", args[15])
	assert.Equal(t, "content", args[16])
	assert.Equal(t, "term", args[17])
	assert.Equal(t, "path = ? AND entry_path = ? AND exit_path = ? AND language = ? AND country_code = ? AND city = ? AND referrer = ? AND referrer_name = ? AND os = ? AND os_version = ? AND browser = ? AND browser_version = ? AND screen_class = ? AND utm_source = ? AND utm_medium = ? AND utm_campaign = ? AND utm_content = ? AND utm_term = ? AND desktop = 0 AND mobile = 0 ", query)
	filter.EventName = "event"
	args, query = filter.queryFields()
	assert.Len(t, args, 17)
	assert.Equal(t, "event", args[16])
	assert.Equal(t, "path = ? AND language = ? AND country_code = ? AND city = ? AND referrer = ? AND referrer_name = ? AND os = ? AND os_version = ? AND browser = ? AND browser_version = ? AND screen_class = ? AND utm_source = ? AND utm_medium = ? AND utm_campaign = ? AND utm_content = ? AND utm_term = ? AND event_name = ? AND desktop = 0 AND mobile = 0 ", query)
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
	filter.UTMSource = "!source"
	filter.UTMMedium = "!medium"
	filter.UTMCampaign = "!campaign"
	filter.UTMContent = "!content"
	filter.UTMTerm = "!term"
	filter.validate()
	args, query := filter.queryFields()
	assert.Len(t, args, 18)
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
	assert.Equal(t, "source", args[13])
	assert.Equal(t, "medium", args[14])
	assert.Equal(t, "campaign", args[15])
	assert.Equal(t, "content", args[16])
	assert.Equal(t, "term", args[17])
	assert.Equal(t, "path != ? AND entry_path != ? AND exit_path != ? AND language != ? AND country_code != ? AND city != ? AND referrer != ? AND referrer_name != ? AND os != ? AND os_version != ? AND browser != ? AND browser_version != ? AND screen_class != ? AND utm_source != ? AND utm_medium != ? AND utm_campaign != ? AND utm_content != ? AND utm_term != ? AND (desktop = 1 OR mobile = 1) ", query)
	filter.EventName = "!event"
	args, query = filter.queryFields()
	assert.Len(t, args, 17)
	assert.Equal(t, "event", args[16])
	assert.Equal(t, "path != ? AND language != ? AND country_code != ? AND city != ? AND referrer != ? AND referrer_name != ? AND os != ? AND os_version != ? AND browser != ? AND browser_version != ? AND screen_class != ? AND utm_source != ? AND utm_medium != ? AND utm_campaign != ? AND utm_content != ? AND utm_term != ? AND event_name != ? AND (desktop = 1 OR mobile = 1) ", query)
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
	filter.UTMSource = "null"
	filter.UTMMedium = "null"
	filter.UTMCampaign = "null"
	filter.UTMContent = "null"
	filter.UTMTerm = "null"
	filter.validate()
	args, query := filter.queryFields()
	assert.Len(t, args, 18)

	for i := 0; i < len(args); i++ {
		assert.Empty(t, args[i])
	}

	assert.Equal(t, "path = ? AND entry_path = ? AND exit_path = ? AND language = ? AND country_code = ? AND city = ? AND referrer = ? AND referrer_name = ? AND os = ? AND os_version = ? AND browser = ? AND browser_version = ? AND screen_class = ? AND utm_source = ? AND utm_medium = ? AND utm_campaign = ? AND utm_content = ? AND utm_term = ? AND desktop = 0 AND mobile = 0 ", query)
	filter.EventName = "null"
	filter.validate()
	args, query = filter.queryFields()
	assert.Len(t, args, 17)

	for i := 0; i < len(args); i++ {
		assert.Empty(t, args[i])
	}

	assert.Equal(t, "path = ? AND language = ? AND country_code = ? AND city = ? AND referrer = ? AND referrer_name = ? AND os = ? AND os_version = ? AND browser = ? AND browser_version = ? AND screen_class = ? AND utm_source = ? AND utm_medium = ? AND utm_campaign = ? AND utm_content = ? AND utm_term = ? AND event_name = ? AND desktop = 0 AND mobile = 0 ", query)
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
	_, query = filter.query()
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
	_, query = filter.query()
	assert.Contains(t, query, "(desktop = 1 OR mobile = 1)")
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

func TestFilter_QueryPageOrEvent(t *testing.T) {
	filter := NewFilter(NullClient)
	filter.Path = "/"
	filter.EventName = "event"
	filter.validate()
	args, query := filter.queryPageOrEvent()
	assert.Len(t, args, 2)
	assert.Equal(t, "/", args[0])
	assert.Equal(t, "event", args[1])
	assert.Equal(t, "path = ? AND event_name = ? ", query)
}

func TestFilter_QueryPageOrEventInvert(t *testing.T) {
	filter := NewFilter(NullClient)
	filter.Path = "!/"
	filter.EventName = "!event"
	filter.validate()
	args, query := filter.queryPageOrEvent()
	assert.Len(t, args, 2)
	assert.Equal(t, "/", args[0])
	assert.Equal(t, "event", args[1])
	assert.Equal(t, `path != ? AND event_name != ? `, query)
	filter.Path = ""
	filter.PathPattern = "!pattern"
	filter.validate()
	args, query = filter.queryPageOrEvent()
	assert.Len(t, args, 2)
	assert.Equal(t, "event", args[0])
	assert.Equal(t, "pattern", args[1])
	assert.Equal(t, `event_name != ? AND match("path", ?) = 0`, query)
}

func TestFilter_QueryPageOrEventNull(t *testing.T) {
	filter := NewFilter(NullClient)
	filter.Path = "null"
	filter.EventName = "null"
	filter.validate()
	args, query := filter.queryPageOrEvent()
	assert.Len(t, args, 2)

	for i := 0; i < len(args); i++ {
		assert.Empty(t, args[i])
	}

	assert.Equal(t, "path = ? AND event_name = ? ", query)
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
	assert.Equal(t, "WITH FILL FROM toDate(?) TO toDate(?)+1 ", query)
}

func TestFilter_WithLimit(t *testing.T) {
	filter := NewFilter(NullClient)
	assert.Empty(t, filter.withLimit())
	filter.Limit = 42
	assert.Equal(t, "LIMIT 42 ", filter.withLimit())
}

func TestFilter_GroupByTitle(t *testing.T) {
	filter := NewFilter(NullClient)
	assert.Empty(t, filter.groupByTitle())
	filter.IncludeTitle = true
	assert.Equal(t, ",title", filter.groupByTitle())
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
	assert.Equal(t, "path,language,country_code,city,referrer,referrer_name,os,os_version,browser,browser_version,screen_class,utm_source,utm_medium,utm_campaign,utm_content,utm_term,event_name,desktop,mobile", filter.fields())
}

func pastDay(n int) time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()-n, 0, 0, 0, 0, time.UTC)
}
