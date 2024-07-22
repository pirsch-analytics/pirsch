package model

import "time"

// ImportedBrowser stores imported statistics.
type ImportedBrowser struct {
	ClientID uint64 `db:"client_id"`
	Date     time.Time
	Browser  string
	Visitors int
}

// ImportedCampaign stores imported statistics.
type ImportedCampaign struct {
	ClientID uint64 `db:"client_id"`
	Date     time.Time
	Campaign string
	Visitors int
}

// ImportedCity stores imported statistics.
type ImportedCity struct {
	ClientID uint64 `db:"client_id"`
	Date     time.Time
	City     string
	Visitors int
}

// ImportedCountry stores imported statistics.
type ImportedCountry struct {
	ClientID    uint64 `db:"client_id"`
	Date        time.Time
	CountryCode string `db:"country_code"`
	Visitors    int
}

// ImportedDevice stores imported statistics.
type ImportedDevice struct {
	ClientID uint64 `db:"client_id"`
	Date     time.Time
	Category string
	Visitors int
}

// ImportedEntryPage stores imported statistics.
type ImportedEntryPage struct {
	ClientID  uint64 `db:"client_id"`
	Date      time.Time
	EntryPath string `db:"entry_path"`
	Visitors  int
	Sessions  int
}

// ImportedExitPage stores imported statistics.
type ImportedExitPage struct {
	ClientID uint64 `db:"client_id"`
	Date     time.Time
	ExitPath string `db:"exit_path"`
	Visitors int
	Sessions int
}

// ImportedLanguage stores imported statistics.
type ImportedLanguage struct {
	ClientID uint64 `db:"client_id"`
	Date     time.Time
	Language string
	Visitors int
}

// ImportedMedium stores imported statistics.
type ImportedMedium struct {
	ClientID uint64 `db:"client_id"`
	Date     time.Time
	Medium   string
	Visitors int
}

// ImportedOS stores imported statistics.
type ImportedOS struct {
	ClientID uint64 `db:"client_id"`
	Date     time.Time
	OS       string
	Visitors int
}

// ImportedPage stores imported statistics.
type ImportedPage struct {
	ClientID  uint64 `db:"client_id"`
	Date      time.Time
	Path      string
	Visitors  int
	PageViews int `db:"page_views"`
	Sessions  int
	Bounces   int
}

// ImportedReferrer stores imported statistics.
type ImportedReferrer struct {
	ClientID uint64 `db:"client_id"`
	Date     time.Time
	Referrer string
	Visitors int
	Sessions int
	Bounces  int
}

// ImportedRegion stores imported statistics.
type ImportedRegion struct {
	ClientID uint64 `db:"client_id"`
	Date     time.Time
	Region   string
	Visitors int
}

// ImportedSource stores imported statistics.
type ImportedSource struct {
	ClientID uint64 `db:"client_id"`
	Date     time.Time
	Source   string
	Visitors int
}

// ImportedVisitors stores imported statistics.
type ImportedVisitors struct {
	ClientID        uint64 `db:"client_id"`
	Date            time.Time
	Visitors        int
	PageViews       int `db:"page_views"`
	Sessions        int
	Bounces         int
	SessionDuration int `db:"session_duration"`
}
