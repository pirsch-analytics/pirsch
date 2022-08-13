#!/bin/bash

go test -cover -race github.com/pirsch-analytics/pirsch/v4/analyzer
go test -cover -race github.com/pirsch-analytics/pirsch/v4/db
go test -cover -race github.com/pirsch-analytics/pirsch/v4/tracker/geodb
go test -cover -race github.com/pirsch-analytics/pirsch/v4/tracker/ip
go test -cover -race github.com/pirsch-analytics/pirsch/v4/tracker/language
go test -cover -race github.com/pirsch-analytics/pirsch/v4/tracker/referrer
go test -cover -race github.com/pirsch-analytics/pirsch/v4/tracker/screen
go test -cover -race github.com/pirsch-analytics/pirsch/v4/tracker/session
go test -cover -race github.com/pirsch-analytics/pirsch/v4/tracker/ua
go test -cover -race github.com/pirsch-analytics/pirsch/v4/tracker/utm
go test -cover -race github.com/pirsch-analytics/pirsch/v4/tracker
go test -cover -race github.com/pirsch-analytics/pirsch/v4/util
