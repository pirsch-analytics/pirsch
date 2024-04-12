package analyzer

import (
	"fmt"
	"github.com/pirsch-analytics/pirsch/v6/pkg/db"
	"github.com/pirsch-analytics/pirsch/v6/pkg/model"
	"math"
	"sort"
	"time"
)

// Sessions aggregates statistics regarding single sessions.
type Sessions struct {
	analyzer *Analyzer
	store    db.Store
}

// List returns a list of sessions for given filter.
func (sessions *Sessions) List(filter *Filter) ([]model.Session, error) {
	filter = sessions.analyzer.getFilter(filter)
	filter.Sample = 0
	q, args := filter.buildQuery([]Field{
		FieldSessionsAll,
	}, []Field{
		FieldVisitorID,
		FieldSessionID,
		FieldExitPath,
		FieldSessionExitTitle,
	}, []Field{
		FieldMaxTime,
	})
	query := fmt.Sprintf(`SELECT * EXCEPT (exit_paths, exit_titles)
		FROM (
		    SELECT t.visitor_id,
				t.session_id,
				max(t.session_time),
				min(t.session_start),
				max(t.session_duration_seconds),
				any(t.session_entry_path),
		        groupArray(t.session_exit_path) exit_paths,
				exit_paths[length(exit_paths)] exit_path,
				max(t.session_page_views),
				min(t.session_is_bounce),
				any(t.session_entry_title),
		        groupArray(t.session_exit_title) exit_titles,
				exit_titles[length(exit_titles)] exit_title,
				any(t.session_language),
				any(t.session_country_code),
				any(t.session_city),
				any(t.session_referrer),
				any(t.session_referrer_name),
				any(t.session_referrer_icon),
				any(t.session_os),
				any(t.session_os_version),
				any(t.session_browser),
				any(t.session_browser_version),
				any(t.session_desktop),
				any(t.session_mobile),
				any(t.session_screen_class),
				any(t.session_utm_source),
				any(t.session_utm_medium),
				any(t.session_utm_campaign),
				any(t.session_utm_content),
				any(t.session_utm_term),
		        max(t.session_extended)
		    FROM (%s) t
			GROUP BY visitor_id, session_id
			ORDER BY max(session_time)
		)`, q)
	stats, err := sessions.store.SelectSessions(filter.Ctx, query, args...)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// Breakdown returns the page views and events for a single session in chronological order.
func (sessions *Sessions) Breakdown(filter *Filter) ([]model.SessionStep, error) {
	filter = sessions.analyzer.getFilter(filter)

	if filter.VisitorID == 0 || filter.SessionID == 0 {
		return nil, nil
	}

	f := &Filter{
		Ctx:         filter.Ctx,
		ClientID:    filter.ClientID,
		Timezone:    filter.Timezone,
		From:        filter.From,
		To:          filter.To,
		VisitorID:   filter.VisitorID,
		SessionID:   filter.SessionID,
		IncludeTime: filter.IncludeTime,
	}
	q, args := f.buildQuery([]Field{FieldPageViewsAll}, nil, []Field{FieldTime})
	pageViews, err := sessions.store.SelectPageViews(f.Ctx, q, args...)

	if err != nil {
		return nil, err
	}

	q, args = f.buildQuery([]Field{FieldEventsAll}, nil, nil)
	events, err := sessions.store.SelectEvents(f.Ctx, q, args...)

	if err != nil {
		return nil, err
	}

	stats := make([]model.SessionStep, 0, len(pageViews)+len(events))

	for i := range pageViews {
		if i < len(pageViews)-1 {
			pageViews[i].DurationSeconds = uint32(math.Round(pageViews[i+1].Time.Sub(pageViews[i].Time).Seconds()))
		} else if i == len(pageViews)-1 {
			pageViews[i].DurationSeconds = 0
		}

		stats = append(stats, model.SessionStep{
			PageView: &pageViews[i],
		})
	}

	for i := range events {
		stats = append(stats, model.SessionStep{
			Event: &events[i],
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		var a, b time.Time

		if stats[i].PageView != nil {
			a = stats[i].PageView.Time
		} else {
			a = stats[i].Event.Time
		}

		if stats[j].PageView != nil {
			b = stats[j].PageView.Time
		} else {
			b = stats[j].Event.Time
		}

		return a.Before(b)
	})
	return stats, nil
}
