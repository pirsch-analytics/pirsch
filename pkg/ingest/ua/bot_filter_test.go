package ua

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
	"github.com/stretchr/testify/assert"
)

func TestBotFilterUserAgent(t *testing.T) {
	userAgents := []struct {
		userAgent string
		ignore    string
	}{
		{"This is a bot request", "ua-keyword"},
		{"This is a crawler request", "ua-keyword"},
		{"This is a spider request", "ua-keyword"},
		{"Visit http://spam.com!", "ua-keyword"},
		{"", "ua-chars"},
		{"172.22.0.11:30004", "ua-ip"},
		{"172.22.0.11", "ua-chars"},
		{"2345:0425:2CA1:0000:0000:0567:5673:23b5", "ua-ip"},
		{"2345:425:2CA1:0000:0000:567:5673:23b5", "ua-ip"},
		{"2345:0425:2CA1:0:0:0567:5673:23b5", "ua-ip"},
		{"[2345:0425:2CA1:0:0:0567:5673:23b5]:8080", "ua-ip"},
		{"Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0", ""},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 17_1_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/21B101 Instagram 312.0.1.19.124 (iPhone14,2; iOS 17_1_2; de_FR; de; scale=3.00; 1170x2532; 548339486)", ""},
		{"Mozilla/5.0 (Linux; Android 9; SM-G950F Build/PPR1.180610.011; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/116.0.0.0 Mobile Safari/537.36 trill_2023102050 JsSdk/1.0 NetType/4G Channel/googleplay AppName/musical_ly app_version/31.2.5 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:31.0) Gecko/20130401 Firefox/31.0", "browser"},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 17_1_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 musical_ly_32.7.0 JsSdk/2.0 NetType/WIFI Channel/App Store ByteLocale/de Region/DE isDarkMode/0 WKWebView/1 RevealType/Dialog BytedanceWebview/d8a21c6 FalconTag/523CAEFB-209D-4BCF-A7A7-FEE8BD659140", "ua-keyword"},
		{"Mozilla/5.0 (Linux; Android 12; SM-G973F Build/SP1A.210812.016; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 11; SM-T500 Build/RP1A.200720.012; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.193 Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 14; SM-S918B Build/UP1A.231005.007; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 13; 22081212UG Build/TKQ1.220829.002; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 12; M2102J20SG Build/SKQ1.211006.001; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/ru-RU ByteFullLocale/ru-RU Region/RU AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 13; RMX3511 Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 14; SM-S901B Build/UP1A.231005.007; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 13; SM-A725F Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 trill_2021707000 JsSdk/1.0 NetType/WIFI Channel/tt_eu_samsung2020_yz1 AppName/musical_ly app_version/17.7.0 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE", ""},
		{"Mozilla/5.0 (Linux; Android 14; SM-S916B Build/UP1A.231005.007; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/MOBILE Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 13; SM-A225F Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 Instagram 312.1.0.34.111 Android (33/13; 300dpi; 720x1452; samsung; SM-A225F; a22; mt6769t; de_DE; 548323754)", ""},
		{"Mozilla/5.0 (Linux; Android 14; SM-A546B Build/UP1A.231005.007; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 14; SM-A336B Build/UP1A.231005.007; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 16_7_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 musical_ly_32.7.0 JsSdk/2.0 NetType/WIFI Channel/App Store ByteLocale/de Region/DE isDarkMode/0 WKWebView/1 RevealType/Dialog BytedanceWebview/d8a21c6 FalconTag/D99C3025-1798-4643-9FD5-00CFABA0DA30", "ua-keyword"},
		{"Mozilla/5.0 (Linux; Android 13; SM-A546B Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 13; SM-G780G Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 trill_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0", ""},
		{"Mozilla/5.0 (Linux; Android 13; 23053RN02Y Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/111.0.5563.116 Mobile Safari/537.36 Instagram 312.1.0.34.111 Android (33/13; 440dpi; 1080x2226; Xiaomi/Redmi; 23053RN02Y; heat; mt6768; es_US; 548323755)", ""},
		{"Mozilla/5.0 (Linux; Android 10; M2003J15SC Build/QP1A.190711.020; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/es ByteFullLocale/es Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 13; SM-A145R Build/TP1A.220624.014; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 trill_2023000000 JsSdk/1.0 NetType/WIFI Channel/samsung_preload AppName/musical_ly app_version/30.0.0 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 10; PPA-LX2 Build/HUAWEIPPA-LX2; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/92.0.4515.105 Mobile Safari/537.36 musical_ly_2023206050 JsSdk/1.0 NetType/WIFI Channel/huaweiadsglobal_int AppName/musical_ly app_version/32.6.5 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.7.4-bugfix AppVersion/32.6.5 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 14; SM-S911B Build/UP1A.231005.007; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 9; ZTE Blade A7 2020 Build/PPR1.180610.011; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 17_2_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 musical_ly_32.7.0 JsSdk/2.0 NetType/4G Channel/App Store ByteLocale/de Region/DE isDarkMode/0 WKWebView/1 RevealType/Dialog BytedanceWebview/d8a21c6 FalconTag/6C26B20B-D898-4AA5-9455-688897104628", "ua-keyword"},
		{"Mozilla/5.0 (Linux; Android 13; 21051182G Build/TKQ1.221013.002; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (Linux; Android 14; Pixel 6a Build/UP1A.231128.003; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.6099.144 Mobile Safari/537.36 musical_ly_2023208030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly app_version/32.8.3 ByteLocale/de-DE ByteFullLocale/de-DE Region/DE AppId/1233 Spark/1.4.8.3-bugfix AppVersion/32.8.3 BytedanceWebview/d8a21c6", ""},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 17_4_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/21E236 [FBAN/FBIOS;FBAV/465.0.1.41.103;FBBV/602060281;FBDV/iPhone13,4;FBMD/iPhone;FBSN/iOS;FBSV/17.4.1;FBSS/3;FBID/phone;FBLC/en_US;FBOP/5;FBRV/603032588]", ""},
		{"-8235OR 5208=5208", "ua-regex"},
		{"-4368OR 2918=6019 AND ('Veeg'='Veeg", "ua-regex"},
		{"-2985OR 6255=1124 AND ('lNxX' LIKE 'lNxX", "ua-regex"},
		{"(CASE WHEN 3116=9361 THEN 3116 ELSE NULL END)", "ua-regex"},
		{"IiF(3856=6771,3856,1/0)", "ua-regex"},
		{"Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; Perplexity-User/1.0; +https://perplexity.ai/perplexitybot-user)", "ua-keyword"},
		{"Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; PerplexityBot/1.0; +https://perplexity.ai/perplexitybot)", "ua-keyword"},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36; compatible; OAI-SearchBot/1.0; +https://openai.com/searchbot", "ua-keyword"},
		{"Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; ChatGPT-User/1.0; +https://openai.com/bot", "ua-keyword"},
		{"Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; GPTBot/1.2; +https://openai.com/gptbot)", "ua-keyword"},
		{"Mozilla/5.0 (compatible; PerplexityBot/1.0; +https://perplexity.ai)", "ua-keyword"},
		{"Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; GPTBot/1.0; +https://openai.com/gptbot)  ", "ua-keyword"},
		{"Anthropic-AI-Agent/1.0 (+https://www.anthropic.com/crawler)", "ua-keyword"},
		{"Mozilla/5.0 (compatible; ClaudeBot/1.0; +https://www.anthropic.com/claude)  ", "ua-keyword"},
		{"Mozilla/5.0 (Linux; Android 10; SM-G960F) AppleWebKit/537.36 (KHTML, like Gecko; compatible; Google-Bard/1.0; +https://bard.google.com)  ", "ua-keyword"},
		{"Cohere-Crawler/1.0 (https://cohere.com/crawl)", "ua-keyword"},
		{"Meta AI Bot/1.0 (+https://meta.ai/about)", "ua-keyword"},
		{"Mozilla/5.0 (compatible; LLaMA-Bot/1.0; +https://ai.meta.com/llama)", "ua-keyword"},
		{"Mozilla/5.0 (compatible; BingBot/2.0; +https://www.bing.com/bingbot.htm)", "ua-keyword"},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko; compatible; BingAI/1.0; +https://www.bing.com/bot)  ", "ua-keyword"},
		{"Mozilla/5.0 (compatible; JasperBot/1.0; +https://www.jasper.ai/bot)", "ua-keyword"},
		{"Mozilla/5.0 (compatible; MistralBot/1.0; +https://mistral.ai/bot)  ", "ua-keyword"},

		// rv does not match
		{"Mozilla/5.0 (X11; Linux x86_64; rv:148.0) Gecko/20100101 Firefox/149.0", "ua-rv-mismatch"},
	}
	u := NewUserAgent()
	f := NewBotFilter()

	for _, userAgent := range userAgents {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("User-Agent", userAgent.userAgent)
		r := &ingest.Request{
			Request: req,
		}
		cancel, err := u.Step(r)
		assert.NoError(t, err)
		assert.False(t, cancel)
		cancel, err = f.Step(r)
		assert.NoError(t, err)
		assert.Equalf(t, userAgent.ignore != "", cancel, userAgent.userAgent)
		assert.Equalf(t, userAgent.ignore, r.BotReason, userAgent.userAgent)
	}
}

