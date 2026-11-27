package ip

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"sort"
	"strings"
	"sync"
)

const (
	lordAlfredIPRanges = "https://raw.githubusercontent.com/lord-alfred/ipranges/main/"
	rezmossIPRanges    = "https://raw.githubusercontent.com/rezmoss/cloud-provider-ip-addresses/main/"
)

// RangeSource is a downloadable list of IP addresses and networks.
type RangeSource struct {
	// URL is the URL of the list.
	// Each line must start with an IP address or network in CIDR notation, optionally quoted and followed by a comma (CSV).
	// Everything after it is ignored, as are invalid lines.
	URL string

	// Exclude is a list of case-insensitive keywords. Lines containing one of them are skipped.
	Exclude []string
}

// DefaultRangeSources are the default lists of data center, cloud provider, and crawler IP ranges.
// Lists of providers that are also used by regular visitors (like Cloudflare WARP, Apple Private Relay, VPNs, Tor, CDNs, or corporate proxies) are left out on purpose.
var DefaultRangeSources = defaultRangeSources()

func defaultRangeSources() []RangeSource {
	sources := []RangeSource{
		{
			URL:     "https://raw.githubusercontent.com/jhassine/server-ip-addresses/master/data/datacenters.csv",
			Exclude: []string{"cloudflare"},
		},
	}

	for _, provider := range []string{"amazon", "google", "googlebot", "digitalocean", "linode", "vultr", "github", "facebook", "pingdom"} {
		sources = append(sources,
			RangeSource{URL: lordAlfredIPRanges + provider + "/ipv4_merged.txt"},
			RangeSource{URL: lordAlfredIPRanges + provider + "/ipv6_merged.txt"})
	}

	for _, provider := range []string{"bing", "oracle", "openai", "perplexity", "duckduckbot", "duckassistbot", "statuscake"} {
		sources = append(sources, RangeSource{URL: lordAlfredIPRanges + provider + "/ipv4_merged.txt"})
	}

	for _, provider := range []string{
		"googlecloud", "azure", "amazonbot", "applebot", "baidu", "claudebot", "commoncrawl", "gptbot", "perplexitybot",
		"bingbot", "internetarchive", "uptimerobot", "hetzner", "ovhcloud", "leaseweb", "contabo", "huawei", "ibmcloud",
		"tencent", "alibaba", "rackspace", "scaleway", "upcloud",
	} {
		sources = append(sources, RangeSource{URL: rezmossIPRanges + provider + "/" + provider + "_ips_merged.txt"})
	}

	return sources
}

type addrRange struct {
	from, to netip.Addr
}

// RangeList implements the Filter interface.
// It filters IP addresses using downloadable lists of IP ranges, like those of data centers, cloud providers, and crawlers.
type RangeList struct {
	sources       []RangeSource
	sourceRanges  map[string][]addrRange
	rangesV4      []addrRange
	rangesV6      []addrRange
	downloadMutex sync.Mutex
	m             sync.RWMutex
}

// NewRangeList creates a new RangeList for given sources. If no sources are passed, DefaultRangeSources will be used.
func NewRangeList(sources []RangeSource) *RangeList {
	if len(sources) == 0 {
		sources = DefaultRangeSources
	}

	return &RangeList{
		sources:      sources,
		sourceRanges: make(map[string][]addrRange),
	}
}

// Update implements the Filter interface.
// The RangeList filter uses downloaded lists only, so this is a no-op. Use DownloadAndUpdate to update the lists.
func (l *RangeList) Update([]string, []string, []string, []string, []Range, []Range) {}

// Ignore implements the Filter interface.
func (l *RangeList) Ignore(ip string) bool {
	addr, err := netip.ParseAddr(ip)

	if err != nil {
		return true
	}

	addr = addr.Unmap()
	l.m.RLock()
	defer l.m.RUnlock()

	if addr.Is4() {
		return findAddrRange(l.rangesV4, addr)
	}

	return findAddrRange(l.rangesV6, addr)
}

