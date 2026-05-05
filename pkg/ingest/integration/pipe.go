package integration

import (
	"testing"

	"github.com/pirsch-analytics/pirsch/v7/pkg/db"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest/channel"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest/geo"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest/header"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest/ip"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest/language"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest/referrer"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest/screen"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest/session"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest/ua"
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest/utm"
	"github.com/stretchr/testify/assert"
)

func newPipe(t *testing.T) (*ingest.Pipe, *db.Mock, *session.MemCache) {
	s := db.NewMock()
	c := session.NewMemCache(s, 1000)
	ipFilter := ip.NewList()
	ipFilter.Update([]string{"89.123.21.128"}, nil, nil, nil, nil, nil)
	geoDB, _ := geo.NewGeo("", "", "")
	assert.NoError(t, geoDB.UpdateFromFile("../../../test/GeoIP2-City-Test.mmdb"))
	return ingest.NewPipe(ingest.PipeOptions{
		Storage: s,
	}).Use(header.NewBotFilter(),
		ip.NewIP(ip.DefaultHeaderParser, nil),
		ip.NewBotFilter([]ip.Filter{ipFilter}),
		referrer.NewBotFilter(),
		referrer.NewReferrer(referrer.Groups),
		ua.NewUserAgent(),
		ua.NewBotFilter(),
		geoDB,
		channel.NewChannel(channel.List),
		language.NewLanguage(),
		screen.NewScreen(screen.Classes),
		utm.NewUTM(),
		session.NewSession(1, 2, "salt", c, 200)), s, c
}
