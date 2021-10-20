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

	// Unknown filters for an unknown (empty) value.
	// This is a synonym for "null".
	Unknown = "null"
)

// NullClient is a placeholder for no client (0).
var NullClient = int64(0)

// Filter are all fields that can be used to filter the result sets.
// Fields can be inverted by adding a "!" in front of the string.
// To compare to none/unknown/empty, set the value to "null" (case-insensitive).
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
	// Note that if this and PathPattern are both set, Path will be preferred.
	Path string

	// EntryPath filters for the entry page.
	EntryPath string

	// ExitPath filters for the exit page.
	ExitPath string

	// PathPattern filters for the path using a (ClickHouse supported) regex pattern.
	// Note that if this and Path are both set, Path will be preferred.
	// Examples for useful patterns (all case-insensitive, * is used for every character but slashes, ** is used for all characters including slashes):
	//  (?i)^/path/[^/]+$ // matches /path/*
	//  (?i)^/path/[^/]+/.* // matches /path/*/**
	//  (?i)^/path/[^/]+/slashes$ // matches /path/*/slashes
	//  (?i)^/path/.+/slashes$ // matches /path/**/slashes
	PathPattern string

	// Language filters for the ISO language code.
	Language string

	// Country filters for the ISO country code.
	Country string

	// City filters for the city name.
	City string

	// Referrer filters for the full referrer.
	Referrer string

	// ReferrerName filters for the referrer name.
	ReferrerName string

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

	// EventName filters for an event by its name.
	EventName string

	// EventMetaKey filters for an event meta key.
	// This must be used together with an EventName.
	EventMetaKey string

	// Limit limits the number of results. Less or equal to zero means no limit.
	Limit int

	// IncludeTitle indicates whether the Analyzer.Pages, Analyzer.EntryPages, and Analyzer.ExitPages should contain the page title or not.
	IncludeTitle bool

	// MaxTimeOnPageSeconds is an optional maximum for the time spent on page.
	// Visitors who are idle artificially increase the average time spent on a page, this option can be used to limit the effect.
	// Set to 0 to disable this option (default).
	MaxTimeOnPageSeconds int

	eventFilter bool
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
	} else {
		filter.From = filter.From.In(time.UTC)
	}

	if !filter.To.IsZero() {
		filter.To = filter.toDate(filter.To)
	} else {
		filter.To = filter.To.In(time.UTC)
	}

	if !filter.Day.IsZero() {
		filter.Day = filter.toDate(filter.Day)
	} else {
		filter.Day = filter.Day.In(time.UTC)
	}

	if !filter.Start.IsZero() {
		filter.Start = time.Date(filter.Start.Year(), filter.Start.Month(), filter.Start.Day(), filter.Start.Hour(), filter.Start.Minute(), filter.Start.Second(), 0, time.UTC)
	}

	if !filter.To.IsZero() && filter.From.After(filter.To) {
		filter.From, filter.To = filter.To, filter.From
	}

	// use tomorrow instead of limiting to "today", so that all timezones are included
	now := time.Now().UTC()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)

	if !filter.To.IsZero() && filter.To.After(tomorrow) {
		filter.To = tomorrow
	}

	if filter.Path != "" && filter.PathPattern != "" {
		filter.PathPattern = ""
	}

	if filter.Limit < 0 {
		filter.Limit = 0
	}
}

func (filter *Filter) table() string {
	if filter.EventName != "" || filter.eventFilter {
		return "event"
	}

	return "session"
}

func (filter *Filter) queryTime() ([]interface{}, string) {
	args := make([]interface{}, 0, 5)
	args = append(args, filter.ClientID)
	timezone := filter.Timezone.String()
	var sqlQuery strings.Builder
	sqlQuery.WriteString("client_id = ? ")

	if !filter.From.IsZero() {
		args = append(args, filter.From)
		sqlQuery.WriteString(fmt.Sprintf("AND toDate(time, '%s') >= toDate(?) ", timezone))
	}

	if !filter.To.IsZero() {
		args = append(args, filter.To)
		sqlQuery.WriteString(fmt.Sprintf("AND toDate(time, '%s') <= toDate(?) ", timezone))
	}

	if !filter.Day.IsZero() {
		args = append(args, filter.Day)
		sqlQuery.WriteString(fmt.Sprintf("AND toDate(time, '%s') = toDate(?) ", timezone))
	}

	if !filter.Start.IsZero() {
		args = append(args, filter.Start)
		sqlQuery.WriteString(fmt.Sprintf("AND toDateTime(time, '%s') >= toDateTime(?, '%s') ", timezone, timezone))
	}

	return args, sqlQuery.String()
}

func (filter *Filter) queryFields() ([]interface{}, string) {
	args := make([]interface{}, 0, 22)
	queryFields := make([]string, 0, 22)
	filter.appendQuery(&queryFields, &args, "path", filter.Path)

	if filter.EventName == "" && !filter.eventFilter {
		filter.appendQuery(&queryFields, &args, "entry_path", filter.EntryPath)
		filter.appendQuery(&queryFields, &args, "exit_path", filter.ExitPath)
	}

	filter.appendQuery(&queryFields, &args, "language", filter.Language)
	filter.appendQuery(&queryFields, &args, "country_code", filter.Country)
	filter.appendQuery(&queryFields, &args, "city", filter.City)
	filter.appendQuery(&queryFields, &args, "referrer", filter.Referrer)
	filter.appendQuery(&queryFields, &args, "referrer_name", filter.ReferrerName)
	filter.appendQuery(&queryFields, &args, "os", filter.OS)
	filter.appendQuery(&queryFields, &args, "os_version", filter.OSVersion)
	filter.appendQuery(&queryFields, &args, "browser", filter.Browser)
	filter.appendQuery(&queryFields, &args, "browser_version", filter.BrowserVersion)
	filter.appendQuery(&queryFields, &args, "screen_class", filter.ScreenClass)
	filter.appendQuery(&queryFields, &args, "utm_source", filter.UTMSource)
	filter.appendQuery(&queryFields, &args, "utm_medium", filter.UTMMedium)
	filter.appendQuery(&queryFields, &args, "utm_campaign", filter.UTMCampaign)
	filter.appendQuery(&queryFields, &args, "utm_content", filter.UTMContent)
	filter.appendQuery(&queryFields, &args, "utm_term", filter.UTMTerm)
	filter.appendQuery(&queryFields, &args, "event_name", filter.EventName)
	filter.queryPlatform(&queryFields)
	filter.queryPathPattern(&queryFields, &args)
	return args, strings.Join(queryFields, "AND ")
}

