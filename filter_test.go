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
	assert.Equal(t, Today(), filter.To)
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
	assert.Equal(t, "client_id = ? AND toDate(time, 'UTC') >= toDate(?, 'UTC') AND toDate(time, 'UTC') <= toDate(?, 'UTC') AND toDate(time, 'UTC') = toDate(?, 'UTC') AND toDateTime(time, 'UTC') >= toDateTime(?, 'UTC') ", query)
}

func TestFilter_QueryFields(t *testing.T) {
	filter := NewFilter(NullClient)
	filter.Path = "/"
	filter.PathPattern = "pattern"
	filter.Language = "en"
	filter.Country = "jp"
	filter.Referrer = "ref"
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
	assert.Len(t, args, 14)
	assert.Equal(t, "path = ? AND language = ? AND country_code = ? AND referrer = ? AND os = ? AND os_version = ? AND browser = ? AND browser_version = ? AND screen_class = ? AND utm_source = ? AND utm_medium = ? AND utm_campaign = ? AND utm_content = ? AND utm_term = ? AND desktop = 0 AND mobile = 0 ", query)
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

func TestFilter_QueryFieldsPathPattern(t *testing.T) {
	filter := NewFilter(NullClient)
	filter.PathPattern = "/some/pattern"
	args, query := filter.queryFields()
	assert.Len(t, args, 1)
	assert.Equal(t, "/some/pattern", args[0])
	assert.Equal(t, `match("path", ?) = 1`, query)
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
}

func pastDay(n int) time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()-n, 0, 0, 0, 0, time.UTC)
}
