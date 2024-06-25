package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"log"
)

// Funnel aggregates funnels.
type Funnel struct {
	analyzer *Analyzer
	store    db.Store
}

// Steps returns the funnel steps for given filter list.
func (funnel *Funnel) Steps(filter []Filter) {
	if len(filter) == 0 {
		return
	}

	for i := range filter {
		filter[i] = *funnel.analyzer.getFilter(&filter[i])
		fields := []Field{
			FieldVisitors,
		}
		q, args := filter[i].buildQuery(fields, nil, nil)
		log.Println(q, args)
	}
}
