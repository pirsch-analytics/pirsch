# Changelog

## 6.26.0

* updated referrer groups
* updated dependencies

## 6.25.0

* added claude.ai to the AI channel
* updated referrer blacklist
* updated dependencies

## 6.24.0

* added networks to IP filter list
* updated dependencies

## 6.23.4

* fixed parsing IP when bot filters are disabled

## 6.23.3

* fixed parsing User-Agent when bot filters are disabled
* updated dependencies

## 6.23.2

* updated referrer blacklist
* updated User-Agent blacklist
* updated dependencies

## 6.23.1

* improved bot filter
* added ClickHouse default password
* upgraded to Go version 1.25
* updated dependencies

## 6.23.0

* added an option to disable the bot filter entirely
* updated dependencies

## 6.22.2

* updated referrer blacklist
* updated User-Agent blacklist
* updated dependencies

## 6.22.1

* updated referrer blacklist
* updated User-Agent blacklist
* updated dependencies

## 6.22.0

* added non-interactive events
* added event meta keys to filter options
* added browser and OS to filter options
* added weekday mode configuration to filter
* added the option to add multiple IP address filters
* added a simple flat IP blocklist implementation
* removed the requirement for event to be set to filter for key/value options
* updated dependencies

## 6.21.3

* fixed Android WebView being filtered

## 6.21.2

* fixed assigning some AI referrers/sources to channel

## 6.21.1

* added Bluesky referrer group
* added Viewport-Width and Width header to get screen class
* improved bot filter using browser
* updated and improved channel attribution list
* updated referrer blacklist
* fixed organic search traffic attribution
* updated dependencies

## 6.21.0

* added search field for filter options
* added AI to ignore list (User-Agent only)
* added traffic source channel for AI
* added DuckDuckGo browser detection
* updated referrer blacklist
* updated referrer mapping
* updated dependencies

## 6.20.0

* upgraded required Go version to 1.24
* fixed storing a page view for an event if the session hasn't been created yet
* updated dependencies

## 6.19.8

* changed required Go version back to 1.23

## 6.19.7

* always store a page view, even if bounced
* fixed time on page when first page is reloaded multiple times
* updated to Go 1.24

## 6.19.6

* added `Empty` method to filter

## 6.19.5

* updated User-Agent blacklist
* updated OS mapping
* fixed reading imported statistics with platform filter
* updated dependencies

## 6.19.4

* added function returning details for a request
* updated referrer blacklist
* updated dependencies

## 6.19.3

* added `Equal` method to filter
* optimized funnels
* updated referrer blacklist
* updated User-Agent blacklist
* fixed tests

## 6.19.2

* removed funnel step limit
* updated dependencies

## 6.19.1

* fixed sorting cities with non-ASCII characters in name

## 6.19.0

* added anonymous gclid and msclkid tracking
* added regex User-Agent filtering
* added source attribution channels
* updated User-Agent blacklist
* updated dependencies

## 6.18.1

* fixed migration

## 6.18.0

**This release contains backward incompatible changes!**

* added cluster support for ClickHouse

## 6.17.0

* added max page views configuration per request
* updated referrer blacklist
* updated User-Agent blacklist
* updated dependencies

## 6.16.2

* removed hostname from page, entry page, exit page, and active visitor stats
* updated dependencies

## 6.16.1

* fixed including VPNs based on IP address
* improved logging

## 6.16.0

* removed hostname fallback
* updated dependencies

## 6.15.5

* fixed reading events with path or regex filter

## 6.15.4

* fixed reading events with path or regex filter

## 6.15.3

* fixed overwriting hostname in URL

## 6.15.2

* read entry/exit pages in batch rather than everything at once
* fixed regex filter for imported statistics and platforms

## 6.15.1

* read average time on page in batch rather than everything at once
* updated referrer blacklist
* updated dependencies

## 6.15.0

* added storing hostname
* added reading and filtering for hostnames
* added reading visitors by weekday and hour
* added sorting by entry and exit rate
* improved entry and exit rate calculation
* fixed storing milliseconds for timestamps
* updated dependencies

## 6.14.2

* improved Arc browser detection
* updated WebKit version list
* updated dependencies

## 6.14.1

* removed boundary checks previously necessary to fix incorrect session data
* fixed migration if migration table already exists
* updated User-Agent blacklist
* updated dependencies

## 6.14.0

* switched to VersionedCollapsingMergeTree for session data
* use batch inserts instead of prepared statements
* updated dependencies

