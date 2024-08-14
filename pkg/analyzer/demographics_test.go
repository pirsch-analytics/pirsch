package analyzer

import (
	"fmt"
	"github.com/pirsch-analytics/pirsch/v6/pkg"
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"github.com/pirsch-analytics/pirsch/v6/pkg/util"
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
	analyzer := NewAnalyzer(dbClient)
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
			Direction: pkg.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldLanguage,
			Input: "en,jp",
		},
	}})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)

	// imported statistics
	yesterday := util.PastDay(1).Format(time.DateOnly)
	_, err = dbClient.Exec(fmt.Sprintf(`INSERT INTO "imported_language" (date, language, visitors) VALUES
		('%s', 'ru', 2), ('%s', 'en', 1)`, yesterday, yesterday))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	visitors, err = analyzer.Demographics.Languages(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 4)
	assert.Equal(t, "en", visitors[0].Language)
	assert.Equal(t, "de", visitors[1].Language)
	assert.Equal(t, "ru", visitors[2].Language)
	assert.Equal(t, "jp", visitors[3].Language)
	assert.Equal(t, 4, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 2, visitors[2].Visitors)
	assert.Equal(t, 1, visitors[3].Visitors)
	assert.InDelta(t, 0.4444, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.2222, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.2222, visitors[2].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1111, visitors[3].RelativeVisitors, 0.01)
	visitors, err = analyzer.Demographics.Languages(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
		Language:      []string{"en"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "en", visitors[0].Language)
	assert.Equal(t, 4, visitors[0].Visitors)
	assert.InDelta(t, 0.4444, visitors[0].RelativeVisitors, 0.01)
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
	analyzer := NewAnalyzer(dbClient)
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
	visitors, err = analyzer.Demographics.Countries(&Filter{
		Country: []string{"en"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "en", visitors[0].CountryCode)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	visitors, err = analyzer.Demographics.Countries(&Filter{
		Country: []string{"!en"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)
	assert.Equal(t, "de", visitors[0].CountryCode)
	assert.Equal(t, "jp", visitors[1].CountryCode)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 1, visitors[1].Visitors)
	assert.InDelta(t, 0.33, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[1].RelativeVisitors, 0.01)
	_, err = analyzer.Demographics.Countries(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Demographics.Countries(getMaxFilter("event"))
	assert.NoError(t, err)
	visitors, err = analyzer.Demographics.Countries(&Filter{Offset: 0, Limit: 10, Sort: []Sort{
		{
			Field:     FieldCountry,
			Direction: pkg.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldCountry,
			Input: "en,jp",
		},
	}})
	assert.NoError(t, err)
	assert.Len(t, visitors, 2)

	// imported statistics
	yesterday := util.PastDay(1).Format(time.DateOnly)
	_, err = dbClient.Exec(fmt.Sprintf(`INSERT INTO "imported_country" (date, country_code, visitors) VALUES
		('%s', 'ru', 2), ('%s', 'en', 1)`, yesterday, yesterday))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	visitors, err = analyzer.Demographics.Countries(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 4)
	assert.Equal(t, "en", visitors[0].CountryCode)
	assert.Equal(t, "de", visitors[1].CountryCode)
	assert.Equal(t, "ru", visitors[2].CountryCode)
	assert.Equal(t, "jp", visitors[3].CountryCode)
	assert.Equal(t, 4, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 2, visitors[2].Visitors)
	assert.Equal(t, 1, visitors[3].Visitors)
	assert.InDelta(t, 0.4444, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.2222, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.2222, visitors[2].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1111, visitors[3].RelativeVisitors, 0.01)
	visitors, err = analyzer.Demographics.Countries(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
		Country:       []string{"ru"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "ru", visitors[0].CountryCode)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.InDelta(t, 0.2222, visitors[0].RelativeVisitors, 0.01)
}

func TestAnalyzer_Regions(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), CountryCode: "no", City: "Oslo"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Start: time.Now(), CountryCode: "no", City: "Oslo"},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), CountryCode: "gb", Region: "England", City: "London"},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Start: time.Now(), CountryCode: "de", Region: "Berlin", City: "Berlin"},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), CountryCode: "de"},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), CountryCode: "jp", Region: "Tokyo", City: "Tokyo"},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Start: time.Now(), CountryCode: "gb", Region: "England", City: "London"},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Start: time.Now(), CountryCode: "gb"},
		},
	})
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Demographics.Regions(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 4)
	assert.Empty(t, visitors[0].CountryCode)
	assert.Equal(t, "gb", visitors[1].CountryCode)
	assert.Equal(t, "de", visitors[2].CountryCode)
	assert.Equal(t, "jp", visitors[3].CountryCode)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	assert.Equal(t, 1, visitors[3].Visitors)
	assert.InDelta(t, 0.5, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[2].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, visitors[3].RelativeVisitors, 0.01)
	_, err = analyzer.Demographics.Regions(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.Demographics.Regions(getMaxFilter("event"))
	assert.NoError(t, err)
	visitors, err = analyzer.Demographics.Regions(&Filter{Offset: 1, Limit: 10, Sort: []Sort{
		{
			Field:     FieldRegion,
			Direction: pkg.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldRegion,
			Input: "e",
		},
	}})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "England", visitors[0].Region)
	assert.Equal(t, "gb", visitors[0].CountryCode)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.InDelta(t, 0.3333, visitors[0].RelativeVisitors, 0.01)

	// imported statistics
	yesterday := util.PastDay(1).Format(time.DateOnly)
	_, err = dbClient.Exec(fmt.Sprintf(`INSERT INTO "imported_region" (date, region, visitors) VALUES
		('%s', 'Berlin', 2), ('%s', 'England', 1)`, yesterday, yesterday))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	visitors, err = analyzer.Demographics.Regions(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 4)
	assert.Equal(t, "", visitors[0].CountryCode)
	assert.Equal(t, "de", visitors[1].CountryCode)
	assert.Equal(t, "gb", visitors[2].CountryCode)
	assert.Equal(t, "jp", visitors[3].CountryCode)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 3, visitors[1].Visitors)
	assert.Equal(t, 3, visitors[2].Visitors)
	assert.Equal(t, 1, visitors[3].Visitors)
	assert.InDelta(t, 0.3333, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.3333, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.3333, visitors[2].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1111, visitors[3].RelativeVisitors, 0.01)
	visitors, err = analyzer.Demographics.Regions(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
		Region:        []string{"Berlin"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "de", visitors[0].CountryCode)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.InDelta(t, 0.3333, visitors[0].RelativeVisitors, 0.01)
}

