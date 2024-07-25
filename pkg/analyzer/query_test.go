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
		Referrer:    []string{"~Google", "^face"},
		Platform:    pkg.PlatformDesktop,
		PathPattern: []string{"/some/pattern"},
		Offset:      10,
		Limit:       99,
		Tags: map[string]string{
			"author": "John",
		},
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
	assert.Len(t, args, 27)
	expected := []any{
		int64(42),
		util.PastDay(7).Format(dateFormat),
		util.Today().Format(dateFormat),
		int64(42),
		util.PastDay(7).Format(dateFormat),
		util.Today().Format(dateFormat),
		"",
		"de",
		"ja",
		"%Google%",
		"%face%",
		int64(42),
		util.PastDay(7).Format(dateFormat),
		util.Today().Format(dateFormat),
		"/",
		"/foo",
		"/some/pattern",
		"author",
		"John",
		"",
		"de",
		"ja",
		"%Google%",
		"%face%",
		"%search%",
		util.PastDay(7).Format(dateFormat),
		util.Today().Format(dateFormat),
	}

	for i, arg := range args {
		assert.Equal(t, expected[i], arg)
	}

	assert.Equal(t, `SELECT toDate(time, 'UTC') day,country_code country_code,uniq(t.visitor_id) visitors,toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id) FROM "session" WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) ), 1)) relative_visitors FROM "page_view" t JOIN (SELECT t.visitor_id visitor_id,t.session_id session_id FROM "session" t WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND language = ? AND country_code = ? AND country_code != ? AND (ilike(referrer, ?) = 1 OR ilike(referrer, ?) != 1 ) AND desktop = 1 GROUP BY t.visitor_id,t.session_id HAVING sum(sign) > 0 ) j ON j.visitor_id = t.visitor_id AND j.session_id = t.session_id WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND (path = ? OR path = ? ) AND match("path", ?) = 1 AND tag_values[indexOf(tag_keys, ?)] = ? AND language = ? AND country_code = ? AND country_code != ? AND (ilike(referrer, ?) = 1 OR ilike(referrer, ?) != 1 ) AND desktop = 1 AND ilike(path, ?) = 1 GROUP BY country_code ORDER BY day ASC WITH FILL FROM toDate(?) TO toDate(?)+1 STEP INTERVAL 1 DAY ,visitors DESC,country_code ASC LIMIT 99 OFFSET 10 `, queryStr)
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
	assert.Equal(t, `SELECT path path,uniq(t.visitor_id) visitors FROM "page_view" t WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND path IN (?,?) GROUP BY path `, queryStr)
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
	assert.Equal(t, `SELECT uniqIf(visitor_id, desktop = 1) platform_desktop FROM "session" t WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) HAVING sum(sign) > 0 `, queryStr)
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
	assert.Equal(t, `SELECT toInt64OrDefault((SELECT uniq(t.visitor_id) visitors FROM "page_view" t JOIN (SELECT t.visitor_id visitor_id,t.session_id session_id FROM "session" t WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) GROUP BY t.visitor_id,t.session_id HAVING sum(sign) > 0 ) j ON j.visitor_id = t.visitor_id AND j.session_id = t.session_id WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND desktop = 1 AND mobile = 0 AND path = ? )) platform_desktop,toInt64OrDefault((SELECT uniq(t.visitor_id) visitors FROM "page_view" t JOIN (SELECT t.visitor_id visitor_id,t.session_id session_id FROM "session" t WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) GROUP BY t.visitor_id,t.session_id HAVING sum(sign) > 0 ) j ON j.visitor_id = t.visitor_id AND j.session_id = t.session_id WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND desktop = 0 AND mobile = 1 AND path = ? )) platform_mobile `, queryStr)
}

func TestQueryCustomMetricFloat(t *testing.T) {
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
			FieldEventMetaCustomMetricAvg,
			FieldEventMetaCustomMetricTotal,
			FieldVisitors,
		},
		from: events,
	}
	queryStr, args := q.query()
	assert.Len(t, args, 6)
	assert.Equal(t, "Custom Meta Value", args[0])
	assert.Equal(t, "Custom Meta Value", args[1])
	assert.Equal(t, "Event", args[5])
	assert.Equal(t, `SELECT ifNotFinite(avg(coalesce(toFloat64OrZero(event_meta_values[indexOf(event_meta_keys, ?)]))), 0) custom_metric_avg,sum(coalesce(toFloat64OrZero(event_meta_values[indexOf(event_meta_keys, ?)]))) custom_metric_total,uniq(t.visitor_id) visitors FROM "event" t WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND event_name = ? `, queryStr)
}

