package pirsch

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient("tcp://127.0.0.1:9000", nil)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NoError(t, client.DB.Ping())
}

func TestClient_SaveHit(t *testing.T) {
	client, err := NewClient("tcp://127.0.0.1:9000", nil)
	assert.NoError(t, err)
	cleanupDB()
	assert.NoError(t, client.SaveHits([]Hit{
		{
			TenantID:       NewTenantID(1),
			Fingerprint:    "fp",
			Time:           time.Now(),
			Session:        sql.NullTime{Time: time.Now(), Valid: true},
			UserAgent:      "ua",
			Path:           "/path",
			Language:       sql.NullString{String: "en", Valid: true},
			Referrer:       sql.NullString{String: "ref", Valid: true},
			ReferrerName:   sql.NullString{String: "ref_name", Valid: true},
			ReferrerIcon:   sql.NullString{String: "ref_icon", Valid: true},
			OS:             sql.NullString{String: "os", Valid: true},
			OSVersion:      sql.NullString{String: "10", Valid: true},
			Browser:        sql.NullString{String: "browser", Valid: true},
			BrowserVersion: sql.NullString{String: "89", Valid: true},
			CountryCode:    sql.NullString{String: "en", Valid: true},
			Desktop:        true,
			Mobile:         false,
			ScreenWidth:    1920,
			ScreenHeight:   1080,
			ScreenClass:    sql.NullString{String: "XL", Valid: true},
		},
		{
			Fingerprint: "fp",
			Time:        time.Now(),
			UserAgent:   "ua",
			Path:        "/path",
		},
	}))
}

func TestClient_Session(t *testing.T) {
	client, err := NewClient("tcp://127.0.0.1:9000", nil)
	assert.NoError(t, err)
	cleanupDB()
	tenant := NewTenantID(1)
	fp := "session_fp"
	session := Today()
	assert.NoError(t, client.SaveHits([]Hit{
		{
			TenantID:    tenant,
			Fingerprint: fp,
			Time:        time.Now(),
			Session:     sql.NullTime{Time: session, Valid: true},
			UserAgent:   "ua",
			Path:        "/path",
		},
	}))
	s, err := client.Session(tenant, fp, time.Now().Add(-time.Second))
	assert.NoError(t, err)
	assert.Equal(t, session, s)
}
