package pirsch

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// FilterField is used to group results by fields.
type FilterField string

// FilterDirection is used to set the sorting direction.
type FilterDirection string

// FilterOrder is used to sort results by fields.
type FilterOrder struct {
	Field     FilterField
	Direction FilterDirection
}

var (
	Visitors = FilterField("visitors")
	Path     = FilterField("path")
	Language = FilterField("language")

	ASC  = FilterDirection("ASC")
	DESC = FilterDirection("DESC")
)

// Query is used to select results.
type Query struct {
	// TenantID is the optional tenant ID used to filter results.
	TenantID sql.NullInt64

	// From is the start of the selection.
	From time.Time

	// To is the end of the selection.
	To time.Time

	// Day is the day for the selection.
	Day time.Time

	// Path is the optional path for the selection.
	Path string

	// GroupBy groups the result set by given fields.
	GroupBy []FilterField

	// OrderBy sorts the result set by given fields.
	OrderBy []FilterOrder
}

// NewQuery returns a new query for given tenant.
func NewQuery(tenantID sql.NullInt64) *Query {
	return &Query{TenantID: tenantID}
}

func (query *Query) validate() {
	today := Today()

	if !query.From.IsZero() {
		query.From = time.Date(query.From.Year(), query.From.Month(), query.From.Day(), 0, 0, 0, 0, time.UTC)
	}

	if !query.To.IsZero() {
		query.To = time.Date(query.To.Year(), query.To.Month(), query.To.Day(), 0, 0, 0, 0, time.UTC)
	}

	if !query.To.IsZero() && query.From.After(query.To) {
		query.From, query.To = query.To, query.From
	}

	if !query.To.IsZero() && query.To.After(today) {
		query.To = today
	}

	if !query.Day.IsZero() {
		query.Day = time.Date(query.Day.Year(), query.Day.Month(), query.Day.Day(), 0, 0, 0, 0, time.UTC)
	}
}

func (query *Query) build() ([]interface{}, string) {
	args := make([]interface{}, 0, 5)
	var sqlQuery strings.Builder
	sqlQuery.WriteString(`SELECT count(DISTINCT fingerprint) "visitors", toDate("time") "day" `)
	groupBy := make([]string, 0, len(query.GroupBy))

	if len(query.GroupBy) > 0 {
		for _, group := range query.GroupBy {
			groupBy = append(groupBy, fmt.Sprintf(`"%s"`, group))
		}

		sqlQuery.WriteString(",")
		sqlQuery.WriteString(strings.Join(groupBy, ","))
		sqlQuery.WriteString(" ")
	}

	sqlQuery.WriteString(`FROM "hit" WHERE `)

	if query.TenantID.Valid {
		args = append(args, query.TenantID)
		sqlQuery.WriteString("tenant_id = ? ")
	} else {
		sqlQuery.WriteString("tenant_id IS NULL ")
	}

	if !query.From.IsZero() {
		args = append(args, query.From)
		sqlQuery.WriteString("AND time >= ? ")
	}

	if !query.To.IsZero() {
		args = append(args, query.To)
		sqlQuery.WriteString("AND time <= ? ")
	}

	if !query.Day.IsZero() {
		args = append(args, query.Day)
		sqlQuery.WriteString("AND toDate(time) = ? ")
	}

	if query.Path != "" {
		args = append(args, query.Path)
		sqlQuery.WriteString("AND path = ? ")
	}

	sqlQuery.WriteString(`GROUP BY "day" `)

	if len(query.GroupBy) > 0 {
		sqlQuery.WriteString(",")
		sqlQuery.WriteString(strings.Join(groupBy, ","))
		sqlQuery.WriteString(" ")
	}

	if len(query.OrderBy) > 0 {
		sqlQuery.WriteString("ORDER BY ")
		orderBy := make([]string, 0, len(query.OrderBy))

		for _, order := range query.OrderBy {
			orderBy = append(orderBy, fmt.Sprintf(`"%s" %s`, order.Field, order.Direction))
		}

		sqlQuery.WriteString(strings.Join(orderBy, ","))
	}

	return args, sqlQuery.String()
}
