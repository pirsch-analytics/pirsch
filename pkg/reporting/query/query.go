package query

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/pirsch-analytics/pirsch/v7/pkg"
	"github.com/pirsch-analytics/pirsch/v7/pkg/db"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/dimensions"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/metrics"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/report"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/request"
)

var (
	tablePriority = map[string]int{
		pkg.TableSessions:  1,
		pkg.TablePageViews: 2,
		pkg.TableEvents:    2,
	}
	likeEscaper = strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`)
)

type joinQueries struct {
	primaryQuery        string
	primaryArgs         []any
	primaryMetrics      []metrics.Metric
	secondaryQuery      string
	secondaryArgs       []any
	secondaryMetrics    []metrics.Metric
	sortingByJoinMetric bool
}

type classifiedFilter struct {
	table  string
	filter request.Filter
}

// Query queries results for a report.Report.
type Query struct {
	db                            *db.ClickHouse
	primaryTable                  string
	joinTable                     string
	importedTable                 string
	funnelTimeColumn              string
	funnelTimeAggregationFunction string
	joinStep                      int
	primaryFilter                 []classifiedFilter
	subqueryFilter                []classifiedFilter
}

// NewQuery returns a new Query for given database connection.
func NewQuery(db *db.ClickHouse) *Query {
	if db == nil {
		panic("database connection is nil")
	}

	return &Query{
		db:             db,
		primaryFilter:  make([]classifiedFilter, 0),
		subqueryFilter: make([]classifiedFilter, 0),
	}
}

// Run runs given request.Request and returns the report.Report.
func (q *Query) Run(req request.Request) report.Report {
	if errs := q.prepare(&req); errs != nil {
		return report.Report{
			Meta: report.Meta{
				Errors: errs,
			},
		}
	}

	var r report.Report

	if q.joinTable != "" {
		r = q.runWithJoin(req)
	} else {
		r = q.run(req)
	}

	if req.Period.Compare != nil {
		if err := q.runWithComparison(req, &r); err != nil {
			return report.Report{
				Meta: report.Meta{
					Errors: []error{err},
				},
			}
		}
	}

	return r
}

// Funnel runs given request.FunnelRequest and returns the report.FunnelReport.
func (q *Query) Funnel(req request.FunnelRequest) report.FunnelReport {
	if errs := req.Validate(); errs != nil {
		return report.FunnelReport{
			Meta: report.Meta{
				Errors: errs,
			},
		}
	}

	// query and scan results
	query, args, err := q.prepareFunnel(req)

	if err != nil {
		return report.FunnelReport{
			Meta: report.Meta{
				Errors: []error{err},
			},
		}
	}

	rows, err := q.db.Query(req.Ctx, query, args...)

	if err != nil {
		return report.FunnelReport{
			Meta: report.Meta{
				Errors: []error{err},
			},
		}
	}

	defer func() {
		_ = rows.Close()
	}()
	r := report.FunnelReport{
		Request: req,
		Steps:   make([]report.FunnelStep, 0, len(req.Filter)),
	}

	for rows.Next() {
		var step report.FunnelStep

		if err := rows.Scan(&step.Step, &step.Visitors); err != nil {
			return report.FunnelReport{
				Meta: report.Meta{
					Errors: []error{err},
				},
			}
		}

		r.Steps = append(r.Steps, step)
	}

	if err := rows.Err(); err != nil {
		return report.FunnelReport{
			Meta: report.Meta{
				Errors: []error{err},
			},
		}
	}

	// compute derived metrics
	for i := range r.Steps {
		if i > 0 {
			if r.Steps[0].Visitors > 0 {
				r.Steps[i].RelativeVisitors = float64(r.Steps[i].Visitors) / float64(r.Steps[0].Visitors)
			}

			r.Steps[i].PreviousVisitors = r.Steps[i-1].Visitors
			r.Steps[i].RelativePreviousVisitors = r.Steps[i-1].RelativeVisitors
			r.Steps[i].Dropped = r.Steps[i].PreviousVisitors - r.Steps[i].Visitors

			if r.Steps[i].PreviousVisitors > 0 {
				r.Steps[i].DropOff = 1 - float64(r.Steps[i].Visitors)/float64(r.Steps[i].PreviousVisitors)
			}
		} else if r.Steps[i].Visitors > 0 {
			r.Steps[i].RelativeVisitors = 1
		}
	}

	return r
}

// DryRun dry runs given request.Request and returns the primary and secondary query and arguments.
// This does not include the comparison mode.
func (q *Query) DryRun(req request.Request) (string, []any, string, []any, []error) {
	if errs := q.prepare(&req); errs != nil {
		return "", nil, "", nil, errs
	}

	if q.joinTable != "" {
		join, err := q.prepareRunWithJoinQueries(req)

		if err != nil {
			return "", nil, "", nil, []error{err}
		}

		return join.primaryQuery, join.primaryArgs, join.secondaryQuery, join.secondaryArgs, nil
	}

	var query string
	var args []any

	if q.useImportedStatistics(req) {
		query, args = q.buildUnionQuery(req)
	} else {
		query, args = q.buildQuery(req)
	}

	return query, args, "", nil, nil
}

// DryRunFunnel dry runs given request.FunnelRequest and returns the query and arguments.
func (q *Query) DryRunFunnel(req request.FunnelRequest) (string, []any, []error) {
	if errs := req.Validate(); errs != nil {
		return "", nil, errs
	}

	// generate one subquery for each step
	query, args, err := q.prepareFunnel(req)

	if err != nil {
		return "", nil, []error{err}
	}

	return query, args, nil
}

func (q *Query) buildFunnelStepQuery(req request.FunnelRequest, stepIndex int, stepFilter []request.Filter) (string, []any, error) {
	stepDimensions := []dimensions.Dimension{
		dimensions.VisitorID{},
		dimensions.SessionID{},
	}

	r := request.Request{
		Ctx:        req.Ctx,
		SiteID:     req.SiteID,
		Period:     req.Period,
		Dimensions: stepDimensions,
		Filter:     stepFilter,
		Options:    req.Options,
	}
	query := NewQuery(q.db)

	if errs := query.prepare(&r); errs != nil {
		return "", nil, fmt.Errorf("funnel step %d: %w", stepIndex+1, errors.Join(errs...))
	}

	query.joinStep = stepIndex // join the previous step (starting at 1)

	// Determine the time column for step ordering based on the query filters.
	// Subquery filters will be promoted to primary if they resolve to page views or events.
	// For pure session queries, determine which field to use (time or start), depending on entry/exit path or title filters.
	if len(query.subqueryFilter) > 0 {
		subqueryTable := query.subqueryFilter[0].table
		demoted := make([]classifiedFilter, 0, len(query.primaryFilter))

		for _, f := range query.primaryFilter {
			demoted = append(demoted, classifiedFilter{table: f.table, filter: f.filter})
		}

		promoted := make([]classifiedFilter, 0, len(query.subqueryFilter))

		for _, f := range query.subqueryFilter {
			promoted = append(promoted, classifiedFilter{table: subqueryTable, filter: f.filter})
		}

		query.primaryTable = subqueryTable
		query.primaryFilter = promoted
		query.subqueryFilter = demoted
		query.funnelTimeColumn = "time"
		query.funnelTimeAggregationFunction = "min"
	} else if query.primaryTable == pkg.TableSessions {
		if funnelStepHasExitFilter(stepFilter) {
			query.funnelTimeColumn = "time"
			query.funnelTimeAggregationFunction = "max"
		} else {
			query.funnelTimeColumn = "start"
			query.funnelTimeAggregationFunction = "min"
		}
	} else {
		query.funnelTimeColumn = "time"
		query.funnelTimeAggregationFunction = "min"
	}

	stepQuery, stepArgs := query.buildQuery(r)
	return stepQuery, stepArgs, nil
}

func (q *Query) funnelStepsEqual(a, b []request.Filter) bool {
	return len(a) == len(b) && fmt.Sprint(a) == fmt.Sprint(b)
}

func funnelStepHasExitFilter(filters []request.Filter) bool {
	for _, f := range filters {
		if len(f.Filter) > 0 {
			if funnelStepHasExitFilter(f.Filter) {
				return true
			}

			continue
		}

		switch f.Dimension.(type) {
		case dimensions.ExitPath, dimensions.ExitTitle:
			return true
		}
	}

	return false
}

func (q *Query) prepare(req *request.Request) []error {
	if errs := req.Validate(); errs != nil {
		return errs
	}

	q.resolvePrimaryTable(*req)
	q.resolveJoinTable(*req)

	if err := q.prepareImported(req); err != nil {
		return []error{err}
	}

	for _, filter := range req.Filter {
		if err := q.classifyFilter(filter); err != nil {
			return []error{err}
		}
	}

	return nil
}

func (q *Query) prepareImported(req *request.Request) error {
	if q.useImportedStatistics(*req) {
		if err := q.resolveImportedTable(*req); err != nil {
			return err
		}

		if err := q.checkFilterImportedTable(req.Filter); err != nil {
			return err
		}
	}

	return nil
}

func (q *Query) runWithJoin(req request.Request) report.Report {
	join, err := q.prepareRunWithJoinQueries(req)

	if err != nil {
		return report.Report{
			Meta: report.Meta{
				Errors: []error{err},
			},
		}
	}

	// run both in parallel
	var wg sync.WaitGroup
	var primaryErr, secondaryErr error
	var results, secondaryResults []report.Result
	wg.Go(func() {
		results, primaryErr = q.queryAndScan(req.Ctx, join.primaryQuery, join.primaryArgs, req.Dimensions, join.primaryMetrics)
	})
	wg.Go(func() {
		secondaryResults, secondaryErr = q.queryAndScan(req.Ctx, join.secondaryQuery, join.secondaryArgs, req.Dimensions, join.secondaryMetrics)
	})
	wg.Wait()

	if err := errors.Join(primaryErr, secondaryErr); err != nil {
		return report.Report{
			Meta: report.Meta{
				Errors: []error{
					errors.New("error executing queries"),
					err,
				},
			},
		}
	}

	q.mergeResults(results, secondaryResults, req.Metrics, join.primaryMetrics)

	if join.sortingByJoinMetric {
		q.sortResults(results, req.Dimensions, req.Metrics, req.OrderBy)
		results = q.paginateResults(results, req.Pagination)
	}

	return report.Report{
		Request: req,
		Results: results,
	}
}

func (q *Query) run(req request.Request) report.Report {
	var query string
	var args []any

	if q.useImportedStatistics(req) {
		query, args = q.buildUnionQuery(req)
	} else {
		query, args = q.buildQuery(req)
	}

	results, err := q.queryAndScan(req.Ctx, query, args, req.Dimensions, req.Metrics)

	if err != nil {
		return report.Report{
			Meta: report.Meta{
				Errors: []error{
					errors.New("error executing query"),
					err,
				},
			},
		}
	}

	return report.Report{
		Request: req,
		Results: results,
	}
}

func (q *Query) runWithComparison(req request.Request, currentReport *report.Report) error {
	// determine if the results should be merged in order (positionally) or by keys
	mergeInOrder := false

	if len(req.OrderBy) > 0 && req.OrderBy[0].Dimension != nil {
		switch req.OrderBy[0].Dimension.(type) {
		case dimensions.Time,
			dimensions.Start,
			dimensions.Day,
			dimensions.Week,
			dimensions.Month,
			dimensions.Year,
			dimensions.Hour:
			allTime := true

			for _, d := range req.Dimensions {
				switch d.(type) {
				case dimensions.Time,
					dimensions.Start,
					dimensions.Day,
					dimensions.Week,
					dimensions.Month,
					dimensions.Year,
					dimensions.Hour:
				default:
					allTime = false
				}
			}

			mergeInOrder = allTime
		}
	}

	// get results ignoring pagination and possibly order by
	req.Period.From = req.Period.Compare.From
	req.Period.To = req.Period.Compare.To
	req.Pagination = nil

	if !mergeInOrder {
		req.OrderBy = nil
	}

	// prepare imported statistics table based on the new dates if needed
	if err := q.prepareImported(&req); err != nil {
		return err
	}

	// get the report for the comparison period
	var previousReport report.Report

	if q.joinTable != "" {
		previousReport = q.runWithJoin(req)
	} else {
		previousReport = q.run(req)
	}

	if len(previousReport.Meta.Errors) > 0 {
		return errors.Join(previousReport.Meta.Errors...)
	}

	// merge results
	if mergeInOrder {
		for i := range currentReport.Results {
			if i < len(previousReport.Results) {
				currentReport.Results[i].CompareMetricValues = previousReport.Results[i].MetricValues
			} else {
				currentReport.Results[i].CompareMetricValues = q.zeroValues(req.Metrics)
			}
		}

		for i := len(currentReport.Results); i < len(previousReport.Results); i++ {
			currentReport.Results = append(currentReport.Results, report.Result{
				DimensionValues:     previousReport.Results[i].DimensionValues,
				MetricValues:        q.zeroValues(req.Metrics),
				CompareMetricValues: previousReport.Results[i].MetricValues,
			})
		}
	} else {
		primaryIndex := make(map[string]int, len(currentReport.Results))

		for i, result := range currentReport.Results {
			primaryIndex[q.dimensionKey(result.DimensionValues)] = i
			currentReport.Results[i].CompareMetricValues = q.zeroValues(req.Metrics)
		}

		for _, compResult := range previousReport.Results {
			key := q.dimensionKey(compResult.DimensionValues)

			if i, ok := primaryIndex[key]; ok {
				currentReport.Results[i].CompareMetricValues = compResult.MetricValues
			} else {
				currentReport.Results = append(currentReport.Results, report.Result{
					DimensionValues:     compResult.DimensionValues,
					MetricValues:        q.zeroValues(req.Metrics),
					CompareMetricValues: compResult.MetricValues,
				})
			}
		}
	}

	return nil
}

func (q *Query) queryAndScan(ctx context.Context, query string, args []any, dimensions []dimensions.Dimension, metrics []metrics.Metric) ([]report.Result, error) {
	rows, err := q.db.Query(ctx, query, args...)

	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rows.Close()
	}()
	return q.scanRows(rows, dimensions, metrics)
}

func (q *Query) prepareRunWithJoinQueries(req request.Request) (joinQueries, error) {
	var join joinQueries
	join.primaryMetrics, join.secondaryMetrics = q.splitMetrics(req.Metrics)
	join.sortingByJoinMetric = q.orderByContainsJoinMetric(req.OrderBy)

	// build query for primary metrics
	primaryReq := req
	primaryReq.Metrics = join.primaryMetrics
	primaryReq.OrderBy = q.filterOrderByForPrimary(req.OrderBy, join.primaryMetrics)

	if join.sortingByJoinMetric {
		primaryReq.Pagination = nil
	}

	join.primaryQuery, join.primaryArgs = q.buildQuery(primaryReq)

	// build query for secondary metrics
	savedPrimaryTable := q.primaryTable
	savedPrimaryFilter := q.primaryFilter
	savedSubqueryFilter := q.subqueryFilter
	defer func() {
		q.primaryTable = savedPrimaryTable
		q.primaryFilter = savedPrimaryFilter
		q.subqueryFilter = savedSubqueryFilter
	}()
	q.primaryTable = q.joinTable
	q.primaryFilter = make([]classifiedFilter, 0)
	q.subqueryFilter = make([]classifiedFilter, 0)

	for _, filter := range req.Filter {
		if err := q.classifyFilter(filter); err != nil {
			return joinQueries{}, err
		}
	}

	secondaryReq := req
	secondaryReq.Metrics = join.secondaryMetrics
	secondaryReq.OrderBy = nil
	secondaryReq.Pagination = nil
	join.secondaryQuery, join.secondaryArgs = q.buildQuery(secondaryReq)
	return join, nil
}

func (q *Query) prepareFunnel(req request.FunnelRequest) (string, []any, error) {
	var query strings.Builder
	args := make([]any, 0)

	for i, step := range req.Filter {
		if i == 0 {
			query.WriteString(fmt.Sprintf("WITH step%d AS (", i+1))
		} else {
			query.WriteString(fmt.Sprintf(", step%d AS (", i+1))
		}

		// skip steps that have the same filters as the previous step
		if i > 0 && q.funnelStepsEqual(req.Filter[i], req.Filter[i-1]) {
			query.WriteString(fmt.Sprintf("SELECT * FROM step%d", i))
		} else {
			stepQuery, stepArgs, err := q.buildFunnelStepQuery(req, i, step)

			if err != nil {
				return "", nil, err
			}

			query.WriteString(stepQuery)
			args = append(args, stepArgs...)
		}

		query.WriteString(") ")
	}

	// union all steps to get visitor counts
	query.WriteString("SELECT * FROM (")

	for i := range req.Filter {
		query.WriteString(fmt.Sprintf("SELECT %d step, uniq(visitor_id) visitors FROM step%d ", i+1, i+1))

		if i != len(req.Filter)-1 {
			query.WriteString("UNION ALL ")
		}
	}

	query.WriteString(") ORDER BY step")
	return query.String(), args, nil
}

func (q *Query) useImportedStatistics(req request.Request) bool {
	return req.Options != nil && req.Options.IncludeImportedStatistics &&
		!req.Period.ImportedUntil.IsZero() && req.Period.From.Before(req.Period.ImportedUntil)
}

func (q *Query) zeroValues(metrics []metrics.Metric) []any {
	zeroValues := make([]any, len(metrics))

	for i, m := range metrics {
		zeroValues[i] = m.Zero()
	}

	return zeroValues
}

func (q *Query) resolvePrimaryTable(req request.Request) {
	// dimensions drive the primary table
	if len(req.Dimensions) > 0 {
		q.primaryTable = q.resolveBestTable(q.dimensionTables(req.Dimensions))
		return
	}

	// no dimensions, use metrics instead
	if len(req.Metrics) > 0 {
		q.primaryTable = q.resolveBestTable(q.metricTables(req.Metrics))
		return
	}

	// fall back to sessions for simple queries
	q.primaryTable = pkg.TableSessions
}

func (q *Query) resolveBestTable(tableSets [][]string) string {
	candidates := []string{pkg.TableSessions, pkg.TablePageViews, pkg.TableEvents}
	valid := make([]string, 0, len(candidates))

	for _, candidate := range candidates {
		satisfiesAll := true

		for _, tables := range tableSets {
			if !slices.Contains(tables, candidate) {
				satisfiesAll = false
				break
			}
		}

		if satisfiesAll {
			valid = append(valid, candidate)
		}
	}

	if len(valid) == 0 {
		// no single table works, use highest priority preferred table
		preferred := make([]string, 0, len(tableSets))

		for _, tables := range tableSets {
			if len(tables) > 0 {
				preferred = append(preferred, tables[0])
			}
		}

		return q.highestPriorityTable(preferred)
	}

	if len(valid) == 1 {
		return valid[0]
	}

	// multiple tables satisfy all dimensions, prefer the one most dimensions
	// list as their first choice (preferred table)
	score := make(map[string]int, len(valid))

	for _, tables := range tableSets {
		if len(tables) > 0 && slices.Contains(valid, tables[0]) {
			score[tables[0]]++
		}
	}

	best := valid[0]

	for _, t := range valid[1:] {
		if score[t] > score[best] {
			best = t
		}
	}

	return best
}

func (q *Query) highestPriorityTable(tables []string) string {
	table := pkg.TableSessions
	priority := 0

	for _, t := range tables {
		if p := tablePriority[t]; p > priority {
			table = t
			priority = p
		}
	}

	return table
}

func (q *Query) dimensionTables(dimensions []dimensions.Dimension) [][]string {
	tables := make([][]string, 0, len(dimensions))

	for _, d := range dimensions {
		tables = append(tables, d.Table())
	}

	return tables
}

func (q *Query) metricTables(metrics []metrics.Metric) [][]string {
	tables := make([][]string, 0, len(metrics))

	for _, m := range metrics {
		tables = append(tables, m.Table())
	}

	return tables
}

func (q *Query) resolveJoinTable(req request.Request) {
	for _, m := range req.Metrics {
		joinTable := m.JoinTable()

		if joinTable != "" && joinTable != q.primaryTable {
			q.joinTable = joinTable
			return
		}
	}
}

func (q *Query) resolveImportedTable(req request.Request) error {
	candidates := slices.Clone(pkg.ImportedTables)

	for _, d := range req.Dimensions {
		candidates = slices.DeleteFunc(candidates, func(t string) bool {
			return !slices.Contains(d.TableImported(), t)
		})
	}

	if len(candidates) == 0 {
		return errors.New("no overlapping imported statistics table found")
	}

	q.importedTable = candidates[0]
	return nil
}

func (q *Query) checkFilterImportedTable(filter []request.Filter) error {
	for _, f := range filter {
		if !slices.Contains(f.Dimension.TableImported(), q.importedTable) {
			return fmt.Errorf("filter dimension '%s' does not apply to imported statistics table '%s'", f.Dimension, q.importedTable)
		}

		if err := q.checkFilterImportedTable(f.Filter); err != nil {
			return err
		}
	}

	return nil
}

func (q *Query) splitMetrics(requestMetrics []metrics.Metric) ([]metrics.Metric, []metrics.Metric) {
	primary := make([]metrics.Metric, 0, len(requestMetrics))
	secondary := make([]metrics.Metric, 0)

	for _, m := range requestMetrics {
		if q.isSecondaryMetric(m) {
			secondary = append(secondary, m)
		} else {
			primary = append(primary, m)
		}
	}

	return primary, secondary
}

func (q *Query) mergeResults(primary []report.Result, secondary []report.Result, requestMetrics, primaryMetrics []metrics.Metric) {
	totalCount := len(requestMetrics)
	primaryPos := make(map[string]int, len(primaryMetrics))

	for i, m := range primaryMetrics {
		primaryPos[m.Column()] = i
	}

	secondaryIndices := 0
	secondaryPos := make(map[string]int, totalCount-len(primaryMetrics))

	for _, m := range requestMetrics {
		if q.isSecondaryMetric(m) {
			secondaryPos[m.Column()] = secondaryIndices
			secondaryIndices++
		}
	}

	for i := range primary {
		expanded := make([]any, totalCount)

		for fullIndex, m := range requestMetrics {
			if !q.isSecondaryMetric(m) {
				if pos, ok := primaryPos[m.Column()]; ok && pos < len(primary[i].MetricValues) {
					expanded[fullIndex] = primary[i].MetricValues[pos]
				}
			} else {
				expanded[fullIndex] = m.Zero()
			}
		}

		primary[i].MetricValues = expanded
	}

	index := make(map[string]int, len(primary))

	for i, r := range primary {
		index[q.dimensionKey(r.DimensionValues)] = i
	}

	for _, sec := range secondary {
		primaryIndex, ok := index[q.dimensionKey(sec.DimensionValues)]

		if !ok {
			// secondary row has no matching primary row (can happen with pagination), skip
			continue
		}

		for fullIndex, m := range requestMetrics {
			if !q.isSecondaryMetric(m) {
				continue
			}

			if pos, ok := secondaryPos[m.Column()]; ok && pos < len(sec.MetricValues) {
				primary[primaryIndex].MetricValues[fullIndex] = sec.MetricValues[pos]
			}
		}
	}
}

func (q *Query) orderByContainsJoinMetric(order []request.OrderBy) bool {
	for _, o := range order {
		if o.Metric != nil && q.isSecondaryMetric(o.Metric) {
			return true
		}
	}

	return false
}

func (q *Query) isSecondaryMetric(m metrics.Metric) bool {
	return q.joinTable != "" && m.JoinTable() == q.joinTable
}

func (q *Query) filterOrderByForPrimary(order []request.OrderBy, primaryMetrics []metrics.Metric) []request.OrderBy {
	primaryColumns := make(map[string]bool, len(primaryMetrics))

	for _, m := range primaryMetrics {
		primaryColumns[m.Column()] = true
	}

	result := make([]request.OrderBy, 0, len(order))

	for _, o := range order {
		if o.Metric != nil {
			if primaryColumns[o.Metric.Column()] {
				result = append(result, o)
			}
		} else {
			result = append(result, o)
		}
	}

	return result
}

func (q *Query) sortResults(results []report.Result, requestDimensions []dimensions.Dimension, requestMetrics []metrics.Metric, order []request.OrderBy) {
	if len(order) == 0 {
		return
	}

	// index lookup maps
	metricIndex := make(map[string]int, len(requestMetrics))

	for i, m := range requestMetrics {
		metricIndex[m.Column()] = i
	}

	dimensionIndex := make(map[string]int, len(requestDimensions))

	for i, d := range requestDimensions {
		dimensionIndex[d.Column(q.primaryTable)] = i
	}

	slices.SortStableFunc(results, func(a, b report.Result) int {
		for _, o := range order {
			var compare int

			if o.Metric != nil {
				index, ok := metricIndex[o.Metric.Column()]

				if !ok || index >= len(a.MetricValues) || index >= len(b.MetricValues) {
					continue
				}

				compare = q.compareAny(a.MetricValues[index], b.MetricValues[index])
			} else if o.Dimension != nil {
				index, ok := dimensionIndex[o.Dimension.Column(q.primaryTable)]

				if !ok || index >= len(a.DimensionValues) || index >= len(b.DimensionValues) {
					continue
				}

				compare = q.compareAny(a.DimensionValues[index], b.DimensionValues[index])
			}

			if compare != 0 {
				if o.Direction == request.DirectionASC {
					return compare
				}

				return -compare
			}
		}

		return 0
	})
}

func (q *Query) compareAny(a, b any) int {
	if a == nil || b == nil {
		switch {
		case a == nil && b == nil:
			return 0
		case a == nil:
			return -1
		default:
			return 1
		}
	}

	switch av := a.(type) {
	case uint64:
		return q.compareTyped(av, b)
	case uint32:
		return q.compareTyped(av, b)
	case uint8:
		return q.compareTyped(av, b)
	case int64:
		return q.compareTyped(av, b)
	case int:
		return q.compareTyped(av, b)
	case float64:
		return q.compareTyped(av, b)
	case string:
		return q.compareTyped(av, b)
	case time.Time:
		if bv, ok := b.(time.Time); ok {
			return av.Compare(bv)
		}
	}

	return 0
}

func (q *Query) compareTyped[T cmp.Ordered](a T, b any) int {
	if bv, ok := b.(T); ok {
		return cmp.Compare(a, bv)
	}

	return 0
}

func (q *Query) paginateResults(results []report.Result, pagination *request.Pagination) []report.Result {
	if pagination == nil || pagination.Limit <= 0 {
		return results
	}

	offset := max(pagination.Offset, 0)

	if offset >= len(results) {
		return []report.Result{}
	}

	end := min(offset+pagination.Limit, len(results))

	return results[offset:end]
}

func (q *Query) dimensionKey(values []any) string {
	parts := make([]string, len(values))

	for i, v := range values {
		parts[i] = fmt.Sprint(v)
	}

	return strings.Join(parts, "|")
}

func (q *Query) classifyFilter(filter request.Filter) error {
	table, err := q.filterTable(filter)

	if err != nil {
		return err
	}

	if table == q.primaryTable || table == "" {
		q.primaryFilter = append(q.primaryFilter, classifiedFilter{
			table:  q.primaryTable,
			filter: filter,
		})
	} else {
		q.subqueryFilter = append(q.subqueryFilter, classifiedFilter{
			table:  table,
			filter: filter,
		})
	}

	return nil
}

func (q *Query) filterTable(filter request.Filter) (string, error) {
	// leaf node, dimension must be set
	if len(filter.Filter) == 0 {
		if filter.Dimension == nil {
			return "", nil
		}

		tables := filter.Dimension.Table()

		if slices.Contains(tables, q.primaryTable) {
			return q.primaryTable, nil
		}

		return tables[0], nil
	}

	// logical group, collect tables from all children
	childTables := make(map[string]bool)

	for _, child := range filter.Filter {
		table, err := q.filterTable(child)

		if err != nil {
			return "", err
		}

		if table != "" {
			childTables[table] = true
		}
	}

	if len(childTables) == 0 {
		return "", nil
	}

	// multiple tables in one logical group is only safe for AND at the top level
	if len(childTables) > 1 {
		// check if children are compatible with a single shared table
		sharedTable := q.findSharedTable(filter.Filter)

		if sharedTable != "" {
			return sharedTable, nil
		}

		if filter.Operator == request.OperatorGroupOr || filter.Operator == request.OperatorGroupNot {
			tables := make([]string, 0, len(childTables))

			for t := range childTables {
				tables = append(tables, t)
			}

			return "", fmt.Errorf("cannot mix dimensions from different tables (%s) within an OR/NOT filter group", strings.Join(tables, ", "))
		}

		return "", nil
	}

	// all leaves on the same table
	for t := range childTables {
		return t, nil
	}

	return "", nil
}

func (q *Query) findSharedTable(filter []request.Filter) string {
	// collect all table sets from leaf dimensions in the group
	tableSets := q.collectLeafTableSets(filter)

	if len(tableSets) == 0 {
		return ""
	}

	// try primary table first
	if q.allCompatible(tableSets, q.primaryTable) {
		return q.primaryTable
	}

	// try other tables in priority order
	for _, candidate := range []string{pkg.TablePageViews, pkg.TableEvents, pkg.TableSessions} {
		if candidate != q.primaryTable && q.allCompatible(tableSets, candidate) {
			return candidate
		}
	}

	return ""
}

func (q *Query) collectLeafTableSets(filter []request.Filter) [][]string {
	result := make([][]string, 0)

	for _, f := range filter {
		if len(f.Filter) == 0 {
			if f.Dimension != nil {
				result = append(result, f.Dimension.Table())
			}
		} else {
			result = append(result, q.collectLeafTableSets(f.Filter)...)
		}
	}

	return result
}

func (q *Query) allCompatible(tableSets [][]string, candidate string) bool {
	for _, tables := range tableSets {
		if !slices.Contains(tables, candidate) {
			return false
		}
	}

	return true
}

func (q *Query) buildQuery(req request.Request) (string, []any) {
	var query strings.Builder
	args := make([]any, 0)
	withRequired := q.buildQueryWithRequired(req.Metrics)

	if withRequired {
		withQuery, withArgs := q.buildQueryWith(req)
		query.WriteString(withQuery)
		args = append(args, withArgs...)
	}

	selectQuery, selectArgs := q.buildQuerySelect(req)
	query.WriteString(selectQuery)
	args = append(args, selectArgs...)
	query.WriteString(q.buildQueryFrom(q.primaryTable, req.Options.Sample))

	if withRequired {
		query.WriteString(q.buildQueryWithJoin(req))
	}

	if q.joinStep > 0 {
		query.WriteString(q.buildQueryWithJoinStep())
	}

	whereQuery, whereArgs := q.buildQueryWhere(req)
	query.WriteString(whereQuery)
	args = append(args, whereArgs...)
	query.WriteString(q.buildQueryGroupBy(req.Dimensions))
	query.WriteString(q.buildQueryHaving())
	orderByQuery, orderByArgs := q.buildOrderBy(req)
	query.WriteString(orderByQuery)
	args = append(args, orderByArgs...)
	query.WriteString(q.buildQueryPagination(req.Pagination))
	return query.String(), args
}

func (q *Query) buildUnionQuery(req request.Request) (string, []any) {
	// query natively if required
	var nativeQuery string
	var nativeArgs []any
	queryNative := !req.Period.To.Before(req.Period.ImportedUntil)
	var nativeReq request.Request
	importedReq := req

	if queryNative {
		importedReq.Period.To = req.Period.ImportedUntil.AddDate(0, 0, -1)
		nativeReq = req
		nativeReq.Pagination = nil
		nativeReq.OrderBy = nil
		nativeOptions := *req.Options // copy, the options are shared with the request
		nativeOptions.IncludeImportedStatistics = false
		nativeReq.Options = &nativeOptions
		nativeReq.Period.From = nativeReq.Period.ImportedUntil
		nativeQuery, nativeArgs = q.buildQuery(nativeReq)
	}

	// query imported statistics
	importedQuery, importedArgs := q.buildQueryImported(importedReq)

	// union queries
	var query strings.Builder
	args := make([]any, 0)

	if queryNative {
		outerFields := make([]string, 0, len(req.Metrics)+len(req.Dimensions))

		for _, m := range req.Metrics {
			switch m.(type) {
			case metrics.RelativeVisitors, metrics.RelativeViews:
				whereSiteAndPeriodNative, whereSiteAndPeriodNativeArgs := q.buildQueryWhereSiteAndPeriod(req.SiteID, nativeReq.Period, pkg.TableSessions, false)
				whereSiteAndPeriodImported, whereSiteAndPeriodImportedArgs := q.buildQueryWhereSiteAndPeriod(req.SiteID, importedReq.Period, q.importedTable, true)
				args = append(args, whereSiteAndPeriodNativeArgs...)
				args = append(args, whereSiteAndPeriodImportedArgs...)

				if m.Column() == "relative_visitors" {
					outerFields = append(outerFields, fmt.Sprintf(`toFloat64OrDefault(visitors / greatest((SELECT uniq(visitor_id) FROM "%s" %s) + (SELECT sum(visitors) FROM "%s" %s), 1)) %s`,
						pkg.TableSessions, whereSiteAndPeriodNative, q.importedTable, whereSiteAndPeriodImported, m.Column()))
				} else {
					outerFields = append(outerFields, fmt.Sprintf(`toFloat64OrDefault(page_views / greatest((SELECT sum(page_views * sign) FROM "%s" %s) + (SELECT sum(views) FROM "%s" %s), 1)) %s`,
						pkg.TableSessions, whereSiteAndPeriodNative, q.importedTable, whereSiteAndPeriodImported, m.Column()))
				}
			case metrics.BounceRate:
				outerFields = append(outerFields, fmt.Sprintf("toFloat64OrDefault(bounces / greatest(sessions, 1)) %s", m.Column()))
			default:
				outerFields = append(outerFields, fmt.Sprintf("sum(%s) %s", m.Column(), m.Column()))
			}
		}

		for _, d := range req.Dimensions {
			outerFields = append(outerFields, d.Column(""))
		}

		query.WriteString(fmt.Sprintf("SELECT %s FROM (", strings.Join(outerFields, ", ")))
		query.WriteString(nativeQuery)
		query.WriteString("UNION ALL ")
		query.WriteString(importedQuery)
		query.WriteString(") ")
		groupColumns := make([]string, 0, len(req.Dimensions))

		for _, d := range req.Dimensions {
			groupColumns = append(groupColumns, d.Column(""))
		}

		if len(groupColumns) > 0 {
			query.WriteString(fmt.Sprintf("GROUP BY %s ", strings.Join(groupColumns, ", ")))
		}
	} else {
		query.WriteString(importedQuery)
	}

	args = append(args, nativeArgs...)
	args = append(args, importedArgs...)

	// order by, and offset/limit
	orderByQuery, orderByArgs := q.buildOrderBy(req)
	args = append(args, orderByArgs...)
	query.WriteString(orderByQuery)
	query.WriteString(q.buildQueryPagination(req.Pagination))
	return query.String(), args
}

func (q *Query) buildQueryImported(req request.Request) (string, []any) {
	var query strings.Builder
	args := make([]any, 0)
	fields := make([]string, 0, len(req.Metrics)+len(req.Dimensions))

	for _, m := range req.Metrics {
		expression := m.ExpressionImported()
		column := m.ColumnImported()

		if expression == "" {
			fields = append(fields, fmt.Sprintf("%v %s", m.Zero(), column))
		} else {
			fields = append(fields, fmt.Sprintf("%s %s", expression, column))
		}
	}

	for _, d := range req.Dimensions {
		expression := d.ExpressionImported(&dimensions.DimensionExpressionOptions{
			Timezone:    req.Period.Timezone,
			WeekdayMode: int(req.Period.WeekdayMode),
		})
		column := d.ColumnImported()

		if expression != "" {
			fields = append(fields, fmt.Sprintf("%s %s", expression, column))
		} else {
			fields = append(fields, column)
		}
	}

	whereQuery, whereArgs := q.buildQueryWhereSiteAndPeriod(req.SiteID, req.Period, q.importedTable, true)
	query.WriteString(fmt.Sprintf(`SELECT %s FROM "%s" `, strings.Join(fields, ", "), q.importedTable))
	query.WriteString(whereQuery)
	args = append(args, whereArgs...)
	args = append(args, q.writeQueryFilters(&query, q.importedTable, q.primaryFilter)...)
	groupBy := make([]string, 0, len(req.Dimensions))

	for _, d := range req.Dimensions {
		groupBy = append(groupBy, d.ColumnImported())
	}

	if len(groupBy) > 0 {
		query.WriteString(fmt.Sprintf("GROUP BY %s ", strings.Join(groupBy, ", ")))
	}

	return query.String(), args
}

func (q *Query) buildQueryWithRequired(list []metrics.Metric) bool {
	found := false

	for _, m := range list {
		if _, ok := m.(metrics.AvgTimeOnPage); ok {
			found = true
			break
		}
	}

	return found
}

func (q *Query) buildQueryWith(req request.Request) (string, []any) {
	// store original tables and filter filters
	primaryFilter := q.primaryFilter
	primaryTable := q.primaryTable
	q.primaryTable = pkg.TablePageViews
	q.primaryFilter = q.buildQueryFiltersForTable(primaryFilter, pkg.TablePageViews)

	// build where query
	var query strings.Builder
	args := make([]any, 0)
	whereQuery, whereArgs := q.buildQueryWhere(req)
	args = append(args, whereArgs...)

	// filter dimensions to group by
	groupBy := q.buildQueryWithFilterDimensions(req)
	fields := make([]string, 0, len(groupBy))

	for _, d := range groupBy {
		fields = append(fields, q.buildQuerySelectColumn(d, req.Period.Timezone, req.Period.WeekdayMode))
	}

	selectFields := strings.Join(fields, ", ")

	// build query
	query.WriteString("WITH top AS (SELECT path, ")

	if len(fields) > 0 {
		query.WriteString(selectFields)
		query.WriteString(", ")
	}

	query.WriteString(`leadInFrame(duration_seconds) OVER (
					PARTITION BY visitor_id, session_id
					ORDER BY time ASC
					ROWS BETWEEN CURRENT ROW AND 1 FOLLOWING
				) time_on_page
			FROM ` + pkg.TablePageViews + ` `)
	query.WriteString(whereQuery)
	query.WriteString("), avgtop AS (SELECT path, ")

	if len(fields) > 0 {
		names := make([]string, 0, len(groupBy))

		for _, field := range groupBy {
			names = append(names, field.Column(""))
		}

		query.WriteString(strings.Join(names, ", "))
		query.WriteString(", ")
	}

	query.WriteString("avgOrDefaultIf(time_on_page, time_on_page > 0) avg_time_on_page FROM top ")
	groupBy = append(groupBy, dimensions.Path{})
	query.WriteString(q.buildQueryGroupBy(groupBy))
	query.WriteString(") ")

	// restore original tables
	q.primaryTable = primaryTable
	q.primaryFilter = primaryFilter
	return query.String(), args
}

func (q *Query) buildQueryFiltersForTable(filters []classifiedFilter, table string) []classifiedFilter {
	result := make([]classifiedFilter, 0, len(filters))

	for _, f := range filters {
		if q.buildQueryFilterCompatibleWithTable(f.filter, table) {
			result = append(result, classifiedFilter{
				table:  table,
				filter: f.filter,
			})
		}
	}

	return result
}

func (q *Query) buildQueryFilterCompatibleWithTable(filter request.Filter, table string) bool {
	// leaf node
	if len(filter.Filter) == 0 {
		if filter.Dimension == nil {
			return true
		}

		return slices.Contains(filter.Dimension.Table(), table)
	}

	// logical group, all children must be compatible
	for _, child := range filter.Filter {
		if !q.buildQueryFilterCompatibleWithTable(child, table) {
			return false
		}
	}

	return true
}

func (q *Query) buildQueryWithJoin(req request.Request) string {
	var query strings.Builder

	// use entry_path to join the time on page, exit_path isn't allowed (because it's always 0)
	if q.primaryTable == pkg.TableSessions {
		query.WriteString("LEFT JOIN avgtop ON entry_path = avgtop.path ")
	} else {
		query.WriteString("LEFT JOIN avgtop ON path = avgtop.path ")
	}

	// join dimensions
	join := q.buildQueryWithFilterDimensions(req)

	if len(join) > 0 {
		query.WriteString("AND ")
		joinColumns := make([]string, 0, len(join))

		for _, field := range join {
			column := field.Column("")
			joinColumns = append(joinColumns, fmt.Sprintf(`"%s" = avgtop."%s"`, column, column))
		}

		query.WriteString(strings.Join(joinColumns, " AND "))
		query.WriteString(" ")
	}

	return query.String()
}

func (q *Query) buildQueryWithJoinStep() string {
	return fmt.Sprintf(`JOIN step%d s ON visitor_id = s.visitor_id AND session_id = s.session_id AND "%s" > s.step_time `, q.joinStep, q.funnelTimeColumn)
}

func (q *Query) buildQueryWithFilterDimensions(req request.Request) []dimensions.Dimension {
	groupBy := make([]dimensions.Dimension, 0, len(req.Dimensions))

	for _, d := range req.Dimensions {
		if _, ok := d.(dimensions.Path); ok || !slices.Contains(d.Table(), pkg.TablePageViews) {
			continue
		}

		groupBy = append(groupBy, d)
	}

	return groupBy
}

func (q *Query) buildQuerySelect(req request.Request) (string, []any) {
	fields := make([]string, 0, len(req.Metrics)+len(req.Dimensions))
	args := make([]any, 0)

	for _, metric := range req.Metrics {
		expression, requiresSubquery := metric.Expression(q.primaryTable)

		if requiresSubquery {
			// metric subqueries always select from the sessions table
			subquery, a := q.buildQueryWhereSiteAndPeriod(req.SiteID, req.Period, pkg.TableSessions, false)
			expression = fmt.Sprintf(expression, subquery)
			args = append(args, a...)
		}

		fields = append(fields, fmt.Sprintf("%s %s", expression, metric.Column()))
	}

	for _, dimension := range req.Dimensions {
		fields = append(fields, q.buildQuerySelectColumn(dimension, req.Period.Timezone, req.Period.WeekdayMode))
		args = append(args, dimension.Args()...)
	}

	// expose the step time for funnel ordering so subsequent steps can enforce time > s.step_time
	if q.funnelTimeColumn != "" {
		aggregationFunc := q.funnelTimeAggregationFunction

		if aggregationFunc == "" {
			aggregationFunc = "min"
		}

		fields = append(fields, fmt.Sprintf(`%s("%s") AS step_time`, aggregationFunc, q.funnelTimeColumn))
	}

	return fmt.Sprintf("SELECT %s ", strings.Join(fields, ",")), args
}

func (q *Query) buildQuerySelectColumn(dimension dimensions.Dimension, location *time.Location, weekdayMode request.WeekdayMode) string {
	options := dimensions.DimensionExpressionOptions{
		Timezone:    location,
		WeekdayMode: int(weekdayMode),
	}

	switch d := dimension.(type) {
	case dimensions.EventMeta:
		if d.Path != "" {
			return d.Select(q.buildQueryFilterJSONPath(d.Path))
		}

		return fmt.Sprintf("%s %s", dimension.Expression(&options), dimension.Column(q.primaryTable))
	case dimensions.EventMetaValue:
		if d.Path != "" {
			return d.Select(q.buildQueryFilterJSONPath(d.Path))
		}

		return fmt.Sprintf("%s %s", dimension.Expression(&options), dimension.Column(q.primaryTable))
	default:
		if dimension.Expression(&options) != "" {
			return fmt.Sprintf("%s %s", dimension.Expression(&options), dimension.Column(q.primaryTable))
		}

		return dimension.Column(q.primaryTable)
	}
}

func (q *Query) buildQueryFrom(table string, sample uint) string {
	if sample > 0 {
		return fmt.Sprintf("FROM %s SAMPLE %d ", table, sample)
	}

	return fmt.Sprintf("FROM %s ", table)
}

func (q *Query) buildQueryWhere(req request.Request) (string, []any) {
	var query strings.Builder
	args := make([]any, 0)
	whereQuery, whereArgs := q.buildQueryWhereSiteAndPeriod(req.SiteID, req.Period, q.primaryTable, false)
	query.WriteString(whereQuery)
	args = append(args, whereArgs...)
	args = append(args, q.writeQueryFilters(&query, q.primaryTable, q.primaryFilter)...)

	if len(q.subqueryFilter) > 0 {
		byTable := make(map[string][]classifiedFilter)
		tableOrder := make([]string, 0)

		for _, filter := range q.subqueryFilter {
			if _, exists := byTable[filter.table]; !exists {
				tableOrder = append(tableOrder, filter.table)
			}

			byTable[filter.table] = append(byTable[filter.table], filter)
		}

		for _, table := range tableOrder {
			filters := byTable[table]
			query.WriteString("AND (visitor_id, session_id) IN (SELECT visitor_id, session_id ")
			query.WriteString(q.buildQueryFrom(table, req.Options.Sample))
			subWhereQuery, subWhereArgs := q.buildQueryWhereSiteAndPeriod(req.SiteID, req.Period, table, false)
			query.WriteString(subWhereQuery)
			args = append(args, subWhereArgs...)
			args = append(args, q.writeQueryFilters(&query, table, filters)...)

			// only match the latest state of a session, ignoring canceled rows
			if table == pkg.TableSessions {
				query.WriteString("GROUP BY visitor_id, session_id HAVING sum(sign) > 0 ")
			}

			query.WriteString(") ")
		}
	}

	return query.String(), args
}

func (q *Query) buildQueryWhereSiteAndPeriod(siteID uint64, period request.Period, table string, imported bool) (string, []any) {
	var query strings.Builder
	args := []any{siteID}
	query.WriteString(fmt.Sprintf("WHERE %s = ? ", q.buildQualifiedColumn(table, "site_id")))

	if period.From.IsZero() && period.To.IsZero() {
		return query.String(), args
	}

	timeColumn := "time"

	if imported {
		timeColumn = "date"
	} else if period.IncludeTime && table == pkg.TableSessions {
		timeColumn = "start"
	}

	column := q.buildQualifiedColumn(table, timeColumn)
	param := "toDate(?)"
	format := time.DateOnly

	if !imported {
		tz := q.timezoneName(period)

		if period.IncludeTime {
			column = fmt.Sprintf(`toDateTime(%s, '%s')`, column, tz)
			param = fmt.Sprintf("toDateTime(?, '%s')", tz)
			format = time.DateTime
		} else {
			column = fmt.Sprintf(`toDate(%s, '%s')`, column, tz)
			param = fmt.Sprintf("toDate(?, '%s')", tz)
		}
	}

	switch {
	case period.From.IsZero():
		query.WriteString(fmt.Sprintf("AND %s <= %s ", column, param))
		args = append(args, period.To.Format(format))
	case period.To.IsZero():
		query.WriteString(fmt.Sprintf("AND %s >= %s ", column, param))
		args = append(args, period.From.Format(format))
	case period.From.Equal(period.To):
		query.WriteString(fmt.Sprintf("AND %s = %s ", column, param))
		args = append(args, period.From.Format(format))
	default:
		query.WriteString(fmt.Sprintf("AND %s BETWEEN %s AND %s ", column, param, param))
		args = append(args, period.From.Format(format), period.To.Format(format))
	}

	return query.String(), args
}

func (q *Query) timezoneName(period request.Period) string {
	if period.Timezone == nil {
		return time.UTC.String()
	}

	return period.Timezone.String()
}

func (q *Query) buildQueryFilter(table string, filter request.Filter) (string, []any) {
	// filter for a column
	if len(filter.Filter) == 0 {
		if len(filter.Values) == 0 {
			return "", nil
		}

		return q.buildQueryFilterColumn(table, filter)
	}

	// filter for a group
	groups := make([]string, 0, len(filter.Filter))
	args := make([]any, 0)

	for _, f := range filter.Filter {
		group, groupArgs := q.buildQueryFilter(table, f)

		if group == "" {
			continue
		}

		groups = append(groups, fmt.Sprintf("(%s)", group))
		args = append(args, groupArgs...)
	}

	if len(groups) == 0 {
		return "", nil
	}

	switch filter.Operator {
	case request.OperatorGroupOr:
		return strings.Join(groups, " OR "), args
	case request.OperatorGroupNot:
		return fmt.Sprintf("NOT (%s)", strings.Join(groups, " AND ")), args
	default:
		return strings.Join(groups, " AND "), args
	}
}

func (q *Query) writeQueryFilters(query *strings.Builder, table string, filters []classifiedFilter) []any {
	args := make([]any, 0)

	for _, filter := range filters {
		where, a := q.buildQueryFilter(table, filter.filter)

		if where == "" {
			continue
		}

		query.WriteString("AND (")
		query.WriteString(where)
		query.WriteString(") ")
		args = append(args, a...)
	}

	return args
}

func (q *Query) buildQueryFilterColumn(table string, filter request.Filter) (string, []any) {
	switch filter.Dimension.(type) {
	case dimensions.TagKey:
		tagKeys := fmt.Sprintf("mapKeys(%s)", q.buildQualifiedColumn(table, "tags"))

		switch filter.Operator {
		case request.OperatorIsNot:
			if len(filter.Values) > 1 {
				group := make([]string, 0, len(filter.Values))

				for range filter.Values {
					group = append(group, fmt.Sprintf("has(%s, ?) = 0", tagKeys))
				}

				return strings.Join(group, " OR "), filter.Values
			}

			return fmt.Sprintf("has(%s, ?) = 0", tagKeys), filter.Values
		case request.OperatorContains:
			if len(filter.Values) > 1 {
				group := make([]string, 0, len(filter.Values))

				for range filter.Values {
					group = append(group, fmt.Sprintf("arrayExists(k -> ilike(k, '%%' || ? || '%%'), %s)", tagKeys))
				}

				return strings.Join(group, " OR "), filter.Values
			}

			return fmt.Sprintf("arrayExists(k -> ilike(k, '%%' || ? || '%%'), %s)", tagKeys), filter.Values
		case request.OperatorContainsNot:
			if len(filter.Values) > 1 {
				group := make([]string, 0, len(filter.Values))

				for range filter.Values {
					group = append(group, fmt.Sprintf("arrayExists(k -> ilike(k, '%%' || ? || '%%'), %s) = 0", tagKeys))
				}

				return strings.Join(group, " OR "), filter.Values
			}

			return fmt.Sprintf("arrayExists(k -> ilike(k, '%%' || ? || '%%'), %s) = 0", tagKeys), filter.Values
		case request.OperatorMatches:
			if len(filter.Values) > 1 {
				group := make([]string, 0, len(filter.Values))

				for range filter.Values {
					group = append(group, fmt.Sprintf("arrayExists(k -> match(k, ?), %s)", tagKeys))
				}

				return strings.Join(group, " OR "), filter.Values
			}

			return fmt.Sprintf("arrayExists(k -> match(k, ?), %s)", tagKeys), filter.Values
		case request.OperatorMatchesNot:
			if len(filter.Values) > 1 {
				group := make([]string, 0, len(filter.Values))

				for range filter.Values {
					group = append(group, fmt.Sprintf("arrayExists(k -> match(k, ?), %s) = 0", tagKeys))
				}

				return strings.Join(group, " OR "), filter.Values
			}

			return fmt.Sprintf("arrayExists(k -> match(k, ?), %s) = 0", tagKeys), filter.Values
		default:
			if len(filter.Values) > 1 {
				group := make([]string, 0, len(filter.Values))

				for range filter.Values {
					group = append(group, fmt.Sprintf("has(%s, ?)", tagKeys))
				}

				return strings.Join(group, " OR "), filter.Values
			}

			return fmt.Sprintf("has(%s, ?)", tagKeys), filter.Values
		}
	case dimensions.EventMetaKey:
		expression, _ := q.buildQueryWhereColumn(table, filter)

		switch filter.Operator {
		case request.OperatorIsNot:
			if len(filter.Values) > 1 {
				group := make([]string, 0, len(filter.Values))

				for _, v := range filter.Values {
					group = append(group, fmt.Sprintf(`isNull(%s%s)`, expression, q.buildQueryFilterJSONPath(v.(string))))
				}

				return strings.Join(group, " OR "), nil
			}

			return fmt.Sprintf(`isNull(%s%s)`, expression, q.buildQueryFilterJSONPath(filter.Values[0].(string))), nil
		default:
			if len(filter.Values) > 1 {
				group := make([]string, 0, len(filter.Values))

				for _, v := range filter.Values {
					group = append(group, fmt.Sprintf(`isNotNull(%s%s)`, expression, q.buildQueryFilterJSONPath(v.(string))))
				}

				return strings.Join(group, " OR "), nil
			}

			return fmt.Sprintf(`isNotNull(%s%s)`, expression, q.buildQueryFilterJSONPath(filter.Values[0].(string))), nil
		}
	default:
		expression, args := q.buildQueryWhereColumn(table, filter)

		switch filter.Operator {
		case request.OperatorIsNot:
			if len(filter.Values) > 1 {
				args = append(args, filter.Values)
				return fmt.Sprintf("%s NOT IN (?)", expression), args
			}

			args = append(args, filter.Values...)
			return fmt.Sprintf("%s != ?", expression), args
		case request.OperatorContains:
			values := q.buildQueryFilterILikeValues(filter.Values)

			if len(values) > 1 {
				args = append(args, values)
				return fmt.Sprintf("arrayExists(v -> ilike(%s, '%%' || v || '%%'), ?)", expression), args
			}

			args = append(args, values...)
			return fmt.Sprintf("%s ILIKE '%%' || ? || '%%'", expression), args
		case request.OperatorContainsNot:
			values := q.buildQueryFilterILikeValues(filter.Values)

			if len(values) > 1 {
				args = append(args, values)
				return fmt.Sprintf("arrayExists(v -> ilike(%s, '%%' || v || '%%'), ?) = 0", expression), args
			}

			args = append(args, values...)
			return fmt.Sprintf("%s NOT ILIKE '%%' || ? || '%%'", expression), args
		case request.OperatorMatches:
			if len(filter.Values) > 1 {
				args = append(args, filter.Values)
				return fmt.Sprintf("multiMatchAny(%s, ?)", expression), args
			}

			args = append(args, filter.Values...)
			return fmt.Sprintf("match(%s, ?)", expression), args
		case request.OperatorMatchesNot:
			if len(filter.Values) > 1 {
				args = append(args, filter.Values)
				return fmt.Sprintf("multiMatchAny(%s, ?) = 0", expression), args
			}

			args = append(args, filter.Values...)
			return fmt.Sprintf("match(%s, ?) = 0", expression), args
		default:
			if len(filter.Values) > 1 {
				args = append(args, filter.Values)
				return fmt.Sprintf("%s IN (?)", expression), args
			}

			args = append(args, filter.Values...)
			return fmt.Sprintf("%s = ?", expression), args
		}
	}
}

func (q *Query) buildQueryWhereColumn(table string, filter request.Filter) (string, []any) {
	switch d := filter.Dimension.(type) {
	case dimensions.TagValue:
		return fmt.Sprintf("%s[?]", q.buildQualifiedColumn(table, "tags")), []any{d.Key}
	case dimensions.EventMeta:
		return fmt.Sprintf("%s%s", q.buildQualifiedColumn(table, "meta_data"), q.buildQueryFilterJSONPath(d.Path)), nil
	case dimensions.EventMetaValue:
		return fmt.Sprintf("%s%s", q.buildQualifiedColumn(table, "meta_data"), q.buildQueryFilterJSONPath(d.Path)), nil
	case dimensions.EventMetaKey:
		return q.buildQualifiedColumn(table, "meta_data"), nil
	default:
		return q.buildQualifiedColumn(table, d.Column(table)), nil
	}
}

func (q *Query) buildQualifiedColumn(table, column string) string {
	return fmt.Sprintf(`"%s"."%s"`, table, column)
}

func (q *Query) buildQueryFilterJSONPath(value string) string {
	parts := strings.Split(value, ".")
	var path strings.Builder

	for _, p := range parts {
		if index, err := strconv.Atoi(p); err == nil {
			path.WriteString(fmt.Sprintf("[%d]", index+1))
		} else {
			path.WriteString(fmt.Sprintf(`."%s"`, p))
		}
	}

	return path.String()
}

func (q *Query) buildQueryFilterILikeValues(filterValues []any) []any {
	values := make([]any, 0, len(filterValues))

	for _, v := range filterValues {
		values = append(values, likeEscaper.Replace(fmt.Sprint(v)))
	}

	return values
}

func (q *Query) buildQueryGroupBy(d []dimensions.Dimension) string {
	if q.funnelTimeColumn != "" {
		return "GROUP BY visitor_id, session_id "
	}

	if q.joinStep == 0 && len(d) > 0 {
		fields := make([]string, 0, len(d))

		for _, dimension := range d {
			column := ""

			if eventMeta, ok := dimension.(dimensions.EventMeta); ok {
				if eventMeta.Function == dimensions.EventMetaFunctionNone {
					column = eventMeta.Column("")
				}
			} else if eventMetaKey, ok := dimension.(dimensions.EventMetaKey); ok {
				if !eventMetaKey.Any {
					column = dimension.Column(q.primaryTable)
				}
			} else if country, ok := dimension.(dimensions.Country); ok {
				if !country.Any {
					column = dimension.Column(q.primaryTable)
				}
			} else if region, ok := dimension.(dimensions.Region); ok {
				if !region.Any {
					column = dimension.Column(q.primaryTable)
				}
			} else if referrer, ok := dimension.(dimensions.Referrer); ok {
				if !referrer.Any {
					column = dimension.Column(q.primaryTable)
				}
			} else {
				column = dimension.Column(q.primaryTable)
			}

			if column != "" {
				fields = append(fields, column)
			}
		}

		if len(fields) > 0 {
			return fmt.Sprintf("GROUP BY %s ", strings.Join(fields, ","))
		}
	}

	return ""
}

func (q *Query) buildOrderBy(req request.Request) (string, []any) {
	if len(req.OrderBy) == 0 {
		return "", nil
	}

	fields := make([]string, 0, len(req.OrderBy))
	args := make([]any, 0)

	for _, o := range req.OrderBy {
		direction := o.Direction

		if direction == "" {
			direction = request.DirectionDESC
		}

		field := ""

		if o.Dimension != nil {
			field = fmt.Sprintf("%s %s", o.Dimension.Column(q.primaryTable), direction)
			withFill, a := q.buildQueryWithFill(req, o.Dimension)

			if withFill != "" {
				field += " " + withFill
				args = append(args, a...)
			}
		} else {
			field = fmt.Sprintf("%s %s", o.Metric.Column(), direction)
		}

		fields = append(fields, field)
	}

	return fmt.Sprintf("ORDER BY %s ", strings.Join(fields, ",")), args
}

func (q *Query) buildQueryWithFill(req request.Request, dimension dimensions.Dimension) (string, []any) {
	if req.Period.From.IsZero() || req.Period.To.IsZero() {
		return "", nil
	}

	tz := q.timezoneName(req.Period)

	// use wall clock, so that it won't be converted by the driver
	args := []any{
		req.Period.From.Format(time.DateTime),
		req.Period.To.Format(time.DateTime),
	}

	switch dimension.(type) {
	case dimensions.Minute:
		if req.Period.IncludeTime {
			return fmt.Sprintf("WITH FILL FROM toStartOfMinute(toDateTime(?, '%s')) TO toDateTime(?, '%s') STEP INTERVAL 1 MINUTE", tz, tz), args
		}

		return fmt.Sprintf("WITH FILL FROM toStartOfDay(toDateTime(?, '%s')) TO toStartOfDay(toDateTime(?, '%s')) + INTERVAL 1 DAY STEP INTERVAL 1 MINUTE", tz, tz), args
	case dimensions.Hour:
		if req.Period.IncludeTime {
			return fmt.Sprintf("WITH FILL FROM toStartOfHour(toDateTime(?, '%s')) TO toDateTime(?, '%s') STEP INTERVAL 1 HOUR", tz, tz), args
		}

		return fmt.Sprintf("WITH FILL FROM toStartOfDay(toDateTime(?, '%s')) TO toStartOfDay(toDateTime(?, '%s')) + INTERVAL 1 DAY STEP INTERVAL 1 HOUR", tz, tz), args
	case dimensions.Day:
		return fmt.Sprintf("WITH FILL FROM toDate(?, '%s') TO toDate(?, '%s') + INTERVAL 1 DAY STEP INTERVAL 1 DAY", tz, tz), args
	case dimensions.Week:
		return fmt.Sprintf("WITH FILL FROM toStartOfWeek(toDate(?, '%s'), %d) TO toDate(?, '%s') + INTERVAL 1 DAY STEP INTERVAL 1 WEEK", tz, req.Period.WeekdayMode, tz), args
	case dimensions.Month:
		return fmt.Sprintf("WITH FILL FROM toStartOfMonth(toDate(?, '%s')) TO toDate(?, '%s') + INTERVAL 1 DAY STEP INTERVAL 1 MONTH", tz, tz), args
	case dimensions.Year:
		return fmt.Sprintf("WITH FILL FROM toStartOfYear(toDate(?, '%s')) TO toDate(?, '%s') + INTERVAL 1 DAY STEP INTERVAL 1 YEAR", tz, tz), args
	}

	return "", nil
}

func (q *Query) buildQueryPagination(pagination *request.Pagination) string {
	if pagination != nil && pagination.Limit > 0 {
		if pagination.Offset > 0 {
			return fmt.Sprintf("LIMIT %d, %d", pagination.Offset, pagination.Limit)
		}

		return fmt.Sprintf("LIMIT %d", pagination.Limit)
	}

	return ""
}

func (q *Query) buildQueryHaving() string {
	if q.primaryTable == pkg.TableSessions {
		return "HAVING sum(sign) > 0 "
	}

	return ""
}

func (q *Query) scanRows(rows driver.Rows, dimensions []dimensions.Dimension, metrics []metrics.Metric) ([]report.Result, error) {
	metricCount := len(metrics)
	dimensionCount := len(dimensions)
	columnCount := metricCount + dimensionCount
	columns := make([]any, columnCount)
	columnPtrs := make([]any, columnCount)

	for i := range metricCount {
		columns[i] = metrics[i].ScanType()
		columnPtrs[i] = columns[i]
	}

	for i := range dimensionCount {
		columns[metricCount+i] = dimensions[i].ScanType()
		columnPtrs[metricCount+i] = columns[metricCount+i]
	}

	results := make([]report.Result, 0)

	for rows.Next() {
		if err := rows.Scan(columnPtrs...); err != nil {
			return nil, err
		}

		result := report.Result{
			MetricValues:    make([]any, metricCount),
			DimensionValues: make([]any, dimensionCount),
		}

		for i := range metricCount {
			result.MetricValues[i] = reflect.ValueOf(columnPtrs[i]).Elem().Interface()
		}

		for i := range dimensionCount {
			result.DimensionValues[i] = reflect.ValueOf(columnPtrs[metricCount+i]).Elem().Interface()
		}

		results = append(results, result)
	}

	return results, rows.Err()
}
