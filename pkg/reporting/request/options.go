package request

// Options are optional fields for a Request.
type Options struct {
	// Sample sets the sampling size.
	Sample uint
}

// TODO
/*
	>>> Imported data?

	// IncludeTime sets whether the selected period should contain the time (hour, minute, second).
	IncludeTime bool

	// IncludeTitle indicates that the Analyzer.ByPath, Analyzer.Entry, and Analyzer.Exit should contain the page title.
	IncludeTitle bool

	// IncludeTimeOnPage indicates that the Analyzer.ByPath and Analyzer.Entry should contain the average time on the page.
	IncludeTimeOnPage bool

	// IncludeCR indicates that Analyzer.Total and Analyzer.ByPeriod should contain the conversion rate.
	IncludeCR bool

	// MaxTimeOnPageSeconds is an optional maximum for the time spent on a page.
	// Visitors who are idle artificially increase the average time spent on a page; this option can be used to limit the effect.
	// Set to 0 to disable this option (default).
	MaxTimeOnPageSeconds int

	// Sample sets the (optional) sampling size.
	Sample uint

	Case insensitive
*/
