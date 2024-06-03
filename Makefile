.PHONY: test deps referrer ua

all: deps referrer ua test

test:
	go test -cover -race -p 1 github.com/pirsch-analytics/pirsch/v6/pkg/analyzer
	go test -cover -race -p 1 github.com/pirsch-analytics/pirsch/v6/pkg/db
	go test -cover -race -p 1 github.com/pirsch-analytics/pirsch/v6/pkg/tracker/geodb
	go test -cover -race -p 1 github.com/pirsch-analytics/pirsch/v6/pkg/tracker/ip
	go test -cover -race -p 1 github.com/pirsch-analytics/pirsch/v6/pkg/tracker/referrer
	go test -cover -race -p 1 github.com/pirsch-analytics/pirsch/v6/pkg/tracker/session
	go test -cover -race -p 1 github.com/pirsch-analytics/pirsch/v6/pkg/tracker/ua
	go test -cover -race -p 1 github.com/pirsch-analytics/pirsch/v6/pkg/tracker
	go test -cover -race -p 1 github.com/pirsch-analytics/pirsch/v6/pkg/util

deps:
	go get -u -t ./...

referrer:
	go run scripts/update_referrer_list/update_referrer_list.go

ua:
	go run scripts/update_ua_blacklist/update_ua_blacklist.go
