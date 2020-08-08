package pirsch

import (
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
	days, err := processor.store.Days()
	panicOnErr(err)

	for _, day := range days {
		var wg sync.WaitGroup
		wg.Add(4)
		go func() {
			panicOnErr(processor.visitorCount(day))
			wg.Done()
		}()
		go func() {
			panicOnErr(processor.visitorCountHour(day))
			wg.Done()
		}()
		go func() {
			panicOnErr(processor.languageCount(day))
			wg.Done()
		}()
		go func() {
			panicOnErr(processor.pageViews(day))
			wg.Done()
		}()
		wg.Wait()
		panicOnErr(processor.store.DeleteHitsByDay(day))
	}
}

func (processor *Processor) visitorCount(day time.Time) error {
	visitors, err := processor.store.VisitorsPerDay(day)

	if err != nil {
		return err
	}

	if visitors == 0 {
		return nil
	}

	return processor.store.SaveVisitorsPerDay(&VisitorsPerDay{
		Day:      day,
		Visitors: visitors,
	})
}

func (processor *Processor) visitorCountHour(day time.Time) error {
	visitors, err := processor.store.VisitorsPerDayAndHour(day)

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

func (processor *Processor) languageCount(day time.Time) error {
	visitors, err := processor.store.VisitorsPerLanguage(day)

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

func (processor *Processor) pageViews(day time.Time) error {
	visitors, err := processor.store.VisitorsPerPage(day)

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
