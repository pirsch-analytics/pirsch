package pirsch

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	// The license key needs to be set for the tests!
	testLicensekey = ""
)

func TestGetGeoLite2(t *testing.T) {
	if testLicensekey == "" {
		return
	}

	if err := GetGeoLite2("geodb", testLicensekey); err != nil {
		t.Fatalf("GeoLite2 DB must have been downloaded and unpacked, but was: %v", err)
	}

	if _, err := os.Stat(filepath.Join("geodb", geoLite2Filename)); err != nil {
		t.Fatal("GeoLite2 database must exist")
	}

	if _, err := os.Stat(filepath.Join("geodb", geoLite2TarGzFilename)); !os.IsNotExist(err) {
		t.Fatalf("GeoLite2 database tarball must not exist anymore, but was: %v", err)
	}
}
