package ingest

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

	"github.com/pirsch-analytics/pirsch/v7/pkg/db"
	"github.com/pirsch-analytics/pirsch/v7/pkg/model"
)

// Pipe ingests requests into the system.
// Requests are processed through configurable steps before they are stored.
type Pipe struct {
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	steps    []PipeStep
	requests chan *Request
	storage  db.Storage
	logIP    bool
	logger   *slog.Logger
}

// NewPipe creates a new Pipe for the given PipeOptions.
func NewPipe(options PipeOptions) *Pipe {
	options.validate()
	ctx, cancel := context.WithCancel(context.Background())
	p := &Pipe{
		ctx:      ctx,
		cancel:   cancel,
		steps:    make([]PipeStep, 0),
		requests: make(chan *Request, options.RequestChannelBufferSize),
		storage:  options.Storage,
		logIP:    options.LogIP,
		logger:   options.Logger,
	}

	for range options.Worker {
		p.wg.Go(p.collect(options.WorkerBufferSize, options.WorkerTimeout))
	}

	return p
}

// Use adds a processing step to the Pipe.
func (p *Pipe) Use(f ...PipeStep) *Pipe {
	p.steps = append(p.steps, f...)
	return p
}

// Process processes the given request.
// It can be run in its own Goroutine, so that the client won't have to wait for the request to be processed.
// http.StatusAccepted can be returned in that case for example.
// It will return true if the request has been accepted, or false if it doesn't (due to some filter step for example).
func (p *Pipe) Process(request *Request) (bool, error) {
	// return if the pipe has been halted
	select {
	case <-p.ctx.Done():
		return false, p.ctx.Err()
	default:
	}

	// check if the request should be ignored for any reason
	if p.ignore(request) {
		return false, nil
	}

	// process the request otherwise
	request.validate()

	for _, step := range p.steps {
		cancel, err := step.Step(request)

		if err != nil {
			return false, err
		}

		if cancel {
			// mark the request as canceled, but keep storing it (for bot analysis)
			request.cancelled = true
			break
		}
	}

	// schedule request to be stored in batch
	p.requests <- request
	return true, nil
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
			select {
			case request := <-p.requests:
				if !p.logIP {
					request.IP = ""
				}

				requests = append(requests, model.Request{
					SiteID:      request.SiteID,
					VisitorID:   request.VisitorID,
					Time:        request.Time,
					Hostname:    request.Hostname,
					Path:        request.Path,
					Query:       request.Query,
					IP:          request.IP,
					UserAgent:   request.UserAgent,
					Headers:     request.Headers,
					EventName:   request.EventName,
					Referrer:    request.Referrer,
					UTMSource:   request.UTMSource,
					UTMMedium:   request.UTMMedium,
					UTMCampaign: request.UTMCampaign,
					UTMContent:  request.UTMContent,
					UTMTerm:     request.UTMTerm,
					Bot:         request.IsBot,
					BotReason:   request.BotReason,
				})

				if !request.cancelled {
					if request.CancelSession != nil {
						sessions = append(sessions, *request.CancelSession)
					}

					if request.Session != nil {
						sessions = append(sessions, *request.Session)
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
							Data:            p.dataFromRequest(request),
							DurationSeconds: request.DurationSeconds,
							Path:            request.Path,
							Title:           request.Title,
							Tags:            request.Tags,
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
					timer.Reset(timeout)
				}
			case <-timer.C:
				p.flush(sessions, pageViews, events, requests)
				sessions = sessions[:0]
				pageViews = pageViews[:0]
				events = events[:0]
				requests = requests[:0]
				timer.Reset(timeout)
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
		SiteID:         request.SiteID,
		VisitorID:      request.VisitorID,
		SessionID:      request.SessionID,
		Time:           request.Time,
		Hostname:       request.Hostname,
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
		Platform:       request.Platform,
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
	// copy ingestion data
	sessionsCopy := make([]model.Session, len(sessions))
	pageViewsCopy := make([]model.PageView, len(pageViews))
	eventsCopy := make([]model.Event, len(events))
	requestsCopy := make([]model.Request, len(requests))
	copy(sessionsCopy, sessions)
	copy(pageViewsCopy, pageViews)
	copy(eventsCopy, events)
	copy(requestsCopy, requests)

	// retries run asynchronously, so that we won't block the main ingestion pipeline
	var wg sync.WaitGroup
	wg.Go(func() {
		p.flushWithRetry(func() error {
			return p.storage.SaveSessions(p.ctx, sessionsCopy)
		}, "save sessions")
	})
	wg.Go(func() {
		p.flushWithRetry(func() error {
			return p.storage.SavePageViews(p.ctx, pageViewsCopy)
		}, "save page views")
	})
	wg.Go(func() {
		p.flushWithRetry(func() error {
			return p.storage.SaveEvents(p.ctx, eventsCopy)
		}, "save events")
	})
	wg.Go(func() {
		p.flushWithRetry(func() error {
			return p.storage.SaveRequests(p.ctx, requestsCopy)
		}, "save requests")
	})
	wg.Wait()
}

func (p *Pipe) flushWithRetry(save func() error, operation string) {
	if err := save(); err == nil {
		return
	}

	// run retries asynchronously
	go func() {
		const maxRetries = 5
		var err error

		for attempt := range maxRetries {
			select {
			case <-p.ctx.Done():
				p.logger.Error("Failed saving data",
					"err", "context canceled",
					"operation", operation)
				return
			default:
				if err = save(); err == nil {
					return
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
		}

		p.logger.Error("Failed saving data",
			"err", fmt.Sprintf("%s failed after %d attempts: %s", operation, maxRetries, err),
			"operation", operation)
	}()
}

func (p *Pipe) ignore(request *Request) bool {
	// ignore requests with missing http.Request attributes
	if request.Request == nil {
		return true
	}

	// ignore browsers pre-fetching data
	xMoz := strings.ToLower(request.Request.Header.Get("X-Moz"))
	xPurpose := strings.ToLower(request.Request.Header.Get("X-Purpose"))
	purpose := strings.ToLower(request.Request.Header.Get("Purpose"))

	if xMoz == "prefetch" ||
		xPurpose == "prefetch" ||
		xPurpose == "preview" ||
		purpose == "prefetch" ||
		purpose == "preview" {
		return true
	}

	return false
}
