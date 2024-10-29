# yamon

yamon is a lightweight set of tools for collecting and ingesting monitoring data into [ClickHouse](https://clickhouse.com/). It's designed to ingest metrics, logs, and events in a variety of formats.

## Features

- **collection agent** is lightweight, extendable, and can be compiled for numerous targets
- **powerful storage** via ClickHouse which can easily handle trillions of data points
- **prometheus** support for easy integration into your existing services and tools
- **journald** collect all your system logging data for querying and analysis

## Installation

- build it locally (`just build`)
- [docker](/Dockerfile) just provide a `config.hcl` file for the agent or server

### Agent Configuration

```hcl
target = "http://client:key@my-yamon-server:6691"

// Journal enables processing and forwarding of journald entries
journal {
  enabled = true

  // the cursor will ensure all logs get synced even across agent restarts
  cursor_path = "/var/opt/yamon-journal-cursor.txt"

  // how many log entries to forward before fsyncing the cursor file
  cursor_sync = 128
}

// Collect prometheus metrics from ClickHouse
prometheus {
  url      = "http://localhost:9363/metrics"
  interval = "60s"
}

// Collect prometheus metrics from the Yamon Server
prometheus {
  url      = "http://localhost:6691/metrics"
  interval = "15s"
}

// Collect log data from a file on disk
log_file "/var/log/my_service/example.log" {
  service = "my_service"
  level   = "info"
}
```

### Server Configuration

```hcl
bind       = "0.0.0.0:6691"
keys       = { "client" : "key" }

clickhouse {
  targets  = ["localhost:9000"]
  database = "yamon"
}
```
