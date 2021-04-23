#!/bin/bash

go test -cover github.com/pirsch-analytics/pirsch/analyze
go test -cover github.com/pirsch-analytics/pirsch/geodb
go test -cover github.com/pirsch-analytics/pirsch/hit
go test -cover github.com/pirsch-analytics/pirsch/ua
