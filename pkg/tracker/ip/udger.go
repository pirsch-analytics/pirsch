package ip

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

const (
	udgerDownload = "https://data.udger.com/%s/udgerdb_v4.dat"
	udgerFilename = "udgerdb_v4.dat"
)

type ipRange struct {
	from net.IP
	to   net.IP
}

// Udger implements the Filter interface.
type Udger struct {
	accessKey                      string
	downloadPath                   string
	downloadURL                    string
	ipsV4, ipsV6                   map[string]struct{}
	whitelistIpsV4, whitelistIpsV6 map[string]struct{}
	rangesV4, rangesV6             []ipRange
	m                              sync.RWMutex
}

// NewUdger creates a new Filter using the IP lists provided by udger.com.
// The download URL is optional and will be set to the default if empty. It must contain "%s" for the access key.
func NewUdger(accessKey, downloadPath, downloadURL string) *Udger {
	if downloadURL == "" {
		downloadURL = udgerDownload
	}

	return &Udger{
		accessKey:    accessKey,
		downloadPath: downloadPath,
		downloadURL:  downloadURL,
	}
}

// Update implements the Filter interface.
func (u *Udger) Update(ipsV4, ipsV6, whitelistIpV4, whitelistIpV6 []string, rangesV4, rangesV6 []Range) {
	ipMapV4 := u.ipsToMap(ipsV4)
	ipMapV6 := u.ipsToMap(ipsV6)
	whitelistIpMapV4 := u.ipsToMap(whitelistIpV4)
	whitelistIpMapV6 := u.ipsToMap(whitelistIpV6)
	ipRangesV4 := u.ipRangesToList(rangesV4)
	ipRangesV6 := u.ipRangesToList(rangesV6)
	u.m.Lock()
	defer u.m.Unlock()
	u.ipsV4 = ipMapV4
	u.ipsV6 = ipMapV6
	u.whitelistIpsV4 = whitelistIpMapV4
	u.whitelistIpsV6 = whitelistIpMapV6
	u.rangesV4 = ipRangesV4
	u.rangesV6 = ipRangesV6
}

// Ignore implements the Filter interface.
func (u *Udger) Ignore(ip string) bool {
	parsedIP := net.ParseIP(ip)

	if parsedIP == nil {
		return true
	}

	u.m.RLock()
	defer u.m.RUnlock()

	if parsedIP.To4() != nil {
		if u.findIP(ip, parsedIP, u.whitelistIpsV4, nil) {
			return false
		}

		return u.findIP(ip, parsedIP, u.ipsV4, u.rangesV4)
	}

	if u.findIP(ip, parsedIP, u.whitelistIpsV6, nil) {
		return false
	}

	return u.findIP(ip, parsedIP, u.ipsV6, u.rangesV6)
}

// DownloadAndUpdate downloads and updates the IP list from udger.com.
func (u *Udger) DownloadAndUpdate() error {
	if err := u.download(); err != nil {
		return err
	}

	if err := u.updateFromDB(); err != nil {
		return err
	}

	if err := os.Remove(filepath.Join(u.downloadPath, udgerFilename)); err != nil {
		return err
	}

	return nil
}

func (u *Udger) download() error {
	if err := os.MkdirAll(u.downloadPath, 0755); err != nil {
		return err
	}

	resp, err := http.Get(fmt.Sprintf(u.downloadURL, u.accessKey))

	if err != nil {
		return err
	}

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(u.downloadPath, udgerFilename), data, 0755); err != nil {
		return err
	}

	return nil
}

func (u *Udger) updateFromDB() error {
	db, err := sql.Open("sqlite3", filepath.Join(u.downloadPath, udgerFilename))

	if err != nil {
		return err
	}

	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("Error closing sqlite connection:", "err", err)
		}
	}()
	ipV4, err := u.loadIPs(db, "SELECT ip FROM udger_ip_list WHERE ip NOT LIKE '%:%' AND class_id NOT IN (1, 5, 100)")

	if err != nil {
		return err
	}

	ipV6, err := u.loadIPs(db, "SELECT ip FROM udger_ip_list WHERE ip LIKE '%:%' AND class_id NOT IN (1, 5, 100)")

	if err != nil {
		return err
	}

	whitelistIpV4, err := u.loadIPs(db, "SELECT ip FROM udger_ip_list WHERE ip NOT LIKE '%:%' AND class_id = 5")

	if err != nil {
		return err
	}

	whitelistIpV6, err := u.loadIPs(db, "SELECT ip FROM udger_ip_list WHERE ip LIKE '%:%' AND class_id = 5")

	if err != nil {
		return err
	}

	rangesV4, err := u.loadIPRanges(db, "SELECT ip_from, ip_to FROM udger_datacenter_range")

	if err != nil {
		return err
	}

	rangesV6, err := u.loadIPRanges(db, "SELECT ip_from, ip_to FROM udger_datacenter_range6")

	if err != nil {
		return err
	}

	u.Update(ipV4, ipV6, whitelistIpV4, whitelistIpV6, rangesV4, rangesV6)
	return nil
}

func (u *Udger) loadIPs(db *sql.DB, query string) ([]string, error) {
	var ips []string
	rows, err := db.Query(query)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var ip string

		if err := rows.Scan(&ip); err != nil {
			return nil, err
		}

		ips = append(ips, ip)
	}

	return ips, nil
}

func (u *Udger) loadIPRanges(db *sql.DB, query string) ([]Range, error) {
	var ranges []Range
	rows, err := db.Query(query)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var ipRange Range

		if err := rows.Scan(&ipRange.From, &ipRange.To); err != nil {
			return nil, err
		}

		ranges = append(ranges, ipRange)
	}

	return ranges, nil
}

func (u *Udger) ipsToMap(ips []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ips))

	for _, ip := range ips {
		m[ip] = struct{}{}
	}

	return m
}

func (u *Udger) ipRangesToList(ranges []Range) []ipRange {
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

func (u *Udger) findIP(ip string, parsedIP net.IP, ips map[string]struct{}, ranges []ipRange) bool {
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
