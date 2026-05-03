package referrer

import (
	"net/http"
	"testing"

	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
	"github.com/stretchr/testify/assert"
)

func TestBotFilter(t *testing.T) {
	referrer := []string{
		`(select(0)from(select(sleep(15)))v)/*'+(select(0)from(select(sleep(15)))v)+'"+(select(0)from(select(sleep(15)))v)+"*/`,
		`-1 OR 2+1-1+1=1 AND 136=136 --`,
		`-1 OR 2+136-136-1=0+0+0+1 --`,
		`-1 OR 3*2=6 AND 707=707`,
		`-1 OR 3*2=6 AND 946=946 --`,
		`-1 OR 3*2>(0+5+946-946) --`,
		`-1 OR 3+87-87-1=0+0+0+1`,
		`-1 OR 3+946-946-1=0+0+0+1 --`,
		`-1" OR 2+1-1-1=1 AND 846=846 --`,
		`-1" OR 2+874-874-1=0+0+0+1 --`,
		`-1" OR 3*2<(0+5+846-846) --`,
		`-1" OR 3*2=6 AND 778=778 --`,
		`-1" OR 3*2>(0+5+846-846) --`,
		`-1" OR 3+211-211-1=0+0+0+1 --`,
		`-1' OR 2+1-1-1=1 AND 94=94 --`,
		`-1' OR 3*2>(0+5+912-912) or 'S37EBSa9'='`,
		`-1' OR 3*2>(0+5+94-94) --`,
		`-1' OR 3+142-142-1=0+0+0+1 --`,
		`-1' OR 3+842-842-1=0+0+0+1 or 'XSkGSBJC'='`,
		`-1)) OR 105=(SELECT 105 FROM PG_SLEEP(15))--`,
		`0"XOR(if(now()=sysdate(),sleep(15),0))XOR"Z`,
		`0sjy32e7') OR 259=(SELECT 259 FROM PG_SLEEP(15))--`,
	}
	acknowledged := make([]string, 0)
	filter := NewBotFilter()

	for _, r := range referrer {
		req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
		req.Header.Set("Referer", r)
		cancel, err := filter.Step(&ingest.Request{
			Request: req,
		})
		assert.NoError(t, err)

		if !cancel {
			acknowledged = append(acknowledged, r)
		}
	}

	assert.Empty(t, acknowledged)
}

func TestBotFilterStripSubdomain(t *testing.T) {
	input := []string{
		"",
		".",
		"..",
		"...",
		" ",
		"\t",
		"boring.old",
		"with.subdomain.com",
		"with.multiple.subdomains.com",
	}
	expected := []string{
		"",
		".",
		"..",
		".",
		" ",
		"\t",
		"boring.old",
		"subdomain.com",
		"subdomains.com",
	}
	filter := NewBotFilter()

	for i, in := range input {
		assert.Equal(t, expected[i], filter.stripSubdomain(in))
	}
}

func TestBotFilterAcknowledge(t *testing.T) {
	referrer := []string{
		"https://www.adsensecustomsearchads.com/",
	}
	ignored := make([]string, 0)
	filter := NewBotFilter()

	for _, r := range referrer {
		req, _ := http.NewRequest(http.MethodGet, "https://example.com", nil)
		req.Header.Set("Referer", r)
		cancel, err := filter.Step(&ingest.Request{
			Request: req,
		})
		assert.NoError(t, err)

		if cancel {
			ignored = append(ignored, r)
		}
	}

	assert.Empty(t, ignored)
}
