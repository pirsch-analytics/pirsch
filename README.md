# Pirsch

[![Go Reference](https://pkg.go.dev/badge/github.com/pirsch-analytics/pirsch?status.svg)](https://pkg.go.dev/github.com/pirsch-analytics/pirsch?status)
[![Go Report Card](https://goreportcard.com/badge/github.com/pirsch-analytics/pirsch/v5)](https://goreportcard.com/report/github.com/pirsch-analytics/pirsch/v5)
<a href="https://discord.gg/fAYm4Cz"><img src="https://img.shields.io/discord/739184135649886288?logo=discord" alt="Chat on Discord"></a>

Pirsch is a server side, no-cookie, drop-in and privacy focused tracking solution for Go. Integrated into a Go application it enables you to track HTTP traffic without invading the privacy of your visitors. The visualization of the data (dashboard) is not part of this project.

**If you're looking for a managed solution with an easy-to-use API and JavaScript integration, check out https://pirsch.io/.**

## How does it work?

Pirsch generates a unique fingerprint for each visitor. The fingerprint is a hash of the visitors IP, User-Agent, the date, and a salt.

Each time a visitor opens your page, Pirsch will store a hit. The hits are analyzed using the `Analyzer` to extract meaningful data.

The tracking works without invading the visitor's privacy as no cookies are used nor required. Pirsch can track visitors using ad blockers that block trackers like Google Analytics.

## Features

* unique visitor count per day, path, and hour
* session count
* bounce rate
* view count
* growth (unique visitors, sessions, bounces, views, average session duration)
* average time on page
* average session duration
* languages
* operating system and browser (including versions)
* referrers
* countries
* cities
* platform
* screen size
* UTM query parameters for campaign tracking
* entry and exit pages
* custom event tracking
* conversion goals

All timestamps are stored as UTC. Starting with version 2.1, the results can be transformed to the desired timezone. All data points belongs to an (optional) client, which can be used to split data between multiple domains for example. If you just integrate Pirsch into your application, you don't need to care about that field. **But if you do, you need to set a client ID for all columns!**

## Usage

To store hits and statistics, Pirsch uses ClickHouse. Database migrations can be run manually be executing the migrations steps in `schema` or by using the automatic migration.

**Make sure you read the changelog before upgrading! There are sometimes manual steps required to migrate the data to the new version.**

### Server-side tracking

Please refer to the [demo](demos/backend) for server-side tracking.

### Client-side tracking

Please refer to the [demo](demos/frontend) for client-side tracking and keeping sessions alive.

The scripts are configured using HTML attributes. All of them are optional, except for the `id`. Here is a list of the possible options.

| Option | Description | Default |
| - | - | - |
| data-endpoint | The endpoint to call. This can be a local path, like /tracking, or a complete URL, like http://mywebsite.com/tracking. It must not contain any parameters. | /pirsch |
| data-client-id | The client ID to use, in case you plan to track multiple websites using the same backend, or you want to split the data. Note that the client ID must be validated in the backend. | 0 (no client) |
| data-include | Specifies a list of regular expressions to test against. On a match, the page view or event will be included. This is done before excluding any pages. | (no paths) |
| data-exclude | Specifies a list of regular expressions to test against. On a match, the page view or event will be ignored. This is done after including any pages. | (no paths) |
| data-domain | Specifies a list of additional domains to send data to. | (empty list) |
| data-dev | Enable tracking hits on localhost. This is used for testing purposes only. You can set a hostname to overwrite localhost to something else, like `data-dev="example.com"`. | undefined |
| data-disable-query | Removes all query parameters from the URL. | false |
| data-disable-referrer | Disables the collection of the referrer. | false |
| data-disable-resolution | Disables the collection of the screen resolution. | false |

The scripts can be disabled by setting the `disable_pirsch` variable in localStorage of your browser.

### Mapping IPs to countries and cities

Pirsch uses MaxMind's [GeoLite2](https://dev.maxmind.com/geoip/geoip2/geolite2/) database to map IPs to countries. The database **is not included**, so you need to download it yourself. IP mapping is optional, it must explicitly be enabled by setting the GeoDB attribute of the `TrackerConfig` or through the `HitOptions` when calling `HitFromRequest`.

1. create an account at MaxMind
2. generate a new license key
3. call `geodb.Get` with the path you would like to extract the tarball to and pass your license key
4. create a new GeoDB by using `NewGeoDB` and the file you downloaded and extracted using the step before

The GeoDB should be updated on a regular basis. The Tracker has a method `SetGeoDB` to update the GeoDB at runtime (thread-safe).

### Filtering IP addresses

It's possible to filter bots using an IP address list. Currently, Pirsch supports [Udger](https://udger.com).

1. create an account at Udger
2. subscribe to the local parser (SQLite) IP list
3. create an `ip.Udger` object using your access key
4. pass it to the `Tracker`

The list inside `ip.Udger` can be automatically updated by calling `ip.Udger.Update`.

## Documentation

Read the [full documentation](https://godoc.org/github.com/pirsch-analytics/pirsch) for details, check out `demos`, or read the article at https://marvinblum.de/blog/server-side-tracking-without-cookies-in-go-OxdzmGZ1Bl.

## Building the scripts

To minify `pirsch.js`/`pirsch-events.js`/`pirsch-sessions.js` to `pirsch.min.js`/`pirsch-events.min.js`/`pirsch-sessions.min.js` you need to run `npm i` and `npm run build` inside the `js` directory.

## Things to maintain

The following things need regular maintenance/updates (using the scripts in the `scripts` directory when possible):

* Go and JS dependencies
* referrer blacklist
* User-Agent blacklist
* browser version mapping
* os version mapping
* referrer mapping (grouping)

GeoDB updates itself if used.

## Changelog

See [CHANGELOG.md](CHANGELOG.md).

## Contribution

Contributions are welcome! Please open a pull requests for your changes and tickets in case you would like to discuss something or have a question.

To run the tests you'll need a ClickHouse database, and a schema called `pirschtest`. The user is set to `default` (no password).

Note that we only accept pull requests if you transfer the ownership of your contribution to us. As we also offer a managed commercial solution with this library at its core, we want to make sure we can keep control over the source code.

## License

GNU AGPLv3

<p align="center">
    <img src="gopher.svg" alt="Gopher" width="200px" />
</p>
