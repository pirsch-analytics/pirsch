package request

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/dimensions"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/metrics"
)

func validateSiteID(siteID uint64) error {
	if siteID == 0 {
		return errors.New("SiteID is required")
	}

	return nil
}

func validateFilterValues(filter Filter) []error {
	errs := make([]error, 0)

	switch filter.Dimension.(type) {
	case dimensions.EventMetaKey:
		for _, values := range filter.Values {
			v, ok := values.(string)

			if !ok {
				errs = append(errs, fmt.Errorf("metadata key value must be a string"))
				break
			}

			if err := validateMetadataKey(v); err != nil {
				errs = append(errs, err)
			}
		}
	}

	for _, f := range filter.Filter {
		errs = append(errs, validateFilterValues(f)...)
	}

	return errs
}

func validateMetadataKey(path string) error {
	if path == "" {
		return errors.New("metadata key path must not be empty")
	}

	parts := strings.SplitSeq(path, ".")

	for part := range parts {
		if part == "" {
			return fmt.Errorf("metadata key path '%s' must not contain empty segments", path)
		}

		if !metaKeyValidKeySegment.MatchString(part) {
			return fmt.Errorf("metadata key path '%s' segment '%s' contains invalid characters: only a-z, A-Z, 0-9, _ and - are allowed", path, part)
		}
	}

	return nil
}

func validateOrderBy(order []OrderBy, requestDimensions []dimensions.Dimension, requestMetrics []metrics.Metric) []error {
	errs := make([]error, 0)

	for _, o := range order {
		if o.Dimension != nil {
			if err := validateOrderByDimension(o.Dimension, requestDimensions); err != nil {
				errs = append(errs, err)
			}
		} else if o.Metric != nil {
			if err := validateOrderByMetric(o.Metric, requestMetrics); err != nil {
				errs = append(errs, err)
			}
		} else {
			errs = append(errs, errors.New("order by must have either a dimension or a metric set"))
		}
	}

	return errs
}

func validateOrderByDimension(order dimensions.Dimension, requestDimensions []dimensions.Dimension) error {
	switch o := order.(type) {
	case dimensions.TagValue:
		if o.Key == "" {
			return nil
		}

		// key must match a TagValue dimension
		for _, d := range requestDimensions {
			if tv, ok := d.(dimensions.TagValue); ok && tv.Key == o.Key {
				return nil
			}
		}

		return fmt.Errorf("order by tag value key %q not found in dimensions", o.Key)
	case dimensions.EventMeta:
		if o.Path == "" {
			return nil
		}

		// path must match an EventMeta dimension
		for _, d := range requestDimensions {
			if meta, ok := d.(dimensions.EventMeta); ok && meta.Path == o.Path {
				return nil
			}
		}

		return fmt.Errorf("order by event meta path %q not found in dimensions", o.Path)
	default:
		// for non-parameterized dimensions, check by column name
		column := order.Column("")

		for _, d := range requestDimensions {
			if d.Column("") == column {
				return nil
			}
		}

		return fmt.Errorf("order by dimension %q not found in dimensions", column)
	}
}

func validateOrderByMetric(order metrics.Metric, requestMetrics []metrics.Metric) error {
	column := order.Column()

	for _, m := range requestMetrics {
		if m.Column() == column {
			return nil
		}
	}

	return fmt.Errorf("order by metric %q not found in metrics", column)
}
