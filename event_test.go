package pirsch

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEventOptions_getMetaData(t *testing.T) {
	options := EventOptions{
		Meta: map[string]string{
			"key":   "value",
			"hello": "world",
		},
	}
	k, v := options.getMetaData()
	assert.Len(t, k, 2)
	assert.Len(t, v, 2)
	assert.Contains(t, k, "key")
	assert.Contains(t, k, "hello")
	assert.Contains(t, v, "value")
	assert.Contains(t, v, "world")
}

func TestEventFromRequest(t *testing.T) {
	cleanupDB()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	req.RemoteAddr = "81.2.69.142"
	client := NewMockClient()
	cache := NewSessionCacheMem(client, 100)
	assert.Nil(t, EventFromRequest(req, "salt", nil))
	assert.Nil(t, EventFromRequest(req, "salt", &HitOptions{
		SessionCache: cache,
	}))
	_, sessions, _ := HitFromRequest(req, "salt", &HitOptions{
		SessionCache: cache,
	})
	assert.Len(t, sessions, 1)
	event := EventFromRequest(req, "salt", &HitOptions{
		SessionCache: cache,
	})
	assert.NotNil(t, event)
	assert.Equal(t, sessions[0].VisitorID, event.VisitorID)
}
