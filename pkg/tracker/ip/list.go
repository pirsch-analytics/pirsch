package ip

import (
	"net"
	"sync"
)

// List implements the Filter interface.
// It uses a simple flat IP address list to filter requests.
type List struct {
	blacklist, whitelist map[string]struct{}
	m                    sync.RWMutex
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

	l.m.Lock()
	defer l.m.Unlock()
	l.blacklist = blacklist
	l.whitelist = whitelist
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
	return ignore
}
