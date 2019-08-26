# Pact Broker Exporter

Prometheus exporter for [Pact Broker](https://github.com/pact-foundation/pact_broker).

[![CircleCI](https://circleci.com/gh/ContainerSolutions/pactbroker_exporter.svg?style=svg)](https://circleci.com/gh/ContainerSolutions/pactbroker_exporter)
[![Docker Pulls](https://img.shields.io/docker/pulls/containersol/pactbroker_exporter.svg)](https://hub.docker.com/r/containersol/pactbroker_exporter/tags)

## Quick Start

This package is available for Docker:

1. Example Pact Broker setup you can find in [pperzyna/pact-example](https://github.com/pperzyna/pact-example) repository.

2. Run Pact Broker Exporter

```bash
docker run --net=host -e PB_EXPORTER_PACTBROKER_URI="http://localhost:9292" containersol/pactbroker_exporter
```

## Building and running

The default way to build is:

```bash
go get github.com/ContainerSolutions/pactbroker_exporter
cd ${GOPATH-$HOME/go}/src/github.com/ContainerSolutions/pactbroker_exporter/
go build -o pactbroker_exporter
export PB_EXPORTER_PACTBROKER_URI="http://localhost:9292"
./pactbroker_exporter <flags>
```

### Flags

* `pactbroker.uri`
  Address of Pact Broker. Default is `http://localhost:9292`.

* `pactbroker.timeout`
  Timeout request to Pact Broker. Default is `5s`.

* `web.listen-address`
  Address to listen on for web interface and telemetry. Default is `:9624`.

* `web.telemetry-path`
  Path under which to expose metrics. Default is `/metrics`.

* `log.level`
  Set logging level: one of `debug`, `info`, `warn`, `error`, `fatal`

* `log.format`
  Set the log output target and format. e.g. `logger:syslog?appname=bob&local=7` or `logger:stdout?json=true`
  Defaults to `logger:stderr`.

### Environment Variables

The following environment variables configure the exporter:

* `PB_EXPORTER_PACTBROKER_URI`
  Address of Pact Broker. Default is `http://localhost:9292`.

* `PB_EXPORTER_PACTBROKER_TIMEOUT`
  Timeout reqeust to Pact Broker. Default is `5s`.

* `PB_EXPORTER_WEB_LISTEN_ADDRESS`
  Address to listen on for web interface and telemetry. Default is `:9624`.

* `PB_EXPORTER_WEB_TELEMETRY_PATH`
  Path under which to expose metrics. Default is `/metrics`.

Settings set by environment variables starting with `PB_` will be overwritten by the corresponding CLI flag if given.
