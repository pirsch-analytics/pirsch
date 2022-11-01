package analyzer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQuery(t *testing.T) {
	q := query{
		filter: &Filter{
			ClientID: 42,
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
	assert.Len(t, args, 1)
	assert.Equal(t, "SELECT country_code country_code,uniq(visitor_id) visitors,toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id) FROM session WHERE client_id = ? ), 1)) relative_visitors FROM session GROUP BY country_code HAVING sum(sign) > 0 ORDER BY visitors DESC,country_code ASC", queryStr)
}
