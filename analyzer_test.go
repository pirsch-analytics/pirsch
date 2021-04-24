package pirsch

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAnalyzer_ActiveVisitors(t *testing.T) {
	cleanupDB()
	assert.NoError(t, dbClient.SaveHits([]Hit{
		{Fingerprint: "fp1", Time: time.Now().Add(-time.Minute * 30), Path: "/"},
		{Fingerprint: "fp1", Time: time.Now().Add(-time.Minute * 15), Path: "/"},
		{Fingerprint: "fp1", Time: time.Now().Add(-time.Minute * 5), Path: "/bar"},
		{Fingerprint: "fp2", Time: time.Now().Add(-time.Minute * 4), Path: "/bar"},
		{Fingerprint: "fp2", Time: time.Now().Add(-time.Minute * 3), Path: "/foo"},
		{Fingerprint: "fp3", Time: time.Now().Add(-time.Minute * 3), Path: "/"},
		{Fingerprint: "fp4", Time: time.Now().Add(-time.Minute), Path: "/"},
	}))
	analyzer := NewAnalyzer(dbClient)
	visitors, count, err := analyzer.ActiveVisitors(nil, time.Minute*10)
	assert.NoError(t, err)
	assert.Equal(t, 4, count)
	assert.Len(t, visitors, 3)
	assert.Equal(t, "/", visitors[0].Path.String)
	assert.Equal(t, "/bar", visitors[1].Path.String)
	assert.Equal(t, "/foo", visitors[2].Path.String)
	assert.Equal(t, 2, visitors[0].Visitors)
	assert.Equal(t, 2, visitors[1].Visitors)
	assert.Equal(t, 1, visitors[2].Visitors)
	visitors, count, err = analyzer.ActiveVisitors(&Filter{Path: "/bar"}, time.Minute*10)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Equal(t, "/bar", visitors[0].Path.String)
	assert.Equal(t, 2, visitors[0].Visitors)
}
