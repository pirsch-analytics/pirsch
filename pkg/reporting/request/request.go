package request

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/dimensions"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/metrics"
)

var (
	metaKeyValidKeySegment = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// Request generates a Report.
type Request struct {
	// Ctx can be used to set a timeout or to cancel queries.
	Ctx context.Context

	// SiteID is the site ID for the request.
	SiteID uint64

	// Period is the period and timezone for the Report.
	Period Period

	// Metrics is a list of result fields for the Report.
	Metrics []metrics.Metric

	// Dimensions is a list of attributes for the Request.
	Dimensions []dimensions.Dimension

	// Filter filters the results for a Report.
	// Top-level filters are connected using the AND operator by default.
	// To use other operators, they need to be set in the Filter recursively.
	Filter []Filter

	// OrderBy sorts the result fields of a Report.
	OrderBy []OrderBy

	// Pagination limits the number of results for a Report.
	Pagination *Pagination

	// Options are optional fields for the Request.
	Options *Options
}

// Validate validates the Request and returns an error if a report.Report cannot be constructed for the specified fields.
func (r *Request) Validate() []error {
	if r.Ctx == nil {
		r.Ctx = context.Background()
	}

	errs := make([]error, 0)

	if r.SiteID == 0 {
		errs = append(errs, errors.New("SiteID is required"))
	}

	if r.Period.Timezone == nil {
		r.Period.Timezone = time.UTC
	}

	if r.Period.WeekdayMode == 0 {
		r.Period.WeekdayMode = WeekdayMonday
	}

	if r.Period.From.After(r.Period.To) {
		r.Period.From, r.Period.To = r.Period.To, r.Period.From
	}

	for _, f := range r.Filter {
		errs = append(errs, r.validateFilterValues(f)...)
	}

	// TODO check other relevant fields and filter combinations

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (r *Request) validateFilterValues(filter Filter) []error {
	errs := make([]error, 0)

	switch filter.Dimension.(type) {
	case dimensions.EventMetaKey:
		for _, values := range filter.Values {
			v, ok := values.(string)

			if !ok {
				errs = append(errs, fmt.Errorf("metadata key value must be a string"))
				break
			}

			if err := r.validateMetadataKey(v); err != nil {
				errs = append(errs, err)
			}
		}
	}

	for _, f := range filter.Filter {
		errs = append(errs, r.validateFilterValues(f)...)
	}

	return errs
}

func (r *Request) validateMetadataKey(path string) error {
	if path == "" {
		return errors.New("metadata key path must not be empty")
	}

	parts := strings.Split(path, ".")

	for _, part := range parts {
		if part == "" {
			return fmt.Errorf("metadata key path '%s' must not contain empty segments", path)
		}

		if !metaKeyValidKeySegment.MatchString(part) {
			return fmt.Errorf("metadata key path '%s' segment '%s' contains invalid characters: only a-z, A-Z, 0-9, _ and - are allowed", path, part)
		}
	}

	return nil
}
