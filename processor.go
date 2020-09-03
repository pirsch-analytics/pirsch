package pirsch

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"time"
)

// ProcessorConfig is the optional configuration for the Processor.
type ProcessorConfig struct {
	// ProcessVisitors enables/disabled processing the visitor count and platforms per path.
	// The default is true (enabled).
	ProcessVisitors bool

	// ProcessVisitorTime enables/disabled processing the visitor count per hour.
	// The default is true (enabled).
	ProcessVisitorTime bool

	// ProcessLanguages enables/disabled processing the visitor count per language.
	// The default is true (enabled).
	ProcessLanguages bool

	// ProcessReferrer enables/disabled processing the visitor count per referrer.
	// The default is true (enabled).
	ProcessReferrer bool

	// ProcessOS enables/disabled processing the visitor count per operating system.
	// The default is true (enabled).
	ProcessOS bool

	// ProcessBrowser enables/disabled processing the visitor count per browser.
	// The default is true (enabled).
	ProcessBrowser bool
}

type processorFunc struct {
	f    func(*sqlx.Tx, sql.NullInt64, time.Time, string) error
	exec bool
}

// Processor processes hits to reduce them into meaningful statistics.
type Processor struct {
	store  Store
	config ProcessorConfig
}

// NewProcessor creates a new Processor for given Store and config.
// Pass nil for the config to use the defaults.
func NewProcessor(store Store, config *ProcessorConfig) *Processor {
	if config == nil {
		config = &ProcessorConfig{
			ProcessVisitors:    true,
			ProcessVisitorTime: true,
			ProcessLanguages:   true,
			ProcessReferrer:    true,
			ProcessOS:          true,
			ProcessBrowser:     true,
		}
	}

	return &Processor{
		store:  store,
		config: *config,
	}
}

// Process processes all hits in database and deletes them afterwards.
func (processor *Processor) Process() error {
	return processor.ProcessTenant(NullTenant)
}

// ProcessTenant processes all hits in database for given tenant and deletes them afterwards.
// The tenant can be set to nil if you don't split your data (which is usually the case).
func (processor *Processor) ProcessTenant(tenantID sql.NullInt64) error {
	// this explicitly excludes "today", because we might not have collected all visitors
	// and the hits will be deleted after the processor has finished reducing the data
	days, err := processor.store.Days(tenantID)

	if err != nil {
		return err
	}

	// a list of all processing functions and if they're enabled/disabled
	processors := []processorFunc{
		{processor.visitors, processor.config.ProcessVisitors},
		{processor.visitorHours, processor.config.ProcessVisitorTime},
		{processor.languages, processor.config.ProcessLanguages},
		{processor.referrer, processor.config.ProcessReferrer},
		{processor.os, processor.config.ProcessOS},
		{processor.browser, processor.config.ProcessBrowser},
	}

	for _, day := range days {
		if err := processor.processDay(tenantID, day, processors); err != nil {
			return err
		}
	}

	return nil
}

func (processor *Processor) processDay(tenantID sql.NullInt64, day time.Time, processors []processorFunc) error {
	paths, err := processor.store.Paths(tenantID, day)

	if err != nil {
		return err
	}

	tx := processor.store.NewTx()

	for _, path := range paths {
		if err := processor.processPath(tx, tenantID, day, path, processors); err != nil {
			processor.store.Rollback(tx)
			return err
		}
	}

	if err := processor.store.DeleteHitsByDay(tx, tenantID, day); err != nil {
		processor.store.Rollback(tx)
		return err
	}

	processor.store.Commit(tx)
	return nil
}

func (processor *Processor) processPath(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time, path string, processors []processorFunc) error {
	for _, proc := range processors {
		if proc.exec {
			if err := proc.f(tx, tenantID, day, path); err != nil {
				return err
			}
		}
	}

	return nil
}

func (processor *Processor) visitors(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time, path string) error {
	visitors, err := processor.store.CountVisitorsByPath(tx, tenantID, day, path)

	if err != nil {
		return err
	}

	for _, v := range visitors {
		if err := processor.store.SaveVisitorStats(tx, &v); err != nil {
			return err
		}
	}

	return nil
}

func (processor *Processor) visitorHours(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time, path string) error {
	visitors, err := processor.store.CountVisitorsByPathAndHour(tx, tenantID, day, path)

	if err != nil {
		return err
	}

	for _, v := range visitors {
		if err := processor.store.SaveVisitorTimeStats(tx, &v); err != nil {
			return err
		}
	}

	return nil
}

func (processor *Processor) languages(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time, path string) error {
	visitors, err := processor.store.CountVisitorsByPathAndLanguage(tx, tenantID, day, path)

	if err != nil {
		return err
	}

	for _, v := range visitors {
		if err := processor.store.SaveLanguageStats(tx, &v); err != nil {
			return err
		}
	}

	return nil
}

func (processor *Processor) referrer(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time, path string) error {
	visitors, err := processor.store.CountVisitorsByPathAndReferrer(tx, tenantID, day, path)

	if err != nil {
		return err
	}

	for _, v := range visitors {
		if err := processor.store.SaveReferrerStats(tx, &v); err != nil {
			return err
		}
	}

	return nil
}

func (processor *Processor) os(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time, path string) error {
	visitors, err := processor.store.CountVisitorsByPathAndOS(tx, tenantID, day, path)

	if err != nil {
		return err
	}

	for _, v := range visitors {
		if err := processor.store.SaveOSStats(tx, &v); err != nil {
			return err
		}
	}

	return nil
}

func (processor *Processor) browser(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time, path string) error {
	visitors, err := processor.store.CountVisitorsByPathAndBrowser(tx, tenantID, day, path)

	if err != nil {
		return err
	}

	for _, v := range visitors {
		if err := processor.store.SaveBrowserStats(tx, &v); err != nil {
			return err
		}
	}

	return nil
}
