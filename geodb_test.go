package pirsch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetGeoLite2(t *testing.T) {
	// this test is disabled if the license key is empty
	licenseKey := os.Getenv("GEOLITE2_LICENSE_KEY")

	if licenseKey == "" {
		return
	}

	if err := GetGeoLite2("geodb", licenseKey); err != nil {
		t.Fatalf("GeoLite2 DB must have been downloaded and unpacked, but was: %v", err)
	}

	if _, err := os.Stat(filepath.Join("geodb", GeoLite2Filename)); err != nil {
		t.Fatal("GeoLite2 database must exist")
	}

	if _, err := os.Stat(filepath.Join("geodb", geoLite2TarGzFilename)); !os.IsNotExist(err) {
		t.Fatalf("GeoLite2 database tarball must not exist anymore, but was: %v", err)
	}
}

func TestGeoDB_CountryCode(t *testing.T) {
	db, err := NewGeoDB(filepath.Join("geodb/GeoIP2-Country-Test.mmdb"))

	if err != nil {
		t.Fatalf("Geo DB must have been loaded, but was: %v", err)
	}

	defer db.Close()

	if out := db.CountryCode("81.2.69.142"); out != "gb" {
		t.Fatalf("Country code for GB must have been returned, but was: %v", out)
	}
}
