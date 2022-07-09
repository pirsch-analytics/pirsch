package geodb

import (
	"archive/tar"
	"compress/gzip"
	"github.com/oschwald/maxminddb-golang"
	"github.com/pirsch-analytics/pirsch/v4"
	"github.com/pirsch-analytics/pirsch/v4/util"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	geoLite2Permalink     = "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=LICENSE_KEY&suffix=tar.gz"
	geoLite2LicenseKey    = "LICENSE_KEY"
	geoLite2TarGzFilename = "GeoLite2-City.tar.gz"
)

// Config is the configuration for the GeoDB.
type Config struct {
	// File is the path (including the filename) to the GeoLite2 country database file.
	// See GeoLite2Filename for the required filename.
	File string

	// Logger is the log.Logger used for logging.
	// Note that this will log the IP address and should therefore only be used for debugging.
	// Set it to nil to disable logging for GeoDB.
	Logger *log.Logger
}

// GeoDB maps IPs to their geo location based on MaxMinds GeoLite2 or GeoIP2 database.
type GeoDB struct {
	db     *maxminddb.Reader
	logger *log.Logger
}

// NewGeoDB creates a new GeoDB for given database file.
// The file is loaded into memory, therefore it's not necessary to close the reader (see oschwald/maxminddb-golang documentatio).
// The database should be updated on a regular basis.
func NewGeoDB(config Config) (*GeoDB, error) {
	data, err := os.ReadFile(config.File)

	if err != nil {
		return nil, err
	}

	db, err := maxminddb.FromBytes(data)

	if err != nil {
		return nil, err
	}

	return &GeoDB{
		db:     db,
		logger: config.Logger,
	}, nil
}

// CountryCodeAndCity looks up the country code and city for given IP.
// If the IP is invalid it will return an empty string.
// The country code is returned in lowercase.
func (db *GeoDB) CountryCodeAndCity(ip string) (string, string) {
	parsedIP := net.ParseIP(ip)

	if parsedIP == nil {
		if db.logger != nil {
			db.logger.Printf("error parsing IP address %s", ip)
		}

		return "", ""
	}

	record := struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
		City struct {
			Names struct {
				En string `maxminddb:"en"`
			} `maxminddb:"names"`
		} `maxminddb:"city"`
	}{}

	if err := db.db.Lookup(parsedIP, &record); err != nil {
		if db.logger != nil {
			db.logger.Printf("error looking up IP address %s", parsedIP)
		}

		return "", ""
	}

	return strings.ToLower(record.Country.ISOCode), record.City.Names.En
}

// Get downloads and unpacks the MaxMind GeoLite2 database.
// The tarball is downloaded and unpacked at the provided path. The directories will created if required.
// The license key is used for the download and must be provided for a registered account.
// Please refer to MaxMinds website on how to do that: https://dev.maxmind.com/geoip/geoip2/geolite2/
// The database should be updated on a regular basis.
func Get(path, licenseKey string) error {
	if err := download(path, licenseKey); err != nil {
		return err
	}

	if err := unpack(path); err != nil {
		return err
	}

	if err := cleanupDownload(path); err != nil {
		return err
	}

	return nil
}

func download(path, licenseKey string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	resp, err := http.Get(strings.Replace(geoLite2Permalink, geoLite2LicenseKey, licenseKey, 1))

	if err != nil {
		return err
	}

	tarGz, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(path, geoLite2TarGzFilename), tarGz, 0755); err != nil {
		return err
	}

	return nil
}

func unpack(path string) error {
	file, err := os.Open(filepath.Join(path, geoLite2TarGzFilename))

	if err != nil {
		return err
	}

	logger := util.GetDefaultLogger()
	defer func() {
		if err := file.Close(); err != nil {
			logger.Printf("error closing GeoDB file")
		}
	}()
	gzipFile, err := gzip.NewReader(file)

	if err != nil {
		return err
	}

	defer func() {
		if err := gzipFile.Close(); err != nil {
			logger.Printf("error closing GeoDB zip file")
		}
	}()
	r := tar.NewReader(gzipFile)

	for {
		header, err := r.Next()

		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if filepath.Base(header.Name) == pirsch.GeoLite2Filename {
			out, err := os.Create(filepath.Join(path, pirsch.GeoLite2Filename))

			if err != nil {
				return err
			}

			if _, err := io.Copy(out, r); err != nil {
				if err := out.Close(); err != nil {
					logger.Printf("error closing GeoLite2 database file")
				}

				return err
			}

			if err := out.Close(); err != nil {
				logger.Printf("error closing GeoLite2 database file")
			}

			break
		}
	}

	return nil
}

func cleanupDownload(path string) error {
	if err := os.Remove(filepath.Join(path, geoLite2TarGzFilename)); err != nil {
		return err
	}

	return nil
}
