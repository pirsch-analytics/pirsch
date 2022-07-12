package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v4"
	"github.com/pirsch-analytics/pirsch/v4/db"
	"github.com/pirsch-analytics/pirsch/v4/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAnalyzer_Languages(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), Language: "ru"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Start: time.Now(), Language: "ru"},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), Language: "en"},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Start: time.Now(), Language: "de"},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), Language: "de"},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), Language: "jp"},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Start: time.Now(), Language: "en"},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Start: time.Now(), Language: "en"},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.Demographics.Languages(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "en", visitors[0].Language)
	assert.Equal(t, "de", visitors[1].Language)
	assert.Equal(t, "jp", visitors[2].Language)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	_, err = analyzer.Demographics.Languages(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Demographics.Languages(getMaxFilter("event"))
	assert.NoError(t, err)
	visitors, err = analyzer.Demographics.Languages(&Filter{Offset: 0, Limit: 10, Sort: []Sort{
		{
			Field:     FieldLanguage,
			Direction: pirsch.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldLanguage,
			Input: "en,jp",
		},
	}})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)
}

func TestAnalyzer_Countries(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), CountryCode: "ru"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Start: time.Now(), CountryCode: "ru"},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), CountryCode: "en"},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Start: time.Now(), CountryCode: "de"},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), CountryCode: "de"},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), CountryCode: "jp"},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Start: time.Now(), CountryCode: "en"},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Start: time.Now(), CountryCode: "en"},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.Demographics.Countries(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "en", visitors[0].CountryCode)
	assert.Equal(t, "de", visitors[1].CountryCode)
	assert.Equal(t, "jp", visitors[2].CountryCode)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	_, err = analyzer.Demographics.Countries(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Demographics.Countries(getMaxFilter("event"))
	assert.NoError(t, err)
	visitors, err = analyzer.Demographics.Countries(&Filter{Offset: 0, Limit: 10, Sort: []Sort{
		{
			Field:     FieldCountry,
			Direction: pirsch.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldCountry,
			Input: "en,jp",
		},
	}})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)
}

func TestAnalyzer_Cities(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), CountryCode: "no", City: "Oslo"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Start: time.Now(), CountryCode: "no", City: "Oslo"},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), CountryCode: "gb", City: "London"},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Start: time.Now(), CountryCode: "de", City: "Berlin"},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), CountryCode: "de", City: ""},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), CountryCode: "jp", City: "Tokyo"},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Start: time.Now(), CountryCode: "gb", City: "London"},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Start: time.Now(), CountryCode: "gb", City: ""},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient, nil)
	visitors, err := analyzer.Demographics.Cities(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 4)
	assert.Empty(t, visitors[0].CountryCode)
	assert.Equal(t, "gb", visitors[1].CountryCode)
	assert.Equal(t, "de", visitors[2].CountryCode)
	assert.Equal(t, "jp", visitors[3].CountryCode)
	assert.Empty(t, visitors[0].City)
	assert.Equal(t, "London", visitors[1].City)
	assert.Equal(t, "Berlin", visitors[2].City)
	assert.Equal(t, "Tokyo", visitors[3].City)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.Equal(t, 1, visitors[3].Visitors)
	assert.InDelta(t, 0.33, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[3].RelativeVisitors, 0.01)
	_, err = analyzer.Demographics.Cities(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Demographics.Cities(getMaxFilter("event"))
	assert.NoError(t, err)
	_, err = analyzer.Demographics.Cities(&Filter{Offset: 1, Limit: 10, Sort: []Sort{
		{
			Field:     FieldCity,
			Direction: pirsch.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldCity,
			Input: "New York",
		},
	}})
	assert.NoError(t, err)
}
