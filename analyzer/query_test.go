package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v4"
	"github.com/pirsch-analytics/pirsch/v4/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQuery(t *testing.T) {
	filter := &Filter{
		ClientID:    42,
		From:        util.PastDay(7),
		To:          util.Today(),
		Path:        []string{"/", "/foo"},
		Country:     []string{"de", "!ja"},
		Language:    []string{"nUlL"},
		Referrer:    []string{"~Google"},
		Platform:    pirsch.PlatformDesktop,
		PathPattern: []string{"/some/pattern"},
		Offset:      10,
		Limit:       99,
		minIsBot:    5,
	}
	q := query{
		filter: filter,
		fields: []Field{
			FieldDay,
			FieldCountry,
			FieldVisitors,
			FieldRelativeVisitors,
		},
		from: pageViews,
		join: &query{
			filter: filter,
			fields: []Field{
				FieldVisitorID,
				FieldSessionID,
			},
			from: sessions,
			groupBy: []Field{
				FieldVisitorID,
				FieldSessionID,
			},
		},
		search: []Search{{FieldPath, "search"}},
		orderBy: []Field{
			FieldDay,
			FieldVisitors,
			FieldCountry,
		},
		groupBy: []Field{
			FieldCountry,
		},
		offset: 10,
		limit:  99,
	}
	queryStr, args := q.query()
	assert.Len(t, args, 24)
	assert.Equal(t, int64(42), args[0])
	assert.Equal(t, util.PastDay(7).Format(dateFormat), args[1])
	assert.Equal(t, util.Today().Format(dateFormat), args[2])
	assert.Equal(t, int64(42), args[3])
	assert.Equal(t, util.PastDay(7).Format(dateFormat), args[4])
	assert.Equal(t, util.Today().Format(dateFormat), args[5])
	assert.Equal(t, uint8(5), args[6])
	assert.Equal(t, "", args[7])
	assert.Equal(t, "de", args[8])
	assert.Equal(t, "ja", args[9]) // null
	assert.Equal(t, "%Google%", args[10])
	assert.Equal(t, int64(42), args[11])
	assert.Equal(t, util.PastDay(7).Format(dateFormat), args[12])
	assert.Equal(t, util.Today().Format(dateFormat), args[13])
	assert.Equal(t, "/", args[14])
	assert.Equal(t, "/foo", args[15])
	assert.Equal(t, "/some/pattern", args[16])
	assert.Equal(t, "", args[17])
	assert.Equal(t, "de", args[18])
	assert.Equal(t, "ja", args[19])
	assert.Equal(t, "%Google%", args[20])
	assert.Equal(t, "%search%", args[21])
	assert.Equal(t, util.PastDay(7).Format(dateFormat), args[22])
	assert.Equal(t, util.Today().Format(dateFormat), args[23])
	assert.Equal(t, "SELECT toDate(time, 'UTC') day,country_code country_code,uniq(visitor_id) visitors,toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id) FROM session WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) ), 1)) relative_visitors FROM page_view t JOIN (SELECT visitor_id visitor_id,session_id session_id FROM session t WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?)  AND is_bot < ? AND language = ? AND country_code = ? AND country_code != ? AND ilike(referrer, ?) = 1 AND desktop = 1 GROUP BY visitor_id,session_id HAVING sum(sign) > 0 ) j ON j.visitor_id = t.visitor_id AND j.session_id = t.session_id WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND (path = ? OR path = ? ) AND match(\"path\", ?) = 1 AND language = ? AND country_code = ? AND country_code != ? AND ilike(referrer, ?) = 1 AND desktop = 1 AND ilike(path, ?) = 1 GROUP BY country_code ORDER BY day ASC WITH FILL FROM toDate(?) TO toDate(?)+1 STEP INTERVAL 1 DAY ,visitors DESC,country_code ASC LIMIT 99 OFFSET 10 ", queryStr)
}

func TestQueryAnyPath(t *testing.T) {
	filter := &Filter{
		ClientID: 42,
		From:     util.PastDay(7),
		To:       util.Today(),
		AnyPath:  []string{"/", "/foo"},
	}
	q := query{
		filter: filter,
		fields: []Field{
			FieldPath,
			FieldVisitors,
		},
		from: pageViews,
		groupBy: []Field{
			FieldPath,
		},
	}
	queryStr, args := q.query()
	assert.Len(t, args, 5)
	assert.Equal(t, int64(42), args[0])
	assert.Equal(t, util.PastDay(7).Format(dateFormat), args[1])
	assert.Equal(t, util.Today().Format(dateFormat), args[2])
	assert.Equal(t, "/", args[3])
	assert.Equal(t, "/foo", args[4])
	assert.Equal(t, "SELECT path path,uniq(visitor_id) visitors FROM page_view t WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND path IN (?,?) GROUP BY path ", queryStr)
}
