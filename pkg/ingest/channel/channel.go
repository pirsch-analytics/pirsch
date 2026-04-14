package channel

import (
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest/util"
)

var (
	paidShoppingCampaignRegex = regexp.MustCompile("^(.*(([^a-df-z]|^)shop|shopping).*)$")
	paidMediumRegex           = regexp.MustCompile("^(.*cp.*|ppc|retargeting|paid.*)$")
	organicShoppingRegex      = regexp.MustCompile("^(.*(([^a-df-z]|^)shop|shopping).*)$")
	emailParams               = []string{"email", "e-mail", "e_mail", "e mail", "gmail"}
)

// Channel maps traffic sources to channels.
type Channel struct {
	search   []string
	social   []string
	shopping []string
	video    []string
	ai       []string
}

// NewChannel creates a new Channel for the given list of sources.
func NewChannel(list map[string]string) *Channel {
	searchChannel := make([]string, 0)
	socialChannel := make([]string, 0)
	shoppingChannel := make([]string, 0)
	videoChannel := make([]string, 0)
	aiChannel := make([]string, 0)

	for hostname, c := range list {
		switch c {
		case SourceCategorySearch:
			searchChannel = append(searchChannel, strings.ToLower(hostname))
			break
		case SourceCategorySocial:
			socialChannel = append(socialChannel, strings.ToLower(hostname))
			break
		case SourceCategoryShopping:
			shoppingChannel = append(shoppingChannel, strings.ToLower(hostname))
			break
		case SourceCategoryVideo:
			videoChannel = append(videoChannel, strings.ToLower(hostname))
			break
		case SourceCategoryAI:
			aiChannel = append(aiChannel, strings.ToLower(hostname))
			break
		default:
			panic(fmt.Sprintf("unknown channel type: %s", c))
		}
	}

	return &Channel{
		search:   searchChannel,
		social:   socialChannel,
		shopping: shoppingChannel,
		video:    videoChannel,
		ai:       aiChannel,
	}
}

// Step implements ingest.PipeStep to process a step.
func (c *Channel) Step(request *ingest.Request) (bool, error) {
	referrer := request.Referrer
	u, err := url.Parse(referrer)

	if err == nil && u.Hostname() != "" {
		referrer = util.StripWWW(u.Hostname())
	}

	referrer = strings.ToLower(referrer)
	referrerName := strings.ToLower(request.ReferrerName)
	utmMedium := strings.ToLower(request.UTMMedium)
	utmCampaign := strings.ToLower(request.UTMCampaign)
	utmSource := strings.ToLower(request.UTMSource)

	if strings.Contains(utmCampaign, "cross-network") {
		request.Channel = "Cross-network"
		return false, nil
	}

	isShoppingChannel := slices.Contains(c.shopping, referrer)
	isPaidMedium := paidMediumRegex.MatchString(utmMedium)

	if (isShoppingChannel || paidShoppingCampaignRegex.MatchString(utmCampaign)) && isPaidMedium {
		request.Channel = "Paid Shopping"
		return false, nil
	}

	isSearchChannel := slices.Contains(c.search, referrer) || slices.Contains(c.search, referrerName)

	if isSearchChannel && isPaidMedium ||
		referrerName == "google" && request.ClickID == "(gclid)" ||
		referrerName == "bing" && request.ClickID == "(msclkid)" {
		request.Channel = "Paid Search"
		return false, nil
	}

	isSocialChannel := slices.Contains(c.social, referrer)

	if isSocialChannel && isPaidMedium {
		request.Channel = "Paid Social"
		return false, nil
	}

	isVideoChannel := slices.Contains(c.video, referrer)

	if isVideoChannel && isPaidMedium {
		request.Channel = "Paid Video"
		return false, nil
	}

	if utmMedium == "display" ||
		utmMedium == "banner" ||
		utmMedium == "expandable" ||
		utmMedium == "interstitial" ||
		utmMedium == "cpm" {
		request.Channel = "Display"
		return false, nil
	}

	if isPaidMedium {
		request.Channel = "Paid Other"
		return false, nil
	}

	if isShoppingChannel || organicShoppingRegex.MatchString(utmCampaign) {
		request.Channel = "Organic Shopping"
		return false, nil
	}

	if isSocialChannel ||
		utmMedium == "social" ||
		utmMedium == "social-network" ||
		utmMedium == "social-media" ||
		utmMedium == "sm" ||
		utmMedium == "social network" ||
		utmMedium == "social media" {
		request.Channel = "Organic Social"
		return false, nil
	}

	if isVideoChannel || strings.Contains(utmMedium, "video") {
		request.Channel = "Organic Video"
		return false, nil
	}

	if isSearchChannel || utmMedium == "organic" {
		request.Channel = "Organic Search"
		return false, nil
	}

	if utmMedium == "referral" ||
		utmMedium == "app" ||
		utmMedium == "link" {
		request.Channel = "Referral"
		return false, nil
	}

	if slices.Contains(emailParams, referrer) ||
		slices.Contains(emailParams, utmSource) ||
		slices.Contains(emailParams, utmMedium) {
		request.Channel = "Email"
		return false, nil
	}

	if utmMedium == "affiliate" {
		request.Channel = "Affiliates"
		return false, nil
	}

	if utmMedium == "audio" {
		request.Channel = "Audio"
		return false, nil
	}

	if referrer == "sms" ||
		utmSource == "sms" ||
		utmMedium == "sms" {
		request.Channel = "SMS"
		return false, nil
	}

	if strings.HasSuffix(utmMedium, "push") ||
		strings.Contains(utmMedium, "mobile") ||
		strings.Contains(utmMedium, "notification") ||
		referrer == "firebase" ||
		utmSource == "firebase" {
		request.Channel = "Mobile Push Notifications"
		return false, nil
	}

	if slices.Contains(c.ai, referrer) ||
		slices.Contains(c.ai, utmSource) ||
		slices.Contains(c.ai, util.StripWWW(utmSource)) {
		request.Channel = "AI"
		return false, nil
	}

	request.Channel = "Direct"
	return false, nil
}