## 6.13.3

* fixed types for views and bounces result fields

## 6.13.2

* fixed reading platform with imported statistics and page filter set

## 6.13.1

* upgraded to Go 1.23
* improved creating and cleaning up test data
* speed up tests
* fixed logging error
* fixed relative visitor calculation for platform

## 6.13.0

* added merging imported statistics
* updated and improved referrer blacklist
* updated User-Agent blacklist
* updated dependencies

## 6.12.2

* fixed storing bot traffic reason

## 6.12.1

* added reason for bot traffic
* updated dependencies

## 6.12.0

* added funnel
* updated referrer blacklist
* updated User-Agent blacklist
* updated dependencies

## 6.11.1

* added `not contains` filter option
* fixed filter excluding countries
* updated referrer blacklist
* updated User-Agent blacklist
* updated dependencies

## 6.11.0

* replaced `neighbor` with window functions
* updated dependencies

## 6.10.0

**WARNING: this release has a breaking configuration change for GeoDB and Udger!**

* added configurable download URL for GeoDB and Udger
* fully automated schema migrations (including steps that were previously run manually)
* updated dependencies

## 6.9.2

* improved bot filter

## 6.9.1

* added country filter validation
* updated dependencies

## 6.9.0

* added regions
* updated User-Agent blacklist
* updated referrer blacklist
* updated referrer mapping
* updated dependencies

## 6.8.5

* fixed logging IP address if configured for bot requests

## 6.8.4

* removed honoring DNT (nobody uses it)
* updated dependencies

## 6.8.3

* improved logging bot traffic
* updated dependencies

## 6.8.2

* fixed filter for session breakdown

## 6.8.1

* fixed missing IncludeTime for TotalVisitors filter
* updated dependencies

## 6.8.0

* added listing and reading individual sessions
* added grouping results by minute
* fixed joining time stats
* fixed joining events when grouping by hours/minutes
* updated User-Agent blacklist
* updated dependencies

## 6.7.4

* added boolean to Tracker returning whether a request has been dropped
* updated dependencies

## 6.7.3

* added a fixed limit to the number of filter options that are returned
* improved browser and OS parsing
* updated dependencies

## 6.7.2

* fixed reading event breakdown
* updated dependencies

## 6.7.1

* fixed filtering for custom metrics and tags

## 6.7.0

* added tags for segmentation to page views
* added trimming to metadata keys
* added storing requests for new sessions with or without bot flag
* empty metadata keys and values are now ignored
* renamed bot table to request
* removed `OptionsFromRequest`
* removed `user_agent` table
* fixed SQL injection attack vector
* updated dependencies

## 6.6.8

* updated dependencies

## 6.6.7

* improved User-Agent filter

## 6.6.6

* improved User-Agent filter
* updated User-Agent blacklist
* updated referrer mapping
* updated dependencies

## 6.6.5

* fixed missing filter for time on page
* updated dependencies

## 6.6.4

* added retries to `Tracker` when saving data
* updated dependencies

## 6.6.3

* added User-Agent and client hint detection for Arc
* updated dependencies

## 6.6.2

* fixed exit rate calculation

## 6.6.1

* removed "dalvik" from the User-Agent blacklist
* removed id field from filter fields
* do not log context cancelled errors
* fixed parsing generic Sec-CH-UA for "Not", "Not A", and "Not)A"
* fixed reading event pages with title and sampling
* fixed casting negative UInt64s due to broken data

## 6.6.0

* added optional context to cancel queries
* added sampling
* fixed insert parameter placeholders
* updated User-Agent blacklist
* updated dependencies

## 6.5.2

* improved tests related to time on page
* fixed average time on page calculation
* fixed events resetting time the session was last seen
* updated dependencies

## 6.5.1

* improved tests related to session duration and time on page
* fixed average session duration calculation with path filter

## 6.5.0

**This release requires a manual database migration. Check `pkg/db/schema/0020_primary_keys_sampling.up.sql` for details.**

* added sampling keys for all tables
* added number of events to event statistics
* ignore comments in migration scripts
* fixed primary key for events table
* fixed alias for path when merging page view, session, and event table
* updated dependencies

## 6.4.1

* optimized query builder
* fixed limit for reading total number of sessions per page
* fixed tests

## 6.4.0

* added `TotalPageViews` and `TotalSessions` to `Analyzer`
* updated dependencies

## 6.3.1

* fixed timezone for session duration and time on page

## 6.3.0

