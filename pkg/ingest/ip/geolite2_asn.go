package ip

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/oschwald/maxminddb-golang"
)

const (
	geoLite2ASNPermalink     = "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN&license_key=LICENCE_KEY&suffix=tar.gz"
	geoLite2ASNLicenseKey    = "LICENCE_KEY"
	geoLite2ASNTarGzFilename = "GeoLite2-ASN.tar.gz"
	geoLite2ASNFilename      = "GeoLite2-ASN.mmdb"

	// DefaultDatacenterASNListURL is the default list of data center ASNs maintained by X4BNet.
	DefaultDatacenterASNListURL = "https://raw.githubusercontent.com/X4BNet/lists_vpn/main/input/datacenter/ASN.txt"
)

// DefaultASNOrganizationKeywords is the default list of keywords used to identify data centers by their organization name.
var DefaultASNOrganizationKeywords = []string{
	"hosting",
	"datacenter",
	"data center",
	"data centre",
	"colocation",
	"server",
	"cloud",
	"vps",
	"dedicated",
}

// DefaultAllowASNs is the default list of autonomous system numbers that are never filtered.
// They belong to providers relaying regular visitors, like Cloudflare WARP and Apple Private Relay.
var DefaultAllowASNs = []uint{
	13335, // Cloudflare
	20940, // Akamai International
	16625, // Akamai Technologies
	36183, // Akamai (Apple Private Relay)
}

// GeoLite2ASNOptions are the options for the GeoLite2ASN filter.
type GeoLite2ASNOptions struct {
	// LicenseKey is the MaxMind license key used to download the database.
	LicenseKey string

	// DownloadPath is the directory the database is downloaded to.
	DownloadPath string

	// DownloadURL is the optional URL to download the database (e.g. a proxy).
	// "LICENCE_KEY" will be replaced with the configured license key.
	DownloadURL string

	// AllowASNs is a list of autonomous system numbers that are never filtered.
	// Use this to fix false positives from the downloaded list or organization keywords.
	// Defaults to DefaultAllowASNs. Pass an empty, non-nil slice to disable it.
	AllowASNs []uint

	// ASNListURL is the URL to download a list of data center ASNs from.
	// Each line must start with an ASN, optionally prefixed by "AS". Everything after it is ignored.
	// Defaults to DefaultDatacenterASNListURL.
	ASNListURL string

	// DisableASNList disables downloading the ASN list.
	DisableASNList bool

	// OrganizationKeywords is a list of case-insensitive keywords.
	// If the organization name of an autonomous system contains one of them, it will be filtered.
	// Defaults to DefaultASNOrganizationKeywords.
	OrganizationKeywords []string

	// DisableOrganizationKeywords disables filtering by organization name.
	DisableOrganizationKeywords bool
}

// GeoLite2ASN implements the Filter interface.
// It filters IP addresses belonging to data centers and cloud providers based on MaxMinds GeoLite2-ASN database,
// using a downloaded list of data center ASNs and the organization name.
type GeoLite2ASN struct {
	licenseKey   string
	downloadPath string
	downloadURL  string
	asnListURL   string
	keywords     []string
	listASNs     map[uint]struct{}
	allowASNs    map[uint]struct{}
	db           *maxminddb.Reader
	m            sync.RWMutex
}

// NewGeoLite2ASN creates a new Filter using the MaxMind GeoLite2-ASN database.
func NewGeoLite2ASN(options GeoLite2ASNOptions) *GeoLite2ASN {
	if options.DownloadURL == "" {
		options.DownloadURL = geoLite2ASNPermalink
	}

	if options.DisableASNList {
		options.ASNListURL = ""
	} else if options.ASNListURL == "" {
		options.ASNListURL = DefaultDatacenterASNListURL
	}

	if options.AllowASNs == nil {
		options.AllowASNs = DefaultAllowASNs
	}

	var keywords []string

	if !options.DisableOrganizationKeywords {
		if options.OrganizationKeywords == nil {
			options.OrganizationKeywords = DefaultASNOrganizationKeywords
		}

		keywords = make([]string, 0, len(options.OrganizationKeywords))

		for _, keyword := range options.OrganizationKeywords {
			keyword = strings.ToLower(strings.TrimSpace(keyword))

			if keyword != "" {
				keywords = append(keywords, keyword)
			}
		}
	}

	return &GeoLite2ASN{
		licenseKey:   options.LicenseKey,
		downloadPath: options.DownloadPath,
		downloadURL:  options.DownloadURL,
		asnListURL:   options.ASNListURL,
		keywords:     keywords,
		listASNs:     make(map[uint]struct{}),
		allowASNs:    asnsToMap(options.AllowASNs),
	}
}

// Update implements the Filter interface.
// The GeoLite2-ASN filter does not use plain IP addresses or ranges, so this is a no-op.
// Use DownloadAndUpdate or UpdateFromFile to update the database.
func (g *GeoLite2ASN) Update([]string, []string, []string, []string, []Range, []Range) {}

