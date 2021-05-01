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
	assert.Zero(t, filter.From)
	assert.Zero(t, filter.To)
	filter = &Filter{From: pastDay(2), To: pastDay(5)}
	filter.validate()
	assert.Equal(t, pastDay(5), filter.From)
	assert.Equal(t, pastDay(2), filter.To)
	filter = &Filter{From: pastDay(2), To: Today().Add(time.Hour * 24 * 5)}
	filter.validate()
	assert.Equal(t, pastDay(2), filter.From)
	assert.Equal(t, Today(), filter.To)
	filter = &Filter{Day: time.Now()}
	filter.validate()
	assert.Zero(t, filter.Day.Hour())
	assert.Zero(t, filter.Day.Minute())
	assert.Zero(t, filter.Day.Second())
	assert.Zero(t, filter.Day.Nanosecond())
}

func TestFilter_QueryTime(t *testing.T) {
	filter := NewFilter(NullClient)
	filter.From = pastDay(5)
	filter.To = pastDay(2)
	filter.Day = pastDay(1)
	filter.Start = time.Now()
	args, query := filter.queryTime()
	assert.Len(t, args, 5)
	assert.Equal(t, NullClient, args[0])
	assert.Equal(t, filter.From, args[1])
	assert.Equal(t, filter.To, args[2])
	assert.Equal(t, filter.Day, args[3])
	assert.Equal(t, filter.Start, args[4])
	assert.Equal(t, "client_id = ? AND toDate(time) >= ? AND toDate(time) <= ? AND toDate(time) = ? AND time >= ? ", query)
}

func TestFilter_QueryFields(t *testing.T) {
	filter := NewFilter(NullClient)
	filter.Path = "/"
	filter.Language = "en"
	filter.Country = "jp"
	filter.Referrer = "ref"
	filter.OS = OSWindows
	filter.OSVersion = "10"
	filter.Browser = BrowserEdge
	filter.BrowserVersion = "89"
	filter.Platform = PlatformDesktop
	filter.ScreenClass = "XXL"
	filter.UTMSource = "source"
	filter.UTMMedium = "medium"
	filter.UTMCampaign = "campaign"
	filter.UTMContent = "content"
	filter.UTMTerm = "term"
	args, query := filter.queryFields()
	assert.Len(t, args, 15)
	assert.Equal(t, "path = ? AND language = ? AND country_code = ? AND referrer = ? AND os = ? AND os_version = ? AND browser = ? AND browser_version = ? AND desktop IS TRUE AND screen_class = ? AND utm_source = ? AND utm_medium = ? AND utm_campaign = ? AND utm_content = ? AND utm_term = ? ", query)
}

func pastDay(n int) time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()-n, 0, 0, 0, 0, time.UTC)
}
