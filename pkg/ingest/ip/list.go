package ip

import (
	"bytes"
	"net"
	"sync"
)

// List implements the Filter interface.
// It filters IP addresses directly and by checking them against a list of networks.
type List struct {
	blacklist, whitelist                   map[string]struct{}
	networkV4Blacklist, networkV6Blacklist []ipRange
	m                                      sync.RWMutex
}

// NewList creates a new IP address filter list.
func NewList() *List {
	return new(List)
}

// Update implements the Filter interface.
// The slices will be joined. Ranges are ignored.
func (l *List) Update(ipsV4, ipsV6, whitelistIpV4, whitelistIpV6 []string, rangesV4, rangesV6 []Range) {
	blacklist := make(map[string]struct{}, len(ipsV4)+len(ipsV6))
	whitelist := make(map[string]struct{}, len(whitelistIpV4)+len(whitelistIpV6))

	for _, ip := range ipsV4 {
		blacklist[ip] = struct{}{}
	}

	for _, ip := range ipsV6 {
		blacklist[ip] = struct{}{}
	}

	for _, ip := range whitelistIpV4 {
		whitelist[ip] = struct{}{}
	}

	for _, ip := range whitelistIpV6 {
		whitelist[ip] = struct{}{}
	}

	networkV4Blacklist := l.ipRangesToList(rangesV4)
	networkV6Blacklist := l.ipRangesToList(rangesV6)
	l.m.Lock()
	defer l.m.Unlock()
	l.blacklist = blacklist
	l.whitelist = whitelist
	l.networkV4Blacklist = networkV4Blacklist
	l.networkV6Blacklist = networkV6Blacklist
}

// Ignore implements the Filter interface.
func (l *List) Ignore(ip string) bool {
	parsedIP := net.ParseIP(ip)

	if parsedIP == nil {
		return true
	}

	l.m.RLock()
	defer l.m.RUnlock()

	if _, ok := l.whitelist[ip]; ok {
		return false
	}

	_, ignore := l.blacklist[ip]

	if ignore {
		return true
	}

	if parsedIP.To4() != nil {
		return l.ignoreNetwork(parsedIP, l.networkV4Blacklist)
	}

	return l.ignoreNetwork(parsedIP, l.networkV6Blacklist)
}

func (l *List) ipRangesToList(ranges []Range) []ipRange {
	m := make([]ipRange, 0, len(ranges))

	for _, r := range ranges {
		fromIP := net.ParseIP(r.From)
		toIP := net.ParseIP(r.To)

		if fromIP != nil && toIP != nil {
			m = append(m, ipRange{
				from: net.ParseIP(r.From),
				to:   net.ParseIP(r.To),
			})
		}
	}

	return m
}

func (l *List) ignoreNetwork(parsedIP net.IP, ranges []ipRange) bool {
	for _, r := range ranges {
		if bytes.Compare(parsedIP, r.from) >= 0 && bytes.Compare(parsedIP, r.to) <= 0 {
			return true
		}
	}

	return false
}
