package query

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg/db"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/report"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/request"
)

// Query queries results for a report.Report.
type Query struct {
	db *db.ClickHouse
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

	return report.Report{
		Request: request,
	}
}
