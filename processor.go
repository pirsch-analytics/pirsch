package pirsch

import (
	"sync"
	"time"
)

// TODO docs
type Processor struct {
	store Store
}

func NewProcessor(store Store) *Processor {
	return &Processor{store}
}

func (processor *Processor) Analyze() {
	days, err := processor.store.Days()
	panicOnErr(err)

	for _, day := range days {
		var wg sync.WaitGroup
		wg.Add(5)
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
		go func() {
			panicOnErr(processor.visitorPageFlow(day))
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
		if err := processor.store.SaveVisitorsPerHour(&visitors); err != nil {
			return err
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
		if err := processor.store.SaveVisitorsPerLanguage(&visitors); err != nil {
			return err
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
		if err := processor.store.SaveVisitorsPerPage(&visitors); err != nil {
			return err
		}
	}

	return nil
}

func (processor *Processor) visitorPageFlow(day time.Time) error {
	// TODO
	return nil
}
