package pirsch

import (
	"fmt"
	"strings"
)

// QueryField is a field that can be used in a Query.
type QueryField string

// QueryDirection are the sorting directions that can be used in a Query.
type QueryDirection string

// QueryOrder is the combination of a field and direction that can be used to sort a Query.
type QueryOrder struct {
	Field     QueryField
	Direction QueryDirection
}

var (
	FieldVisitors       = QueryField("visitors")
	FieldPath           = QueryField("path")
	FieldLanguage       = QueryField("language")
	FieldCountryCode    = QueryField("country_code")
	FieldBrowser        = QueryField("browser")
	FieldBrowserVersion = QueryField("browser_version")
	FieldOS             = QueryField("os")
	FieldOSVersion      = QueryField("os_version")

	ASC  = QueryDirection("ASC")
	DESC = QueryDirection("DESC")
)

// Query is used to build SQL queries and the corresponding parameters.
type Query struct {
	Filter
	fields  []string
	groupBy []QueryField
	orderBy []QueryOrder
}

// NewQuery creates a new empty query for given filter.
func NewQuery(filter *Filter) *Query {
	return &Query{
		Filter: *filter,
	}
}

// Fields adds given fields to the result set (including functions, name changes, ...).
func (query *Query) Fields(fields ...string) *Query {
	query.fields = append(query.fields, fields...)
	return query
}

// Group groups the result set by given fields.
func (query *Query) Group(fields ...QueryField) *Query {
	query.groupBy = append(query.groupBy, fields...)
	return query
}

// Order sorts the result set by given fields.
func (query *Query) Order(fields ...QueryOrder) *Query {
	query.orderBy = append(query.orderBy, fields...)
	return query
}

// Build builds the query and returns a list of parameters and the query itself.
func (query *Query) Build() ([]interface{}, string) {
	args := make([]interface{}, 0, 5)
	var sqlQuery strings.Builder
	sqlQuery.WriteString("SELECT ")
	sqlQuery.WriteString(strings.Join(query.fields, ","))
	sqlQuery.WriteString(" ")
	groupBy := make([]string, 0, len(query.groupBy))

	if len(query.groupBy) > 0 {
		for _, group := range query.groupBy {
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

	if len(query.groupBy) > 0 {
		sqlQuery.WriteString(`GROUP BY `)
		sqlQuery.WriteString(strings.Join(groupBy, ","))
		sqlQuery.WriteString(" ")
	}

	if len(query.orderBy) > 0 {
		sqlQuery.WriteString("ORDER BY ")
		orderBy := make([]string, 0, len(query.orderBy))

		for _, order := range query.orderBy {
			orderBy = append(orderBy, fmt.Sprintf(`"%s" %s`, order.Field, order.Direction))
		}

		sqlQuery.WriteString(strings.Join(orderBy, ","))
	}

	return args, sqlQuery.String()
}
