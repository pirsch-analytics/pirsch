package reporting

// Pagination limits the number of results for a Report.
type Pagination struct {
	// Offset limits the number of results. Offset <= 0 means no offset.
	Offset int

	// Limit limits the number of results. Limit <= 0 means unlimited.
	Limit int
}
