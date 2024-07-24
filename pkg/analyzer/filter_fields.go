package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v6/pkg"
)

var (
	// FieldSessionsAll is a query result column.
	FieldSessionsAll = Field{
		querySessions:  querySessionFields,
		queryPageViews: querySessionFields,
	}

	// FieldPageViewsAll is a query result column.
	FieldPageViewsAll = Field{
		querySessions:  "t.visitor_id, t.session_id, time, duration_seconds, path, title, language, country_code, region, city, referrer, referrer_name, referrer_icon, os, os_version, browser, browser_version, desktop, mobile, screen_class, utm_source, utm_medium, utm_campaign, utm_content, utm_term, tag_keys, tag_values",
		queryPageViews: "t.visitor_id, t.session_id, time, duration_seconds, path, title, language, country_code, region, city, referrer, referrer_name, referrer_icon, os, os_version, browser, browser_version, desktop, mobile, screen_class, utm_source, utm_medium, utm_campaign, utm_content, utm_term, tag_keys, tag_values",
	}

	// FieldEventsAll is a query result column.
	FieldEventsAll = Field{
		querySessions:  "t.visitor_id, time, t.session_id, event_name, event_meta_keys, event_meta_values, duration_seconds, path, title, language, country_code, region, city, referrer, referrer_name, referrer_icon, os, os_version, browser, browser_version, desktop, mobile, screen_class, utm_source, utm_medium, utm_campaign, utm_content, utm_term",
		queryPageViews: "t.visitor_id, time, t.session_id, event_name, event_meta_keys, event_meta_values, duration_seconds, path, title, language, country_code, region, city, referrer, referrer_name, referrer_icon, os, os_version, browser, browser_version, desktop, mobile, screen_class, utm_source, utm_medium, utm_campaign, utm_content, utm_term",
	}

	// FieldClientID is a query result column.
	FieldClientID = Field{
		querySessions:  "t.client_id",
		queryPageViews: "t.client_id",
		Name:           "client_id",
	}

	// FieldVisitorID is a query result column.
	FieldVisitorID = Field{
		querySessions:  "t.visitor_id",
		queryPageViews: "t.visitor_id",
		Name:           "visitor_id",
	}

	// FieldSessionID is a query result column.
	FieldSessionID = Field{
		querySessions:  "t.session_id",
		queryPageViews: "t.session_id",
		Name:           "session_id",
	}

	// FieldCount is a query result column.
	FieldCount = Field{
		querySessions:  "count(*)",
		queryPageViews: "count(*)",
		Name:           "count",
		queryDirection: "DESC",
		sampleType:     sampleTypeInt,
	}

	// FieldPath is a query result column.
	FieldPath = Field{
		querySessions:  "path",
		queryPageViews: "path",
		queryImported:  "coalesce(nullif(t.path, ''), imp.path)",
		queryDirection: "ASC",
		Name:           "path",
	}

	// FieldEventPath is a query result column.
	FieldEventPath = Field{
		querySessions:  "path",
		queryPageViews: "t.path",
		queryDirection: "ASC",
		Name:           "path",
	}

	// FieldEntryPath is a query result column.
	FieldEntryPath = Field{
		querySessions:  "entry_path",
		queryPageViews: "entry_path",
		queryImported:  "coalesce(nullif(t.entry_path, ''), imp.entry_path)",
		queryDirection: "ASC",
		Name:           "entry_path",
	}

	// FieldEntries is a query result column.
	FieldEntries = Field{
		querySessions:  "sum(sign)",
		queryPageViews: "uniq(t.visitor_id, t.session_id)",
		queryDirection: "DESC",
		sampleType:     sampleTypeInt,
		Name:           "entries",
	}

	// FieldExitPath is a query result column.
	FieldExitPath = Field{
		querySessions:  "exit_path",
		queryPageViews: "exit_path",
		queryImported:  "coalesce(nullif(t.exit_path, ''), imp.exit_path)",
		queryDirection: "ASC",
		Name:           "exit_path",
	}

	// FieldExits is a query result column.
	FieldExits = Field{
		querySessions:  "sum(sign)",
		queryPageViews: "uniq(t.visitor_id, t.session_id)",
		queryDirection: "DESC",
		sampleType:     sampleTypeInt,
		Name:           "exits",
	}

	// FieldVisitors is a query result column.
	FieldVisitors = Field{
		querySessions:  "uniq(t.visitor_id)",
		queryPageViews: "uniq(t.visitor_id)",
		queryImported:  "sum(t.visitors + imp.visitors)",
		queryPeriod:    "sum(visitors)",
		queryDirection: "DESC",
		sampleType:     sampleTypeInt,
		Name:           "visitors",
	}

	// FieldVisitorsRaw is a query result column.
	FieldVisitorsRaw = Field{
		querySessions:  "uniq(visitor_id)",
		queryPageViews: "uniq(visitor_id)",
		queryDirection: "DESC",
		sampleType:     sampleTypeInt,
		Name:           "visitors",
	}

	// FieldRelativeVisitors is a query result column.
	FieldRelativeVisitors = Field{
		querySessions:  `toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id)%s FROM "session"%s WHERE %s), 1))`,
		queryPageViews: `toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id)%s FROM "session"%s WHERE %s), 1))`,
		queryImported:  `toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id)%s FROM "session"%s WHERE %s) + (SELECT sum(visitors) FROM "%s" WHERE %s), 1))`,
		queryDirection: "DESC",
		filterTime:     true,
		Name:           "relative_visitors",
	}

	// FieldCR is a query result column.
	FieldCR = Field{
		querySessions:  `toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id)%s FROM "session"%s WHERE %s), 1))`,
		queryPageViews: `toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id)%s FROM "session"%s WHERE %s), 1))`,
		queryDirection: "DESC",
		filterTime:     true,
		Name:           "cr",
	}

	// FieldCRPeriod is a query result column.
	FieldCRPeriod = Field{
		querySessions:  `toFloat64OrDefault(visitors / greatest(ifNull(max(uvd.visitors), visitors), 1))`,
		queryPageViews: `toFloat64OrDefault(visitors / greatest(ifNull(max(uvd.visitors), visitors), 1))`,
		queryDirection: "DESC",
		sampleType:     sampleTypeFloat,
		Name:           "cr",
	}

	// FieldSessions is a query result column.
	FieldSessions = Field{
		querySessions:  "uniq(t.visitor_id, t.session_id)",
		queryPageViews: "uniq(t.visitor_id, t.session_id)",
		queryPeriod:    "sum(sessions)",
		queryDirection: "DESC",
		sampleType:     sampleTypeInt,
		Name:           "sessions",
	}

	// FieldViews is a query result column.
	FieldViews = Field{
		querySessions:  "sum(page_views*sign)",
		queryPageViews: "count(1)",
		queryEvents:    "sum(views)",
		queryPeriod:    "sum(views)",
		queryDirection: "DESC",
		sampleType:     sampleTypeInt,
		Name:           "views",
	}

	// FieldRelativeViews is a query result column.
	FieldRelativeViews = Field{
		querySessions:  `toFloat64OrDefault(views / greatest((SELECT sum(page_views*sign)%s views FROM "session"%s WHERE %s), 1))`,
		queryPageViews: `toFloat64OrDefault(views / greatest((SELECT sum(page_views*sign)%s views FROM "session"%s WHERE %s), 1))`,
		queryDirection: "DESC",
		filterTime:     true,
		Name:           "relative_views",
	}

	// FieldBounces is a query result column.
	FieldBounces = Field{
		querySessions:  "sum(is_bounce*sign)",
		queryPageViews: "uniqIf((t.visitor_id, t.session_id), bounces = 1)",
		queryPeriod:    "sum(bounces)",
		queryDirection: "DESC",
		sampleType:     sampleTypeInt,
		Name:           "bounces",
	}

	// FieldBounceRate is a query result column.
	FieldBounceRate = Field{
		querySessions:  "bounces / IF(sessions = 0, 1, sessions)",
		queryPageViews: "bounces / IF(sessions = 0, 1, sessions)",
		queryPeriod:    "ifNotFinite(avg(bounce_rate), 0)",
		queryDirection: "DESC",
		Name:           "bounce_rate",
	}

	// FieldReferrer is a query result column.
	FieldReferrer = Field{
		querySessions:  "referrer",
		queryPageViews: "referrer",
		queryImported:  "coalesce(nullif(t.referrer, ''), imp.referrer)",
		queryDirection: "ASC",
		Name:           "referrer",
	}

	// FieldAnyReferrer is a query result column.
	FieldAnyReferrer = Field{
		querySessions:  "any(referrer)",
		queryPageViews: "any(referrer)",
		queryDirection: "ASC",
		Name:           "referrer",
	}

	// FieldReferrerName is a query result column.
	FieldReferrerName = Field{
		querySessions:  "referrer_name",
		queryPageViews: "referrer_name",
		queryDirection: "ASC",
		Name:           "referrer_name",
	}

	// FieldReferrerIcon is a query result column.
	FieldReferrerIcon = Field{
		querySessions:  "any(referrer_icon)",
		queryPageViews: "any(referrer_icon)",
		queryDirection: "ASC",
		Name:           "referrer_icon",
	}

	// FieldLanguage is a query result column.
	FieldLanguage = Field{
		querySessions:  "language",
		queryPageViews: "language",
		queryImported:  "coalesce(nullif(t.language, ''), imp.language)",
		queryDirection: "ASC",
		Name:           "language",
	}

	// FieldCountryCity is a query result column.
	// This field can only be used in combination with the FieldCity.
	FieldCountryCity = Field{
		querySessions:  "if(city = '', '', country_code)",
		queryPageViews: "if(city = '', '', country_code)",
		queryImported:  "coalesce(nullif(t.country_code, ''), imp.city)",
		queryDirection: "ASC",
		Name:           "country_code",
	}

	// FieldCountryRegion is a query result column.
	// This field can only be used in combination with the FieldCountry.
	FieldCountryRegion = Field{
		querySessions:  "if(region = '', '', country_code)",
		queryPageViews: "if(region = '', '', country_code)",
		queryImported:  "coalesce(nullif(t.country_code, ''), imp.region)",
		queryDirection: "ASC",
		Name:           "country_code",
	}

	// FieldCountry is a query result column.
	FieldCountry = Field{
		querySessions:  "country_code",
		queryPageViews: "country_code",
		queryImported:  "coalesce(nullif(t.country_code, ''), imp.country_code)",
		queryDirection: "ASC",
		Name:           "country_code",
	}

	// FieldRegionCity is a query result column.
	// This field can only be used in combination with the FieldCity.
	FieldRegionCity = Field{
		querySessions:  "if(city = '', '', region)",
		queryPageViews: "if(city = '', '', region)",
		queryImported:  "coalesce(nullif(t.region, ''), imp.region)",
		queryDirection: "ASC",
		Name:           "region",
	}

	// FieldRegion is a query result column.
	FieldRegion = Field{
		querySessions:  "region",
		queryPageViews: "region",
		queryImported:  "coalesce(nullif(t.region, ''), imp.region)",
		queryDirection: "ASC",
		Name:           "region",
	}

	// FieldCity is a query result column.
	FieldCity = Field{
		querySessions:  "city",
		queryPageViews: "city",
		queryImported:  "coalesce(nullif(t.city, ''), imp.city)",
		queryDirection: "ASC",
		Name:           "city",
	}

	// FieldBrowser is a query result column.
	FieldBrowser = Field{
		querySessions:  "browser",
		queryPageViews: "browser",
		queryImported:  "coalesce(nullif(t.browser, ''), imp.browser)",
		queryDirection: "ASC",
		Name:           "browser",
	}

	// FieldBrowserVersion is a query result column.
	FieldBrowserVersion = Field{
		querySessions:  "browser_version",
		queryPageViews: "browser_version",
		queryDirection: "ASC",
		Name:           "browser_version",
	}

	// FieldOS is a query result column.
	FieldOS = Field{
		querySessions:  "os",
		queryPageViews: "os",
		queryImported:  "coalesce(nullif(t.os, ''), imp.os)",
		queryDirection: "ASC",
		Name:           "os",
	}

	// FieldOSVersion is a query result column.
	FieldOSVersion = Field{
		querySessions:  "os_version",
		queryPageViews: "os_version",
		queryDirection: "ASC",
		Name:           "os_version",
	}

	// FieldScreenClass is a query result column.
	FieldScreenClass = Field{
		querySessions:  "screen_class",
		queryPageViews: "screen_class",
		queryDirection: "ASC",
		Name:           "screen_class",
	}

	// FieldUTMSource is a query result column.
	FieldUTMSource = Field{
		querySessions:  "utm_source",
		queryPageViews: "utm_source",
		queryImported:  "coalesce(nullif(t.utm_source, ''), imp.utm_source)",
		queryDirection: "ASC",
		Name:           "utm_source",
	}

	// FieldUTMMedium is a query result column.
	FieldUTMMedium = Field{
		querySessions:  "utm_medium",
		queryPageViews: "utm_medium",
		queryImported:  "coalesce(nullif(t.utm_medium, ''), imp.utm_medium)",
		queryDirection: "ASC",
		Name:           "utm_medium",
	}

	// FieldUTMCampaign is a query result column.
	FieldUTMCampaign = Field{
		querySessions:  "utm_campaign",
		queryPageViews: "utm_campaign",
		queryImported:  "coalesce(nullif(t.utm_campaign, ''), imp.utm_campaign)",
		queryDirection: "ASC",
		Name:           "utm_campaign",
	}

	// FieldUTMContent is a query result column.
	FieldUTMContent = Field{
		querySessions:  "utm_content",
		queryPageViews: "utm_content",
		queryDirection: "ASC",
		Name:           "utm_content",
	}

	// FieldUTMTerm is a query result column.
	FieldUTMTerm = Field{
		querySessions:  "utm_term",
		queryPageViews: "utm_term",
		queryDirection: "ASC",
		Name:           "utm_term",
	}

	// FieldTagKeysRaw is a query result column.
	FieldTagKeysRaw = Field{
		querySessions:  "tag_keys",
		queryPageViews: "tag_keys",
		Name:           "tag_keys",
	}

	// FieldTagValuesRaw is a query result column.
	FieldTagValuesRaw = Field{
		querySessions:  "tag_values",
		queryPageViews: "tag_values",
		Name:           "tag_values",
	}

	// FieldTagKey is a query result column.
	FieldTagKey = Field{
		querySessions:  "arrayJoin(tag_keys)",
		queryPageViews: "arrayJoin(tag_keys)",
		Name:           "key",
	}

	// FieldTagValue is a query result column.
	FieldTagValue = Field{
		querySessions:  "tag_values[indexOf(tag_keys, ?)]",
		queryPageViews: "tag_values[indexOf(tag_keys, ?)]",
		Name:           "value",
	}

	// FieldTitle is a query result column.
	FieldTitle = Field{
		querySessions:  "title",
		queryPageViews: "title",
		queryDirection: "ASC",
		Name:           "title",
	}

	// FieldEventTitle is a query result column.
	FieldEventTitle = Field{
		querySessions:  "title",
		queryPageViews: "t.title",
		queryEvents:    "title",
		queryDirection: "ASC",
		Name:           "title",
	}

	// FieldEntryTitle is a query result column.
	FieldEntryTitle = Field{
		querySessions:  "entry_title",
		queryPageViews: "entry_title",
		queryDirection: "ASC",
		Name:           "title",
	}

	// FieldExitTitle is a query result column.
	FieldExitTitle = Field{
		querySessions:  "exit_title",
		queryPageViews: "exit_title",
		queryDirection: "ASC",
		Name:           "title",
	}

	// FieldSessionExitTitle is a query result column.
	FieldSessionExitTitle = Field{
		querySessions:  "exit_title",
		queryPageViews: "exit_title",
		queryDirection: "ASC",
		Name:           "exit_title",
	}

	// FieldTime is a query result column.
	FieldTime = Field{
		querySessions:  "t.time",
		queryPageViews: "t.time",
		queryDirection: "ASC",
		Name:           "time",
	}

	// FieldMaxTime is a query result column.
	FieldMaxTime = Field{
		querySessions:  "max(time)",
		queryPageViews: "max(time)",
		queryDirection: "ASC",
		Name:           "max(time)",
	}

	// FieldDay is a query result column.
	FieldDay = Field{
		querySessions:  "toDate(time, '%s')",
		queryPageViews: "toDate(time, '%s')",
		queryDirection: "ASC",
		withFill:       true,
		timezone:       true,
		Name:           "day",
	}

	// FieldHour is a query result column.
	FieldHour = Field{
		querySessions:  "toHour(time, '%s')",
		queryPageViews: "toHour(time, '%s')",
		queryDirection: "ASC",
		queryWithFill:  "WITH FILL FROM 0 TO 24",
		timezone:       true,
		Name:           "hour",
	}

	// FieldMinute is a query result column.
	FieldMinute = Field{
		querySessions:  "toMinute(time, '%s')",
		queryPageViews: "toMinute(time, '%s')",
		queryDirection: "ASC",
		queryWithFill:  "WITH FILL FROM 0 TO 60",
		timezone:       true,
		Name:           "minute",
	}

	// FieldEventName is a query result column.
	FieldEventName = Field{
		querySessions:  "event_name",
		queryPageViews: "event_name",
		Name:           "event_name",
		queryDirection: "ASC",
	}

	// FieldEventMeta is a query result column.
	FieldEventMeta = Field{
		querySessions:  "cast(arraySort(arrayZip(event_meta_keys, event_meta_values)), 'Map(String, String)')",
		queryPageViews: "cast(arraySort(arrayZip(event_meta_keys, event_meta_values)), 'Map(String, String)')",
		Name:           "meta",
	}

	// FieldEventMetaKeys is a query result column.
	FieldEventMetaKeys = Field{
		querySessions:  "groupUniqArrayArray(event_meta_keys)",
		queryPageViews: "groupUniqArrayArray(event_meta_keys)",
		Name:           "meta_keys",
	}

	// FieldEventMetaKeysRaw is a query result column.
	FieldEventMetaKeysRaw = Field{
		querySessions:  "event_meta_keys",
		queryPageViews: "event_meta_keys",
		Name:           "event_meta_keys",
	}

	// FieldEventMetaValues is a query result column.
	FieldEventMetaValues = Field{
		querySessions:  "event_meta_values[indexOf(event_meta_keys, ?)]",
		queryPageViews: "event_meta_values[indexOf(event_meta_keys, ?)]",
		Name:           "meta_value",
	}

	// FieldEventMetaValuesRaw is a query result column.
	FieldEventMetaValuesRaw = Field{
		querySessions:  "event_meta_values",
		queryPageViews: "event_meta_values",
		Name:           "event_meta_values",
	}

	// FieldEventTimeSpent is a query result column.
	FieldEventTimeSpent = Field{
		querySessions:  "toUInt64(greatest(ifNotFinite(avg(duration_seconds), 0), 0))",
		queryPageViews: "toUInt64(greatest(ifNotFinite(avg(duration_seconds), 0), 0))",
		sampleType:     sampleTypeInt,
		Name:           "average_time_spent_seconds",
	}

	// FieldEventMetaCustomMetricAvg is a query result column.
	FieldEventMetaCustomMetricAvg = Field{
		querySessions:  "ifNotFinite(avg(coalesce(%s(event_meta_values[indexOf(event_meta_keys, ?)]))), 0)",
		queryPageViews: "ifNotFinite(avg(coalesce(%s(event_meta_values[indexOf(event_meta_keys, ?)]))), 0)",
		sampleType:     sampleTypeFloat,
		Name:           "custom_metric_avg",
	}

	// FieldEventMetaCustomMetricTotal is a query result column.
	FieldEventMetaCustomMetricTotal = Field{
		querySessions:  "sum(coalesce(%s(event_meta_values[indexOf(event_meta_keys, ?)])))",
		queryPageViews: "sum(coalesce(%s(event_meta_values[indexOf(event_meta_keys, ?)])))",
		sampleType:     sampleTypeAuto,
		Name:           "custom_metric_total",
	}

	// FieldPlatformDesktop is a query result column.
	FieldPlatformDesktop = Field{
		querySessions:  "uniqIf(visitor_id, desktop = 1)",
		queryPageViews: "desktop = 1,mobile = 0",
		sampleType:     sampleTypeInt,
		Name:           "platform_desktop",
	}

	// FieldPlatformMobile is a query result column.
	FieldPlatformMobile = Field{
		querySessions:  "uniqIf(visitor_id, mobile = 1)",
		queryPageViews: "desktop = 0,mobile = 1",
		sampleType:     sampleTypeInt,
		Name:           "platform_mobile",
	}

	// FieldPlatformUnknown is a query result column.
	FieldPlatformUnknown = Field{
		querySessions:  "uniq(visitor_id)-platform_desktop-platform_mobile",
		queryPageViews: "desktop = 0,mobile = 0",
		sampleType:     sampleTypeInt,
		Name:           "platform_unknown",
	}

	// FieldRelativePlatformDesktop is a query result column.
	FieldRelativePlatformDesktop = Field{
		querySessions:  "platform_desktop / IF(platform_desktop + platform_mobile + platform_unknown = 0, 1, platform_desktop + platform_mobile + platform_unknown)",
		queryPageViews: "platform_desktop / IF(platform_desktop + platform_mobile + platform_unknown = 0, 1, platform_desktop + platform_mobile + platform_unknown)",
		Name:           "relative_platform_desktop",
	}

	// FieldRelativePlatformMobile is a query result column.
	FieldRelativePlatformMobile = Field{
		querySessions:  "platform_mobile / IF(platform_desktop + platform_mobile + platform_unknown = 0, 1, platform_desktop + platform_mobile + platform_unknown)",
		queryPageViews: "platform_mobile / IF(platform_desktop + platform_mobile + platform_unknown = 0, 1, platform_desktop + platform_mobile + platform_unknown)",
		Name:           "relative_platform_mobile",
	}

	// FieldRelativePlatformUnknown is a query result column.
	FieldRelativePlatformUnknown = Field{
		querySessions:  "platform_unknown / IF(platform_desktop + platform_mobile + platform_unknown = 0, 1, platform_desktop + platform_mobile + platform_unknown)",
		queryPageViews: "platform_unknown / IF(platform_desktop + platform_mobile + platform_unknown = 0, 1, platform_desktop + platform_mobile + platform_unknown)",
		Name:           "relative_platform_unknown",
	}

	// FieldEventDurationSeconds is a query result column.
	FieldEventDurationSeconds = Field{
		querySessions:  "sum(duration_seconds)",
		queryPageViews: "sum(duration_seconds)",
		sampleType:     sampleTypeInt,
		Name:           "duration_seconds",
	}
)

