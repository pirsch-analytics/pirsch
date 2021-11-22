# Changelog

## 3.4.6

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
* removed date from fingerprint (this is a GDPR compliant change)
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
