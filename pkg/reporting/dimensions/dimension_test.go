package dimensions

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeDimensionExpressions(t *testing.T) {
	tz, err := time.LoadLocation("Europe/Berlin")
	assert.NoError(t, err)
	withTimezone := &DimensionExpressionOptions{Timezone: tz, WeekdayMode: 2}
	withoutTimezone := &DimensionExpressionOptions{WeekdayMode: 2}

	for _, testCase := range []struct {
		dimension             Dimension
		nilOptions            string
		withoutTimezone       string
		withTimezone          string
		importedWithTimezone  string
		importedWithoutOption string
	}{
		{
			dimension:             Week{},
			nilOptions:            `toStartOfWeek("time", 1)`,
			withoutTimezone:       `toStartOfWeek("time", 2)`,
			withTimezone:          `toStartOfWeek("time", 2, 'Europe/Berlin')`,
			importedWithTimezone:  `toStartOfWeek("date", 2)`,
			importedWithoutOption: `toStartOfWeek("date", 1)`,
		},
		{
			dimension:             Weekday{},
			nilOptions:            `toDayOfWeek("time", 1)`,
			withoutTimezone:       `toDayOfWeek("time", 2)`,
			withTimezone:          `toDayOfWeek("time", 2, 'Europe/Berlin')`,
			importedWithTimezone:  `toDayOfWeek("date", 2)`,
			importedWithoutOption: `toDayOfWeek("date", 1)`,
		},
		{
			dimension:             Day{},
			nilOptions:            `toDate("time")`,
			withoutTimezone:       `toDate("time")`,
			withTimezone:          `toDate("time", 'Europe/Berlin')`,
			importedWithTimezone:  `toDate("date")`,
			importedWithoutOption: `toDate("date")`,
		},
		{
			dimension:             Hour{},
			nilOptions:            `toStartOfHour("time")`,
			withoutTimezone:       `toStartOfHour("time")`,
			withTimezone:          `toStartOfHour("time", 'Europe/Berlin')`,
			importedWithTimezone:  "",
			importedWithoutOption: "",
		},
		{
			dimension:             Minute{},
			nilOptions:            `toStartOfMinute("time")`,
			withoutTimezone:       `toStartOfMinute("time")`,
			withTimezone:          `toStartOfMinute("time", 'Europe/Berlin')`,
			importedWithTimezone:  "",
			importedWithoutOption: "",
		},
		{
			dimension:             Month{},
			nilOptions:            `toStartOfMonth("time")`,
			withoutTimezone:       `toStartOfMonth("time")`,
			withTimezone:          `toStartOfMonth("time", 'Europe/Berlin')`,
			importedWithTimezone:  `toStartOfMonth("date")`,
			importedWithoutOption: `toStartOfMonth("date")`,
		},
		{
			dimension:             Year{},
			nilOptions:            `toStartOfYear("time")`,
			withoutTimezone:       `toStartOfYear("time")`,
			withTimezone:          `toStartOfYear("time", 'Europe/Berlin')`,
			importedWithTimezone:  `toStartOfYear("date")`,
			importedWithoutOption: `toStartOfYear("date")`,
		},
		{
			dimension:       Time{},
			withTimezone:    `toTimezone("time", 'Europe/Berlin')`,
			withoutTimezone: "",
		},
		{
			dimension:       Start{},
			withTimezone:    `toTimezone("start", 'Europe/Berlin')`,
			withoutTimezone: "",
		},
	} {
		name := testCase.dimension.String()
		assert.Equal(t, testCase.nilOptions, testCase.dimension.Expression(nil), name)
		assert.Equal(t, testCase.withoutTimezone, testCase.dimension.Expression(withoutTimezone), name)
		assert.Equal(t, testCase.withTimezone, testCase.dimension.Expression(withTimezone), name)
		assert.Equal(t, testCase.importedWithTimezone, testCase.dimension.ExpressionImported(withTimezone), name)
		assert.Equal(t, testCase.importedWithoutOption, testCase.dimension.ExpressionImported(nil), name)
	}
}
