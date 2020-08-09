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
func (processor *Processor) Process() error {
	return processor.ProcessTenant(NullTenant)
}

// ProcessTenant processes all hits in database for given tenant and deletes them afterwards.
// The tenant can be set to nil if you don't split your data (which is usually the case).
// It will panic in case of an error.
func (processor *Processor) ProcessTenant(tenantID sql.NullInt64) error {
	days, err := processor.store.Days(tenantID)

	if err != nil {
		return err
	}

	for _, day := range days {
		waitChan := make(chan struct{})
		errChan := make(chan error, 4)

		go func() {
			var wg sync.WaitGroup
			wg.Add(4)

			go func() {
				if err := processor.visitorCount(tenantID, day); err != nil {
					errChan <- err
				}

				wg.Done()
			}()

			go func() {
				if err := processor.visitorCountHour(tenantID, day); err != nil {
					errChan <- err
				}

				wg.Done()
			}()

			go func() {
				if err := processor.languageCount(tenantID, day); err != nil {
					errChan <- err
				}

				wg.Done()
			}()

			go func() {
				if err := processor.pageViews(tenantID, day); err != nil {
					errChan <- err
				}

				wg.Done()
			}()

			wg.Wait()
			close(waitChan)
		}()

		select {
		case <-waitChan:
			// nothing to do...
		case err := <-errChan:
			return err
		}

		if err := processor.store.DeleteHitsByDay(tenantID, day); err != nil {
			return err
		}
	}

	return nil
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
