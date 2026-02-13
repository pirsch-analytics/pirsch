.PHONY: test deps referrer referrer_blacklist ua

all: deps referrer referrer_blacklist ua browser test

test:
	go test -cover -race -p 1 github.com/pirsch-analytics/pirsch/v6/pkg/...

deps:
	go get -u -t ./...
	go mod vendor

fix:
	go fix ./...

referrer:
	go run scripts/update_referrer_list/update_referrer_list.go

referrer_blacklist:
	go run scripts/update_referrer_blacklist/update_referrer_blacklist.go

ua:
	go run scripts/update_ua_blacklist/update_ua_blacklist.go

browser:
	go run scripts/update_browser_blacklist/update_browser_blacklist.go
