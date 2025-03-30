package channel

import (
	"net/url"
	"regexp"
	"slices"
	"strings"
)

var (
	paidShoppingCampaignRegex = regexp.MustCompile("^(.*(([^a-df-z]|^)shop|shopping).*)$")
	paidMediumRegex           = regexp.MustCompile("^(.*cp.*|ppc|retargeting|paid.*)$")
	organicShoppingRegex      = regexp.MustCompile("^(.*(([^a-df-z]|^)shop|shopping).*)$")
	emailParams               = []string{"email", "e-mail", "e_mail", "e mail"}
)

// Get returns the acquisition channel.
func Get(referrer, referrerName, utmMedium, utmCampaign, utmSource, clickID string) string {
	u, err := url.Parse(referrer)

	if err == nil && u.Hostname() != "" {
		referrer = u.Hostname()
	}

	utmMedium = strings.ToLower(utmMedium)
	utmCampaign = strings.ToLower(utmCampaign)
	utmSource = strings.ToLower(utmSource)

	if strings.Contains(utmCampaign, "cross-network") {
		return "Cross-network"
	}

	isShoppingChannel := slices.Contains(shoppingChannel, referrer)
	isPaidMedium := paidMediumRegex.MatchString(utmMedium)

	if (isShoppingChannel || paidShoppingCampaignRegex.MatchString(utmCampaign)) && isPaidMedium {
		return "Paid Shopping"
	}

	isSearchChannel := slices.Contains(searchChannel, referrer)

	if isSearchChannel && isPaidMedium ||
		referrerName == "Google" && clickID == "(gclid)" ||
		referrerName == "Bing" && clickID == "(msclkid)" {
		return "Paid Search"
	}

	isSocialChannel := slices.Contains(socialChannel, referrer)

	if isSocialChannel && isPaidMedium {
		return "Paid Social"
	}

	isVideoChannel := slices.Contains(videoChannel, referrer)

	if isVideoChannel && isPaidMedium {
		return "Paid Video"
	}

	if utmMedium == "display" ||
		utmMedium == "banner" ||
		utmMedium == "expandable" ||
		utmMedium == "interstitial" ||
		utmMedium == "cpm" {
		return "Display"
	}

	if isPaidMedium {
		return "Paid Other"
	}

	if isShoppingChannel || organicShoppingRegex.MatchString(utmCampaign) {
		return "Organic Shopping"
	}

	if isSocialChannel ||
		utmMedium == "social" ||
		utmMedium == "social-network" ||
		utmMedium == "social-media" ||
		utmMedium == "sm" ||
		utmMedium == "social network" ||
		utmMedium == "social media" {
		return "Organic Social"
	}

	if isVideoChannel || strings.Contains(utmMedium, "video") {
		return "Organic Video"
	}

	if isSearchChannel || utmMedium == "organic" {
		return "Organic Search"
	}

	if utmMedium == "referral" ||
		utmMedium == "app" ||
		utmMedium == "link" {
		return "Referral"
	}

	if slices.Contains(emailParams, referrer) ||
		slices.Contains(emailParams, utmSource) ||
		slices.Contains(emailParams, utmMedium) {
		return "Email"
	}

	if utmMedium == "affiliate" {
		return "Affiliates"
	}

	if utmMedium == "audio" {
		return "Audio"
	}

	if referrer == "sms" ||
		utmSource == "sms" ||
		utmMedium == "sms" {
		return "SMS"
	}

	if strings.HasSuffix(utmMedium, "push") ||
		strings.Contains(utmMedium, "mobile") ||
		strings.Contains(utmMedium, "notification") ||
		referrer == "firebase" ||
		utmSource == "firebase" {
		return "Mobile Push Notifications"
	}

	if slices.Contains(aiChannel, referrer) || slices.Contains(aiChannel, utmSource) {
		return "AI"
	}

	return "Direct"
}