func TestQueryCustomMetricInt(t *testing.T) {
	filter := &Filter{
		ClientID:         42,
		From:             util.PastDay(7),
		To:               util.Today(),
		EventName:        []string{"Event"},
		CustomMetricType: pkg.CustomMetricTypeInteger,
		CustomMetricKey:  "Custom Meta Value",
	}
	q := queryBuilder{
		filter: filter,
		fields: []Field{
			FieldEventMetaCustomMetricAvg,
			FieldEventMetaCustomMetricTotal,
			FieldVisitors,
		},
		from: events,
	}
	queryStr, args := q.query()
	assert.Len(t, args, 6)
	assert.Equal(t, "Custom Meta Value", args[0])
	assert.Equal(t, "Custom Meta Value", args[1])
	assert.Equal(t, "Event", args[5])
	assert.Equal(t, `SELECT ifNotFinite(avg(coalesce(toInt64OrZero(event_meta_values[indexOf(event_meta_keys, ?)]))), 0) custom_metric_avg,sum(coalesce(toInt64OrZero(event_meta_values[indexOf(event_meta_keys, ?)]))) custom_metric_total,uniq(t.visitor_id) visitors FROM "event" t WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND event_name = ? `, queryStr)
}

func TestQuerySampling(t *testing.T) {
	filter := &Filter{
		ClientID: 42,
		From:     util.PastDay(7),
		To:       util.Today(),
	}
	q := queryBuilder{
		filter: filter,
		fields: []Field{
			FieldPath,
			FieldVisitors,
			FieldViews,
			FieldRelativeVisitors,
			FieldRelativeViews,
			FieldCR,
		},
		from: pageViews,
		groupBy: []Field{
			FieldPath,
		},
		sample: 10_000_000,
	}
	queryStr, args := q.query()
	assert.Len(t, args, 12)
	assert.Equal(t, `SELECT path path,toUInt64(greatest(uniq(t.visitor_id)*any(_sample_factor), 0)) visitors,toUInt64(greatest(count(1)*any(_sample_factor), 0)) views,toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id)*any(_sample_factor) FROM "session" SAMPLE 10000000 WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) ), 1)) relative_visitors,toFloat64OrDefault(views / greatest((SELECT sum(page_views*sign)*any(_sample_factor) views FROM "session" SAMPLE 10000000 WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) ), 1)) relative_views,toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id)*any(_sample_factor) FROM "session" SAMPLE 10000000 WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) ), 1)) cr FROM "page_view" t SAMPLE 10000000 WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) GROUP BY path `, queryStr)
}

func TestQueryPlatformSampling(t *testing.T) {
	filter := &Filter{
		ClientID: 42,
		From:     util.PastDay(7),
		To:       util.Today(),
	}
	q := queryBuilder{
		filter: filter,
		fields: []Field{
			FieldPlatformDesktop,
			FieldPlatformMobile,
			FieldPlatformUnknown,
		},
		from:   pageViews,
		sample: 10_000_000,
	}
	queryStr, args := q.query()
	assert.Len(t, args, 9)
	assert.Equal(t, `SELECT toInt64OrDefault((SELECT toUInt64(greatest(uniq(t.visitor_id)*any(_sample_factor), 0)) visitors FROM "page_view" t SAMPLE 10000000 WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND desktop = 1 AND mobile = 0 )) platform_desktop,toInt64OrDefault((SELECT toUInt64(greatest(uniq(t.visitor_id)*any(_sample_factor), 0)) visitors FROM "page_view" t SAMPLE 10000000 WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND desktop = 0 AND mobile = 1 )) platform_mobile,toInt64OrDefault((SELECT toUInt64(greatest(uniq(t.visitor_id)*any(_sample_factor), 0)) visitors FROM "page_view" t SAMPLE 10000000 WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND desktop = 0 AND mobile = 0 )) platform_unknown `, queryStr)
}

func TestQueryCustomMetricFloatSampling(t *testing.T) {
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
			FieldEventMetaCustomMetricAvg,
			FieldEventMetaCustomMetricTotal,
			FieldVisitors,
		},
		from:   events,
		sample: 10_000_000,
	}
	queryStr, args := q.query()
	assert.Len(t, args, 6)
	assert.Equal(t, `SELECT ifNotFinite(avg(coalesce(toFloat64OrZero(event_meta_values[indexOf(event_meta_keys, ?)]))), 0)*any(_sample_factor) custom_metric_avg,sum(coalesce(toFloat64OrZero(event_meta_values[indexOf(event_meta_keys, ?)])))*any(_sample_factor) custom_metric_total,toUInt64(greatest(uniq(t.visitor_id)*any(_sample_factor), 0)) visitors FROM "event" t SAMPLE 10000000 WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND event_name = ? `, queryStr)
}

