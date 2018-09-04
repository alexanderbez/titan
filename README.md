# Titan

[![GoDoc](https://godoc.org/github.com/alexanderbez/titan?status.svg)](https://godoc.org/github.com/alexanderbez/titan)
[![Build Status](https://travis-ci.org/alexanderbez/titan.svg?branch=master)](https://travis-ci.org/alexanderbez/titan)
[![Go Report Card](https://goreportcard.com/badge/github.com/alexanderbez/titan)](https://goreportcard.com/report/github.com/alexanderbez/titan)

A lightweight and simple Cosmos network validator monitoring and alerting tool.
The primary goal of Titan is to trigger simple configurable alerts when monitored
events occur. These events include global(non-validator) events such as new
governance proposals and governance proposals that have transitioned into a voting
stage. In addition, Titan will track when a specific validator(s) misses blocks,
changes in power, and gets slashed.

Titan aims to be a minimal utility ran as a daemon alongside a validator. It uses
[BadgerDB](https://github.com/dgraph-io/badger) as an embedded key/value store
and [SendGrid](https://sendgrid.com/) for alerting email and SMS messages.

## TODO

- [ ] Implement unit tests...
- [ ] Integrate simple RPC service to get certain metrics and metadata
- [ ] Implement Makefile
- [ ] Implement remaining monitor events
- [ ] Integrate gometalinter and TravisCI config
- [ ] Release!

Nice-to-have down the line:

- [ ] Allow more flexible alerting targets and filters

## Build & Usage

TODO

### Example Configuration

TODO

## Tests

TODO

## Contributing

1. [Fork it](https://github.com/alexanderbez/titan/fork)
2. Create your feature branch (`git checkout -b feature/my-new-feature`)
3. Commit your changes (`git commit -m 'Add some feature'`)
4. Push to the branch (`git push origin feature/my-new-feature`)
5. Create a new Pull Request
