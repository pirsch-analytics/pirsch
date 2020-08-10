package pirsch

import (
	"database/sql"
	"sync"
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
			ProcessVisitors:       true,
			ProcessVisitorPerHour: true,
			ProcessLanguages:      true,
			ProcessPageViews:      true,
		}
	}

	return &Processor{
		store:  store,
		config: *config,
	}
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

			if processor.config.ProcessVisitors {
				wg.Add(1)
				go func() {
					if err := processor.countVisitors(tenantID, day); err != nil {
						errChan <- err
					}

					wg.Done()
				}()
			}

			if processor.config.ProcessVisitorPerHour {
				wg.Add(1)
				go func() {
					if err := processor.countVisitorPerHour(tenantID, day); err != nil {
						errChan <- err
					}

					wg.Done()
				}()
			}

			if processor.config.ProcessLanguages {
				wg.Add(1)
				go func() {
					if err := processor.countLanguages(tenantID, day); err != nil {
						errChan <- err
					}

					wg.Done()
				}()
			}

			if processor.config.ProcessPageViews {
				wg.Add(1)
				go func() {
					if err := processor.countPageViews(tenantID, day); err != nil {
						errChan <- err
					}

					wg.Done()
				}()
			}

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

func (processor *Processor) countVisitors(tenantID sql.NullInt64, day time.Time) error {
	visitors, err := processor.store.CountVisitorsPerDay(tenantID, day)

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

func (processor *Processor) countVisitorPerHour(tenantID sql.NullInt64, day time.Time) error {
	visitors, err := processor.store.CountVisitorsPerDayAndHour(tenantID, day)

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

func (processor *Processor) countLanguages(tenantID sql.NullInt64, day time.Time) error {
	visitors, err := processor.store.CountVisitorsPerLanguage(tenantID, day)

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

func (processor *Processor) countPageViews(tenantID sql.NullInt64, day time.Time) error {
	visitors, err := processor.store.CountVisitorsPerPage(tenantID, day)

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
