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
				Operator:  OperatorGroupOr,
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

func TestRequestString(t *testing.T) {
	r := Request{
		SiteID: 1,
		Filter: []Filter{
			{
				Operator:  OperatorGroupOr,
				Dimension: dimensions.City{},
				Values:    []any{"London"},
			},
		},
	}
	assert.Equal(t, `{"Ctx":null,"SiteID":1,"Period":{"From":"0001-01-01T00:00:00Z","To":"0001-01-01T00:00:00Z","Timezone":null,"WeekdayMode":0,"IncludeTime":false,"Compare":null},"Metrics":[],"Dimensions":[],"Filter":[{"Operator":1,"Dimension":{},"Values":["London"],"Filter":[]}],"OrderBy":[],"Pagination":null,"Options":null}`, r.String())
}
