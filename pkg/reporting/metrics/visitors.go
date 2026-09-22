package metrics

import "github.com/pirsch-analytics/pirsch/v7/pkg"

// Visitors is a Metric.
type Visitors struct{}

// Table implements the Metric interface.
func (m Visitors) Table() []string {
	return []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
}

// TableImported implements the Metric interface.
func (m Visitors) TableImported() []string {
	return []string{
		pkg.TableImportedBrowser,
		pkg.TableImportedCity,
		pkg.TableImportedCountry,
		pkg.TableImportedDevice,
		pkg.TableImportedEntryPage,
		pkg.TableImportedExitPage,
		pkg.TableImportedLanguage,
		pkg.TableImportedOS,
		pkg.TableImportedPage,
		pkg.TableImportedReferrer,
		pkg.TableImportedRegion,
		pkg.TableImportedUTMCampaign,
		pkg.TableImportedUTMMedium,
		pkg.TableImportedUTMSource,
		pkg.TableImportedVisitors,
	}
}

// JoinTable implements the Metric interface.
func (m Visitors) JoinTable() string {
	return ""
}

// Column implements the Metric interface.
func (m Visitors) Column() string {
	return "visitors"
}

// ColumnImported implements the Metric interface.
func (m Visitors) ColumnImported() string {
	return "visitors"
}

// Expression implements the Metric interface.
func (m Visitors) Expression(_ string) (string, bool) {
	return "uniq(visitor_id)", false
}

// ExpressionImported implements the Metric interface.
func (m Visitors) ExpressionImported() string {
	return "sum(visitors)"
}

// ScanType implements the Metric interface.
func (m Visitors) ScanType() any {
	return new(uint64)
}

// Zero implements the Metric interface.
func (m Visitors) Zero() any {
	return uint64(0)
}

// String implements the Metric interface.
func (m Visitors) String() string {
	return "visitors"
}
