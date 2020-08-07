<p align="center">
    <img src="gopher.svg" width="200px" />
</p>

# Pirsch

[![GoDoc](https://godoc.org/github.com/emvi/pirsch?status.svg)](https://godoc.org/github.com/emvi/pirsch)
[![Go Report Card](https://goreportcard.com/badge/github.com/emvi/pirsch)](https://goreportcard.com/report/github.com/emvi/pirsch)
<a href="https://discord.gg/fAYm4Cz"><img src="https://img.shields.io/discord/739184135649886288?logo=discord" alt="Chat on Discord"></a>

**State of the project**: we are currently testing how precise Pirsch is by comparing it to other solutions.

Pirsch is a server side, no-cookie, drop-in and privacy focused tracking solution for Go. Integrated into a Go application it enables you to track HTTP traffic without invading the privacy of your visitors. The visualization of the data must be implemented by yourself. We might add a UI for Pirsch in the future.

The name is in German and refers to a special kind of hunt: *the hunter carefully and quietly enters the area to be hunted, he stalks against the wind in order to get as close as possible to the prey without being noticed.*

## How does it work?

Pirsch generates a unique fingerprint for each visitor. The fingerprint is a hash of the visitors IP, User-Agent and a salt. The salt is re-generated at midnight to separate data for each day.

Each time a visitor opens your page, Pirsch will store a hit. The hits are analyzed later to extract meaningful data and reduce storage usage.

This all works without invading the visitors privacy, as no cookies are used and individual users cannot be tracked, as the fingerprint does anonymize the data points. At the same time Pirsch can track visitors using blockers that otherwise would block tracking. uBlock blocks Google Analytics for example.

## Features and limitations

Pirsch tracks the following data points at the moment:

* total visitors per day
* visitors per day and hour
* visitors per day and page
* visitors per day and language

All timestamps are stored as UTC.

It's theoretically possible to track the visitor flow (which page was seen first, which one was opened next and so one), but this is not implemented at the moment. Here is a list of the limitations of Pirsch:

* track sessions, as the salt will prevent you from tracking a user across two days
* bots might not always be filtered out
* rare cases where two fingerprints collide, if two visitors are behind the same proxy for example and the User-Agent is the same (or empty)

## Usage

To store hits and statistics, Pirsch uses a database. Right now only Postgres is supported, but new ones can easily be integrated by implementing the Store interface. The schema can be found within the schema directory. Changes will be added to migrations scripts, so that you can add them to your projects database migration or run them manually.

Here is a quick demo on how to use the library:

```Go
// create a Postgres store using the sql.DB database connection "db"
store := pirsch.NewPostgresStore(db)

// Tracker is the main component of Pirsch
// the salt is used to prevent anyone from generating fingerprints like yours (to prevent man in the middle attacks), pick something random
// an optional configuration can be used to change things like worker count, timeouts and so on
tracker := pirsch.NewTracker(store, "secret_salt", nil)

// the Processor analyzes hits and stores the reduced data points in store
// it's recommended to run the Process method once a day
processor := pirsch.NewProcessor(store)
pirsch.RunAtMidnight(processor.Process) // helper function to run a function at midnight

http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // a call to Hit will track the request
    // note that Pirsch stores the path and URL, therefor you should make sure you only call it for the endpoints you're interersted in
    if r.URL.Path == "/" {
        tracker.Hit(r)
    }

    w.Write([]byte("<h1>Hello World!</h1>"))
}))
```

Instead of calling `Hit`, you can also call `HitPage`, which allows you to specify an alternative path independent of the one provided in the request.
That can be used to implement a single tracking endpoint which you call using an Ajax request providing the path of the current page.

To analyze hits and processed data you can use the analyzer, which provides some functions to extract useful data.

The secret salt passed to `NewTracker` should not be known outside of your organization, as it can be used to generate fingerprints that are like yours.
This is used to prevent people from outside your organization to track your visitors and gain data from it.
Note that while you can generate the salt at random, the fingerprints will change too. To get reliable data configure a fixed salt and treat it like a password.

```Go
// this also needs access to the store
analyzer := pirsch.NewAnalyzer(store)

// as an example, lets extract the total number of visitors
// the filter is used to specify the time frame you're looking at (days) and is optional
// if you pass nil, the Analyzer returns data for the past week including today
visitors, err := analyzer.Visitors(&pirsch.Filter{
    From: yesterday(),
    To: today()
})
```

Read the full documentation for more details or check out [this](https://marvinblum.de/blog/how-i-built-my-website-using-emvi-as-a-headless-cms-RGaqOqK18w) article.

## Changelog

### 1.1.1

* fixed error in case values are too long
* fixed language statistics not including today

### 1.1.0

* added a secret salt to prevent generating fingerprints to identify visitors on other websites (man in the middle)
* extended bot list

### 1.0.0

Initial release.

## Contribution

All contributions are welcome! You can extend the bot list or processor for example, to extract more useful data. Please open a pull requests for your changes and tickets in case you would like to discuss something or have a question.

To run the tests you'll need a Postgres database and a schema called `pirsch`. The user and password is set to `postgres`. To add another data store, the Store interface must be implemented. Pirsch makes heavy use of SQL to aggregate and analyze data.

## License

MIT