const (
	sampleTypeInt   = sampleType(1)
	sampleTypeFloat = sampleType(2)
	sampleTypeAuto  = sampleType(3) // selected by query builder

	querySessionFields = `t.visitor_id visitor_id,
		t.session_id session_id,
		max(t.time) session_time,
		min(t.start) session_start,
		max(t.duration_seconds) session_duration_seconds,
		any(t.entry_path) session_entry_path,
		t.exit_path session_exit_path,
		max(t.page_views) session_page_views,
		min(t.is_bounce) session_is_bounce,
		any(t.entry_title) session_entry_title,
		t.exit_title session_exit_title,
		any(t.language) session_language,
		any(t.country_code) session_country_code,
		any(t.region) session_region,
		any(t.city) session_city,
		any(t.referrer) session_referrer,
		any(t.referrer_name) session_referrer_name,
		any(t.referrer_icon) session_referrer_icon,
		any(t.os) session_os,
		any(t.os_version) session_os_version,
		any(t.browser) session_browser,
		any(t.browser_version) session_browser_version,
		any(t.desktop) session_desktop,
		any(t.mobile) session_mobile,
		any(t.screen_class) session_screen_class,
		any(t.utm_source) session_utm_source,
		any(t.utm_medium) session_utm_medium,
		any(t.utm_campaign) session_utm_campaign,
		any(t.utm_content) session_utm_content,
		any(t.utm_term) session_utm_term,
		max(t.extended) session_extended`
)

type sampleType int

// Field is a column for a query.
type Field struct {
	querySessions  string
	queryPageViews string
	queryEvents    string
	queryImported  string
	queryPeriod    string
	queryDirection pkg.Direction
	queryWithFill  string
	withFill       bool
	timezone       bool
	filterTime     bool
	sampleType     sampleType
	Name           string
}
