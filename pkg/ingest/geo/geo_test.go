package geo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	// this test is disabled if the licence key is empty
	licenseKey := os.Getenv("GEOLITE2_LICENCE_KEY")

	if licenseKey != "" {
		geoDB, err := NewGeo(licenseKey, "tmp", "")
		assert.NoError(t, err)
		assert.NotNil(t, geoDB)
		assert.NoFileExists(t, filepath.Join("tmp", geoLite2TarGzFilename))
		req := &ingest.Request{IP: "81.2.69.142"}
		cancel, err := geoDB.Step(req)
		assert.False(t, cancel)
		assert.NoError(t, err)
		assert.NotEmpty(t, req.CountryCode)
		assert.NotEmpty(t, req.Region)
		assert.NotEmpty(t, req.City)
	}
}

func TestGeoLocation(t *testing.T) {
	geoDB, _ := NewGeo("", "", "")
	assert.NoError(t, geoDB.UpdateFromFile("../../../test/GeoIP2-City-Test.mmdb"))
	req := &ingest.Request{IP: "81.2.69.142"}
	cancel, err := geoDB.Step(req)
	assert.False(t, cancel)
	assert.NoError(t, err)
	assert.Equal(t, "gb", req.CountryCode)
	assert.Equal(t, "England", req.Region)
	assert.Equal(t, "London", req.City)
}
