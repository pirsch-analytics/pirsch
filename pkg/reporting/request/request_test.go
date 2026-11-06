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

func TestRequestValidateEventMetaPath(t *testing.T) {
	valid := dimensions.EventMeta{Path: "items.0.price", ColumnName: "price_sum", Function: dimensions.EventMetaFunctionSum}
	r := Request{
		SiteID:     1,
		Dimensions: []dimensions.Dimension{valid, dimensions.EventMetaValue{Path: "label"}},
		Filter: []Filter{
			{Dimension: dimensions.EventMeta{Path: "position"}, Values: []any{"hero"}},
		},
		OrderBy: []OrderBy{{Dimension: valid}},
	}
	assert.Empty(t, r.Validate())

	for _, r := range []Request{
		{SiteID: 1, Dimensions: []dimensions.Dimension{dimensions.EventMeta{Path: `x") OR 1=1 --`}}},
		{SiteID: 1, Dimensions: []dimensions.Dimension{dimensions.EventMeta{Path: "price", ColumnName: "x; DROP TABLE y"}}},
		{SiteID: 1, Dimensions: []dimensions.Dimension{dimensions.EventMetaValue{Path: "a..b"}}},
		{SiteID: 1, Dimensions: []dimensions.Dimension{dimensions.EventMetaValue{Path: "price", ColumnName: "1abc"}}},
		{SiteID: 1, Filter: []Filter{{
			Operator: OperatorGroupOr,
			Filter:   []Filter{{Dimension: dimensions.EventMetaValue{Path: `a"b`}, Values: []any{"x"}}},
		}}},
		{SiteID: 1, OrderBy: []OrderBy{{Dimension: dimensions.EventMeta{Path: `a"b`}}}},
	} {
		assert.Len(t, r.Validate(), 1, r.String())
	}
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
	assert.Equal(t, `{"Ctx":null,"SiteID":1,"Period":{"From":"0001-01-01T00:00:00Z","To":"0001-01-01T00:00:00Z","Timezone":null,"WeekdayMode":0,"IncludeTime":false,"Compare":null,"ImportedUntil":"0001-01-01T00:00:00Z"},"Metrics":[],"Dimensions":[],"Filter":[{"Operator":1,"Dimension":{},"Values":["London"],"Filter":[]}],"OrderBy":[],"Pagination":null,"Options":null}`, r.String())
}
