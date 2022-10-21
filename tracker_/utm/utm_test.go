package utm

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/path?utm_source=test&utm_medium=email&utm_campaign=newsletter&utm_content=sign%20up&utm_term=key+words", nil)
	params := Get(req)
	assert.Equal(t, "test", params.Source)
	assert.Equal(t, "email", params.Medium)
	assert.Equal(t, "newsletter", params.Campaign)
	assert.Equal(t, "sign up", params.Content)
	assert.Equal(t, "key words", params.Term)
	req = httptest.NewRequest(http.MethodGet, "/path?utm_source=test", nil)
	params = Get(req)
	assert.Equal(t, "test", params.Source)
	assert.True(t, params.Medium == "")
	assert.True(t, params.Campaign == "")
	assert.True(t, params.Content == "")
	assert.True(t, params.Term == "")
}