func TestQueryCustomMetricIntSampling(t *testing.T) {
	filter := &Filter{
		ClientID:         42,
		From:             util.PastDay(7),
		To:               util.Today(),
		EventName:        []string{"Event"},
		CustomMetricType: pkg.CustomMetricTypeInteger,
		CustomMetricKey:  "Custom Meta Value",
	}
	q := queryBuilder{
		filter: filter,
		fields: []Field{
			FieldEventMetaCustomMetricAvg,
			FieldEventMetaCustomMetricTotal,
			FieldVisitors,
		},
		from:   events,
		sample: 10_000_000,
	}
	queryStr, args := q.query()
	assert.Len(t, args, 6)
	assert.Equal(t, `SELECT ifNotFinite(avg(coalesce(toInt64OrZero(event_meta_values[indexOf(event_meta_keys, ?)]))), 0)*any(_sample_factor) custom_metric_avg,toUInt64(greatest(sum(coalesce(toInt64OrZero(event_meta_values[indexOf(event_meta_keys, ?)])))*any(_sample_factor), 0)) custom_metric_total,toUInt64(greatest(uniq(t.visitor_id)*any(_sample_factor), 0)) visitors FROM "event" t SAMPLE 10000000 WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND event_name = ? `, queryStr)
}

func TestQuerySelectFieldPageViewsSampling(t *testing.T) {
	q := queryBuilder{
		from:   pageViews,
		sample: 10_000_000,
	}
	fields := []struct {
		field    Field
		expected string
	}{
		{FieldCount, "toUInt64(greatest(count(*)*any(_sample_factor), 0))"},
		{FieldEntries, "toUInt64(greatest(uniq(t.visitor_id, t.session_id)*any(_sample_factor), 0))"},
		{FieldExits, "toUInt64(greatest(uniq(t.visitor_id, t.session_id)*any(_sample_factor), 0))"},
		{FieldVisitors, "toUInt64(greatest(uniq(t.visitor_id)*any(_sample_factor), 0))"},
		{FieldVisitorsRaw, "toUInt64(greatest(uniq(visitor_id)*any(_sample_factor), 0))"},
		{FieldCRPeriod, "toFloat64OrDefault(visitors / greatest(ifNull(max(uvd.visitors), visitors), 1))*any(_sample_factor)"},
		{FieldSessions, "toUInt64(greatest(uniq(t.visitor_id, t.session_id)*any(_sample_factor), 0))"},
		{FieldViews, "toUInt64(greatest(count(1)*any(_sample_factor), 0))"},
		{FieldBounces, "toUInt64(greatest(uniqIf((t.visitor_id, t.session_id), bounces = 1)*any(_sample_factor), 0))"},
		{FieldEventTimeSpent, "toUInt64(greatest(toUInt64(greatest(ifNotFinite(avg(duration_seconds), 0), 0))*any(_sample_factor), 0))"},
		{FieldEventDurationSeconds, "toUInt64(greatest(sum(duration_seconds)*any(_sample_factor), 0))"},
	}

	for _, field := range fields {
		assert.Equal(t, field.expected, q.selectField(field.field))
	}
}

func TestQuerySelectFieldSessionsSampling(t *testing.T) {
	q := queryBuilder{
		from:   sessions,
		sample: 10_000_000,
	}
	fields := []struct {
		field    Field
		expected string
	}{
		{FieldCount, "toUInt64(greatest(count(*)*any(_sample_factor), 0))"},
		{FieldEntries, "toUInt64(greatest(sum(sign)*any(_sample_factor), 0))"},
		{FieldExits, "toUInt64(greatest(sum(sign)*any(_sample_factor), 0))"},
		{FieldVisitors, "toUInt64(greatest(uniq(t.visitor_id)*any(_sample_factor), 0))"},
		{FieldVisitorsRaw, "toUInt64(greatest(uniq(visitor_id)*any(_sample_factor), 0))"},
		{FieldCRPeriod, "toFloat64OrDefault(visitors / greatest(ifNull(max(uvd.visitors), visitors), 1))*any(_sample_factor)"},
		{FieldSessions, "toUInt64(greatest(uniq(t.visitor_id, t.session_id)*any(_sample_factor), 0))"},
		{FieldViews, "toUInt64(greatest(sum(page_views*sign)*any(_sample_factor), 0))"},
		{FieldBounces, "toUInt64(greatest(sum(is_bounce*sign)*any(_sample_factor), 0))"},
		{FieldEventTimeSpent, "toUInt64(greatest(toUInt64(greatest(ifNotFinite(avg(duration_seconds), 0), 0))*any(_sample_factor), 0))"},
		{FieldEventDurationSeconds, "toUInt64(greatest(sum(duration_seconds)*any(_sample_factor), 0))"},
	}

	for _, field := range fields {
		assert.Equal(t, field.expected, q.selectField(field.field))
	}
}

