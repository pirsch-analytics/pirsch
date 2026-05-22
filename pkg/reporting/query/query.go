package query

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg"
	"github.com/pirsch-analytics/pirsch/v7/pkg/db"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/report"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/request"
)

// Query queries results for a report.Report.
type Query struct {
	db *db.ClickHouse

	// TODO
	/*
		filter             *Filter
		fields             []Field
		fieldsImported     []Field
		from               table
		fromImported       string
		parent             *queryBuilder
		join               *queryBuilder
		joinSecond         *queryBuilder
		joinThird          *queryBuilder
		leftJoin           *queryBuilder
		joinStep           int
		search             []Search
		groupBy            []Field
		orderBy            []Field
		limit              int
		offset             int
		includeEventFilter bool
		sample             uint
		final              bool

		where []where
		q     strings.Builder
		args  []any
	*/
}

// NewQuery returns a new Query for given database connection.
func NewQuery(db *db.ClickHouse) *Query {
	return &Query{
		db: db,
	}
}

// Run runs given request.Request and returns the report.Report.
func (q *Query) Run(request request.Request) report.Report {
	if errs := request.Validate(); errs != nil {
		return report.Report{
			Meta: report.Meta{
				Errors: errs,
			},
		}
	}

	// TODO build query and generate report
	/*
		All metrics and dimensions → sessions    = query sessions only
		Any dimension  → page_views              = query page_views (+ subquery sessions if needed)
		Any dimension  → events                  = query events (+ subquery sessions if needed)
		Mix of page_views + sessions metrics     = subquery or join needed
	*/

	return report.Report{
		Request: request,
	}
}

func (q *Query) resolveTable(req request.Request) string {
	tables := map[string]bool{}

	for _, m := range req.Metrics {
		tables[m.Table()] = true
	}

	for _, d := range req.Dimensions {
		tables[d.Table()] = true
	}

	// TODO
	// check filter tables
	//collectFilterTables(req.Filter, tables)

	if tables[pkg.TableEvents] {
		return pkg.TableEvents
	}

	if tables[pkg.TablePageViews] {
		return pkg.TablePageViews
	}

	return pkg.TableSessions
}
