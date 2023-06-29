package analyzer

import (
	"github.com/pirsch-analytics/pirsch/v5"
	"github.com/pirsch-analytics/pirsch/v5/db"
	"github.com/pirsch-analytics/pirsch/v5/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAnalyzer_UTM(t *testing.T) {
	db.CleanupDB(t, dbClient)
	saveSessions(t, [][]model.Session{
		{
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), UTMSource: "sourceX", UTMMedium: "mediumX", UTMCampaign: "campaignX", UTMContent: "contentX", UTMTerm: "termX"},
		},
		{
			{Sign: -1, VisitorID: 1, Time: time.Now(), Start: time.Now(), UTMSource: "sourceX", UTMMedium: "mediumX", UTMCampaign: "campaignX", UTMContent: "contentX", UTMTerm: "termX"},
			{Sign: 1, VisitorID: 1, Time: time.Now(), Start: time.Now(), UTMSource: "source1", UTMMedium: "medium1", UTMCampaign: "campaign1", UTMContent: "content1", UTMTerm: "term1"},
			{Sign: 1, VisitorID: 2, Time: time.Now(), Start: time.Now(), UTMSource: "source2", UTMMedium: "medium2", UTMCampaign: "campaign2", UTMContent: "content2", UTMTerm: "term2"},
			{Sign: 1, VisitorID: 3, Time: time.Now(), Start: time.Now(), UTMSource: "source2", UTMMedium: "medium2", UTMCampaign: "campaign2", UTMContent: "content2", UTMTerm: "term2"},
			{Sign: 1, VisitorID: 4, Time: time.Now(), Start: time.Now(), UTMSource: "source3", UTMMedium: "medium3", UTMCampaign: "campaign3", UTMContent: "content3", UTMTerm: "term3"},
			{Sign: 1, VisitorID: 5, Time: time.Now(), Start: time.Now(), UTMSource: "source1", UTMMedium: "medium1", UTMCampaign: "campaign1", UTMContent: "content1", UTMTerm: "term1"},
			{Sign: 1, VisitorID: 6, Time: time.Now(), Start: time.Now(), UTMSource: "source1", UTMMedium: "medium1", UTMCampaign: "campaign1", UTMContent: "content1", UTMTerm: "term1"},
		},
	})
	time.Sleep(time.Millisecond * 20)
	analyzer := NewAnalyzer(dbClient)
	source, err := analyzer.UTM.Source(nil)
	assert.NoError(t, err)
	assert.Len(t, source, 3)
	assert.Equal(t, "source1", source[0].UTMSource)
	assert.Equal(t, "source2", source[1].UTMSource)
	assert.Equal(t, "source3", source[2].UTMSource)
	assert.Equal(t, 3, source[0].Visitors)
	assert.Equal(t, 2, source[1].Visitors)
	assert.Equal(t, 1, source[2].Visitors)
	assert.InDelta(t, 0.5, source[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, source[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, source[2].RelativeVisitors, 0.01)
	_, err = analyzer.UTM.Source(getMaxFilter(""))
	assert.NoError(t, err)
	medium, err := analyzer.UTM.Medium(nil)
	assert.NoError(t, err)
	assert.Len(t, medium, 3)
	assert.Equal(t, "medium1", medium[0].UTMMedium)
	assert.Equal(t, "medium2", medium[1].UTMMedium)
	assert.Equal(t, "medium3", medium[2].UTMMedium)
	assert.Equal(t, 3, medium[0].Visitors)
	assert.Equal(t, 2, medium[1].Visitors)
	assert.Equal(t, 1, medium[2].Visitors)
	assert.InDelta(t, 0.5, medium[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, medium[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, medium[2].RelativeVisitors, 0.01)
	_, err = analyzer.UTM.Medium(getMaxFilter(""))
	assert.NoError(t, err)
	campaign, err := analyzer.UTM.Campaign(nil)
	assert.NoError(t, err)
	assert.Len(t, campaign, 3)
	assert.Equal(t, "campaign1", campaign[0].UTMCampaign)
	assert.Equal(t, "campaign2", campaign[1].UTMCampaign)
	assert.Equal(t, "campaign3", campaign[2].UTMCampaign)
	assert.Equal(t, 3, campaign[0].Visitors)
	assert.Equal(t, 2, campaign[1].Visitors)
	assert.Equal(t, 1, campaign[2].Visitors)
	assert.InDelta(t, 0.5, campaign[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, campaign[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, campaign[2].RelativeVisitors, 0.01)
	_, err = analyzer.UTM.Campaign(getMaxFilter(""))
	assert.NoError(t, err)
	content, err := analyzer.UTM.Content(nil)
	assert.NoError(t, err)
	assert.Len(t, content, 3)
	assert.Equal(t, "content1", content[0].UTMContent)
	assert.Equal(t, "content2", content[1].UTMContent)
	assert.Equal(t, "content3", content[2].UTMContent)
	assert.Equal(t, 3, content[0].Visitors)
	assert.Equal(t, 2, content[1].Visitors)
	assert.Equal(t, 1, content[2].Visitors)
	assert.InDelta(t, 0.5, content[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, content[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, content[2].RelativeVisitors, 0.01)
	_, err = analyzer.UTM.Content(getMaxFilter(""))
	assert.NoError(t, err)
	term, err := analyzer.UTM.Term(nil)
	assert.NoError(t, err)
	assert.Len(t, term, 3)
	assert.Equal(t, "term1", term[0].UTMTerm)
	assert.Equal(t, "term2", term[1].UTMTerm)
	assert.Equal(t, "term3", term[2].UTMTerm)
	assert.Equal(t, 3, term[0].Visitors)
	assert.Equal(t, 2, term[1].Visitors)
	assert.Equal(t, 1, term[2].Visitors)
	assert.InDelta(t, 0.5, term[0].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.33, term[1].RelativeVisitors, 0.01)
	assert.InDelta(t, 0.1666, term[2].RelativeVisitors, 0.01)
	_, err = analyzer.UTM.Term(getMaxFilter(""))
	assert.NoError(t, err)
	_, err = analyzer.UTM.Term(getMaxFilter("event"))
	assert.NoError(t, err)
	_, err = analyzer.UTM.Medium(&Filter{Offset: 1, Limit: 10, Sort: []Sort{
		{
			Field:     FieldUTMMedium,
			Direction: pirsch.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldUTMMedium,
			Input: "medium",
		},
	}})
	assert.NoError(t, err)
	_, err = analyzer.UTM.Campaign(&Filter{Offset: 1, Limit: 10, Sort: []Sort{
		{
			Field:     FieldUTMCampaign,
			Direction: pirsch.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldUTMCampaign,
			Input: "campaign",
		},
	}})
	assert.NoError(t, err)
	_, err = analyzer.UTM.Source(&Filter{Offset: 1, Limit: 10, Sort: []Sort{
		{
			Field:     FieldUTMSource,
			Direction: pirsch.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldUTMSource,
			Input: "source",
		},
	}})
	assert.NoError(t, err)
	_, err = analyzer.UTM.Term(&Filter{Offset: 1, Limit: 10, Sort: []Sort{
		{
			Field:     FieldUTMTerm,
			Direction: pirsch.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldUTMTerm,
			Input: "term",
		},
	}})
	assert.NoError(t, err)
	_, err = analyzer.UTM.Content(&Filter{Offset: 1, Limit: 10, Sort: []Sort{
		{
			Field:     FieldUTMContent,
			Direction: pirsch.DirectionASC,
		},
	}, Search: []Search{
		{
			Field: FieldUTMContent,
			Input: "content",
		},
	}})
	assert.NoError(t, err)
}
