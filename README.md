# Pact Broker Exporter

Prometheus exporter for [Pact Broker](https://github.com/pact-foundation/pact_broker).

[![Build Status](https://travis-ci.org/pperzyna/pactbroker_exporter.svg?branch=master)](https://travis-ci.org/pperzyna/pactbroker_exporter)
[![Docker Pulls](https://img.shields.io/docker/pulls/pperzyna/pactbroker_exporter.svg)](https://hub.docker.com/r/pperzyna/pactbroker_exporter/tags)

## Quick Start

This package is available for Docker:

1. Example Pact Broker setup you can find in [pperzyna/pact-example](https://github.com/pperzyna/pact-example) repository.

2. Run Pact Broker Exporter

```bash
docker run --net=host -e DATA_SOURCE_NAME="http://localhost:9292" pperzyna/pactbroker_exporter
```

## Building and running

The default way to build is:

```bash
go get github.com/pperzyna/pactbroker_exporter
cd ${GOPATH-$HOME/go}/src/github.com/pperzyna/pactbroker_exporter/src/
go build -o pactbroker_exporter
export DATA_SOURCE_NAME="http://localhost:9292"
./pactbroker_exporter <flags>
```

### Flags

* `web.listen-address`
  Address to listen on for web interface and telemetry. Default is `:9623`.

* `web.telemetry-path`
  Path under which to expose metrics. Default is `/metrics`.

* `data-source-name`
  Address of Pact Broker. Default is `http://localhost:9292`.

### Environment Variables

The following environment variables configure the exporter:

* `DATA_SOURCE_NAME`
  Address of Pact Broker

* `PB_EXPORTER_WEB_LISTEN_ADDRESS`
  Address to listen on for web interface and telemetry. Default is `:9187`.

* `PB_EXPORTER_WEB_TELEMETRY_PATH`
  Path under which to expose metrics. Default is `/metrics`.

Settings set by environment variables starting with `PB_` will be overwritten by the corresponding CLI flag if given.
