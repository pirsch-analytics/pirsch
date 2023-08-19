package geodb

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestGet(t *testing.T) {
	// this test is disabled if the license key is empty
	licenseKey := os.Getenv("GEOLITE2_LICENSE_KEY")

	if licenseKey != "" {
		geoDB, err := NewGeoDB(licenseKey, "tmp")
		assert.NoError(t, err)
		assert.NotNil(t, geoDB)
		assert.NoFileExists(t, filepath.Join("tmp", geoLite2TarGzFilename))
		countryCode, city := geoDB.GetLocation("81.2.69.142")
		assert.NotEmpty(t, countryCode)
		assert.NotEmpty(t, city)
	}
}

func TestGeoDB_GetLocation(t *testing.T) {
	geoDB, _ := NewGeoDB("", "")
	assert.NoError(t, geoDB.UpdateFromFile("../../../test/GeoIP2-City-Test.mmdb"))
	countryCode, city := geoDB.GetLocation("81.2.69.142")
	assert.Equal(t, "gb", countryCode)
	assert.Equal(t, "London", city)
}
