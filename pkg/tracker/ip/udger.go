package ip

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
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
	accessKey          string
	downloadPath       string
	downloadURL        string
	ipsV4, ipsV6       map[string]struct{}
	rangesV4, rangesV6 []ipRange
	m                  sync.RWMutex
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

// DownloadAndUpdate downloads and updates the IP list from udger.com.
func (udger *Udger) DownloadAndUpdate() error {
	if err := udger.download(); err != nil {
		return err
	}

	if err := udger.updateFromDB(); err != nil {
		return err
	}

	if err := os.Remove(filepath.Join(udger.downloadPath, udgerFilename)); err != nil {
		return err
	}

	return nil
}

func (udger *Udger) download() error {
	if err := os.MkdirAll(udger.downloadPath, 0755); err != nil {
		return err
	}

	resp, err := http.Get(fmt.Sprintf(udger.downloadURL, udger.accessKey))

	if err != nil {
		return err
	}

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(udger.downloadPath, udgerFilename), data, 0755); err != nil {
		return err
	}

	return nil
}

func (udger *Udger) updateFromDB() error {
	db, err := sql.Open("sqlite3", filepath.Join(udger.downloadPath, udgerFilename))

	if err != nil {
		return err
	}

	defer db.Close()
	var ipV4 []string
	rows, err := db.Query("SELECT ip FROM udger_ip_list WHERE ip NOT LIKE '%:%' AND class_id NOT IN (1, 5, 100)")

	if err != nil {
		return err
	}

	for rows.Next() {
		var ip string

		if err := rows.Scan(&ip); err != nil {
			return err
		}

		ipV4 = append(ipV4, ip)
	}

	var ipV6 []string
	rows, err = db.Query("SELECT ip FROM udger_ip_list WHERE ip LIKE '%:%' AND class_id NOT IN (1, 5, 100)")

	if err != nil {
		return err
	}

	for rows.Next() {
		var ip string

		if err := rows.Scan(&ip); err != nil {
			return err
		}

		ipV6 = append(ipV6, ip)
	}

	var rangesV4 []Range
	rows, err = db.Query("SELECT ip_from, ip_to FROM udger_datacenter_range")

	if err != nil {
		return err
	}

	for rows.Next() {
		var ipRange Range

		if err := rows.Scan(&ipRange.From, &ipRange.To); err != nil {
			return err
		}

		rangesV4 = append(rangesV4, ipRange)
	}

	var rangesV6 []Range
	rows, err = db.Query("SELECT ip_from, ip_to FROM udger_datacenter_range6")

	if err != nil {
		return err
	}

	for rows.Next() {
		var ipRange Range

		if err := rows.Scan(&ipRange.From, &ipRange.To); err != nil {
			return err
		}

		rangesV6 = append(rangesV6, ipRange)
	}

	udger.Update(ipV4, ipV6, rangesV4, rangesV6)
	return nil
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
