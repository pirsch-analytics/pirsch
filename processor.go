package pirsch

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"time"
)

// Processor processes hits to reduce them into meaningful statistics.
type Processor struct {
	store Store
}

// NewProcessor creates a new Processor for given Store.
func NewProcessor(store Store) *Processor {
	return &Processor{
		store: store,
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
	days, err := processor.store.HitDays(QueryParams{TenantID: tenantID})

	if err != nil {
		return err
	}

	for _, day := range days {
		if err := processor.processDay(tenantID, day); err != nil {
			return err
		}
	}

	return nil
}

func (processor *Processor) processDay(tenantID sql.NullInt64, day time.Time) error {
	paths, err := processor.store.HitPaths(QueryParams{TenantID: tenantID}, day)

	if err != nil {
		return err
	}

	tx := processor.store.NewTx()

	for _, path := range paths {
		if err := processor.processPath(tx, tenantID, day, path); err != nil {
			processor.store.Rollback(tx)
			return err
		}
	}

	if err := processor.screen(tx, tenantID, day); err != nil {
		processor.store.Rollback(tx)
		return err
	}

	if err := processor.country(tx, tenantID, day); err != nil {
		processor.store.Rollback(tx)
		return err
	}

	if err := processor.store.DeleteHitsByDay(tx, QueryParams{TenantID: tenantID}, day); err != nil {
		processor.store.Rollback(tx)
		return err
	}

	processor.store.Commit(tx)
	return nil
}

func (processor *Processor) processPath(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time, path string) error {
	if err := processor.visitors(tx, tenantID, day, path); err != nil {
		return err
	}

	if err := processor.visitorHours(tx, tenantID, day, path); err != nil {
		return err
	}

	if err := processor.languages(tx, tenantID, day, path); err != nil {
		return err
	}

	if err := processor.referrer(tx, tenantID, day, path); err != nil {
		return err
	}

	if err := processor.os(tx, tenantID, day, path); err != nil {
		return err
	}

	if err := processor.browser(tx, tenantID, day, path); err != nil {
		return err
	}

	return nil
}

func (processor *Processor) visitors(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time, path string) error {
	visitors, err := processor.store.CountVisitorsByPath(tx, QueryParams{TenantID: tenantID}, day, path, true)

	if err != nil {
		return err
	}

	bounces := processor.store.CountVisitorsByPathAndMaxOneHit(tx, QueryParams{TenantID: tenantID}, day, path)

	for _, v := range visitors {
		v.Bounces = bounces

		if err := processor.store.SaveVisitorStats(tx, &v); err != nil {
			return err
		}
	}

	return nil
}

func (processor *Processor) visitorHours(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time, path string) error {
	visitors, err := processor.store.CountVisitorsByPathAndHour(tx, QueryParams{TenantID: tenantID}, day, path)

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
	visitors, err := processor.store.CountVisitorsByPathAndLanguage(tx, QueryParams{TenantID: tenantID}, day, path)

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
	visitors, err := processor.store.CountVisitorsByPathAndReferrer(tx, QueryParams{TenantID: tenantID}, day, path)

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
	visitors, err := processor.store.CountVisitorsByPathAndOS(tx, QueryParams{TenantID: tenantID}, day, path)

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
	visitors, err := processor.store.CountVisitorsByPathAndBrowser(tx, QueryParams{TenantID: tenantID}, day, path)

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

func (processor *Processor) screen(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) error {
	visitors, err := processor.store.CountVisitorsByScreenSize(tx, QueryParams{TenantID: tenantID}, day)

	if err != nil {
		return err
	}

	for _, v := range visitors {
		if err := processor.store.SaveScreenStats(tx, &v); err != nil {
			return err
		}
	}

	return nil
}

func (processor *Processor) country(tx *sqlx.Tx, tenantID sql.NullInt64, day time.Time) error {
	visitors, err := processor.store.CountVisitorsByCountryCode(tx, QueryParams{TenantID: tenantID}, day)

	if err != nil {
		return err
	}

	for _, v := range visitors {
		if err := processor.store.SaveCountryStats(tx, &v); err != nil {
			return err
		}
	}

	return nil
}
