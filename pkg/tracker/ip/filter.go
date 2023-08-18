package ip

// Range is an IP address range.
type Range struct {
	From string
	To   string
}

// Filter filters requests using the IP address.
type Filter interface {
	// Update updates the filter list for given plain IP addresses and IP address ranges (v4 and v6).
	Update([]string, []string, []Range, []Range)

	// Ignore reports whether an IP address should be ignored.
	Ignore(string) bool
}
