# TODO

[ ] Optimise request log (store all relevant fields)
[ ] Ignore bots
    [ ] User-Agent + browser version
    [ ] Firefox/version == rv:version
    [ ] Firefox no leading 0 in version number (regexp.MustCompile(`Firefox/(\d+\.\d+\.\d+)`))
    [ ] UUID + referrer
    [x] Sec-Fetch-Site: none + referrer set = bot (check same site hostname)
    [x] Upgrade-Insecure-Requests: 1 + Sec-Fetch-Mode: cors = bot
[ ] Pipeline integration tests
[ ] New reporting system
[ ] Allow setting required headers via API (override in request)