* create page view when event is fired on a page that's different from the last page visited
* update page views for session when event is fired on a page that's different from the last page visited
* fixed missing page view when event creates a session
* updated referrer mapping
* updated User-Agent blacklist
* updated dependencies

## 6.2.1

* fixed filter for total unique number of visitors

## 6.2.0

* added `TotalVisitors` to `Analyzer`
* updated dependencies

## 6.1.1

* fixed time shift in daily visitor statistics

## 6.1.0

* ignore www to get the referrer
* updated referrer list
* updated referrer blacklist
* updated dependencies

## 6.0.0

* refactored package structure and a few method and struct names
* switched to `log/slog` for logging
* removed deprecated `rand.Seed` initialization
* updated browser minimum versions
* added support for client hints
* added Chrome OS and Windows 11 detection
* added conversion rate to total, by period, and growth
* added custom metrics for events
* updated User-Agent blacklist
* updated referrer blacklist
* upgraded to Go 1.21
* updated dependencies

## 5.10.8

* added referrers to mapping

## 5.10.7

* fixed resetting sessions when referrer name changes

## 5.10.6

* fixed query parameter preference

## 5.10.5

* prefer query parameters over header to extract referrer

## 5.10.4

* added custom referrer mapping
* updated referrer mapping
* fixed script to update referrer mapping

## 5.10.3

* fixed calculating average session duration
* updated dependencies

## 5.10.2

* updated dependencies

## 5.10.1

* fixed grouping events by metadata
* updated dependencies

## 5.10.0

* added bot table for traffic that has been filtered
* removed `is_bot` counter

## 5.9.2

* do not increase page view count on bounce
* do not store page view if the session bounced before
* fixed calculating average session duration and time on page
* updated User-Agent blacklist
* updated dependencies

## 5.9.1

* use less aggressive bot filtering defaults
* updated dependencies

## 5.9.0

* improved string utility functions
* increased minimum delay between page views from 75ms to 500ms
* added option for maximum number of page views
* updated User-Agent blacklist
* updated referrer blacklist
* updated dependencies

## 5.8.2

* fixed page view count when session is created on an event
* updated dependencies

## 5.8.1

* fixed session locks (?)
* updated dependencies

## 5.8.0

* optimized field comparisons
* fixed entries/exits when filtering for a path pattern
* removed JavaScript scripts so that they don't have to be maintained in two places
* removed demo code
* updated User-Agent blacklist
* updated referrer blacklist
* updated dependencies

## 5.7.0

* added `Analyzer.Visitors.TotalVisitorsPageViews`
* updated dependencies

## 5.6.0

* added parsing referrer from query parameters (ref, utm_source, ...)
* added Safari version 16 Webkit version mapping
* fixed escaped URLs in Referer header
* fixed query when multiple metadata keys were set
* updated User-Agent blacklist
* updated dependencies

## 5.5.2

* fixed ClickHouse driver concurrency issues

## 5.5.1

* downgraded ClickHouse driver

## 5.5.0

* added region for US cities
* renamed `GeoDB.CountryCodeAndCity` to `GeoDB.GetLocation`
* updated User-Agent blacklist
* updated referrer blacklist
* updated dependencies

## 5.4.1

* added metadata options to analyzer

## 5.4.0

* added cache for Android app referrer mapping
* added User-Agents tests for recent browsers
* fixed Android app referrer mapping with trailing slashes
* upgraded to Go 1.20
* updated User-Agent blacklist
* updated referrer blacklist
* updated dependencies

## 5.3.0

* improved referrer mapping
* updated dependencies

## 5.2.0

* removed screen width and height (only the class is used)
* updated User-Agent blacklist
* updated referrer blacklist
* updated referrer mapping
* updated dependencies

## 5.1.0

* added parameter to overwrite the time a page view/event/session extension has been sent to support batch inserts
* updated User-Agent blacklist
* updated referrer blacklist
* updated referrer mapping
* updated dependencies

## 5.0.3

* added a rewrite option using the `data-dev` attribute

## 5.0.2

* removed unused custom parameters (`data-param-*`)
* updated User-Agent blacklist
* updated referrer blacklist
* updated referrer mapping
* updated dependencies

## 5.0.1

* fixed searching by path
* updated dependencies

## 5.0.0

