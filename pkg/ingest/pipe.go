package ingest

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg/db"
	"github.com/pirsch-analytics/pirsch/v7/pkg/model"
)

// Pipe ingests requests into the system.
// Requests are processed through steps before they are stored.
type Pipe struct {
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	steps    []PipeFunc
	requests chan *Request
	storage  db.Storage
	logger   *slog.Logger
}

// NewPipe creates a new Pipe for the given PipeOptions.
func NewPipe(options PipeOptions) *Pipe {
	options.validate()
	ctx, cancel := context.WithCancel(context.Background())
	p := &Pipe{
		ctx:      ctx,
		cancel:   cancel,
		steps:    make([]PipeFunc, 0),
		requests: make(chan *Request, options.RequestChannelBufferSize),
		storage:  options.Storage,
		logger:   options.Logger,
	}

	for range options.Worker {
		p.wg.Go(p.collect(options.WorkerBufferSize, options.WorkerTimeout))
	}

	return p
}

// Use adds a processing step to the Pipe.
func (p *Pipe) Use(f PipeFunc) *Pipe {
	p.steps = append(p.steps, f)
	return p
}

// Process processes the given request.
// It can be run in its own Goroutine, so that the client won't have to wait for the request to be processed.
// http.StatusAccepted can be returned in that case for example.
func (p *Pipe) Process(request *Request) error {
	// return if the pipe has been halted
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	default:
	}

	// process the request otherwise
	request.validate()

	for _, step := range p.steps {
		cancel, err := step(request)

		if err != nil {
			return err
		}

		if cancel {
			// mark the request as cancelled, but keep storing it (for bot analysis)
			request.cancelled = true
			break
		}
	}

	// schedule request to be stored in batch
	p.requests <- request
	return nil
}

// Stop flushes all data currently within the pipe and stops processing new data.
func (p *Pipe) Stop() {
	p.cancel()
	p.wg.Wait()
}

func (p *Pipe) collect(bufferSize int, timeout time.Duration) func() {
	return func() {
		// double the session buffer size because we always update the session while cancelling the previous row
		sessions := make([]model.Session, 0, bufferSize*2)
		pageViews := make([]model.PageView, 0, bufferSize)
		events := make([]model.Event, 0, bufferSize)
		requests := make([]model.Request, 0, bufferSize)
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		for {
			timer.Reset(timeout)

			select {
			case request := <-p.requests:
				requests = append(requests, model.Request{
					ClientID:    request.ClientID,
					VisitorID:   request.VisitorID,
					Time:        request.Time,
					IP:          request.IP,
					UserAgent:   request.UserAgent,
					Hostname:    request.Hostname,
					Path:        request.Path,
					Event:       request.EventName,
					Referrer:    request.Referrer,
					UTMSource:   request.UTMSource,
					UTMMedium:   request.UTMMedium,
					UTMCampaign: request.UTMCampaign,
					Bot:         request.IsBot,
					BotReason:   request.BotReason,
				})

				if !request.cancelled {
					if request.cancelSession != nil {
						sessions = append(sessions, *request.cancelSession)
					}

					if request.session != nil {
						sessions = append(sessions, *request.session)
					}

					if request.EventName != "" {
						events = append(events, model.Event{
							Data:     p.dataFromRequest(request),
							Name:     request.EventName,
							MetaData: request.EventMetaData,
							Path:     request.Path,
							Title:    request.Title,
						})
					} else {
						pageViews = append(pageViews, model.PageView{
							Data:  p.dataFromRequest(request),
							Path:  request.Path,
							Title: request.Title,
							Tags:  request.Tags,
						})
					}
				}

				if len(sessions) >= bufferSize*2 ||
					len(pageViews) >= bufferSize ||
					len(events) >= bufferSize ||
					len(requests) >= bufferSize {
					p.flush(sessions, pageViews, events, requests)
					sessions = sessions[:0]
					pageViews = pageViews[:0]
					events = events[:0]
					requests = requests[:0]
				}
			case <-timer.C:
				p.flush(sessions, pageViews, events, requests)
				sessions = sessions[:0]
				pageViews = pageViews[:0]
				events = events[:0]
				requests = requests[:0]
			case <-p.ctx.Done():
				p.flush(sessions, pageViews, events, requests)
				sessions = sessions[:0]
				pageViews = pageViews[:0]
				events = events[:0]
				requests = requests[:0]
				return
			}
		}
	}
}

func (p *Pipe) dataFromRequest(request *Request) model.Data {
	return model.Data{
		ClientID:       request.ClientID,
		VisitorID:      request.VisitorID,
		SessionID:      request.SessionID,
		Time:           request.Time,
		Start:          request.Start,
		Hostname:       request.Hostname,
		PageViews:      request.PageViews,
		IsBounce:       request.IsBounce,
		Language:       request.Language,
		CountryCode:    request.CountryCode,
		Region:         request.Region,
		City:           request.City,
		Referrer:       request.Referrer,
		ReferrerName:   request.ReferrerName,
		ReferrerIcon:   request.ReferrerIcon,
		OS:             request.OS,
		OSVersion:      request.OSVersion,
		Browser:        request.Browser,
		BrowserVersion: request.BrowserVersion,
		Desktop:        request.Desktop,
		Mobile:         request.Mobile,
		ScreenClass:    request.ScreenClass,
		UTMSource:      request.UTMSource,
		UTMMedium:      request.UTMMedium,
		UTMCampaign:    request.UTMCampaign,
		UTMContent:     request.UTMContent,
		UTMTerm:        request.UTMTerm,
		Channel:        request.Channel,
	}
}

func (p *Pipe) flush(sessions []model.Session, pageViews []model.PageView, events []model.Event, requests []model.Request) {
	var wg sync.WaitGroup
	wg.Go(func() {
		if err := p.flushWithRetry(func() error {
			return p.storage.SaveSessions(p.ctx, sessions)
		}, "save sessions"); err != nil {
			p.logger.Error("Failed saving sessions", "err", err)
		}
	})
	wg.Go(func() {
		if err := p.flushWithRetry(func() error {
			return p.storage.SavePageViews(p.ctx, pageViews)
		}, "save page views"); err != nil {
			p.logger.Error("Failed saving page views", "err", err)
		}
	})
	wg.Go(func() {
		if err := p.flushWithRetry(func() error {
			return p.storage.SaveEvents(p.ctx, events)
		}, "save events"); err != nil {
			p.logger.Error("Failed saving events", "err", err)
		}
	})
	wg.Go(func() {
		if err := p.flushWithRetry(func() error {
			return p.storage.SaveRequests(p.ctx, requests)
		}, "save requests"); err != nil {
			p.logger.Error("Failed saving requests", "err", err)
		}
	})
	wg.Wait()
}

func (p *Pipe) flushWithRetry(save func() error, operation string) error {
	const maxRetries = 5
	var err error

	for attempt := range maxRetries {
		if err = save(); err == nil {
			return nil
		}

		remaining := maxRetries - attempt - 1

		if remaining == 0 {
			break
		}

		backoff := time.Duration(attempt+1) * 20 * time.Second
		jitter := time.Duration(rand.N(5)) * time.Second
		wait := backoff + jitter
		p.logger.Error("Storage error, retrying", "err", err,
			"operation", operation,
			"retries_remaining", remaining,
			"wait", wait)

		time.Sleep(wait)
	}

	return fmt.Errorf("%s failed after %d attempts: %w", operation, maxRetries, err)
}
