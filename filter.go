package pirsch

import (
	"fmt"
	"strings"
	"time"
)

const (
	// PlatformDesktop filters for everything on desktops.
	PlatformDesktop = "desktop"

	// PlatformMobile filters for everything on mobile devices.
	PlatformMobile = "mobile"

	// PlatformUnknown filters for everything where the platform is unspecified.
	PlatformUnknown = "unknown"
)

// NullClient is a placeholder for no client (0).
var NullClient = int64(0)

// Filter are all fields that can be used to filter the result sets.
type Filter struct {
	// ClientID is the optional.
	ClientID int64

	// Timezone sets the timezone used to interpret dates and times.
	// It will be set to UTC by default.
	Timezone *time.Location

	// From is the start date of the selected period.
	From time.Time

	// To is the end date of the selected period.
	To time.Time

	// Day is an exact match for the result set ("on this day").
	Day time.Time

	// Start is the start date and time of the selected period.
	Start time.Time

	// Path filters for the path.
	Path string

	// Language filters for the ISO language code.
	Language string

	// Country filters for the ISO country code.
	Country string

	// Referrer filters for the referrer.
	Referrer string

	// OS filters for the operating system.
	OS string

	// OSVersion filters for the operating system version.
	OSVersion string

	// Browser filters for the browser.
	Browser string

	// BrowserVersion filters for the browser version.
	BrowserVersion string

	// Platform filters for the platform (desktop, mobile, unknown).
	Platform string

	// ScreenClass filters for the screen class.
	ScreenClass string

	// UTMSource filters for the utm_source query parameter.
	UTMSource string

	// UTMMedium filters for the utm_medium query parameter.
	UTMMedium string

	// UTMCampaign filters for the utm_campaign query parameter.
	UTMCampaign string

	// UTMContent filters for the utm_content query parameter.
	UTMContent string

	// UTMTerm filters for the utm_term query parameter.
	UTMTerm string

	// Limit limits the number of results. Less or equal to zero means no limit.
	Limit int

	// IncludeAvgTimeOnPage indicates whether Analyzer.Pages should contain the average time on page or not.
	IncludeAvgTimeOnPage bool
}

// NewFilter creates a new filter for given client ID.
func NewFilter(clientID int64) *Filter {
	return &Filter{
		ClientID: clientID,
	}
}

func (filter *Filter) validate() {
	if filter.Timezone == nil {
		filter.Timezone = time.UTC
	}

	if !filter.From.IsZero() {
		filter.From = filter.toDate(filter.From)
	}

	if !filter.To.IsZero() {
		filter.To = filter.toDate(filter.To)
	}

	if !filter.Day.IsZero() {
		filter.Day = filter.toDate(filter.Day)
	}

	if !filter.Start.IsZero() {
		filter.Start = time.Date(filter.Start.Year(), filter.Start.Month(), filter.Start.Day(), filter.Start.Hour(), filter.Start.Minute(), filter.Start.Second(), 0, filter.Timezone)
	}

	if !filter.To.IsZero() && filter.From.After(filter.To) {
		filter.From, filter.To = filter.To, filter.From
	}

	now := time.Now().In(filter.Timezone)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, filter.Timezone)

	if !filter.To.IsZero() && filter.To.After(today) {
		filter.To = today
	}

	if filter.Limit < 0 {
		filter.Limit = 0
	}
}

func (filter *Filter) queryTime() ([]interface{}, string) {
	args := make([]interface{}, 0, 5)
	args = append(args, filter.ClientID)
	timezone := filter.Timezone.String()
	var sqlQuery strings.Builder
	sqlQuery.WriteString("client_id = ? ")

	if !filter.From.IsZero() {
		args = append(args, filter.From)
		sqlQuery.WriteString(fmt.Sprintf("AND toDate(time, '%s') >= ? ", timezone))
	}

	if !filter.To.IsZero() {
		args = append(args, filter.To)
		sqlQuery.WriteString(fmt.Sprintf("AND toDate(time, '%s') <= ? ", timezone))
	}

	if !filter.Day.IsZero() {
		args = append(args, filter.Day)
		sqlQuery.WriteString(fmt.Sprintf("AND toDate(time, '%s') = ? ", timezone))
	}

	if !filter.Start.IsZero() {
		args = append(args, filter.Start)
		sqlQuery.WriteString(fmt.Sprintf("AND toDateTime(time, '%s') >= ? ", timezone))
	}

	return args, sqlQuery.String()
}

func (filter *Filter) queryFields() ([]interface{}, string) {
	args := make([]interface{}, 0, 14)
	fields := make([]string, 0, 14)
	filter.appendQuery(&fields, &args, "path", filter.Path)
	filter.appendQuery(&fields, &args, "language", filter.Language)
	filter.appendQuery(&fields, &args, "country_code", filter.Country)
	filter.appendQuery(&fields, &args, "referrer", filter.Referrer)
	filter.appendQuery(&fields, &args, "os", filter.OS)
	filter.appendQuery(&fields, &args, "os_version", filter.OSVersion)
	filter.appendQuery(&fields, &args, "browser", filter.Browser)
	filter.appendQuery(&fields, &args, "browser_version", filter.BrowserVersion)
	filter.appendQuery(&fields, &args, "screen_class", filter.ScreenClass)
	filter.appendQuery(&fields, &args, "utm_source", filter.UTMSource)
	filter.appendQuery(&fields, &args, "utm_medium", filter.UTMMedium)
	filter.appendQuery(&fields, &args, "utm_campaign", filter.UTMCampaign)
	filter.appendQuery(&fields, &args, "utm_content", filter.UTMContent)
	filter.appendQuery(&fields, &args, "utm_term", filter.UTMTerm)

	if filter.Platform != "" {
		if filter.Platform == PlatformDesktop {
			fields = append(fields, "desktop = 1 ")
		} else if filter.Platform == PlatformMobile {
			fields = append(fields, "mobile = 1 ")
		} else {
			fields = append(fields, "desktop = 0 AND mobile = 0 ")
		}
	}

	return args, strings.Join(fields, "AND ")
}

func (filter *Filter) withFill() ([]interface{}, string) {
	if !filter.From.IsZero() && !filter.To.IsZero() {
		timezone := filter.Timezone.String()
		return []interface{}{filter.From, filter.To}, fmt.Sprintf("WITH FILL FROM toDate(?, '%s') TO toDate(?, '%s')+1 ", timezone, timezone)
	}

	return nil, ""
}

func (filter *Filter) withLimit() string {
	if filter.Limit > 0 {
		return fmt.Sprintf("LIMIT %d ", filter.Limit)
	}

	return ""
}

func (filter *Filter) query() ([]interface{}, string) {
	args, query := filter.queryTime()
	fieldArgs, queryFields := filter.queryFields()
	args = append(args, fieldArgs...)

	if queryFields != "" {
		query += "AND " + queryFields
	}

	return args, query
}

func (filter *Filter) appendQuery(fields *[]string, args *[]interface{}, field, value string) {
	if value != "" {
		*args = append(*args, value)
		*fields = append(*fields, fmt.Sprintf("%s = ? ", field))
	}
}

func (filter *Filter) toDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, filter.Timezone)
}

func (filter *Filter) boolean(b bool) int8 {
	if b {
		return 1
	}

	return 0
}
