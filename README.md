# Pirsch

[![Go Reference](https://pkg.go.dev/badge/github.com/pirsch-analytics/pirsch?status.svg)](https://pkg.go.dev/github.com/pirsch-analytics/pirsch?status)
[![Go Report Card](https://goreportcard.com/badge/github.com/pirsch-analytics/pirsch)](https://goreportcard.com/report/github.com/pirsch-analytics/pirsch)
[![CircleCI](https://circleci.com/gh/pirsch-analytics/pirsch.svg?style=svg)](https://circleci.com/gh/pirsch-analytics/pirsch)
<a href="https://discord.gg/fAYm4Cz"><img src="https://img.shields.io/discord/739184135649886288?logo=discord" alt="Chat on Discord"></a>

Pirsch is a server side, no-cookie, drop-in and privacy focused tracking solution for Go. Integrated into a Go application it enables you to track HTTP traffic without invading the privacy of your visitors. The visualization of the data (dashboard) is not part of this project.

The name is in German and refers to a special kind of hunt: *the hunter carefully and quietly enters the area to be hunted, he stalks against the wind in order to get as close as possible to the prey without being noticed.*

**If you're looking for a managed solution with an easy-to-use API and JavaScript integration, check out https://pirsch.io/.**

## How does it work?

Pirsch generates a unique fingerprint for each visitor. The fingerprint is a hash of the visitors IP, User-Agent, the date, and a salt. The date guarantees that the data is separated by day, so visitors can only be tracked for up to one day.

Each time a visitor opens your page, Pirsch will store a hit. The hits are analyzed later to extract meaningful data and reduce storage usage by aggregation.

The tracking works without invading the visitor's privacy as no cookies are used nor required. Pirsch can track visitors using ad blockers that block trackers like Google Analytics.

## Features

Pirsch tracks the following data:

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
* platform
* screen size

All timestamps are stored as UTC. All data points belongs to an (optional) tenant, which can be used to split data between multiple domains for example. If you just integrate Pirsch into your application, you don't need to care about that field. **But if you do, you need to set a tenant ID for all columns!**

## Usage

To store hits and statistics, Pirsch uses a database. Right now only Postgres is supported, but new ones can easily be added by implementing the Store interface. The schema can be found within the schema directory. Changes will be added to migrations scripts, so that you can add them to your projects database migration or run them manually.

### Server-side tracking

Here is a quick demo on how to use the library:

```Go
// Create a new Postgres store to save statistics and hits.
store := pirsch.NewPostgresStore(db, nil)

// Set up a default tracker with a salt.
// This will buffer and store hits and generate sessions by default.
tracker := pirsch.NewTracker(store, "salt", nil)

// Create a new process and run it each day on midnight (UTC) to process the stored hits.
// The processor also cleans up the hits.
processor := pirsch.NewProcessor(store)
pirsch.RunAtMidnight(func() {
    if err := processor.Process(); err != nil {
        panic(err)
    }
})

// Create a handler to serve traffic.
// We prevent tracking resources by checking the path. So a file on /my-file.txt won't create a new hit
// but all page calls will be tracked.
http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/" {
        go tracker.Hit(r, nil)
    }

    w.Write([]byte("<h1>Hello World!</h1>"))
}))

// And finally, start the server.
// We don't flush hits on shutdown but you should add that in a real application by calling Tracker.Flush().
log.Println("Starting server on port 8080...")
http.ListenAndServe(":8080", nil)
```

To analyze hits and processed data you can use the analyzer, which provides convenience functions to extract useful information.

The secret salt passed to `NewTracker` should not be known outside your organization as it can be used to generate fingerprints equal to yours.
Note that while you can generate the salt at random, the fingerprints will change too. To get reliable data configure a fixed salt and treat it like a password.

```Go
// This also needs access to the store.
// You can set a time zone through the configuration to display local times.
analyzer := pirsch.NewAnalyzer(store, nil)

// As an example, lets extract the total number of visitors.
// The filter is used to specify the time frame you're looking at (days) and is optional.
// If you pass nil, the Analyzer returns data for the past week including today.
visitors, err := analyzer.Visitors(&pirsch.Filter{
    From: yesterday(),
    To: today()
})
```

### Client-side tracking

You can also track visitors on the client side by adding `pirsch.js` to your website. It will perform a GET request to the configured endpoint.

```HTML
<!-- add the tracking script to the head area and configure it using attributes -->
<script type="text/javascript" src="js/pirsch.js" id="pirschjs"
        data-endpoint="/count"
        data-tenant-id="42"
        data-track-localhost
        data-param-optional-param="test"></script>
```

The parameters are configured through HTML attributes. All of them are optional, except for the `id`. Here is a list of the possible options.

| Option | Description | Default |
| - | - | - |
| data-endpoint | The endpoint to call. This can be a local path, like /tracking, or a complete URL, like http://mywebsite.com/tracking. It must not contain any parameters. | /pirsch |
| data-tenant-id | The tenant ID to use, in case you plan to track multiple websites using the same backend, or you want to split the data. Note that the tenant ID must be validated in the backend. | 0 (no tenant) |
| data-track-localhost | Enable tracking hits on localhost. This is used for testing purposes only. | false |
| data-param-* | Additional parameters to send with the request. The name send is everything after `data-param-`. | (no parameters) |

To track the hits you need to call `Hit` from the endpoint that you configured for `pirsch.js`. Here is a simple example.

```Go
// Create an endpoint to handle client tracking requests.
// HitOptionsFromRequest is a utility function to process the required parameters.
// You might want to additional checks, like for the tenant ID.
http.Handle("/count", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    tracker.Hit(r, pirsch.HitOptionsFromRequest(r))
}))
```

`HitOptionsFromRequest` will read the parameters send by `pirsch.js` and returns a new `HitOptions` object that can be passed to `Hit`. You might want to split these steps into two, to run additional checks for the parameters that were sent by the user.

## Mapping IPs to countries

Pirsch uses MaxMind's [GeoLite2](https://dev.maxmind.com/geoip/geoip2/geolite2/) database to map IPs to countries. The database **is not included**, so you need to download it yourself. IP mapping is optional, it must explicitly be enabled by setting the GeoDB attribute of the `TrackerConfig` or through the `HitOptions` when calling `HitFromRequest`.

1. create an account at MaxMind
2. generate a new license key
3. call `GetGeoLite2` with the path you would like to extract the tarball to and pass your license key
4. create a new GeoDB by using `NewGeoDB` and the file you downloaded and extracted using the step before

The GeoDB should be updated on a regular basis. The Tracker has a method `SetGeoDB` to update the GeoDB at runtime (thread-safe).

## Documentation

Read the [full documentation](https://godoc.org/github.com/pirsch-analytics/pirsch) for details, check out `demos`, or read the article at https://marvinblum.de/blog/server-side-tracking-without-cookies-in-go-OxdzmGZ1Bl.

## Build pirsch.js

To minify `pirsch.js` to `pirsch.min.js` you need to run `npm i` and `npm run minify` inside the `js` directory.

## Changelog

See [CHANGELOG.md](CHANGELOG.md).

## Contribution

Contributions are welcome! Please open a pull requests for your changes and tickets in case you would like to discuss something or have a question.

To run the tests you'll need a Postgres database, and a schema called `pirsch`. The user and password are set to `postgres`.

Note that we only accept pull requests if you transfer the ownership of your contribution to us. As we also offer a managed commercial solution with this library at its core, we want to make sure we can keep control over the source code.

## License

GNU AGPLv3

<p align="center">
    <img src="gopher.svg" width="200px" />
</p>
