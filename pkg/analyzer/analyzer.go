package analyzer

import (
	"context"
	"fmt"
	"github.com/pirsch-analytics/pirsch/v6/pkg"
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
)

// Analyzer provides an interface to analyze statistics.
type Analyzer struct {
	Visitors     Visitors
	Pages        Pages
	Demographics Demographics
	Device       Device
	UTM          UTM
	Events       Events
	Time         Time
	Options      FilterOptions
}

// NewAnalyzer returns a new Analyzer for given Store.
func NewAnalyzer(store db.Store) *Analyzer {
	analyzer := new(Analyzer)
	analyzer.Visitors = Visitors{
		analyzer: analyzer,
		store:    store,
	}
	analyzer.Pages = Pages{
		analyzer: analyzer,
		store:    store,
	}
	analyzer.Demographics = Demographics{
		analyzer: analyzer,
		store:    store,
	}
	analyzer.Device = Device{
		analyzer: analyzer,
		store:    store,
	}
	analyzer.UTM = UTM{
		analyzer: analyzer,
		store:    store,
	}
	analyzer.Events = Events{
		analyzer: analyzer,
		store:    store,
	}
	analyzer.Time = Time{
		analyzer: analyzer,
		store:    store,
	}
	analyzer.Options = FilterOptions{
		analyzer: analyzer,
		store:    store,
	}
	return analyzer
}

func (analyzer *Analyzer) timeOnPageQuery(filter *Filter) string {
	timeOnPage := "neighbor(duration_seconds, 1, 0)"

	if filter.MaxTimeOnPageSeconds > 0 {
		timeOnPage = fmt.Sprintf("least(neighbor(duration_seconds, 1, 0), %d)", filter.MaxTimeOnPageSeconds)
	}

	return timeOnPage
}

func (analyzer *Analyzer) selectByAttribute(filter *Filter, attr ...Field) (context.Context, string, []any) {
	fields := make([]Field, 0, len(attr)+2)
	fields = append(fields, attr...)
	fields = append(fields, FieldVisitors, FieldRelativeVisitors)
	orderBy := make([]Field, 0, len(attr)+1)
	orderBy = append(orderBy, FieldVisitors)
	orderBy = append(orderBy, attr...)
	filter = analyzer.getFilter(filter)
	query, args := filter.buildQuery(fields, attr, orderBy)
	return filter.Ctx, query, args
}

func (analyzer *Analyzer) getFilter(filter *Filter) *Filter {
	if filter == nil {
		filter = NewFilter(pkg.NullClient)
	}

	filter.validate()
	filterCopy := *filter
	return &filterCopy
}
