package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v6/pkg"
	"github.com/pirsch-analytics/pirsch/v6/pkg/util"
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
		Platform:    pkg.PlatformDesktop,
		PathPattern: []string{"/some/pattern"},
		Offset:      10,
		Limit:       99,
	}
	q := queryBuilder{
		filter: filter,
		fields: []Field{
			FieldDay,
			FieldCountry,
			FieldVisitors,
			FieldRelativeVisitors,
		},
		from: pageViews,
		join: &queryBuilder{
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
	assert.Len(t, args, 23)
	assert.Equal(t, int64(42), args[0])
	assert.Equal(t, util.PastDay(7).Format(dateFormat), args[1])
	assert.Equal(t, util.Today().Format(dateFormat), args[2])
	assert.Equal(t, int64(42), args[3])
	assert.Equal(t, util.PastDay(7).Format(dateFormat), args[4])
	assert.Equal(t, util.Today().Format(dateFormat), args[5])
	assert.Equal(t, "", args[6])
	assert.Equal(t, "de", args[7])
	assert.Equal(t, "ja", args[8]) // null
	assert.Equal(t, "%Google%", args[9])
	assert.Equal(t, int64(42), args[10])
	assert.Equal(t, util.PastDay(7).Format(dateFormat), args[11])
	assert.Equal(t, util.Today().Format(dateFormat), args[12])
	assert.Equal(t, "/", args[13])
	assert.Equal(t, "/foo", args[14])
	assert.Equal(t, "/some/pattern", args[15])
	assert.Equal(t, "", args[16])
	assert.Equal(t, "de", args[17])
	assert.Equal(t, "ja", args[18])
	assert.Equal(t, "%Google%", args[19])
	assert.Equal(t, "%search%", args[20])
	assert.Equal(t, util.PastDay(7).Format(dateFormat), args[21])
	assert.Equal(t, util.Today().Format(dateFormat), args[22])
	assert.Equal(t, "SELECT toDate(time, 'UTC') day,country_code country_code,uniq(t.visitor_id) visitors,toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id) FROM session WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) ), 1)) relative_visitors FROM page_view t JOIN (SELECT visitor_id visitor_id,session_id session_id FROM session t WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND language = ? AND country_code = ? AND country_code != ? AND ilike(referrer, ?) = 1 AND desktop = 1 GROUP BY visitor_id,session_id HAVING sum(sign) > 0 ) j ON j.visitor_id = t.visitor_id AND j.session_id = t.session_id WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND (path = ? OR path = ? ) AND match(\"path\", ?) = 1 AND language = ? AND country_code = ? AND country_code != ? AND ilike(referrer, ?) = 1 AND desktop = 1 AND ilike(path, ?) = 1 GROUP BY country_code ORDER BY day ASC WITH FILL FROM toDate(?) TO toDate(?)+1 STEP INTERVAL 1 DAY ,visitors DESC,country_code ASC LIMIT 99 OFFSET 10 ", queryStr)
}

func TestQueryAnyPath(t *testing.T) {
	q := queryBuilder{
		filter: &Filter{
			ClientID: 42,
			From:     util.PastDay(7),
			To:       util.Today(),
			AnyPath:  []string{"/", "/foo"},
		},
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
	assert.Equal(t, "SELECT path path,uniq(t.visitor_id) visitors FROM page_view t WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND path IN (?,?) GROUP BY path ", queryStr)
}

func TestQueryPlatformSession(t *testing.T) {
	q := queryBuilder{
		filter: &Filter{
			ClientID: 42,
			From:     util.PastDay(7),
			To:       util.Today(),
		},
		fields: []Field{
			FieldPlatformDesktop,
		},
		from: sessions,
	}
	queryStr, args := q.query()
	assert.Len(t, args, 3)
	assert.Equal(t, int64(42), args[0])
	assert.Equal(t, util.PastDay(7).Format(dateFormat), args[1])
	assert.Equal(t, util.Today().Format(dateFormat), args[2])
	assert.Equal(t, "SELECT uniqIf(visitor_id, desktop = 1) platform_desktop FROM session t WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) HAVING sum(sign) > 0 ", queryStr)
}

func TestQueryPlatformPageView(t *testing.T) {
	filter := &Filter{
		ClientID: 42,
		From:     util.PastDay(7),
		To:       util.Today(),
		Path:     []string{"/foo"},
	}
	q := queryBuilder{
		filter: filter,
		fields: []Field{
			FieldPlatformDesktop,
			FieldPlatformMobile,
		},
		from: pageViews,
		join: &queryBuilder{
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
	}
	queryStr, args := q.query()
	assert.Len(t, args, 14)
	assert.Equal(t, int64(42), args[0])
	assert.Equal(t, util.PastDay(7).Format(dateFormat), args[1])
	assert.Equal(t, util.Today().Format(dateFormat), args[2])
	assert.Equal(t, int64(42), args[3])
	assert.Equal(t, util.PastDay(7).Format(dateFormat), args[4])
	assert.Equal(t, util.Today().Format(dateFormat), args[5])
	assert.Equal(t, "/foo", args[6])
	assert.Equal(t, int64(42), args[7])
	assert.Equal(t, util.PastDay(7).Format(dateFormat), args[8])
	assert.Equal(t, util.Today().Format(dateFormat), args[9])
	assert.Equal(t, int64(42), args[10])
	assert.Equal(t, util.PastDay(7).Format(dateFormat), args[11])
	assert.Equal(t, util.Today().Format(dateFormat), args[12])
	assert.Equal(t, "/foo", args[13])
	assert.Equal(t, "SELECT toInt64OrDefault((SELECT uniq(t.visitor_id) visitors FROM page_view t JOIN (SELECT visitor_id visitor_id,session_id session_id FROM session t WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) GROUP BY visitor_id,session_id HAVING sum(sign) > 0 ) j ON j.visitor_id = t.visitor_id AND j.session_id = t.session_id WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND desktop = 1 AND mobile = 0 AND path = ? )) platform_desktop,toInt64OrDefault((SELECT uniq(t.visitor_id) visitors FROM page_view t JOIN (SELECT visitor_id visitor_id,session_id session_id FROM session t WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) GROUP BY visitor_id,session_id HAVING sum(sign) > 0 ) j ON j.visitor_id = t.visitor_id AND j.session_id = t.session_id WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND desktop = 0 AND mobile = 1 AND path = ? )) platform_mobile ", queryStr)
}

func TestQueryCustomMetric(t *testing.T) {
	filter := &Filter{
		ClientID:         42,
		From:             util.PastDay(7),
		To:               util.Today(),
		EventName:        []string{"Event"},
		CustomMetricType: pkg.CustomMetricTypeFloat,
		CustomMetricKey:  "Custom Meta Value",
	}
	q := queryBuilder{
		filter: filter,
		fields: []Field{
			FieldEventMetaCustomMetric,
			FieldVisitors,
		},
		from: events,
	}
	queryStr, args := q.query()
	assert.Len(t, args, 5)
	assert.Equal(t, "Custom Meta Value", args[0])
	assert.Equal(t, "Event", args[4])
	assert.Equal(t, "SELECT coalesce(toFloat64OrZero(event_meta_values[indexOf(event_meta_keys, ?)])) custom_metric,uniq(t.visitor_id) visitors FROM event t WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND event_name = ? ", queryStr)
}
