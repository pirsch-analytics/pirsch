<p align="center">
    <img src="logo.svg" alt="Pirsch Logo" width="78px" />
</p>

# Pirsch Analytics

[![Go Reference](https://pkg.go.dev/badge/github.com/pirsch-analytics/pirsch?status.svg)](https://pkg.go.dev/github.com/pirsch-analytics/pirsch/v5)
[![Go Report Card](https://goreportcard.com/badge/github.com/pirsch-analytics/pirsch/v5)](https://goreportcard.com/report/github.com/pirsch-analytics/pirsch/v5)
<a href="https://discord.gg/fAYm4Cz"><img src="https://img.shields.io/discord/739184135649886288?logo=discord" alt="Chat on Discord"></a>

This is the open-source core of [Pirsch Analytics](https://pirsch.io), a privacy-friendly web analytics solution. It originally started as an [experiment](https://marvinblum.de/blog/server-side-tracking-without-cookies-in-go-OxdzmGZ1Bl) to reliably analyze web traffic from the server-side using Go.

Pirsch is made in the EU 🇪🇺 and hosted on german 🇩🇪 owned servers at [Hetzner](https://www.hetzner.com/). You can find an interactive demo of what the dashboard looks like today [here](https://pirsch.pirsch.io).

## How does it work?

Pirsch generates a unique fingerprint for each visitor. The fingerprint is a hash of the visitors IP address, User-Agent, the date, and a salt.  The tracking works without invading the visitor's privacy. It doesn't use cookies and no personal information is stored, making it GDPR-, CCPA-, and PECR-compliant. If used on the server-side, Pirsch can track visitors using ad blockers.

[Learn more about privacy on our documentation.](https://docs.pirsch.io/privacy)

## Documentation

You can find our documentation [here](https://docs.pirsch.io). The code reference can be found on [go.dev](https://pkg.go.dev/github.com/pirsch-analytics/pirsch/v5).

## Contributions

Contributions are welcome! Please open a pull requests for your changes and tickets in case you would like to discuss something or have a question.

To run the tests you'll need a ClickHouse database and a schema called `pirschtest`. The user is set to `default` (no password).

Note that we only accept pull requests if you transfer the ownership of your contribution to us. As we also offer a managed commercial solution with this library at its core, we want to make sure we can keep control over the source code.

## License

GNU AGPLv3