* added `pirsch-sessions.js` script to automatically extend sessions
* added field to count how often a session has been extended
* added bot filter based on IP address
* added event pages as new statistic
* improved interfaces and defaults
* improved query builder
* fixed inverting event filter
* fixed extending sessions in database
* fixed automatic GeoDB update
* updated User-Agent blacklist
* updated referrer blacklist
* updated referrer mapping
* updated browser version mapping
* updated dependencies

## 4.5.1

* fixed page title for page views

## 4.5.0

* added options to disable the collection of query parameters, referrers, and resolution in scripts
* added `HitOptions.MaxPageViews` for maximum number of page views to flag bots
* updated User-Agent blacklist
* updated referrer blacklist
* updated referrer mapping
* updated browser version mapping
* updated dependencies

## 4.4.0

* connect negated parameters using logical AND instead of OR
* use nil instead of empty slices when removing filter options
* updated dependencies

## 4.3.2

* fixed platform statistics returning sessions instead of unique visitors
* fixed fingerprint test
* updated dependencies

## 4.3.1

* optimization reading session for previous day

## 4.3.0

* added 24 hour limit to fingerprint (previously closed source as part of the salt)
* updated dependencies

## 4.2.0

* ignore User-Agents containing non-ASCII characters
* ignore User-Agents containing less than 10 or more than 300 characters
* switched to `sendBeacon` for events
* limited session lifetime to 24 hours
* added sorting entry/exit pages by the number of visitors
* fixed `data-dev` attribute in `pirsch-events.js`
* fixed `SessionMaxAge` being used for the maximum session lifetime
* updated User-Agent blacklist
* updated referrer blacklist
* updated referrer mapping
* updated dependencies

## 4.1.0

* added visitors, sessions, entry/exit rate, and average time on page to entry/exit page statistics when filtering for an event
* upgraded to Go version 1.19
* switched from `atomic.LoadInt32` to `atomic.Bool`
* updated dependencies

## 4.0.0

* added multiple filters for the same field connected with OR
* added support for "contains" for filter fields
* added reading filter options using the `Analyzer`
* refactored code into modules
* refactored `Analyzer` to group methods
* ignore event metadata when event name is not set
* fixed entry/exit rate calculation
* fixed some concurrency issues
* updated User-Agent blacklist
* updated referrer blacklist
* updated referrer mapping
* updated dependencies

## 3.10.2

* fixed missing spaces in query builder
* updated dependencies

## 3.10.1

* changed weekly timescale to start on Monday
* fixed visitor statistics having the right date set depending on the timescale
* updated dependencies

## 3.10.0

* added database connection configuration and defaults
* added whitelisting pages
* removed jmoiron/sql and optimized reading results
* saving page views, events, or sessions will now panic on error
* fixed grouping cities and countries when city is unknown
* refactored `Tracker`
* updated dependencies

## 3.9.1

* always return clean IP from `getIP`
* allow overwriting parser using `HitOptions`

## 3.9.0

* changed `session.is_bounce` type to `Int8`, so that it is the same for all booleans
* updated ClickHouse driver to v2
* moved database configuration from connection string to struct
* added a custom schema migrator and removed golang-migrate/migrate
* use the rightmost IP from X-Forwarded-For and Forwarded headers
* check if the visitor IP address is valid
* overwrite the default header parser list for the `Tracker` and allow no header usage at all
* configure valid proxy IP address ranges for IP headers
* added detection for Safari 15
* updated User-Agent blacklist
* updated referrer blacklist
* updated referrer mapping
* updated dependencies

## 3.8.3

* fixed passing dates without timezone

## 3.8.2

* always return referrer URL in lower-case
* updated dependencies

## 3.8.1

* fixed statistics when using a different period other than "day"
* updated User-Agent blacklist
* updated referrer blacklist
* updated referrer mapping
* updated dependencies

## 3.8.0

* added `Search`, `Sort`, and `Offset` to `Filter`
* changed min delay for flagging bots to 50ms
* exposed fields and field names
* updated referrer blacklist
* updated referrer mapping
* updated dependencies

## 3.7.8

* replace `substr` with `substring` in scripts
* added option to send data to multiple clients in scripts

## 3.7.7

* fixed filtering entry/exit pages by path pattern
* upgraded to Go version 1.18
* updated referrer mapping
* updated dependencies

## 3.7.6

* fixed calculating growth for today not including time

## 3.7.5

* the growth for a single day is now compared to the same day of the previous week
* fixed building entry/exit page queries when filtering for path and grouping by page titles
* updated dependencies

## 3.7.4

