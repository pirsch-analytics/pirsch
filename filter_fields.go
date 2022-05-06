package pirsch

var (
	// FieldCount is a query result column.
	FieldCount = Field{
		querySessions:  "count(*)",
		queryPageViews: "count(*)",
		name:           "count",
		queryDirection: "DESC",
	}

	// FieldPath is a query result column.
	FieldPath = Field{
		querySessions:  "path",
		queryPageViews: "path",
		queryDirection: "ASC",
		name:           "path",
	}

	// FieldEntryPath is a query result column.
	FieldEntryPath = Field{
		querySessions:  "entry_path",
		queryPageViews: "entry_path",
		queryDirection: "ASC",
		name:           "entry_path",
	}

	// FieldEntries is a query result column.
	FieldEntries = Field{
		querySessions:  "sum(sign)",
		queryPageViews: "uniq(visitor_id, session_id)",
		queryDirection: "DESC",
		name:           "entries",
	}

	// FieldExitPath is a query result column.
	FieldExitPath = Field{
		querySessions:  "exit_path",
		queryPageViews: "exit_path",
		queryDirection: "ASC",
		name:           "exit_path",
	}

	// FieldExits is a query result column.
	FieldExits = Field{
		querySessions:  "sum(sign)",
		queryPageViews: "uniq(visitor_id, session_id)",
		queryDirection: "DESC",
		name:           "exits",
	}

	// FieldVisitors is a query result column.
	FieldVisitors = Field{
		querySessions:  "uniq(visitor_id)",
		queryPageViews: "uniq(visitor_id)",
		queryPeriod:    "sum(visitors)",
		queryDirection: "DESC",
		name:           "visitors",
	}

	// FieldRelativeVisitors is a query result column.
	FieldRelativeVisitors = Field{
		querySessions:  "visitors / greatest((SELECT uniq(visitor_id) FROM session WHERE %s), 1)",
		queryPageViews: "visitors / greatest((SELECT uniq(visitor_id) FROM session WHERE %s), 1)",
		queryDirection: "DESC",
		filterTime:     true,
		name:           "relative_visitors",
	}

	// FieldCR is a query result column.
	FieldCR = Field{
		querySessions:  "visitors / greatest((SELECT uniq(visitor_id) FROM session WHERE %s), 1)",
		queryPageViews: "visitors / greatest((SELECT uniq(visitor_id) FROM session WHERE %s), 1)",
		queryDirection: "DESC",
		filterTime:     true,
		name:           "cr",
	}

	// FieldSessions is a query result column.
	FieldSessions = Field{
		querySessions:  "uniq(visitor_id, session_id)",
		queryPageViews: "uniq(visitor_id, session_id)",
		queryPeriod:    "sum(sessions)",
		queryDirection: "DESC",
		name:           "sessions",
	}

	// FieldViews is a query result column.
	FieldViews = Field{
		querySessions:  "sum(page_views*sign)",
		queryPageViews: "count(1)",
		queryPeriod:    "sum(views)",
		queryDirection: "DESC",
		name:           "views",
	}

	// FieldRelativeViews is a query result column.
	FieldRelativeViews = Field{
		querySessions:  "views / greatest((SELECT sum(page_views*sign) views FROM session WHERE %s), 1)",
		queryPageViews: "views / greatest((SELECT sum(page_views*sign) views FROM session WHERE %s), 1)",
		queryDirection: "DESC",
		filterTime:     true,
		name:           "relative_views",
	}

	// FieldBounces is a query result column.
	FieldBounces = Field{
		querySessions:  "sum(is_bounce*sign)",
		queryPageViews: "uniqIf((visitor_id, session_id), is_bounce = 1)",
		queryPeriod:    "sum(bounces)",
		queryDirection: "DESC",
		name:           "bounces",
	}

	// FieldBounceRate is a query result column.
	FieldBounceRate = Field{
		querySessions:  "bounces / IF(sessions = 0, 1, sessions)",
		queryPageViews: "bounces / IF(sessions = 0, 1, sessions)",
		queryPeriod:    "avg(bounce_rate)",
		queryDirection: "DESC",
		name:           "bounce_rate",
	}

	// FieldReferrer is a query result column.
	FieldReferrer = Field{
		querySessions:  "referrer",
		queryPageViews: "referrer",
		queryDirection: "ASC",
		name:           "referrer",
	}

	// FieldAnyReferrer is a query result column.
	FieldAnyReferrer = Field{
		querySessions:  "any(referrer)",
		queryPageViews: "any(referrer)",
		queryDirection: "ASC",
		name:           "referrer",
	}

	// FieldReferrerName is a query result column.
	FieldReferrerName = Field{
		querySessions:  "referrer_name",
		queryPageViews: "referrer_name",
		queryDirection: "ASC",
		name:           "referrer_name",
	}

	// FieldReferrerIcon is a query result column.
	FieldReferrerIcon = Field{
		querySessions:  "any(referrer_icon)",
		queryPageViews: "any(referrer_icon)",
		queryDirection: "ASC",
		name:           "referrer_icon",
	}

	// FieldLanguage is a query result column.
	FieldLanguage = Field{
		querySessions:  "language",
		queryPageViews: "language",
		queryDirection: "ASC",
		name:           "language",
	}

	// FieldCountry is a query result column.
	FieldCountry = Field{
		querySessions:  "country_code",
		queryPageViews: "country_code",
		queryDirection: "ASC",
		name:           "country_code",
	}

	// FieldCity is a query result column.
	FieldCity = Field{
		querySessions:  "city",
		queryPageViews: "city",
		queryDirection: "ASC",
		name:           "city",
	}

	// FieldBrowser is a query result column.
	FieldBrowser = Field{
		querySessions:  "browser",
		queryPageViews: "browser",
		queryDirection: "ASC",
		name:           "browser",
	}

	// FieldBrowserVersion is a query result column.
	FieldBrowserVersion = Field{
		querySessions:  "browser_version",
		queryPageViews: "browser_version",
		queryDirection: "ASC",
		name:           "browser_version",
	}

	// FieldOS is a query result column.
	FieldOS = Field{
		querySessions:  "os",
		queryPageViews: "os",
		queryDirection: "ASC",
		name:           "os",
	}

	// FieldOSVersion is a query result column.
	FieldOSVersion = Field{
		querySessions:  "os_version",
		queryPageViews: "os_version",
		queryDirection: "ASC",
		name:           "os_version",
	}

	// FieldScreenClass is a query result column.
	FieldScreenClass = Field{
		querySessions:  "screen_class",
		queryPageViews: "screen_class",
		queryDirection: "ASC",
		name:           "screen_class",
	}

	// FieldUTMSource is a query result column.
	FieldUTMSource = Field{
		querySessions:  "utm_source",
		queryPageViews: "utm_source",
		queryDirection: "ASC",
		name:           "utm_source",
	}

	// FieldUTMMedium is a query result column.
	FieldUTMMedium = Field{
		querySessions:  "utm_medium",
		queryPageViews: "utm_medium",
		queryDirection: "ASC",
		name:           "utm_medium",
	}

	// FieldUTMCampaign is a query result column.
	FieldUTMCampaign = Field{
		querySessions:  "utm_campaign",
		queryPageViews: "utm_campaign",
		queryDirection: "ASC",
		name:           "utm_campaign",
	}

	// FieldUTMContent is a query result column.
	FieldUTMContent = Field{
		querySessions:  "utm_content",
		queryPageViews: "utm_content",
		queryDirection: "ASC",
		name:           "utm_content",
	}

	// FieldUTMTerm is a query result column.
	FieldUTMTerm = Field{
		querySessions:  "utm_term",
		queryPageViews: "utm_term",
		queryDirection: "ASC",
		name:           "utm_term",
	}

	// FieldTitle is a query result column.
	FieldTitle = Field{
		querySessions:  "title",
		queryPageViews: "title",
		queryDirection: "ASC",
		name:           "title",
	}

	// FieldEntryTitle is a query result column.
	FieldEntryTitle = Field{
		querySessions:  "entry_title",
		queryPageViews: "entry_title",
		queryDirection: "ASC",
		name:           "title",
	}

	// FieldExitTitle is a query result column.
	FieldExitTitle = Field{
		querySessions:  "exit_title",
		queryPageViews: "exit_title",
		queryDirection: "ASC",
		name:           "title",
	}

	// FieldDay is a query result column.
	FieldDay = Field{
		querySessions:  "toDate(time, '%s')",
		queryPageViews: "toDate(time, '%s')",
		queryDirection: "ASC",
		withFill:       true,
		timezone:       true,
		name:           "day",
	}

	// FieldHour is a query result column.
	FieldHour = Field{
		querySessions:  "toHour(time, '%s')",
		queryPageViews: "toHour(time, '%s')",
		queryDirection: "ASC",
		queryWithFill:  "WITH FILL FROM 0 TO 24",
		timezone:       true,
		name:           "hour",
	}

	// FieldEventName is a query result column.
	FieldEventName = Field{
		querySessions:  "event_name",
		queryPageViews: "event_name",
		name:           "event_name",
		queryDirection: "ASC",
	}

	// FieldEventMeta is a query result column.
	FieldEventMeta = Field{
		// TODO optimize once maps are supported in the driver (v2)
		/*querySessions:  "cast((event_meta_keys, event_meta_values), 'Map(String, String)')",
		queryPageViews: "cast((event_meta_keys, event_meta_values), 'Map(String, String)')",*/
		querySessions:  "arrayZip(event_meta_keys, event_meta_values)",
		queryPageViews: "arrayZip(event_meta_keys, event_meta_values)",
		name:           "meta",
	}

	// FieldEventMetaKeys is a query result column.
	FieldEventMetaKeys = Field{
		querySessions:  "groupUniqArrayArray(event_meta_keys)",
		queryPageViews: "groupUniqArrayArray(event_meta_keys)",
		name:           "meta_keys",
	}

	// FieldEventMetaValues is a query result column.
	FieldEventMetaValues = Field{
		querySessions:  "event_meta_values[indexOf(event_meta_keys, ?)]",
		queryPageViews: "event_meta_values[indexOf(event_meta_keys, ?)]",
		name:           "meta_value",
	}

	// FieldEventTimeSpent is a query result column.
	FieldEventTimeSpent = Field{
		querySessions:  "ifNull(toUInt64(avg(nullIf(duration_seconds, 0))), 0)",
		queryPageViews: "ifNull(toUInt64(avg(nullIf(duration_seconds, 0))), 0)",
		name:           "average_time_spent_seconds",
	}
)

// Field is a column for a query.
type Field struct {
	querySessions  string
	queryPageViews string
	queryPeriod    string
	queryDirection string
	queryWithFill  string
	withFill       bool
	timezone       bool
	filterTime     bool
	name           string
}