func TestQuerySelectFieldEventsSampling(t *testing.T) {
	q := queryBuilder{
		from:   events,
		sample: 10_000_000,
	}
	fields := []struct {
		field    Field
		expected string
	}{
		{FieldCount, "toUInt64(greatest(count(*)*any(_sample_factor), 0))"},
		{FieldEntries, "toUInt64(greatest(uniq(t.visitor_id, t.session_id)*any(_sample_factor), 0))"},
		{FieldExits, "toUInt64(greatest(uniq(t.visitor_id, t.session_id)*any(_sample_factor), 0))"},
		{FieldVisitors, "toUInt64(greatest(uniq(t.visitor_id)*any(_sample_factor), 0))"},
		{FieldVisitorsRaw, "toUInt64(greatest(uniq(visitor_id)*any(_sample_factor), 0))"},
		{FieldCRPeriod, "toFloat64OrDefault(visitors / greatest(ifNull(max(uvd.visitors), visitors), 1))*any(_sample_factor)"},
		{FieldSessions, "toUInt64(greatest(uniq(t.visitor_id, t.session_id)*any(_sample_factor), 0))"},
		{FieldViews, "toUInt64(greatest(sum(views)*any(_sample_factor), 0))"},
		{FieldBounces, "toUInt64(greatest(uniqIf((t.visitor_id, t.session_id), bounces = 1)*any(_sample_factor), 0))"},
		{FieldEventTimeSpent, "toUInt64(greatest(toUInt64(greatest(ifNotFinite(avg(duration_seconds), 0), 0))*any(_sample_factor), 0))"},
		{FieldEventDurationSeconds, "toUInt64(greatest(sum(duration_seconds)*any(_sample_factor), 0))"},
	}

	for _, field := range fields {
		assert.Equal(t, field.expected, q.selectField(field.field))
	}
}

func TestQueryImported(t *testing.T) {
	filter := &Filter{
		ClientID:      42,
		From:          util.PastDay(14),
		To:            util.Today(),
		ImportedUntil: util.PastDay(7),
		Country:       []string{"jp"},
	}
	filter.validate()
	q := queryBuilder{
		filter: filter,
		fields: []Field{
			FieldCountry,
			FieldVisitors,
			FieldRelativeVisitors,
		},
		fieldsImported: []Field{
			FieldCountry,
			FieldVisitors,
		},
		from:         sessions,
		fromImported: "imported_country",
		orderBy: []Field{
			FieldVisitors,
		},
		groupBy: []Field{
			FieldCountry,
		},
		limit: 10,
	}
	queryStr, args := q.query()
	assert.Len(t, args, 17)
	assert.Equal(t, `SELECT coalesce(nullif(t.country_code, ''), imp.country_code) country_code,sum(t.visitors + imp.visitors) visitors,toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id) FROM "session" WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) ) + (SELECT sum(visitors) FROM "imported_country" WHERE client_id = ? AND toDate(date, 'UTC') >= toDate(?) AND toDate(date, 'UTC') <= toDate(?) ), 1)) relative_visitors FROM (SELECT country_code country_code,uniq(t.visitor_id) visitors,toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id) FROM "session" WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) ), 1)) relative_visitors FROM "session" t WHERE client_id = ? AND toDate(time, 'UTC') >= toDate(?) AND toDate(time, 'UTC') <= toDate(?) AND country_code = ? GROUP BY country_code HAVING sum(sign) > 0 ORDER BY visitors DESC ) t FULL JOIN (SELECT country_code,visitors FROM "imported_country" WHERE client_id = ? AND toDate(date, 'UTC') >= toDate(?) AND toDate(date, 'UTC') <= toDate(?)  AND country_code = ? ) imp ON t.country_code = imp.country_code GROUP BY country_code ORDER BY visitors DESC LIMIT 10 `, queryStr)
}
