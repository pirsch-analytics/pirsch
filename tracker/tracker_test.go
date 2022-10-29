package tracker

import (
	"github.com/pirsch-analytics/pirsch/v4/model"
	"github.com/pirsch-analytics/pirsch/v4/tracker/ua"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTracker_PageView(t *testing.T) {
	// TODO
}

func TestTracker_Event(t *testing.T) {
	// TODO
}

func TestTracker_ExtendSession(t *testing.T) {
	// TODO
}

func TestTracker_ignorePrefetch(t *testing.T) {
	tracker := NewTracker(Config{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0")
	req.Header.Set("X-Moz", "prefetch")

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Session with X-Moz header must be ignored")
	}

	req.Header.Del("X-Moz")
	req.Header.Set("X-Purpose", "prefetch")

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Session with X-Purpose header must be ignored")
	}

	req.Header.Set("X-Purpose", "preview")

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Session with X-Purpose header must be ignored")
	}

	req.Header.Del("X-Purpose")
	req.Header.Set("Purpose", "prefetch")

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Session with Purpose header must be ignored")
	}

	req.Header.Set("Purpose", "preview")

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Session with Purpose header must be ignored")
	}

	req.Header.Del("Purpose")

	if _, _, ignore := tracker.ignore(req); ignore {
		t.Fatal("Session must not be ignored")
	}
}

func TestTracker_ignoreUserAgent(t *testing.T) {
	userAgents := []struct {
		userAgent string
		ignore    bool
	}{
		{"This is a bot request", true},
		{"This is a crawler request", true},
		{"This is a spider request", true},
		{"Visit http://spam.com!", true},
		{"", true},
		{"Mozilla/123.0", false},
	}

	tracker := NewTracker(Config{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	for _, userAgent := range userAgents {
		req.Header.Set("User-Agent", userAgent.userAgent)

		if _, _, ignore := tracker.ignore(req); ignore != userAgent.ignore {
			t.Fatalf("Request with User-Agent '%s' must be ignored", userAgent.userAgent)
		}
	}
}

func TestTracker_ignoreBotUserAgent(t *testing.T) {
	tracker := NewTracker(Config{})

	for _, botUserAgent := range ua.Blacklist {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("User-Agent", botUserAgent)

		if _, _, ignore := tracker.ignore(req); !ignore {
			t.Fatalf("Request with user agent '%v' must have been ignored", botUserAgent)
		}
	}
}

func TestTracker_ignoreReferrer(t *testing.T) {
	tracker := NewTracker(Config{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "ua")
	req.Header.Set("Referer", "2your.site")

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Request must have been ignored")
	}

	req.Header.Set("Referer", "subdomain.2your.site")

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Request for subdomain must have been ignored")
	}

	req = httptest.NewRequest(http.MethodGet, "/?ref=2your.site", nil)

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Request must have been ignored")
	}
}

func TestTracker_ignoreBrowserVersion(t *testing.T) {
	tracker := NewTracker(Config{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.4147.135 Safari/537.36")

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Request must have been ignored")
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36")

	if _, _, ignore := tracker.ignore(req); ignore {
		t.Fatal("Request must not have been ignored")
	}
}

func TestTracker_ignoreDoNotTrack(t *testing.T) {
	tracker := NewTracker(Config{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36")

	if _, _, ignore := tracker.ignore(req); ignore {
		t.Fatal("Request must not have been ignored")
	}

	req.Header.Set("DNT", "1")

	if _, _, ignore := tracker.ignore(req); !ignore {
		t.Fatal("Request must have been ignored")
	}
}

func TestTracker_getLanguage(t *testing.T) {
	input := []string{
		"",
		"  \t ",
		"fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5",
		"en-us, en",
		"en-gb, en",
		"invalid",
	}
	expected := []string{
		"",
		"",
		"fr",
		"en",
		"en",
		"",
	}
	tracker := NewTracker(Config{})

	for i, in := range input {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Language", in)

		if lang := tracker.getLanguage(req); lang != expected[i] {
			t.Fatalf("Expected '%v', but was: %v", expected[i], lang)
		}
	}
}

func TestTracker_getScreenClass(t *testing.T) {
	tracker := NewTracker(Config{})
	assert.Equal(t, "", tracker.getScreenClass(0))
	assert.Equal(t, "XS", tracker.getScreenClass(42))
	assert.Equal(t, "XL", tracker.getScreenClass(1024))
	assert.Equal(t, "XL", tracker.getScreenClass(1025))
	assert.Equal(t, "HD", tracker.getScreenClass(1919))
	assert.Equal(t, "Full HD", tracker.getScreenClass(2559))
	assert.Equal(t, "WQHD", tracker.getScreenClass(3839))
	assert.Equal(t, "UHD 4K", tracker.getScreenClass(5119))
	assert.Equal(t, "UHD 5K", tracker.getScreenClass(5120))
}

func TestTracker_referrerOrCampaignChanged(t *testing.T) {
	tracker := NewTracker(Config{})
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Referer", "https://referrer.com")
	session := &model.Session{Referrer: "https://referrer.com"}
	assert.False(t, tracker.referrerOrCampaignChanged(req, session, ""))
	session.Referrer = ""
	assert.True(t, tracker.referrerOrCampaignChanged(req, session, ""))
	session.Referrer = "https://referrer.com"
	req = httptest.NewRequest(http.MethodGet, "/test?ref=https://different.com", nil)
	assert.True(t, tracker.referrerOrCampaignChanged(req, session, ""))
	req = httptest.NewRequest(http.MethodGet, "/test?utm_source=Referrer", nil)
	assert.True(t, tracker.referrerOrCampaignChanged(req, session, ""))
	session.UTMSource = "Referrer"
	assert.False(t, tracker.referrerOrCampaignChanged(req, session, ""))
}