func TestBotFilterBotUserAgent(t *testing.T) {
	u := NewUserAgent()
	f := NewBotFilter()

	for _, botUserAgent := range UserAgentBlacklist {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("User-Agent", botUserAgent)
		r := &ingest.Request{
			Request: req,
		}
		cancel, err := u.Step(r)
		assert.NoError(t, err)
		assert.False(t, cancel)
		cancel, err = f.Step(r)
		assert.NoError(t, err)
		assert.True(t, cancel)
		assert.NotEmptyf(t, r.BotReason, botUserAgent)
	}

	botUserAgent := []string{
		"-1' OR 2+990-990-1=0+0+0+1 or 'J2HdM1AB'='",
		"0'XOR(if(now()=sysdate(),sleep(15),0))XOR'Z",
		"1 waitfor delay '0:0:15' --",
		"14wpthYh' OR 294=(SELECT 294 FROM PG_SLEEP(15))--",
		"{{2959082-1}}",
		"{{ 2959082-1 }}",
		"7144de67-08ee-4fce-9997-49ef5af582d8",
		"a9c4b36c-71e8-4cdf-81a8-e178edcc7f30",
		"5b49398f-26bd-4bc3-ab16-3ca223d4d218",
	}

	for _, userAgent := range botUserAgent {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("User-Agent", userAgent)
		r := &ingest.Request{
			Request: req,
		}
		cancel, err := u.Step(r)
		assert.NoError(t, err)
		assert.False(t, cancel)
		cancel, err = f.Step(r)
		assert.NoError(t, err)
		assert.True(t, cancel)
		assert.NotEmptyf(t, r.BotReason, userAgent)
	}
}

