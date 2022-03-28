package pirsch

var (
	fieldCount = field{
		querySessions:  "count(*)",
		queryPageViews: "count(*)",
		name:           "count",
		queryDirection: "DESC",
	}
	fieldPath = field{
		querySessions:  "path",
		queryPageViews: "path",
		queryDirection: "ASC",
		name:           "path",
	}
	fieldEntryPath = field{
		querySessions:  "entry_path",
		queryPageViews: "entry_path",
		queryDirection: "ASC",
		name:           "entry_path",
	}
	fieldEntries = field{
		querySessions:  "sum(sign)",
		queryPageViews: "uniq(visitor_id, session_id)",
		queryDirection: "DESC",
		name:           "entries",
	}
	fieldExitPath = field{
		querySessions:  "exit_path",
		queryPageViews: "exit_path",
		queryDirection: "ASC",
		name:           "exit_path",
	}
	fieldExits = field{
		querySessions:  "sum(sign)",
		queryPageViews: "uniq(visitor_id, session_id)",
		queryDirection: "DESC",
		name:           "exits",
	}
	fieldVisitors = field{
		querySessions:  "uniq(visitor_id)",
		queryPageViews: "uniq(visitor_id)",
		queryDirection: "DESC",
		name:           "visitors",
	}
	fieldRelativeVisitors = field{
		querySessions:  "visitors / greatest((SELECT uniq(visitor_id) FROM session WHERE %s), 1)",
		queryPageViews: "visitors / greatest((SELECT uniq(visitor_id) FROM session WHERE %s), 1)",
		queryDirection: "DESC",
		filterTime:     true,
		name:           "relative_visitors",
	}
	fieldCR = field{
		querySessions:  "visitors / greatest((SELECT uniq(visitor_id) FROM session WHERE %s), 1)",
		queryPageViews: "visitors / greatest((SELECT uniq(visitor_id) FROM session WHERE %s), 1)",
		queryDirection: "DESC",
		filterTime:     true,
		name:           "cr",
	}
	fieldSessions = field{
		querySessions:  "uniq(visitor_id, session_id)",
		queryPageViews: "uniq(visitor_id, session_id)",
		queryDirection: "DESC",
		name:           "sessions",
	}
	fieldViews = field{
		querySessions:  "sum(page_views*sign)",
		queryPageViews: "count(1)",
		queryDirection: "DESC",
		name:           "views",
	}
	fieldRelativeViews = field{
		querySessions:  "views / greatest((SELECT sum(page_views*sign) views FROM session WHERE %s), 1)",
		queryPageViews: "views / greatest((SELECT sum(page_views*sign) views FROM session WHERE %s), 1)",
		queryDirection: "DESC",
		filterTime:     true,
		name:           "relative_views",
	}
	fieldBounces = field{
		querySessions:  "sum(is_bounce*sign)",
		queryPageViews: "uniqIf((visitor_id, session_id), is_bounce = 1)",
		queryDirection: "DESC",
		name:           "bounces",
	}
	fieldBounceRate = field{
		querySessions:  "bounces / IF(sessions = 0, 1, sessions)",
		queryPageViews: "bounces / IF(sessions = 0, 1, sessions)",
		queryDirection: "DESC",
		name:           "bounce_rate",
	}
	fieldReferrer = field{
		querySessions:  "referrer",
		queryPageViews: "referrer",
		queryDirection: "ASC",
		name:           "referrer",
	}
	fieldAnyReferrer = field{
		querySessions:  "any(referrer)",
		queryPageViews: "any(referrer)",
		queryDirection: "ASC",
		name:           "referrer",
	}
	fieldReferrerName = field{
		querySessions:  "referrer_name",
		queryPageViews: "referrer_name",
		queryDirection: "ASC",
		name:           "referrer_name",
	}
	fieldReferrerIcon = field{
		querySessions:  "any(referrer_icon)",
		queryPageViews: "any(referrer_icon)",
		queryDirection: "ASC",
		name:           "referrer_icon",
	}
	fieldLanguage = field{
		querySessions:  "language",
		queryPageViews: "language",
		queryDirection: "ASC",
		name:           "language",
	}
	fieldCountry = field{
		querySessions:  "country_code",
		queryPageViews: "country_code",
		queryDirection: "ASC",
		name:           "country_code",
	}
	fieldCity = field{
		querySessions:  "city",
		queryPageViews: "city",
		queryDirection: "ASC",
		name:           "city",
	}
	fieldBrowser = field{
		querySessions:  "browser",
		queryPageViews: "browser",
		queryDirection: "ASC",
		name:           "browser",
	}
	fieldBrowserVersion = field{
		querySessions:  "browser_version",
		queryPageViews: "browser_version",
		queryDirection: "ASC",
		name:           "browser_version",
	}
	fieldOS = field{
		querySessions:  "os",
		queryPageViews: "os",
		queryDirection: "ASC",
		name:           "os",
	}
	fieldOSVersion = field{
		querySessions:  "os_version",
		queryPageViews: "os_version",
		queryDirection: "ASC",
		name:           "os_version",
	}
	fieldScreenClass = field{
		querySessions:  "screen_class",
		queryPageViews: "screen_class",
		queryDirection: "ASC",
		name:           "screen_class",
	}
	fieldUTMSource = field{
		querySessions:  "utm_source",
		queryPageViews: "utm_source",
		queryDirection: "ASC",
		name:           "utm_source",
	}
	fieldUTMMedium = field{
		querySessions:  "utm_medium",
		queryPageViews: "utm_medium",
		queryDirection: "ASC",
		name:           "utm_medium",
	}
	fieldUTMCampaign = field{
		querySessions:  "utm_campaign",
		queryPageViews: "utm_campaign",
		queryDirection: "ASC",
		name:           "utm_campaign",
	}
	fieldUTMContent = field{
		querySessions:  "utm_content",
		queryPageViews: "utm_content",
		queryDirection: "ASC",
		name:           "utm_content",
	}
	fieldUTMTerm = field{
		querySessions:  "utm_term",
		queryPageViews: "utm_term",
		queryDirection: "ASC",
		name:           "utm_term",
	}
	fieldTitle = field{
		querySessions:  "title",
		queryPageViews: "title",
		queryDirection: "ASC",
		name:           "title",
	}
	fieldEntryTitle = field{
		querySessions:  "entry_title",
		queryPageViews: "entry_title",
		queryDirection: "ASC",
		name:           "title",
	}
	fieldExitTitle = field{
		querySessions:  "exit_title",
		queryPageViews: "exit_title",
		queryDirection: "ASC",
		name:           "title",
	}
	fieldDay = field{
		querySessions:  "toDate(time, '%s')",
		queryPageViews: "toDate(time, '%s')",
		queryDirection: "ASC",
		withFill:       true,
		timezone:       true,
		name:           "day",
	}
	fieldHour = field{
		querySessions:  "toHour(time, '%s')",
		queryPageViews: "toHour(time, '%s')",
		queryDirection: "ASC",
		queryWithFill:  "WITH FILL FROM 0 TO 24",
		timezone:       true,
		name:           "hour",
	}
	fieldEventName = field{
		querySessions:  "event_name",
		queryPageViews: "event_name",
		name:           "event_name",
		queryDirection: "ASC",
	}
	fieldEventMeta = field{
		// TODO optimize once maps are supported in the driver (v2)
		/*querySessions:  "cast((event_meta_keys, event_meta_values), 'Map(String, String)')",
		queryPageViews: "cast((event_meta_keys, event_meta_values), 'Map(String, String)')",*/
		querySessions:  "arrayZip(event_meta_keys, event_meta_values)",
		queryPageViews: "arrayZip(event_meta_keys, event_meta_values)",
		name:           "meta",
	}
	fieldEventMetaKeys = field{
		querySessions:  "groupUniqArrayArray(event_meta_keys)",
		queryPageViews: "groupUniqArrayArray(event_meta_keys)",
		name:           "meta_keys",
	}
	fieldEventMetaValues = field{
		querySessions:  "event_meta_values[indexOf(event_meta_keys, ?)]",
		queryPageViews: "event_meta_values[indexOf(event_meta_keys, ?)]",
		name:           "meta_value",
	}
	fieldEventTimeSpent = field{
		querySessions:  "ifNull(toUInt64(avg(nullIf(duration_seconds, 0))), 0)",
		queryPageViews: "ifNull(toUInt64(avg(nullIf(duration_seconds, 0))), 0)",
		name:           "average_time_spent_seconds",
	}
)

type field struct {
	querySessions  string
	queryPageViews string
	queryDirection string
	queryWithFill  string
	withFill       bool
	timezone       bool
	filterTime     bool
	name           string
}

func fieldsContain(haystack []field, needle string) bool {
	for i := range haystack {
		if haystack[i].name == needle {
			return true
		}
	}

	return false
}
