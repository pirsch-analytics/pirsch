package pirsch

import (
	"database/sql"
	"sync"
	"time"
)

// Processor processes hits to reduce them into meaningful statistics.
type Processor struct {
	store Store
}

// NewProcessor creates a new Processor for given Store.
func NewProcessor(store Store) *Processor {
	return &Processor{store}
}

// Process processes all hits in database and deletes them afterwards.
// It will panic in case of an error.
func (processor *Processor) Process() {
	processor.ProcessTenant(NullTenant)
}

// ProcessTenant processes all hits in database for given tenant and deletes them afterwards.
// The tenant can be set to nil if you don't split your data (which is usually the case).
// It will panic in case of an error.
func (processor *Processor) ProcessTenant(tenantID sql.NullInt64) {
	days, err := processor.store.Days(tenantID)
	panicOnErr(err)

	for _, day := range days {
		var wg sync.WaitGroup
		wg.Add(4)
		go func() {
			panicOnErr(processor.visitorCount(tenantID, day))
			wg.Done()
		}()
		go func() {
			panicOnErr(processor.visitorCountHour(tenantID, day))
			wg.Done()
		}()
		go func() {
			panicOnErr(processor.languageCount(tenantID, day))
			wg.Done()
		}()
		go func() {
			panicOnErr(processor.pageViews(tenantID, day))
			wg.Done()
		}()
		wg.Wait()
		panicOnErr(processor.store.DeleteHitsByDay(tenantID, day))
	}
}

func (processor *Processor) visitorCount(tenantID sql.NullInt64, day time.Time) error {
	visitors, err := processor.store.VisitorsPerDay(tenantID, day)

	if err != nil {
		return err
	}

	if visitors == 0 {
		return nil
	}

	return processor.store.SaveVisitorsPerDay(&VisitorsPerDay{
		TenantID: tenantID,
		Day:      day,
		Visitors: visitors,
	})
}

func (processor *Processor) visitorCountHour(tenantID sql.NullInt64, day time.Time) error {
	visitors, err := processor.store.VisitorsPerDayAndHour(tenantID, day)

	if err != nil {
		return err
	}

	for _, visitors := range visitors {
		if visitors.Visitors > 0 {
			if err := processor.store.SaveVisitorsPerHour(&visitors); err != nil {
				return err
			}
		}
	}

	return nil
}

func (processor *Processor) languageCount(tenantID sql.NullInt64, day time.Time) error {
	visitors, err := processor.store.VisitorsPerLanguage(tenantID, day)

	if err != nil {
		return err
	}

	for _, visitors := range visitors {
		if visitors.Visitors > 0 {
			if err := processor.store.SaveVisitorsPerLanguage(&visitors); err != nil {
				return err
			}
		}
	}

	return nil
}

func (processor *Processor) pageViews(tenantID sql.NullInt64, day time.Time) error {
	visitors, err := processor.store.VisitorsPerPage(tenantID, day)

	if err != nil {
		return err
	}

	for _, visitors := range visitors {
		if visitors.Visitors > 0 {
			if err := processor.store.SaveVisitorsPerPage(&visitors); err != nil {
				return err
			}
		}
	}

	return nil
}
