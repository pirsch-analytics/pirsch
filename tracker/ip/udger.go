package ip

import (
	"bytes"
	"net"
	"sync"
)

type ipRange struct {
	from net.IP
	to   net.IP
}

// Udger implements the Filter interface.
type Udger struct {
	ipsV4, ipsV6       map[string]struct{}
	rangesV4, rangesV6 []ipRange
	m                  sync.RWMutex
}

// NewUdger creates a new Filter using the IP lists provided by udger.com.
func NewUdger() *Udger {
	return &Udger{}
}

// Update implements the Filter interface.
func (udger *Udger) Update(ipsV4, ipsV6 []string, rangesV4, rangesV6 []Range) {
	ipMapV4 := make(map[string]struct{}, len(ipsV4))

	for _, ip := range ipsV4 {
		ipMapV4[ip] = struct{}{}
	}

	ipMapV6 := make(map[string]struct{}, len(ipsV6))

	for _, ip := range ipsV6 {
		ipMapV6[ip] = struct{}{}
	}

	ipRangesV4 := make([]ipRange, 0, len(rangesV4))

	for _, r := range rangesV4 {
		fromIP := net.ParseIP(r.From)
		toIP := net.ParseIP(r.To)

		if fromIP != nil && toIP != nil {
			ipRangesV4 = append(ipRangesV4, ipRange{
				from: net.ParseIP(r.From),
				to:   net.ParseIP(r.To),
			})
		}
	}

	ipRangesV6 := make([]ipRange, 0, len(rangesV6))

	for _, r := range rangesV6 {
		fromIP := net.ParseIP(r.From)
		toIP := net.ParseIP(r.To)

		if fromIP != nil && toIP != nil {
			ipRangesV6 = append(ipRangesV6, ipRange{
				from: net.ParseIP(r.From),
				to:   net.ParseIP(r.To),
			})
		}
	}

	udger.m.Lock()
	defer udger.m.Unlock()
	udger.ipsV4 = ipMapV4
	udger.ipsV6 = ipMapV6
	udger.rangesV4 = ipRangesV4
	udger.rangesV6 = ipRangesV6
}

// Ignore implements the Filter interface.
func (udger *Udger) Ignore(ip string) bool {
	parsedIP := net.ParseIP(ip)

	if parsedIP == nil {
		return true
	}

	udger.m.RLock()
	defer udger.m.RUnlock()

	if parsedIP.To4() != nil {
		return udger.findIP(ip, parsedIP, udger.ipsV4, udger.rangesV4)
	}

	return udger.findIP(ip, parsedIP, udger.ipsV6, udger.rangesV6)
}

func (udger *Udger) findIP(ip string, parsedIP net.IP, ips map[string]struct{}, ranges []ipRange) bool {
	if _, ok := ips[ip]; ok {
		return true
	}

	for _, r := range ranges {
		if bytes.Compare(parsedIP, r.from) >= 0 && bytes.Compare(parsedIP, r.to) <= 0 {
			return true
		}
	}

	return false
}
