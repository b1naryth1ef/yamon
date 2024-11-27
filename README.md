# yamon

yamon is a lightweight set of tools for collecting and ingesting monitoring data
into [ClickHouse](https://clickhouse.com/). It's designed to ingest metrics,
logs, and events in a variety of formats.

## Features

- **collection agent** is lightweight, extendable, and can be compiled for
  numerous targets
- **powerful storage** via ClickHouse which can easily handle trillions of data
  points
- **prometheus** support for easy integration into your existing services and
  tools
- **journald** collect all your system logging data for querying and analysis

## Installation

- build it locally (`just build`)
- [docker](/Dockerfile) just provide a `config.hcl` file for the agent or server

### Configuration

As a starting point take a look at the example
[agent config](./examples/config-agent.hcl) and
[server config](./examples/config-server.hcl).

### Custom Scripts

yamon supports generating data from custom scripts. These scripts can be
anything exec-able that produce JSON data in a specified format. The intention
is to allow using scripting languages for generating ad-hoc or domain specific
observability data. For an example of the expected format see the
[qbittorrent](./examples/qbittorrent.ts) script.