// DownloadAndUpdate downloads all lists and updates the filter.
// If a list cannot be downloaded or is empty, the previously downloaded ranges for it are kept and an error is returned
// after all other lists have been updated.
func (l *RangeList) DownloadAndUpdate() error {
	l.downloadMutex.Lock()
	defer l.downloadMutex.Unlock()
	var errs []error

	for _, source := range l.sources {
		ranges, err := l.download(source)

		if err != nil {
			errs = append(errs, fmt.Errorf("error downloading IP ranges from %s: %w", source.URL, err))
			continue
		}

		l.sourceRanges[source.URL] = ranges
	}

	var all []addrRange

	for _, ranges := range l.sourceRanges {
		all = append(all, ranges...)
	}

	v4, v6 := mergeAddrRanges(all)
	l.m.Lock()
	l.rangesV4 = v4
	l.rangesV6 = v6
	l.m.Unlock()
	return errors.Join(errs...)
}

func (l *RangeList) download(source RangeSource) ([]addrRange, error) {
	resp, err := http.Get(source.URL)

	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if err := checkDownloadResponse(resp); err != nil {
		return nil, err
	}

	ranges, err := parseRanges(resp.Body, source.Exclude)

	if err != nil {
		return nil, err
	}

	if len(ranges) == 0 {
		return nil, errors.New("list is empty")
	}

	return ranges, nil
}

func parseRanges(r io.Reader, exclude []string) ([]addrRange, error) {
	excludeLower := make([]string, 0, len(exclude))

	for _, keyword := range exclude {
		excludeLower = append(excludeLower, strings.ToLower(keyword))
	}

	exclude = excludeLower
	var ranges []addrRange
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		if len(exclude) > 0 {
			lower := strings.ToLower(line)
			skip := false

			for _, keyword := range exclude {
				if strings.Contains(lower, keyword) {
					skip = true
					break
				}
			}

			if skip {
				continue
			}
		}

		if r, ok := parseRangeLine(line); ok {
			ranges = append(ranges, r)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ranges, nil
}

func parseRangeLine(line string) (addrRange, bool) {
	line = strings.TrimSpace(line)

	if i := strings.IndexAny(line, ", \t#"); i != -1 {
		line = line[:i]
	}

	line = strings.Trim(line, "\"'")

	if line == "" {
		return addrRange{}, false
	}

	if strings.Contains(line, "/") {
		prefix, err := netip.ParsePrefix(line)

		if err != nil {
			return addrRange{}, false
		}

		prefix = prefix.Masked()
		from := prefix.Addr().Unmap()
		bits := prefix.Bits()

		if prefix.Addr().Is4In6() {
			bits -= 96

			if bits < 0 {
				return addrRange{}, false
			}
		}

		return addrRange{from: from, to: lastAddr(from, bits)}, true
	}

	addr, err := netip.ParseAddr(line)

	if err != nil {
		return addrRange{}, false
	}

	addr = addr.Unmap()
	return addrRange{from: addr, to: addr}, true
}

func lastAddr(addr netip.Addr, bits int) netip.Addr {
	b := addr.AsSlice()

	for i := range b {
		networkBits := bits - i*8

		if networkBits <= 0 {
			b[i] = 0xff
		} else if networkBits < 8 {
			b[i] |= 0xff >> networkBits
		}
	}

	last, _ := netip.AddrFromSlice(b)
	return last
}

func mergeAddrRanges(ranges []addrRange) ([]addrRange, []addrRange) {
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].from.Less(ranges[j].from)
	})
	var v4, v6 []addrRange

	for _, r := range ranges {
		list := &v6

		if r.from.Is4() {
			list = &v4
		}

		if n := len(*list); n > 0 {
			last := &(*list)[n-1]

			if next := last.to.Next(); last.to.Compare(r.from) >= 0 || (next.IsValid() && next == r.from) {
				if r.to.Compare(last.to) > 0 {
					last.to = r.to
				}

				continue
			}
		}

		*list = append(*list, r)
	}

	return v4, v6
}

func findAddrRange(ranges []addrRange, addr netip.Addr) bool {
	i := sort.Search(len(ranges), func(i int) bool {
		return ranges[i].to.Compare(addr) >= 0
	})
	return i < len(ranges) && ranges[i].from.Compare(addr) <= 0
}
