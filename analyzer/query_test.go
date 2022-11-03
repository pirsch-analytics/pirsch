package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v4"
	"github.com/pirsch-analytics/pirsch/v4/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQuery(t *testing.T) {
	q := query{
		filter: &Filter{
			ClientID:    42,
			From:        util.PastDay(7),
			To:          util.Today(),
			Path:        []string{"/", "/foo"},
			Country:     []string{"de", "!ja"},
			Language:    []string{"nUlL"},
			Referrer:    []string{"~Google"},
			Platform:    pirsch.PlatformDesktop,
			PathPattern: []string{"/some/pattern"},
			Search:      []Search{{FieldPath, "search"}},
			Offset:      10,
			Limit:       99,
			minIsBot:    5,
		},
		fields: []Field{
			FieldDay,
			FieldCountry,
			FieldVisitors,
			FieldRelativeVisitors,
		},
		from: sessions,
		orderBy: []Field{
			FieldDay,
			FieldVisitors,
			FieldCountry,
		},
		groupBy: []Field{
			FieldCountry,
		},
	}
	queryStr, args := q.query()
	assert.Len(t, args, 17)
	assert.Equal(t, int64(42), args[0])
	assert.Equal(t, util.PastDay(7).Format(dateFormat), args[1])
	assert.Equal(t, util.Today().Format(dateFormat), args[2])
	assert.Equal(t, int64(42), args[3])
	assert.Equal(t, util.PastDay(7).Format(dateFormat), args[4])
	assert.Equal(t, util.Today().Format(dateFormat), args[5])
	assert.Equal(t, uint8(5), args[6])
	assert.Equal(t, "/", args[7])
	assert.Equal(t, "/foo", args[8])
	assert.Equal(t, "", args[9]) // null
	assert.Equal(t, "de", args[10])
	assert.Equal(t, "ja", args[11])
	assert.Equal(t, "%Google%", args[12])
	assert.Equal(t, "/some/pattern", args[13])
	assert.Equal(t, "%search%", args[14])
	assert.Equal(t, util.PastDay(7).Format(dateFormat), args[15])
	assert.Equal(t, util.Today().Format(dateFormat), args[16])
	assert.Equal(t, "SELECT toDate(time, 'UTC') day,country_code country_code,uniq(visitor_id) visitors,toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id) FROM session WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) ), 1)) relative_visitors FROM session WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?)  AND is_bot < ? AND (path = ? OR path = ? ) AND language = ? AND country_code = ? AND country_code != ? AND ilike(referrer, ?) = 1 AND desktop = 1 AND match(\"path\", ?) = 1 AND ilike(path, ?) = 1 GROUP BY country_code HAVING sum(sign) > 0 ORDER BY day ASC WITH FILL FROM toDate(?) TO toDate(?)+1 STEP INTERVAL 1 DAY ,visitors DESC,country_code ASC LIMIT 99 OFFSET 10 ", queryStr)
}
