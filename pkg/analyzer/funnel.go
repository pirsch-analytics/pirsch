package analyzer

import (
	"context"
	"errors"
	"fmt"
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"strings"
)

const (
	minFunnelSteps = 2
	maxFunnelSteps = 8
)

// Funnel aggregates funnels.
type Funnel struct {
	analyzer *Analyzer
	store    db.Store
}

// Steps returns the funnel steps for given filter list.
func (funnel *Funnel) Steps(ctx context.Context, filter []Filter) ([]model.FunnelStep, error) {
	if len(filter) < minFunnelSteps {
		return nil, errors.New("not enough steps")
	} else if len(filter) > maxFunnelSteps {
		return nil, errors.New("too many steps")
	}

	var query strings.Builder
	args := make([]any, 0)

	for i := range filter {
		f := funnel.analyzer.getFilter(&filter[i])
		f.funnelStep = i + 1
		fields := []Field{
			FieldClientID,
			FieldVisitorID,
			FieldSessionID,
			FieldTime,
		}
		q, a := f.buildQuery(fields, nil, nil, nil, "")
		args = append(args, a...)

		if i == 0 {
			query.WriteString(fmt.Sprintf("WITH step%d AS ( ", i+1))
		} else {
			query.WriteString(fmt.Sprintf("step%d AS ( ", i+1))
		}

		query.WriteString(q)
		query.WriteString(") ")

		if i != len(filter)-1 {
			query.WriteString(", ")
		}
	}

	query.WriteString("SELECT * FROM ( ")

	for i := 0; i < len(filter); i++ {
		query.WriteString(fmt.Sprintf("SELECT %d step, uniq(visitor_id) FROM step%d ", i+1, i+1))

		if i != len(filter)-1 {
			query.WriteString("UNION ALL ")
		}
	}

	query.WriteString(") ORDER BY step")
	stats, err := funnel.store.SelectFunnelSteps(ctx, query.String(), args...)

	if err != nil {
		return nil, err
	}

	for i := range stats {
		if i > 0 {
			if stats[0].Visitors > 0 {
				stats[i].RelativeVisitors = float64(stats[i].Visitors) / float64(stats[0].Visitors)
			}

			stats[i].PreviousVisitors = stats[i-1].Visitors
			stats[i].RelativePreviousVisitors = stats[i-1].RelativeVisitors
			stats[i].Dropped = stats[i].PreviousVisitors - stats[i].Visitors

			if stats[i].PreviousVisitors > 0 {
				stats[i].DropOff = 1 - float64(stats[i].Visitors)/float64(stats[i].PreviousVisitors)
			}
		} else if stats[i].Visitors > 0 {
			stats[i].RelativeVisitors = 1
		}
	}

	return stats, nil
}