func TestAnalyzer_Cities(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), CountryCode: "no", City: "Oslo"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Start: time.Now(), CountryCode: "no", City: "Oslo"},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), CountryCode: "gb", Region: "England", City: "London"},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Start: time.Now(), CountryCode: "de", Region: "Berlin", City: "Berlin"},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), CountryCode: "de"},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), CountryCode: "jp", Region: "Tokyo", City: "Tokyo"},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Start: time.Now(), CountryCode: "gb", Region: "England", City: "London"},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Start: time.Now(), CountryCode: "gb"},
		},
	})
	analyzer := NewAnalyzer(dbClient)
	visitors, err := analyzer.Demographics.Cities(nil)
	assert.NoError(t, err)
	assert.Len(t, visitors, 4)
	assert.Empty(t, visitors[0].CountryCode)
	assert.Equal(t, "gb", visitors[1].CountryCode)
	assert.Equal(t, "de", visitors[2].CountryCode)
	assert.Equal(t, "jp", visitors[3].CountryCode)
	assert.Empty(t, visitors[0].Region)
	assert.Equal(t, "England", visitors[1].Region)
	assert.Equal(t, "Berlin", visitors[2].Region)
	assert.Equal(t, "Tokyo", visitors[3].Region)
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
			Direction: pkg.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldCity,
			Input: "New York",
		},
	}})
	assert.NoError(t, err)

	// imported statistics
	yesterday := util.PastDay(1).Format(time.DateOnly)
	_, err = dbClient.Exec(fmt.Sprintf(`INSERT INTO "imported_city" (date, city, visitors) VALUES
		('%s', 'Berlin', 2), ('%s', 'London', 1)`, yesterday, yesterday))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 100)
	visitors, err = analyzer.Demographics.Cities(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 4)
	assert.Equal(t, "de", visitors[0].CountryCode)
	assert.Equal(t, "gb", visitors[1].CountryCode)
	assert.Equal(t, "", visitors[2].CountryCode)
	assert.Equal(t, "jp", visitors[3].CountryCode)
	assert.Equal(t, "Berlin", visitors[0].Region)
	assert.Equal(t, "England", visitors[1].Region)
	assert.Equal(t, "", visitors[2].Region)
	assert.Equal(t, "Tokyo", visitors[3].Region)
	assert.Equal(t, "Berlin", visitors[0].City)
	assert.Equal(t, "London", visitors[1].City)
	assert.Equal(t, "", visitors[2].City)
	assert.Equal(t, "Tokyo", visitors[3].City)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.Equal(t, 3, visitors[1].Visitors)
	assert.Equal(t, 2, visitors[2].Visitors)
	assert.Equal(t, 1, visitors[3].Visitors)
	assert.InDelta(t, 0.3333, visitors[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.3333, visitors[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.2222, visitors[2].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1111, visitors[3].RelativeVisitors, 0.01)
	visitors, err = analyzer.Demographics.Cities(&Filter{
		From:          util.PastDay(1),
		To:            util.Today(),
		ImportedUntil: util.Today(),
		City:          []string{"London"},
	})
	assert.NoError(t, err)
	assert.Len(t, visitors, 1)
	assert.Equal(t, "gb", visitors[0].CountryCode)
	assert.Equal(t, "England", visitors[0].Region)
	assert.Equal(t, "London", visitors[0].City)
	assert.Equal(t, 3, visitors[0].Visitors)
	assert.InDelta(t, 0.3333, visitors[0].RelativeVisitors, 0.01)
}
