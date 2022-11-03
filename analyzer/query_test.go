package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v4/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQuery(t *testing.T) {
	q := query{
		filter: &Filter{
			ClientID: 42,
			From:     util.PastDay(7),
			To:       util.Today(),
			Path:     []string{"/", "/foo"},
			Country:  []string{"de", "!ja"},
			Search:   []Search{{FieldPath, "search"}},
		},
		fields: []Field{
			FieldCountry,
			FieldVisitors,
			FieldRelativeVisitors,
		},
		from: sessions,
		orderBy: []Field{
			FieldVisitors,
			FieldCountry,
		},
		groupBy: []Field{
			FieldCountry,
		},
	}
	queryStr, args := q.query()
	assert.Len(t, args, 11)
	assert.Equal(t, int64(42), args[0])
	assert.Equal(t, util.PastDay(7).Format(dateFormat), args[1])
	assert.Equal(t, util.Today().Format(dateFormat), args[2])
	assert.Equal(t, int64(42), args[3])
	assert.Equal(t, util.PastDay(7).Format(dateFormat), args[4])
	assert.Equal(t, util.Today().Format(dateFormat), args[5])
	assert.Equal(t, "/", args[6])
	assert.Equal(t, "/foo", args[7])
	assert.Equal(t, "de", args[8])
	assert.Equal(t, "ja", args[9])
	assert.Equal(t, "%search%", args[10])
	assert.Equal(t, "SELECT country_code country_code,uniq(visitor_id) visitors,toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id) FROM session WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) ), 1)) relative_visitors FROM session WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND (path = ? OR path = ? ) AND country_code = ? AND country_code != ? AND ilike(path, ?) = 1 GROUP BY country_code HAVING sum(sign) > 0 ORDER BY visitors DESC,country_code ASC", queryStr)
}
