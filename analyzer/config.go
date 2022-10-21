package analyzer

const (
	defaultIsBotThreshold = 5 // also defined for the tracker_.Tracker
)

// Config is the optional configuration for the Analyzer.
type Config struct {
	// IsBotThreshold see HitOptions.IsBotThreshold.
	IsBotThreshold uint8

	// DisableBotFilter disables IsBotThreshold (otherwise these would be set to the default value).
	DisableBotFilter bool
}

func (config *Config) validate() {
	if config.DisableBotFilter {
		config.IsBotThreshold = 0
	} else if config.IsBotThreshold == 0 {
		config.IsBotThreshold = defaultIsBotThreshold
	}
}
