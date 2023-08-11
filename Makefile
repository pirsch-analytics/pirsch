.PHONY: test deps referrer ua

all: deps referrer ua test

test:
	go test -cover -race github.com/pirsch-analytics/pirsch/v5/analyzer
	go test -cover -race github.com/pirsch-analytics/pirsch/v5/db
	go test -cover -race github.com/pirsch-analytics/pirsch/v5/tracker/geodb
	go test -cover -race github.com/pirsch-analytics/pirsch/v5/tracker/ip
	go test -cover -race github.com/pirsch-analytics/pirsch/v5/tracker/referrer
	go test -cover -race github.com/pirsch-analytics/pirsch/v5/tracker/session
	go test -cover -race github.com/pirsch-analytics/pirsch/v5/tracker/ua
	go test -cover -race github.com/pirsch-analytics/pirsch/v5/tracker
	go test -cover -race github.com/pirsch-analytics/pirsch/v5/util

deps:
	go get -u -t ./...

referrer:
	go run scripts/update_referrer_list/update_referrer_list.go

ua:
	go run scripts/update_ua_blacklist/update_ua_blacklist.go