* changed default min delay for bot detection to 200ms
* use the same buffer size for everything in `Tracker`
* fixed releasing session lock when request is cancelled because it's a bot

## 3.7.3

* added options to `Client`
* added debug logging to `Client`
* changed default `Tracker` timeout from 10 to 3 seconds before hits/events are saved
* updated dependencies

## 3.7.2

* fixed calculating growth when including time in filter

## 3.7.1

* fixed including time in filter for live view

## 3.7.0

* added `True-Client-IP` header to receive the real visitor IP (Cloudflare)
* added new bot list by Atmire (https://github.com/atmire/COUNTER-Robots/blob/master/COUNTER_Robots_list.json)
* added script to update and clean up the User-Agent blacklist
* added flagging bots based on how quickly page views are being sent
* added options to `Analyzer`
* added grouping results by week, month, and year and new fields `Week`, `Month`, and `Year` to `VisitorStats` and `TimeSpentStats`
* calculating the growth for today will now take the time into account, and not compare to the full past day
* optimized JS scripts and added page filtering and option to disable the script using localStorage
* removed `Filter.Day` and `Filter.Start` and added `Filter.IncludeTime` instead, allowing to filter everything by date and time
* `VisitorStats.Day` and `TimeSpentStats.Day` are now nullable fields (new week, month, and year are also of type `sql.NullTime`)
* fixed grouping referrers with path in URL
* fixed missing timezone
* updated User-Agent blacklist
* updated referrer blacklist
* updated referrer mapping
* updated dependencies

## 3.6.4

* added counting multiple page views without switching the actual page as bounced again
* fixed filtering time range when joining the session table to page views

## 3.6.3

* fixed counting bounces multiple times when grouping by pages

## 3.6.2

* clear session buffer before it's full or exceeds the buffer size
* fixed concurrent access to redis session cache
* updated dependencies

## 3.6.1

* include country code in cities statistics
* make sure path is not included more than once in filter list
* moved capturing the page view time after looking up the session to prevent overlaps
* added JSON field names for page views, sessions, and events
* smaller optimizations
* updated User-Agent blacklist
* updated browser version mapping
* updated referrer mapping
* fixed time on page test when grouping by page title

## 3.6.0

* added sessions, views, bounces, and bounce rate to hourly visitor statistics
* fixed average session duration calculation
* updated referrer name list
* updated dependencies

## 3.5.7

* drop backup tables
* fixed github.com/containerd/containerd security issue
* set minimum Go version to 1.17
* updated dependencies

## 3.5.6

* migrated to DateTime64 for millisecond precision dates
* added optimistic locking for session cache (in-memory and Redis)
* fixed session collapsing order by
* updated User-Agent blacklist
* updated dependencies

## 3.5.5

* use `strconv.ParseUint` instead of `strconv.Atoi` where possible
* updated referrer list
* updated referrer blacklist
* updated dependencies

## 3.5.4

* fixed Windows 11 User-Agent version

## 3.5.3

* optimized query builder
* fixed missing session collapsing
* updated dependencies

## 3.5.2

* fixed grouping referrers with trailing slashes
* updated dependencies

## 3.5.1

* added event filter for entry/exit pages

## 3.5.0

* create new session when referrer or UTM parameter changes
* update session on event and set bounced to false (keep everything else)
* added list to group referrer domains
* added entry/exit page filter for events
* added screen width/height to filter (exact match)
* added listing events and filtering for event meta key and value
* optimized building queries
* removed TTL from tables
* group OS versions by minor version instead of full length
* fixed User-Agent blacklist (must be lowercase)
* fixed growing sessions buffer in tracker
* updated dependencies

## 3.4.5

* updated User-Agent blacklist
* updated OS and browser versions

## 3.4.4

* updated referrer spam list
* ignore no rows found error when selecting a single result

## 3.4.3

* added `Analyzer.TotalVisitors` (sums not grouped by day)
* don't return an error when no result is found when expecting a single row
* fixed conversion goals test

## 3.4.2

* added logger to Redis session cache
* return nil if path pattern is not set for `Analyzer.PageConversions`
* fixed order in which session cancel state gets stored
* updated dependencies

## 3.4.1

* added tuple for session state cancellation
* removed the requirement to send a page view before an event can be tracked
* fixed collapsing sessions for statistics
* fixed reading growth when there is no data

## 3.4.0

* switched to CollapsingMergeTree in favor of materialized view (which didn't work)
* rewrote all queries

## 3.3.2

* fixed reading active visitors directly from hit table
* fixed reading full referrer (for favicon URL)

## 3.3.1

* fixed limit for `Analyzer.Visitors`

## 3.3.0

* added materialized view for sessions and events
* added optional salt per request to `HitOptions`
* optimized queries using materialized views
* switched to SipHash for fingerprints (64 bit instead of 256)
* removed `Filter.IncludeAvgTimeOnPage` (now always included)
* removed entry_path, page_views, and is_bounce from events

## 3.2.1

* small optimization for `ExtendSession`
* added convenience method `Tracker.ExtendSession`

## 3.2.0

* added function to manually extend sessions
* fixed events counting up page views

## 3.1.4

* fixed timezone for active visitors

## 3.1.3

* fixed limiting filter to today (UTC)
* some query optimizations

## 3.1.2

* added limit for maximum duration on page

## 3.1.1

* fixed filtering for entry/exit pages (path)
* fixed calculation of entry and exit rates

## 3.1.0

* added interface for session cache
* added Redis session cache for distributed systems
* added filtering for entry/exit page paths back in
* updated dependencies

## 3.0.2

* fixed filtering for exit pages in Analyzer.AvgTimeOnPage

## 3.0.1

* fixed filtering for entry/exit pages
* updated dependencies

## 3.0.0

* added rolling forward page view information for deeper analysis of sessions
* added filtering for entry and exit page
* added filtering for "none"/"unknown" (empty strings) by setting a filter to "null"
* added all statistics available for hits to events as well
* added optional limit for active visitor statistics
* added merging referrers by hostname and detailed statistics by filtering for the referrer name
* added city statistics
* added method to clear session cache to tracker
* optimized data layout
* optimized statistics queries
* optimized filter (non required fields are now longer selected)
* the User-Agent header is now stored in a separate table for later analysis (filtering bots)
* removed unused UserAgent and full URL from hit and event table
* remove trailing slashes from referrer URLs
* updated dependencies

## 2.6.3

* fixed session cache

## 2.6.2

* feed hash through MD5 again to shorten it instead of shortening the SHA256 string directly
* updated dependencies

## 2.6.1

* changed hashing algorithm to SHA256

## 2.6.0

* added cache to read sessions
* added filter inversion
* fixed session query order by
* updated dependencies

## 2.5.2

* fixed usage of TrackerConfig.SessionMaxAge

## 2.5.1

* encoded all URL parameters in pirsch.js where it is required
* modernized pirsch.js
* added reference implementation for pirsch-events.js
* added more screen classes
* fixed sending UTM parameters using pirsch.js
* updated dependencies

## 2.5.0

* added collection of page title
* added grouping results by page title additionally to path
* optimized pirsch.js
* updated referrer blacklist
* updated User-Agent blacklist
* updated dependencies

## 2.4.0

* added custom event tracking
* added order by unique visitors to conversion goals
* fixed relative visitors/views and conversion rate if right side is zero
* updated User-Agent blacklist
* updated dependencies

## 2.3.0

* added `MaxTimeOnPageSeconds` option to filter
* removed date from fingerprint (this is a GDPR-compliant change)
* removed timezone from `HitOptions` (no longer needed)
* updated User-Agent blacklist
* updated dependencies

## 2.2.7

* fixed schema migration for non-nullable fields

## 2.2.6

* fixed buffering hits in tracker

## 2.2.5

* added page conversion function to analyzer
* fixed calculating relative values (they were previously calculated using a GROUP BY clause, which didn't make sense)

## 2.2.4

* added `PathPattern` to filter and analyzer to query paths for a regex

## 2.2.3

All `sql.Null...` fields have been changed to non-nullable fields!

* removed nullable fields from schema and model for better ClickHouse performance
* improved and fixed User-Agent blacklist
* updated dependencies

## 2.2.2

* added missing os and browser version methods to analyzer

## 2.2.1

Just tagging a new version, no changes.

## 2.2.0

* added entry and exit pages

## 2.1.0

* added timezone support (UTC by default)
* updated referrer blacklist
* updated dependencies

## 2.0.1

* added SaaSHub to User-Agent blacklist
* fixed analyzer return types (some used sql.NullString for normal strings)

## 2.0.0

Version 2 brings some fundamental changes and is incompatible with version 1.

* switched to ClickHouse in favor of Postgres, as it far better fits the problem domain
* removed data aggregation and the processor
* tenant ID -> client ID
* automatic schema migration (using the Go 1.16 embedding feature, x-multi-statement must be set to true)
* added UTM query parameter tracking (for campaign tracking)
* added option to limit result sets to filter

## 1.14.4

* updated referrer blacklist
* updated dependencies

## 1.14.3

* removed `GeoDB.Close`, as it is no longer required to close the file resource
* fixed filtering by path for growth statistics
* fixed concurrent access to GeoDB, the database file is now loaded into memory
* fixed dividing by 0 when calculating bounce rate growth

## 1.14.2

* optimized Postgres transaction handling
* ignore IP addresses as referrer

## 1.14.1

* fixed average time on page calculation for today

## 1.14.0

**This release requires Go version 1.16!**

* removed deprecated package io/ioutil
* updated dependencies

## 1.13.4

* fixed checking all values before calculating relative values by division

## 1.13.3

* added unit (seconds) to `PathVisitors.AverageTimeOnPageSeconds`

## 1.13.2

* fixed calculating relative page views per day

## 1.13.1

* fixed calculating relative page views

## 1.13.0

**You need to update the schema by running the `v1.13.0.sql` migration script!**

* changed the default session max age from two hours to 15 minutes (you can keep two hours by setting the `TrackerConfig.SessionMaxAge` option)
* added page views to statistics (including relative views and growth)
* added "today" to growth rates, if it's within period (visitors, sessions, bounces, views)
* added average session duration per day
* added session duration growth to analyzer
* added time spent on page statistics
* updated dependencies
* updated referrer blacklist
* updated Safari User-Agent versions

## 1.12.2

* added absolute number of bounces to page visitor statistics
* updated dependencies

## 1.12.1

* added total visitor count, relative visitors, and bounce rate to page visitor statistics
* fixed sorting page visitor statistics

## 1.12.0

**You need to update the schema by running the `v1.12.0.sql` migration script!**

* database indices optimization
* always save language and country codes in lowercase
* test refactorings
* updated dependencies

## 1.11.1

* added relative visitor count to `Analyer.PageVisitors`

## 1.11.0

**You need to update the schema by running the `v1.11.0.sql` migration script!**

* updated dependencies
* updated referrer blacklist
* map Android referrers to app name and icon from the Google Play Store
* added bounce rate per referrer

## 1.10.7

* fixed `RunAtMidnight` not firing

## 1.10.6

* fixed filtering referrers if the hostname cannot be parsed

## 1.10.5

* accept non-url referrers (like for utm_source)

## 1.10.4

* added *source* and *utm_source* as referrer URL parameters
* updated referrer blacklist
* updated dependencies

## 1.10.3

* respect the "do not track" header (backend and js)

## 1.10.2

* more logging for GeoDB
* updated dependencies

## 1.10.1

**You need to update the schema by running the `v1.10.1.sql` migration script!**

* renamed screen classes

## 1.10.0

**You need to update the schema by running the `v1.10.0.sql` migration script!**

* improved logging
* `NewGeoDB` now takes a config as its parameter with optional logger for debugging
* added screen classes to hits and statistics
* screen sizes with 0 width or height are no longer processed and stored

## 1.9.3

**You need to update the schema by running the `v1.9.3.sql` migration script!**

* removed path from time of day statistics
* fixed visitors being counted multiple times in statistics for paths

## 1.9.2

* fixed timer running immediately in `RunAtMidnight`
* fixed running `tracker.Stop` more than once
* fixed flushing hits in tracker correctly (more reliable tests)

## 1.9.1

**You need to update the schema by running the `v1.9.1.sql` migration script!**

* normalized empty paths for referrers
* removed sessions from visitor stats by hours as that does not make sense
* updated dependencies

## 1.9.0

**You need to update the schema by running the `v1.9.0.sql` migration script!**

* hit path is no longer optional and will be set to "/" if empty
* improved Chrome vs Safari detection
* updated referrer blacklist
* ignore hits made by browser versions before 2018
* fixed and shorten URL in js integration script
* fixed relative growth calculation for bounce rate
* fixed null paths

## 1.8.3

* use time.NewTimer instead of time.After for more efficiency and better garbage collection

## 1.8.2

* updated dependencies
* updated referrer blacklist

## 1.8.1

* fixed calculating bounce rate growth

## 1.8.0

* group languages (en-us, en-gb, ... all become en) and check for valid ISO codes
* filter limits the time range for up until today, as it doesn't make any sense to go beyond that
* added growth calculation to Analyzer

## 1.7.5

* fixed filtering visitors by time of day by tenant ID

## 1.7.4

* fixed 0 sum in Analyzer for platform statistics

## 1.7.3

* added json to Analyzer structs, so that they can be added to an API

## 1.7.2

* changed license to GNU AGPLv3
* renamed package from emvi/pirsch to pirsch-analytics/pirsch
* updated dependencies

## 1.7.1

* fixed filtering referrer spam subdomains

## 1.7.0

* `Tracker.Hit` does no longer spawn its own goroutine, so you should do that yourself
* added visitors statistics for time and day for a range of days to Analyzer
* added optional time zone to Analyzer
* fixed reading sessions without tenant ID
* fixed reading hit days without time zone

## 1.6.0

**You need to update the schema by running the `v1.6.0.sql` migration script!**

* added client side tracking (pirsch.js)
* added screen size to Hit, Processor and Anlayzer for client side tracking
* Tracker.Stop now processes all hits in queue before shutting down (Tracker.Flush does not!)
* improved documentation and demos
* fixed counting bounces for path
* fixed counting platforms

## 1.5.2

* fixed grouping language, referrer, OS, and browser statistics

## 1.5.1

* fixed counting active visitors
* fixed counting platforms
* fixed reading statistics for today if no history exists

## 1.5.0

**You need to update the schema by running the `v1.5.0.sql` migration script!**
**WARNING: this release uses a new data structure to store statistics and is incompatible with previous versions. You need to migrate and drop the unused tables using the following statements (migration steps NOT included):**

```SQL
DROP TABLE "visitor_platform";
DROP TABLE "visitors_per_browser";
DROP TABLE "visitors_per_day";
DROP TABLE "visitors_per_hour";
DROP TABLE "visitors_per_language";
DROP TABLE "visitors_per_os";
DROP TABLE "visitors_per_page";
DROP TABLE "visitors_per_referrer";
```

* implemented a new data model
* added session tracking
* added referrer spam protection
* added bounce rates
* re-implemented the Analyzer to support the new data model and make it easier to use

## 1.4.3

**You need to update the schema by running the `v1.4.3.sql` migration script!**

* fixed saving processed data

## 1.4.2

* fixed null fields in model
* fixed counting visitors multiple times (by using a transaction to rollback changes in case the processor fails)
* added optional log.Logger to Tracker and PostgresStore
* removed all panics and log errors instead

## 1.4.1

* added relative visitor statistics for OS and browser usage

## 1.4.0

**You need to update the schema by running the `v1.4.0.sql` migration script!**

* added parsing the User-Agent header to extract the OS, OS version, browser, and browser version
* added OS, browser and platform statistics to Processor and Analyzer
* Pirsch now uses a single struct for all statistics called `Stats`
* fixed error channel size in Processor
* a few smaller refactorings

## 1.3.3

* fixed extracting Referer header
* added ref, referer and referrer query parameters for referrers, when Referer header is not present

## 1.3.2

**You need to update the schema by running the `v1.3.2.sql` migration script!**

* fixed helper function `RunAtMidnight` not using UTC for all times
* referer -> referrer

## 1.3.1

* added statistics for visitor count per page and referrer

## 1.3.0

**You need to update the schema by running the `v1.3.0.sql` migration script!**

* added cancel function to `RunAtMidnight`
* added helper function for tenant IDs
* hits for an empty User-Agent will be ignored from now on
* added configuration options to `Processor`
* `IgnoreHit` and `HitFromRequest` are now exported functions
* added options to filter for unwanted referrer, like your own domain
* added referrer statistics to `Processor` and `Analyzer`
* added method to `Analyzer` to extract active visitor pages
* `Analyzer` results are now sorted by visitors in descending order instead of path and referrer length

## 1.2.0

**You need to update the schema by running the `v1.2.0.sql` migration script!**

* the processor now returns an error
* the processor now updates existing statistics in case it has been run before, but keep in mind that it drops hits and therefor breaks tracking users that return on the same day. It's recommended to run the processor for days in the past excluding today
* (optional) multi-tenancy support to track multiple domains using the same database. In case you don't want to use it, use null as the `client_id`
* improved IP extraction from X-Forwarded-For, Forwarded and X-Real-IP headers

## 1.1.1

* fixed error in case values are too long
* fixed language statistics not including today

## 1.1.0

* added a secret salt to prevent generating fingerprints to identify visitors on other websites (man in the middle)
* extended bot list

## 1.0.0

Initial release.
