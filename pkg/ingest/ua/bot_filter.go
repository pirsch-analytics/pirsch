package ua

import (
	"github.com/pirsch-analytics/pirsch/v7/pkg/ingest"
)

// BotFilter filters bot requests based on their User-Agent information.
// This step operates on the ingest.Request User-Agent fields, which must be set before this step is run.
type BotFilter struct{}

// NewBotFilter creates a new BotFilter.
func NewBotFilter() *BotFilter {
	return new(BotFilter)
}

// Step implements the ingest.PipeStep interface.
func (f *BotFilter) Step(request *ingest.Request) (bool, error) {
	if request.DisableBotFilter {
		return false, nil
	}

	// TODO

	return false, nil
}
