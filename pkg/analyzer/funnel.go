package analyzer

import (
	"errors"
	"fmt"
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"log"
	"strings"
)

// Funnel aggregates funnels.
type Funnel struct {
	analyzer *Analyzer
	store    db.Store
}

// Steps returns the funnel steps for given filter list.
func (funnel *Funnel) Steps(filter []Filter) ([]model.FunnelStep, error) {
	if len(filter) < 2 {
		return nil, errors.New("not enough steps")
	}

	var query strings.Builder
	args := make([]any, 0)

	for i := range filter {
		f := funnel.analyzer.getFilter(&filter[i])
		fields := []Field{
			FieldClientID,
			FieldVisitorID,
			FieldSessionID,
			FieldTime,
		}
		q, a := f.buildQuery(fields, nil, nil)
		args = append(args, a...)
		query.WriteString(fmt.Sprintf("WITH step%d AS ( ", i))
		query.WriteString(q)
		query.WriteString(") ")

		if i != len(filter)-1 {
			query.WriteString(", ")
		}
	}

	query.WriteString("SELECT * FROM ( ")

	for i := 0; i < len(filter); i++ {
		query.WriteString(fmt.Sprintf("SELECT %d step, uniq(visitor_id) FROM step%d ", i, i))

		if i != len(filter)-1 {
			query.WriteString("UNION ALL ")
		}
	}

	query.WriteString(") ORDER BY step")
	log.Println(query.String())

	return nil, nil // TODO
}
