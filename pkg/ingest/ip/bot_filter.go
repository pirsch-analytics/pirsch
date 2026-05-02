package ip

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
)

// BotFilter filters bot requests based on their IP addresses.
// This step operates on the ingest.Request IP, which must be set before this step is run.
type BotFilter struct {
	filter []Filter
}

// NewBotFilter creates a new BotFilter.
func NewBotFilter(filter []Filter) *BotFilter {
	if filter == nil {
		filter = make([]Filter, 0)
	}

	return &BotFilter{
		filter: filter,
	}
}

// Step implements the ingest.PipeStep interface.
func (f *BotFilter) Step(request *ingest.Request) (bool, error) {
	if request.DisableBotFilter {
		return false, nil
	}

	for _, filter := range f.filter {
		if filter.Ignore(request.IP) {
			request.BotReason = "ip"
			return true, nil
		}
	}

	return false, nil
}
