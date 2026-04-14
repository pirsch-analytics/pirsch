package geo

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/oschwald/maxminddb-golang"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
)

const (
	geoLite2Permalink     = "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=LICENCE_KEY&suffix=tar.gz"
	geoLite2LicenseKey    = "LICENCE_KEY"
	geoLite2TarGzFilename = "GeoLite2-City.tar.gz"
	geoLite2Filename      = "GeoLite2-City.mmdb"
)

// Geo maps IPs to their geological location based on MaxMinds GeoLite2 or GeoIP2 database.
type Geo struct {
	licenseKey   string
	downloadPath string
	downloadURL  string
	db           *maxminddb.Reader
	m            sync.RWMutex
}

// NewGeo creates a new Geo for the given licence key.
// The download URL is optional and will be set to the default if empty.
// "LICENCE_KEY" will be replaced with the configured licence key.
func NewGeo(licenseKey, downloadPath, downloadURL string) (*Geo, error) {
	if downloadURL == "" {
		downloadURL = geoLite2Permalink
	}

	geoDB := &Geo{
		licenseKey:   licenseKey,
		downloadPath: downloadPath,
		downloadURL:  downloadURL,
	}

	if licenseKey != "" && downloadPath != "" {
		if err := geoDB.Update(); err != nil {
			return nil, err
		}
	}

	return geoDB, nil
}

// Step implements ingest.PipeStep to process a step.
// It looks up the country code, subdivision (region), and city for given IP.
// If the IP is invalid, it won't do anything.
func (geo *Geo) Step(request *ingest.Request) (bool, error) {
	parsedIP := net.ParseIP(request.IP)

	if parsedIP == nil {
		return false, nil
	}

	record := struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
		Subdivisions []struct {
			ISOCode string `maxminddb:"iso_code"`
			Names   struct {
				En string `maxminddb:"en"`
			} `maxminddb:"names"`
		} `maxminddb:"subdivisions"`
		City struct {
			Names struct {
				En string `maxminddb:"en"`
			} `maxminddb:"names"`
		} `maxminddb:"city"`
	}{}

	geo.m.RLock()
	defer geo.m.RUnlock()

	if err := geo.db.Lookup(parsedIP, &record); err != nil {
		return false, nil
	}

	if record.Country.ISOCode == "US" && len(record.Subdivisions) > 0 && record.Subdivisions[0].ISOCode != "" {
		record.City.Names.En += fmt.Sprintf(" (%s)", record.Subdivisions[0].ISOCode)
	}

	subdivision := ""

	if len(record.Subdivisions) > 0 {
		subdivision = record.Subdivisions[0].Names.En
	}

	request.CountryCode = strings.ToLower(record.Country.ISOCode)
	request.Region = subdivision
	request.City = record.City.Names.En
	return false, nil
}

// Update downloads and unpacks the MaxMind GeoLite2 database.
func (geo *Geo) Update() error {
	if err := geo.download(); err != nil {
		return err
	}

	if err := geo.unpackAndUpdate(); err != nil {
		return err
	}

	if err := os.Remove(filepath.Join(geo.downloadPath, geoLite2TarGzFilename)); err != nil {
		return err
	}

	return nil
}

// UpdateFromFile updates Geo from a given file instead of downloading the database.
func (geo *Geo) UpdateFromFile(path string) error {
	data, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	geo.m.Lock()
	defer geo.m.Unlock()
	geoDB, err := maxminddb.FromBytes(data)

	if err != nil {
		return err
	}

	geo.db = geoDB
	return nil
}

func (geo *Geo) download() error {
	if err := os.MkdirAll(geo.downloadPath, 0755); err != nil {
		return err
	}

	resp, err := http.Get(strings.Replace(geo.downloadURL, geoLite2LicenseKey, geo.licenseKey, 1))

	if err != nil {
		return err
	}

	tarGz, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(geo.downloadPath, geoLite2TarGzFilename), tarGz, 0755); err != nil {
		return err
	}

	return nil
}

func (geo *Geo) unpackAndUpdate() error {
	file, err := os.Open(filepath.Join(geo.downloadPath, geoLite2TarGzFilename))

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

		if filepath.Base(header.Name) == geoLite2Filename {
			if _, err := io.Copy(&out, r); err != nil {
				return err
			}

			break
		}
	}

	geo.m.Lock()
	defer geo.m.Unlock()
	geoDB, err := maxminddb.FromBytes(out.Bytes())

	if err != nil {
		return err
	}

	geo.db = geoDB
	return nil
}
