# Changelog

## 1.9.0

**You need to update the schema by running the `v1.9.0.sql` migration script!**

* hit path is no longer optional and will be set to "/" if empty
* improved Chrome vs Safari detection
* updated referrer spam list
* ignore hits made by browser versions before 2018
* fixed and shorten URL in js integration script
* fixed relative growth calculation for bounce rate
* fixed null paths

## 1.8.3

* use time.NewTimer instead of time.After for more efficiency and better garbage collection

## 1.8.2

* updated dependencies
* updated referrer spam list

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
* (optional) multi-tenancy support to track multiple domains using the same database. In case you don't want to use it, use null as the `tenant_id`
* improved IP extraction from X-Forwarded-For, Forwarded and X-Real-IP headers

## 1.1.1

* fixed error in case values are too long
* fixed language statistics not including today

## 1.1.0

* added a secret salt to prevent generating fingerprints to identify visitors on other websites (man in the middle)
* extended bot list

## 1.0.0

Initial release.
