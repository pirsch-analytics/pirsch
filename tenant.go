package pirsch

import "database/sql"

// NullTenant is the default (no) tenant.
var NullTenant = NewTenantID(0)

// NewTenantID is a helper function to return a sql.NullInt64.
// The ID is considered valid if greater than 0.
func NewTenantID(id int64) sql.NullInt64 {
	return sql.NullInt64{Int64: id, Valid: id > 0}
}
