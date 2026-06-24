package request

import (
	"testing"

	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/dimensions"
	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/metrics"
	"github.com/stretchr/testify/assert"
)

func TestValidateOrderBy(t *testing.T) {
	// tag value
	errs := validateOrderBy([]OrderBy{
		{
			Dimension: dimensions.TagValue{
				Key: "author",
			},
		},
	}, []dimensions.Dimension{
		// key is missing
		dimensions.TagValue{},
	}, []metrics.Metric{})
	assert.Len(t, errs, 1)

	errs = validateOrderBy([]OrderBy{
		{
			Dimension: dimensions.TagValue{
				Key: "author",
			},
		},
	}, []dimensions.Dimension{
		dimensions.TagValue{
			Key: "author",
		},
	}, []metrics.Metric{})
	assert.Empty(t, errs)

	// event metadata path
	errs = validateOrderBy([]OrderBy{
		{
			Dimension: dimensions.EventMeta{
				Path: "author",
			},
		},
	}, []dimensions.Dimension{
		// path is missing
		dimensions.EventMeta{},
	}, []metrics.Metric{})
	assert.Len(t, errs, 1)

	errs = validateOrderBy([]OrderBy{
		{
			Dimension: dimensions.EventMeta{
				Path: "author",
			},
		},
	}, []dimensions.Dimension{
		dimensions.EventMeta{
			Path: "author",
		},
	}, []metrics.Metric{})
	assert.Empty(t, errs)
}