// Ignore implements the Filter interface.
func (g *GeoLite2ASN) Ignore(ip string) bool {
	parsedIP := net.ParseIP(ip)

	if parsedIP == nil {
		return true
	}

	g.m.RLock()
	defer g.m.RUnlock()

	if g.db == nil {
		return false
	}

	record := struct {
		ASN          uint   `maxminddb:"autonomous_system_number"`
		Organization string `maxminddb:"autonomous_system_organization"`
	}{}

	if err := g.db.Lookup(parsedIP, &record); err != nil || record.ASN == 0 {
		return false
	}

	if _, ok := g.allowASNs[record.ASN]; ok {
		return false
	}

	if _, ok := g.listASNs[record.ASN]; ok {
		return true
	}

	if record.Organization != "" && len(g.keywords) > 0 {
		organization := strings.ToLower(record.Organization)

		for _, keyword := range g.keywords {
			if strings.Contains(organization, keyword) {
				return true
			}
		}
	}

	return false
}

// DownloadAndUpdate downloads and updates the GeoLite2-ASN database and the data center ASN list (unless disabled).
func (g *GeoLite2ASN) DownloadAndUpdate() error {
	if err := g.download(); err != nil {
		return err
	}

	if err := g.unpackAndUpdate(); err != nil {
		return err
	}

	if err := os.Remove(filepath.Join(g.downloadPath, geoLite2ASNTarGzFilename)); err != nil {
		return err
	}

	return g.downloadASNList()
}

// UpdateFromFile updates the database from a given .mmdb file instead of downloading it.
// The data center ASN list is not updated.
func (g *GeoLite2ASN) UpdateFromFile(path string) error {
	data, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	return g.updateDB(data)
}

// downloadASNList downloads and updates the data center ASN list only. It does nothing if the list is disabled.
func (g *GeoLite2ASN) downloadASNList() error {
	if g.asnListURL == "" {
		return nil
	}

	resp, err := http.Get(g.asnListURL)

	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if err := checkDownloadResponse(resp); err != nil {
		return err
	}

	return g.updateASNList(resp.Body)
}

func (g *GeoLite2ASN) updateASNList(r io.Reader) error {
	asns := make(map[uint]struct{})
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		if asn := parseASNListLine(scanner.Text()); asn != 0 {
			asns[asn] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if len(asns) == 0 {
		return errors.New("ASN list is empty")
	}

	g.m.Lock()
	defer g.m.Unlock()
	g.listASNs = asns
	return nil
}

func (g *GeoLite2ASN) download() error {
	if err := os.MkdirAll(g.downloadPath, 0755); err != nil {
		return err
	}

	resp, err := http.Get(strings.Replace(g.downloadURL, geoLite2ASNLicenseKey, g.licenseKey, 1))

	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if err := checkDownloadResponse(resp); err != nil {
		return err
	}

	tarGz, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(g.downloadPath, geoLite2ASNTarGzFilename), tarGz, 0755); err != nil {
		return err
	}

	return nil
}

func (g *GeoLite2ASN) unpackAndUpdate() error {
	file, err := os.Open(filepath.Join(g.downloadPath, geoLite2ASNTarGzFilename))

	if err != nil {
		return err
	}

	defer func() {
		_ = file.Close()
	}()
	gzipFile, err := gzip.NewReader(file)

	if err != nil {
		return err
	}

	defer func() {
		_ = gzipFile.Close()
	}()
	r := tar.NewReader(gzipFile)
	var out bytes.Buffer

	for {
		header, err := r.Next()

		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if filepath.Base(header.Name) == geoLite2ASNFilename {
			if _, err := io.Copy(&out, r); err != nil {
				return err
			}

			break
		}
	}

	if out.Len() == 0 {
		return errors.New("GeoLite2-ASN database not found in archive")
	}

	return g.updateDB(out.Bytes())
}

func (g *GeoLite2ASN) updateDB(data []byte) error {
	db, err := maxminddb.FromBytes(data)

	if err != nil {
		return err
	}

	g.m.Lock()
	defer g.m.Unlock()
	g.db = db
	return nil
}

func parseASNListLine(line string) uint {
	line = strings.TrimSpace(line)

	if len(line) > 2 && strings.EqualFold(line[:2], "AS") {
		line = line[2:]
	}

	end := 0

	for end < len(line) && line[end] >= '0' && line[end] <= '9' {
		end++
	}

	if end == 0 {
		return 0
	}

	asn, err := strconv.ParseUint(line[:end], 10, 32)

	if err != nil {
		return 0
	}

	return uint(asn)
}

func asnsToMap(asns []uint) map[uint]struct{} {
	m := make(map[uint]struct{}, len(asns))

	for _, asn := range asns {
		m[asn] = struct{}{}
	}

	return m
}
