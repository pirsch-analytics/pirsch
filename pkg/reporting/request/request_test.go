package request

import (
	"testing"

	"github.com/pirsch-analytics/pirsch/v7/pkg/reporting/dimensions"
	"github.com/stretchr/testify/assert"
)

func TestRequestValidateEventMetaKey(t *testing.T) {
	r := Request{
		SiteID: 1,
		Filter: []Filter{
			{
				Operator:  OperatorOr,
				Dimension: dimensions.EventMetaKey{},
				Values:    []any{"this.is.fine"},
				Filter: []Filter{
					{
						Dimension: dimensions.EventMetaKey{},
						Values:    []any{"this.is.not.(DELETE FROM"},
					},
				},
			},
		},
	}
	errs := r.Validate()
	assert.Len(t, errs, 1)
	assert.Equal(t, "metadata key path 'this.is.not.(DELETE FROM' segment '(DELETE FROM' contains invalid characters: only a-z, A-Z, 0-9, _ and - are allowed", errs[0].Error())
}