func (filter *Filter) queryPlatform(queryFields *[]string) {
	if filter.Platform != "" {
		if strings.HasPrefix(filter.Platform, "!") {
			platform := filter.Platform[1:]

			if platform == PlatformDesktop {
				*queryFields = append(*queryFields, "desktop != 1 ")
			} else if platform == PlatformMobile {
				*queryFields = append(*queryFields, "mobile != 1 ")
			} else {
				*queryFields = append(*queryFields, "(desktop = 1 OR mobile = 1) ")
			}
		} else {
			if filter.Platform == PlatformDesktop {
				*queryFields = append(*queryFields, "desktop = 1 ")
			} else if filter.Platform == PlatformMobile {
				*queryFields = append(*queryFields, "mobile = 1 ")
			} else {
				*queryFields = append(*queryFields, "desktop = 0 AND mobile = 0 ")
			}
		}
	}
}

func (filter Filter) queryPathPattern(queryFields *[]string, args *[]interface{}) {
	if filter.PathPattern != "" {
		if strings.HasPrefix(filter.PathPattern, "!") {
			*args = append(*args, filter.PathPattern[1:])
			*queryFields = append(*queryFields, `match("path", ?) = 0`)
		} else {
			*args = append(*args, filter.PathPattern)
			*queryFields = append(*queryFields, `match("path", ?) = 1`)
		}
	}
}

func (filter *Filter) fields() string {
	// do not include exit_path, as it is selected using argMax
	fields := make([]string, 0, 20)
	filter.appendField(&fields, "path", filter.Path)

	if filter.EventName == "" && !filter.eventFilter {
		filter.appendField(&fields, "entry_path", filter.EntryPath)
	}

	filter.appendField(&fields, "language", filter.Language)
	filter.appendField(&fields, "country_code", filter.Country)
	filter.appendField(&fields, "city", filter.City)
	filter.appendField(&fields, "referrer", filter.Referrer)
	filter.appendField(&fields, "referrer_name", filter.ReferrerName)
	filter.appendField(&fields, "os", filter.OS)
	filter.appendField(&fields, "os_version", filter.OSVersion)
	filter.appendField(&fields, "browser", filter.Browser)
	filter.appendField(&fields, "browser_version", filter.BrowserVersion)
	filter.appendField(&fields, "screen_class", filter.ScreenClass)
	filter.appendField(&fields, "utm_source", filter.UTMSource)
	filter.appendField(&fields, "utm_medium", filter.UTMMedium)
	filter.appendField(&fields, "utm_campaign", filter.UTMCampaign)
	filter.appendField(&fields, "utm_content", filter.UTMContent)
	filter.appendField(&fields, "utm_term", filter.UTMTerm)
	filter.appendField(&fields, "event_name", filter.EventName)

	if filter.Platform != "" {
		platform := filter.Platform

		if strings.HasPrefix(platform, "!") {
			platform = filter.Platform[1:]
		}

		if platform == PlatformDesktop {
			fields = append(fields, "desktop")
		} else if platform == PlatformMobile {
			fields = append(fields, "mobile")
		} else {
			fields = append(fields, "desktop")
			fields = append(fields, "mobile")
		}
	}

	if filter.Path == "" && filter.PathPattern != "" {
		fields = append(fields, "path")
	}

	return strings.Join(fields, ",")
}

func (filter *Filter) appendField(fields *[]string, field, value string) {
	if value != "" {
		*fields = append(*fields, field)
	}
}

func (filter *Filter) withFill() ([]interface{}, string) {
	if !filter.From.IsZero() && !filter.To.IsZero() {
		return []interface{}{filter.From, filter.To}, "WITH FILL FROM toDate(?) TO toDate(?)+1 "
	}

	return nil, ""
}

func (filter *Filter) withLimit() string {
	if filter.Limit > 0 {
		return fmt.Sprintf("LIMIT %d ", filter.Limit)
	}

	return ""
}

func (filter *Filter) groupByTitle() string {
	if filter.IncludeTitle {
		return ",title"
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

func (filter *Filter) appendQuery(queryFields *[]string, args *[]interface{}, field, value string) {
	if value != "" {
		if strings.HasPrefix(value, "!") {
			value = filter.nullValue(value[1:])
			*args = append(*args, value)
			*queryFields = append(*queryFields, fmt.Sprintf("%s != ? ", field))
		} else {
			*args = append(*args, filter.nullValue(value))
			*queryFields = append(*queryFields, fmt.Sprintf("%s = ? ", field))
		}
	}
}

func (filter *Filter) nullValue(value string) string {
	if strings.ToLower(value) == "null" {
		return ""
	}

	return value
}

func (filter *Filter) toDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
}

func (filter *Filter) boolean(b bool) int8 {
	if b {
		return 1
	}

	return 0
}
