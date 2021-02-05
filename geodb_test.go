package pirsch

import (
	"github.com/stretchr/testify/assert"
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

	assert.NoError(t, GetGeoLite2("geodb", licenseKey))
	_, err := os.Stat(filepath.Join("geodb", GeoLite2Filename))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join("geodb", geoLite2TarGzFilename))
	assert.True(t, os.IsNotExist(err))
}

func TestGeoDB_CountryCode(t *testing.T) {
	db, err := NewGeoDB(GeoDBConfig{
		File: filepath.Join("geodb/GeoIP2-Country-Test.mmdb"),
	})
	assert.NoError(t, err)
	defer db.Close()
	assert.Equal(t, "gb", db.CountryCode("81.2.69.142"))
}
