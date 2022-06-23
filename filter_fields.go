package pirsch

var (
	// FieldCount is a query result column.
	FieldCount = Field{
		querySessions:  "count(*)",
		queryPageViews: "count(*)",
		Name:           "count",
		queryDirection: "DESC",
	}

	// FieldPath is a query result column.
	FieldPath = Field{
		querySessions:  "path",
		queryPageViews: "path",
		queryDirection: "ASC",
		Name:           "path",
	}

	// FieldEntryPath is a query result column.
	FieldEntryPath = Field{
		querySessions:  "entry_path",
		queryPageViews: "entry_path",
		queryDirection: "ASC",
		Name:           "entry_path",
	}

	// FieldEntries is a query result column.
	FieldEntries = Field{
		querySessions:  "sum(sign)",
		queryPageViews: "uniq(visitor_id, session_id)",
		queryDirection: "DESC",
		Name:           "entries",
	}

	// FieldExitPath is a query result column.
	FieldExitPath = Field{
		querySessions:  "exit_path",
		queryPageViews: "exit_path",
		queryDirection: "ASC",
		Name:           "exit_path",
	}

	// FieldExits is a query result column.
	FieldExits = Field{
		querySessions:  "sum(sign)",
		queryPageViews: "uniq(visitor_id, session_id)",
		queryDirection: "DESC",
		Name:           "exits",
	}

	// FieldVisitors is a query result column.
	FieldVisitors = Field{
		querySessions:  "uniq(visitor_id)",
		queryPageViews: "uniq(visitor_id)",
		queryPeriod:    "sum(visitors)",
		queryDirection: "DESC",
		Name:           "visitors",
	}

	// FieldRelativeVisitors is a query result column.
	FieldRelativeVisitors = Field{
		querySessions:  "toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id) FROM session WHERE %s), 1))",
		queryPageViews: "toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id) FROM session WHERE %s), 1))",
		queryDirection: "DESC",
		filterTime:     true,
		Name:           "relative_visitors",
	}

	// FieldCR is a query result column.
	FieldCR = Field{
		querySessions:  "toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id) FROM session WHERE %s), 1))",
		queryPageViews: "toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id) FROM session WHERE %s), 1))",
		queryDirection: "DESC",
		filterTime:     true,
		Name:           "cr",
	}

	// FieldSessions is a query result column.
	FieldSessions = Field{
		querySessions:  "uniq(visitor_id, session_id)",
		queryPageViews: "uniq(visitor_id, session_id)",
		queryPeriod:    "sum(sessions)",
		queryDirection: "DESC",
		Name:           "sessions",
	}

	// FieldViews is a query result column.
	FieldViews = Field{
		querySessions:  "sum(page_views*sign)",
		queryPageViews: "count(1)",
		queryPeriod:    "sum(views)",
		queryDirection: "DESC",
		Name:           "views",
	}

	// FieldRelativeViews is a query result column.
	FieldRelativeViews = Field{
		querySessions:  "toFloat64OrDefault(views / greatest((SELECT sum(page_views*sign) views FROM session WHERE %s), 1))",
		queryPageViews: "toFloat64OrDefault(views / greatest((SELECT sum(page_views*sign) views FROM session WHERE %s), 1))",
		queryDirection: "DESC",
		filterTime:     true,
		Name:           "relative_views",
	}

	// FieldBounces is a query result column.
	FieldBounces = Field{
		querySessions:  "sum(is_bounce*sign)",
		queryPageViews: "uniqIf((visitor_id, session_id), is_bounce = 1)",
		queryPeriod:    "sum(bounces)",
		queryDirection: "DESC",
		Name:           "bounces",
	}

	// FieldBounceRate is a query result column.
	FieldBounceRate = Field{
		querySessions:  "bounces / IF(sessions = 0, 1, sessions)",
		queryPageViews: "bounces / IF(sessions = 0, 1, sessions)",
		queryPeriod:    "avg(bounce_rate)",
		queryDirection: "DESC",
		Name:           "bounce_rate",
	}

	// FieldReferrer is a query result column.
	FieldReferrer = Field{
		querySessions:  "referrer",
		queryPageViews: "referrer",
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
		queryDirection: "ASC",
		Name:           "language",
	}

	// FieldCountry is a query result column.
	FieldCountry = Field{
		querySessions:  "country_code",
		queryPageViews: "country_code",
		queryDirection: "ASC",
		Name:           "country_code",
	}

	// FieldCity is a query result column.
	FieldCity = Field{
		querySessions:  "city",
		queryPageViews: "city",
		queryDirection: "ASC",
		Name:           "city",
	}

	// FieldBrowser is a query result column.
	FieldBrowser = Field{
		querySessions:  "browser",
		queryPageViews: "browser",
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
		queryDirection: "ASC",
		Name:           "utm_source",
	}

	// FieldUTMMedium is a query result column.
	FieldUTMMedium = Field{
		querySessions:  "utm_medium",
		queryPageViews: "utm_medium",
		queryDirection: "ASC",
		Name:           "utm_medium",
	}

	// FieldUTMCampaign is a query result column.
	FieldUTMCampaign = Field{
		querySessions:  "utm_campaign",
		queryPageViews: "utm_campaign",
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

	// FieldTitle is a query result column.
	FieldTitle = Field{
		querySessions:  "title",
		queryPageViews: "title",
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

	// FieldEventName is a query result column.
	FieldEventName = Field{
		querySessions:  "event_name",
		queryPageViews: "event_name",
		Name:           "event_name",
		queryDirection: "ASC",
	}

	// FieldEventMeta is a query result column.
	FieldEventMeta = Field{
		// TODO optimize once maps are supported in the driver (v2)
		/*querySessions:  "cast((event_meta_keys, event_meta_values), 'Map(String, String)')",
		queryPageViews: "cast((event_meta_keys, event_meta_values), 'Map(String, String)')",*/
		querySessions:  "arrayZip(event_meta_keys, event_meta_values)",
		queryPageViews: "arrayZip(event_meta_keys, event_meta_values)",
		Name:           "meta",
	}

	// FieldEventMetaKeys is a query result column.
	FieldEventMetaKeys = Field{
		querySessions:  "groupUniqArrayArray(event_meta_keys)",
		queryPageViews: "groupUniqArrayArray(event_meta_keys)",
		Name:           "meta_keys",
	}

	// FieldEventMetaValues is a query result column.
	FieldEventMetaValues = Field{
		querySessions:  "event_meta_values[indexOf(event_meta_keys, ?)]",
		queryPageViews: "event_meta_values[indexOf(event_meta_keys, ?)]",
		Name:           "meta_value",
	}

	// FieldEventTimeSpent is a query result column.
	FieldEventTimeSpent = Field{
		querySessions:  "ifNull(toUInt64(avg(nullIf(duration_seconds, 0))), 0)",
		queryPageViews: "ifNull(toUInt64(avg(nullIf(duration_seconds, 0))), 0)",
		Name:           "average_time_spent_seconds",
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
	Name           string
}
