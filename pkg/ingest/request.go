package ingest

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest/util"
	"github.com/pirsch-analytics/pirsch/v7/pkg/model"
)

// Request is the data for a visitor request (page view or event).
type Request struct {
	// ClientID is the ID for the tenant.
	ClientID uint64

	// VisitorID is the visitor ID for the request.
	VisitorID uint64

	// SessionID is the session ID for the request.
	SessionID uint32

	// Request is the http.Request for the visitor.
	// This field is mandatory.
	Request *http.Request

	// Time overrides the time the page view should be recorded for at UTC.
	// By default, it will be set to the time the request is started to being processed by the Pipe.
	Time time.Time

	// Start is the time a visitor has first been seen for the session.
	Start time.Time

	DurationSeconds uint64

	// URL is the full request URL.
	// If not set, it will be extracted from the Request.
	URL string

	// Hostname overrides the hostname and is used to check the Referrer.
	// If the referrer is the same as the hostname, it will be ignored.
	Hostname string

	// Path overrides the path.
	// If not set, it will be extracted from the Request.
	Path string

	// PageViews is the number of page views for the session.
	PageViews uint16

	// IsBounce sets the session as bounced.
	IsBounce bool

	// Title sets the page title.
	Title string

	// Referrer overrides the referrer header.
	// If it is the same as the Hostname, it will be ignored.
	Referrer string

	// ScreenWidth is the screen width that will be translated to a screen class.
	ScreenWidth uint16

	// ScreenHeight is the screen height that will be translated to a screen class.
	ScreenHeight uint16

	// Tags are optional fields used to break down page views into segments.
	Tags map[string]string

	// EventName is optional.
	// If set, the Request will be stored as an event.
	EventName string

	// EventMetaData are optional event metadata fields.
	EventMetaData map[string]any

	// EventNonInteractive is an optional field marking the event as non-interactive.
	// A non-interactive event will keep the session marked as bounced.
	EventNonInteractive bool

	// DisableBotFilter disables all bot filters if set to true.
	DisableBotFilter bool

	// Language is the language for the request.
	// This should be set by a PipeStep.
	Language string

	// CountryCode is the country ISO code for the request.
	// This should be set by a PipeStep.
	CountryCode string

	// Region is the region for the request.
	// This should be set by a PipeStep.
	Region string

	// City is the city for the request.
	// This should be set by a PipeStep.
	City string

	// ReferrerName is the referrer name (group) for the request.
	// This should be set by a PipeStep.
	ReferrerName string

	// ReferrerIcon is the referrer icon URL for the request.
	// This should be set by a PipeStep.
	ReferrerIcon string

	// OS is the OS for the request.
	// This should be set by a PipeStep.
	OS string

	// OSVersion is the OS version for the request.
	// This should be set by a PipeStep.
	OSVersion string

	// Browser is the browser for the request.
	// This should be set by a PipeStep.
	Browser string

	// BrowserVersion is the browser version for the request.
	// This should be set by a PipeStep.
	BrowserVersion string

	// Desktop indicates a desktop device for the request.
	// This should be set by a PipeStep.
	Desktop bool

	// Mobile indicates a mobile device for the request.
	// This should be set by a PipeStep.
	Mobile bool

	// ScreenClass is the screen class for the request.
	// This should be set by a PipeStep.
	ScreenClass string

	// IP is the IP for the request.
	// This should be set by a PipeStep.
	IP string

	// UserAgent is the User Agent for the request.
	// This should be set by a PipeStep.
	UserAgent string

	// UTMSource is the UTM source for the request.
	// This should be set by a PipeStep.
	UTMSource string

	// UTMMedium is the UTM medium for the request.
	// This should be set by a PipeStep.
	UTMMedium string

	// UTMCampaign is the UTM campaign for the request.
	// This should be set by a PipeStep.
	UTMCampaign string

	// UTMContent is the UTM content for the request.
	// This should be set by a PipeStep.
	UTMContent string

	// UTMTerm is the UTM term for the request.
	// This should be set by a PipeStep.
	UTMTerm string

	// ClickID is the gclid (Google) or msclkid (Microsoft) click ID.
	ClickID string

	// Channel is the channel for the request.
	// This should be set by a PipeStep.
	Channel string

	// IsBot marks the request as a bot.
	// This should be set by a PipeStep.
	IsBot bool

	// BotReason is the reason why a request has been blocked for a bot.
	// This should be set by a PipeStep.
	BotReason string

	cancelled     bool
	session       *model.Session
	cancelSession *model.Session
}

func (request *Request) validate() {
	if request.Time.IsZero() {
		request.Time = time.Now().UTC()
	}

	if request.URL == "" {
		request.URL = request.Request.URL.String()
	}

	u, err := url.ParseRequestURI(request.URL)

	if err == nil {
		if request.Hostname == "" {
			request.Hostname = u.Hostname()
		}

		// change the path and re-assemble the URL if override is set
		if request.Path != "" {
			u.Path = request.Path
			request.URL = u.String()
		} else {
			request.Path = u.Path
		}
	}

	request.Title = util.Shorten(request.Title, 512)
	request.Path = util.Shorten(request.Path, 2000)

	if request.Path == "" {
		request.Path = "/"
	}

	request.EventName = strings.TrimSpace(request.EventName)
}
