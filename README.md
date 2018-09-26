# Titan

[![GoDoc](https://godoc.org/github.com/alexanderbez/titan?status.svg)](https://godoc.org/github.com/alexanderbez/titan)
[![Build Status](https://travis-ci.org/alexanderbez/titan.svg?branch=master)](https://travis-ci.org/alexanderbez/titan)
[![Go Report Card](https://goreportcard.com/badge/github.com/alexanderbez/titan)](https://goreportcard.com/report/github.com/alexanderbez/titan)

A lightweight and simple Cosmos network validator monitoring and alerting tool.
The primary goal of Titan is to trigger simple configurable alerts when monitored
events occur. These events include global (non validator specific) events such as new
governance proposals and governance proposals that have transitioned into a voting
phase. In addition, Titan will track when a specific validator(s) misses signing
(pre-committing) a block, becomes jailed or double signs a block.

Titan aims to be a minimal utility ran as a daemon alongside a validator. It uses
[BadgerDB](https://github.com/dgraph-io/badger) as an embedded key/value store
and [SendGrid](https://sendgrid.com/) for alerting email and SMS messages.

The latest release of Titan currently operates and supports `v0.24.2` of the
[Cosmos SDK](https://github.com/cosmos/cosmos-sdk/) and the
[gaia-8001](https://github.com/cosmos/testnets/tree/master/gaia-8001) testnet.

Nice-to-have features down the line:

- [ ] Allow more flexible alerting targets and filters
- [ ] More granular monitoring of when validators miss a certain % of pre-commits/signatures
- [ ] Monitor when a validator gets slashed
  - NOTE: We can monitor when a validator decreases in power, but this would get
  very noisy.

## Build & Usage

To build the binary, which will get all tools and vendor dependencies:

```shell
$ make
```

Titan has a simple CLI. To run the binary:

```shell
$ titan --config=path/to/config.toml --out=path/to/log/file --debug=true|false
```

See `$ titan --help` for further usage.

Note:

- Titan by default looks for configuration in `$HOME/.titan/config.toml`.
- Titan operates through a series of provided LCD clients. More than a single
  client should be provided and each client should be up-to-date and trusted.

## API

Titan also exposes a very simple JSON REST service exposing information on the
latest monitor execution. This service is exposed on `listen_addr` and has a single
endpoint of: `executions/latest`.

## Example Configuration

See `config/template.go` for the full configuration template.

```toml
poll_interval = 15

monitors = [
  "new_proposals",
  "active_proposals",
  "jailed_validators",
  "double_signing",
  "missing_signatures"
  # or you can simply pass "*" to enable all monitors
]

[database]
data_dir = "/path/to/.titan/data"

[network]
# Address to run JSON REST service
listen_addr = "0.0.0.0:36655"

# NOTE: These will be used in a round-robin fashion
clients = ["https://gaia-seeds.interblock.io:1317"]

[targets]
sms_recipients = ["+11234567890"]
email_recipients = ["foo@bar.com"]

[filters]
  [filters.validator]
    operator = "cosmosaccaddr1rvm0em6w3qkzcwnzf9hkqvksujl895dfww4ecn"
    address = "EBC613967F66F4EC306852CDF58B4F151CF16738"

[integrations]
  [integrations.sendgrid]
    api_key = "your-API-key"
    from_name = "Cosmos Titan"
```

## Tests

To run unit tests and linting:

```shell
$ make test
```

## Contributing

1. [Fork it](https://github.com/alexanderbez/titan/fork)
2. Create your feature branch (`git checkout -b feature/my-new-feature`)
3. Commit your changes (`git commit -m 'Add some feature'`)
4. Push to the branch (`git push origin feature/my-new-feature`)
5. Create a new Pull Request
