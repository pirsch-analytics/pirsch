package dimensions

// Dimension is a field results are grouped by.
type Dimension interface {
	// Table returns the database table for the Dimension.
	Table() string

	// Column returns the database column name for the Dimension.
	Column() string

	// Expression returns the SQL expression for aggregation.
	// If empty, the Column name will be used instead.
	Expression() string
}

// TODO
/*
	>>> Periods (hour, day, ...)
	>>> Search?

	// CustomMetricKey is used to calculate the average and total for an event metadata field.
	// This must be used together with EventName and CustomMetricType.
	CustomMetricKey string

	// CustomMetricType is used to calculate the average and total for an event metadata field.
	CustomMetricType pkg.CustomMetricType

	// VisitorID filters for a visitor.
	// Must be used together with SessionID.
	VisitorID uint64

	// SessionID filters for a session.
	// Must be used together with VisitorID.
	SessionID uint32

	// Hostname filters for the hostname.
	Hostname []string

	// Path filters for the path.
	Path []string

	// EntryPath filters for the entry page.
	EntryPath []string

	// ExitPath filters for the exit page.
	ExitPath []string

	// Language filters for the ISO language code.
	Language []string

	// Country filters for the ISO country code.
	Country []string

	// Region filters for the region.
	Region []string

	// City filters for the city name.
	City []string

	// Referrer filters for the full referrer.
	Referrer []string

	// ReferrerName filters for the referrer name.
	ReferrerName []string

	// Channel filters for the channel query parameter.
	Channel []string

	// OS filters for the operating system.
	OS []string

	// OSVersion filters for the operating system version.
	OSVersion []string

	// Browser filters for the browser.
	Browser []string

	// BrowserVersion filters for the browser version.
	BrowserVersion []string

	// Platform filters for the platform (desktop, mobile, unknown).
	Platform string

	// ScreenClass filters for the screen class.
	ScreenClass []string

	// UTMSource filters for the utm_source query parameter.
	UTMSource []string

	// UTMMedium filters for the utm_medium query parameter.
	UTMMedium []string

	// UTMCampaign filters for the utm_campaign query parameter.
	UTMCampaign []string

	// UTMContent filters for the utm_content query parameter.
	UTMContent []string

	// UTMTerm filters for the utm_term query parameter.
	UTMTerm []string

	// Tags filters for tag key-value pairs.
	Tags map[string]string

	// Tag filters for tags by their keys.
	Tag []string

	// EventName filters for events by their name.
	EventName []string

	// EventMetaKey filters for an event meta-key.
	// This must be used together with an EventName.
	EventMetaKey []string

	// EventMeta filters for event metadata.
	EventMeta map[string]string
*/
