# TODO

[ ] Optimise request log (store all relevant fields)
[ ] Ignore bots
    [ ] User-Agent + browser version
    [ ] UUID + referrer
    [ ] Sec-Fetch-Site: none + referrer set = bot (check same site hostname)
    [ ] Upgrade-Insecure-Requests: 1 + Sec-Fetch-Mode: cors = bot
    [ ] Firefox/version == rv:version
    [ ] Firefox no leading 0 in version number (regexp.MustCompile(`Firefox/(\d+\.\d+\.\d+)`))
[ ] Pipeline integration tests
[ ] New reporting system
