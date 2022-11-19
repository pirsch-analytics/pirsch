#!/bin/bash

go test -cover -race github.com/pirsch-analytics/pirsch/v5/analyzer
go test -cover -race github.com/pirsch-analytics/pirsch/v5/db
go test -cover -race github.com/pirsch-analytics/pirsch/v5/tracker/geodb
go test -cover -race github.com/pirsch-analytics/pirsch/v5/tracker/ip
go test -cover -race github.com/pirsch-analytics/pirsch/v5/tracker/referrer
go test -cover -race github.com/pirsch-analytics/pirsch/v5/tracker/session
go test -cover -race github.com/pirsch-analytics/pirsch/v5/tracker/ua
go test -cover -race github.com/pirsch-analytics/pirsch/v5/tracker
go test -cover -race github.com/pirsch-analytics/pirsch/v5/util