func TestBotFilterBrowserCH(t *testing.T) {
	u := NewUserAgent()
	f := NewBotFilter()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36")
	req.Header.Set("Sec-CH-UA", `"Chromium";v="135", "Google Chrome";v="135", "HeadlessChrome";v="135", " Not;A Brand";v="99"`)
	r := &ingest.Request{
		Request: req,
	}
	cancel, err := u.Step(r)
	assert.NoError(t, err)
	assert.False(t, cancel)
	cancel, err = f.Step(r)
	assert.NoError(t, err)
	assert.True(t, cancel)
	assert.Equal(t, "ch-browser", r.BotReason)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36")
	req.Header.Set("Sec-CH-UA", `"Android WebView";v="135", " Not;A Brand";v="99"`)
	r = &ingest.Request{
		Request: req,
	}
	cancel, err = u.Step(r)
	assert.NoError(t, err)
	assert.False(t, cancel)
	cancel, err = f.Step(r)
	assert.NoError(t, err)
	assert.False(t, cancel)
	assert.Empty(t, r.BotReason)
}

func TestBotFilterBrowserVersion(t *testing.T) {
	u := NewUserAgent()
	f := NewBotFilter()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.4147.135 Safari/537.36")
	r := &ingest.Request{
		Request: req,
	}
	cancel, err := u.Step(r)
	assert.NoError(t, err)
	assert.False(t, cancel)
	cancel, err = f.Step(r)
	assert.NoError(t, err)
	assert.True(t, cancel)
	assert.Equal(t, "browser", r.BotReason)
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0")
	r = &ingest.Request{
		Request: req,
	}
	cancel, err = u.Step(r)
	assert.NoError(t, err)
	assert.False(t, cancel)
	cancel, err = f.Step(r)
	assert.NoError(t, err)
	assert.False(t, cancel)
	assert.Empty(t, r.BotReason)
}
