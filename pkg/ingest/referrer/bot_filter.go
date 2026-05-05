package referrer

import (
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
)

// BotFilter filters bot requests based on the referrer.
type BotFilter struct{}

// NewBotFilter creates a new BotFilter.
func NewBotFilter() *BotFilter {
	return new(BotFilter)
}

// Step implements the ingest.PipeStep interface.
func (f *BotFilter) Step(request *ingest.Request) (bool, error) {
	if request.DisableBotFilter {
		return false, nil
	}

	referrer := ""

	if request.Referrer != "" {
		referrer = request.Referrer
	} else {
		referrer = fromHeaderOrQuery(request.Request)
	}

	if referrer == "" {
		return false, nil
	}

	if _, err := uuid.Parse(referrer); err == nil {
		request.BotReason = "ref-uuid"
		return true, nil
	}

	u, err := url.ParseRequestURI(referrer)

	if err == nil {
		referrer = u.Hostname()
	}

	referrer = f.stripSubdomain(referrer)
	_, found := HostnameBlacklist[referrer]

	if found {
		request.BotReason = "ref-blacklist"
		return true, nil
	}

	// filter for bot keywords
	referrer = strings.ToLower(referrer)

	for _, botReferrer := range referrerBlacklist {
		if strings.Contains(referrer, botReferrer) {
			request.BotReason = "ref-keyword"
			return true, nil
		}
	}

	return false, nil
}

func (f *BotFilter) stripSubdomain(hostname string) string {
	if hostname == "" {
		return ""
	}

	runes := []rune(hostname)
	index := len(runes) - 1
	dots := 0

	for i := index; i > 0; i-- {
		if runes[i] == '.' {
			dots++

			if dots == 2 {
				index++
				break
			}
		}

		index--
	}

	return hostname[index:]
}
