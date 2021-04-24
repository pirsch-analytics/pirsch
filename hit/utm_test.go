package hit

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUTMParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/path?utm_source=test&utm_medium=email&utm_campaign=newsletter&utm_content=sign%20up&utm_term=key+words", nil)
	params := getUTMParams(req)
	assert.Equal(t, "test", params.source)
	assert.Equal(t, "email", params.medium)
	assert.Equal(t, "newsletter", params.campaign)
	assert.Equal(t, "sign up", params.content)
	assert.Equal(t, "key words", params.term)
	req = httptest.NewRequest(http.MethodGet, "/path?utm_source=test", nil)
	params = getUTMParams(req)
	assert.Equal(t, "test", params.source)
	assert.True(t, params.medium == "")
	assert.True(t, params.campaign == "")
	assert.True(t, params.content == "")
	assert.True(t, params.term == "")
}
