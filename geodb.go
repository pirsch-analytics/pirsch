package pirsch

import (
	"archive/tar"
	"compress/gzip"
	"github.com/oschwald/maxminddb-golang"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	geoLite2Permalink     = "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-Country&license_key=LICENSE_KEY&suffix=tar.gz"
	geoLite2LicenseKey    = "LICENSE_KEY"
	geoLite2TarGzFilename = "GeoLite2-Country.tar.gz"

	// GeoLite2Filename is the default filename of the GeoLite2 database.
	GeoLite2Filename = "GeoLite2-Country.mmdb"
)

// GeoDB maps IPs to their geo location based on MaxMinds GeoLite2 or GeoIP2 database.
type GeoDB struct {
	db *maxminddb.Reader
}

// NewGeoDB creates a new GeoDB for given database file.
// Make sure you call GeoDB.Close to release the system resources!
// If you use this in combination with GetGeoLite2, you should pass in the path to GeoLite2Filename (including the filename).
// The database should be updated on a regular basis.
func NewGeoDB(file string) (*GeoDB, error) {
	db, err := maxminddb.Open(file)

	if err != nil {
		return nil, err
	}

	geoDB := &GeoDB{
		db: db,
	}
	return geoDB, nil
}

// Close closes the database file handle and frees the system resources.
// It's important to call this when you don't need the GeoDB anymore!
func (db *GeoDB) Close() error {
	return db.db.Close()
}

// CountryCode looks up the country code for given IP.
// If the IP is invalid it will return an empty string.
// The country code is returned in lowercase.
func (db *GeoDB) CountryCode(ip string) string {
	parsedIP := net.ParseIP(ip)

	if parsedIP == nil {
		return ""
	}

	record := struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	}{}

	if err := db.db.Lookup(parsedIP, &record); err != nil {
		return ""
	}

	return strings.ToLower(record.Country.ISOCode)
}

// GetGeoLite2 downloads and unpacks the MaxMind GeoLite2 database.
// The tarball is downloaded and unpacked at the provided path. The directories will created if required.
// The license key is used for the download and must be provided for a registered account.
// Please refer to MaxMinds website on how to do that: https://dev.maxmind.com/geoip/geoip2/geolite2/
// The database should be updated on a regular basis.
func GetGeoLite2(path, licenseKey string) error {
	if err := downloadGeoLite2(path, licenseKey); err != nil {
		return err
	}

	if err := unpackGeoLite2(path); err != nil {
		return err
	}

	if err := cleanupGeoLite2Download(path); err != nil {
		return err
	}

	return nil
}

func downloadGeoLite2(path, licenseKey string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	resp, err := http.Get(strings.Replace(geoLite2Permalink, geoLite2LicenseKey, licenseKey, 1))

	if err != nil {
		return err
	}

	tarGz, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(path, geoLite2TarGzFilename), tarGz, 0755); err != nil {
		return err
	}

	return nil
}

func unpackGeoLite2(path string) error {
	file, err := os.Open(filepath.Join(path, geoLite2TarGzFilename))

	if err != nil {
		return err
	}

	defer file.Close()
	gzipFile, err := gzip.NewReader(file)

	if err != nil {
		return err
	}

	defer gzipFile.Close()
	r := tar.NewReader(gzipFile)

	for {
		header, err := r.Next()

		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if filepath.Base(header.Name) == GeoLite2Filename {
			out, err := os.Create(filepath.Join(path, GeoLite2Filename))

			if err != nil {
				return err
			}

			if _, err := io.Copy(out, r); err != nil {
				out.Close()
				return err
			}

			out.Close()
			break
		}
	}

	return nil
}

func cleanupGeoLite2Download(path string) error {
	if err := os.Remove(filepath.Join(path, geoLite2TarGzFilename)); err != nil {
		return err
	}

	return nil
}
