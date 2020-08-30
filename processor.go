package pirsch

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"time"
)

// ProcessorConfig is the optional configuration for the Processor.
type ProcessorConfig struct {
	// ProcessVisitors enables/disabled processing the visitor count.
	// The default is true (enabled).
	ProcessVisitors bool

	// ProcessVisitorPerHour enables/disabled processing the visitor count per hour.
	// The default is true (enabled).
	ProcessVisitorPerHour bool

	// ProcessLanguages enables/disabled processing the language count.
	// The default is true (enabled).
	ProcessLanguages bool

	// ProcessPageViews enables/disabled processing the page views.
	// The default is true (enabled).
	ProcessPageViews bool

	// ProcessVisitorPerReferrer enables/disabled processing the visitor count per referrer.
	// The default is true (enabled).
	ProcessVisitorPerReferrer bool

	// ProcessVisitorPerOS enables/disabled processing the visitor count per operating system.
	// The default is true (enabled).
	ProcessVisitorPerOS bool

	// ProcessVisitorPerBrowser enables/disabled processing the visitor count per browser.
	// The default is true (enabled).
	ProcessVisitorPerBrowser bool

	// ProcessPlatform enables/disabled processing the visitor platform.
	// The default is true (enabled).
	ProcessPlatform bool
}

type processorFunc struct {
	f    func(*sqlx.Tx, sql.NullInt64, time.Time) error
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
			ProcessVisitors:           true,
			ProcessVisitorPerHour:     true,
			ProcessLanguages:          true,
			ProcessPageViews:          true,
			ProcessVisitorPerReferrer: true,
			ProcessVisitorPerOS:       true,
			ProcessVisitorPerBrowser:  true,
			ProcessPlatform:           true,
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
		{processor.countVisitors, processor.config.ProcessVisitors},
		{processor.countVisitorPerHour, processor.config.ProcessVisitorPerHour},
		{processor.countLanguages, processor.config.ProcessLanguages},
		{processor.countPageViews, processor.config.ProcessPageViews},
		{processor.countVisitorPerReferrer, processor.config.ProcessVisitorPerReferrer},
		{processor.countVisitorPerOS, processor.config.ProcessVisitorPerOS},
		{processor.countVisitorPerBrowser, processor.config.ProcessVisitorPerBrowser},
		{processor.countVisitorPlatform, processor.config.ProcessPlatform},
	}

	for _, day := range days {
		if err := processor.processDay(tenantID, day, processors); err != nil {
			return err
		}
	}

	return nil
}

func (processor *Processor) processDay(tenantID sql.NullInt64, day time.Time, processors []processorFunc) error {
	tx := processor.store.NewTx()

	for _, proc := range processors {
		if proc.exec {
			if err := proc.f(tx, tenantID, day); err != nil {
				processor.store.Rollback(tx)
				return err
			}
		}
	}

	if err := processor.store.DeleteHitsByDay(tx, tenantID, day); err != nil {
		processor.store.Rollback(tx)
		return err
	}

	processor.store.Commit(tx)
	return nil
}

func (processor *Processor) countVisitors(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) error {
	visitors, err := processor.store.CountVisitorsPerDay(tx, tenantID, day)

	if err != nil {
		return err
	}

	if visitors == 0 {
		return nil
	}

	return processor.store.SaveVisitorsPerDay(tx, &VisitorsPerDay{
		BaseEntity: BaseEntity{TenantID: tenantID},
		Day:        day,
		Visitors:   visitors,
	})
}

func (processor *Processor) countVisitorPerHour(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) error {
	visitors, err := processor.store.CountVisitorsPerDayAndHour(tx, tenantID, day)

	if err != nil {
		return err
	}

	for _, visitors := range visitors {
		if visitors.Visitors > 0 {
			if err := processor.store.SaveVisitorsPerHour(tx, &visitors); err != nil {
				return err
			}
		}
	}

	return nil
}

func (processor *Processor) countLanguages(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) error {
	visitors, err := processor.store.CountVisitorsPerLanguage(tx, tenantID, day)

	if err != nil {
		return err
	}

	for _, visitors := range visitors {
		if visitors.Visitors > 0 {
			if err := processor.store.SaveVisitorsPerLanguage(tx, &visitors); err != nil {
				return err
			}
		}
	}

	return nil
}

func (processor *Processor) countPageViews(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) error {
	visitors, err := processor.store.CountVisitorsPerPage(tx, tenantID, day)

	if err != nil {
		return err
	}

	for _, visitors := range visitors {
		if visitors.Visitors > 0 {
			if err := processor.store.SaveVisitorsPerPage(tx, &visitors); err != nil {
				return err
			}
		}
	}

	return nil
}

func (processor *Processor) countVisitorPerReferrer(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) error {
	visitors, err := processor.store.CountVisitorsPerReferrer(tx, tenantID, day)

	if err != nil {
		return err
	}

	for _, visitors := range visitors {
		if visitors.Visitors > 0 {
			if err := processor.store.SaveVisitorsPerReferrer(tx, &visitors); err != nil {
				return err
			}
		}
	}

	return nil
}

func (processor *Processor) countVisitorPerOS(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) error {
	visitors, err := processor.store.CountVisitorsPerOSAndVersion(tx, tenantID, day)

	if err != nil {
		return err
	}

	for _, visitors := range visitors {
		if visitors.Visitors > 0 {
			if err := processor.store.SaveVisitorsPerOS(tx, &visitors); err != nil {
				return err
			}
		}
	}

	return nil
}

func (processor *Processor) countVisitorPerBrowser(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) error {
	visitors, err := processor.store.CountVisitorsPerBrowserAndVersion(tx, tenantID, day)

	if err != nil {
		return err
	}

	for _, visitors := range visitors {
		if visitors.Visitors > 0 {
			if err := processor.store.SaveVisitorsPerBrowser(tx, &visitors); err != nil {
				return err
			}
		}
	}

	return nil
}

func (processor *Processor) countVisitorPlatform(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) error {
	platform, err := processor.store.CountVisitorPlatforms(tx, tenantID, day)

	if err != nil {
		return err
	}

	return processor.store.SaveVisitorPlatform(tx, platform)
}
