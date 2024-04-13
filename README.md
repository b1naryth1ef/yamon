# yamon

yamon is a simple but powerful monitoring solution designed for small to medium
size infrastructure deployments. Yamon stores all of its data in Clickhouse and
can be queried using tools like Grafana.

## Features

- fast & configurable storage engine via Clickhouse
- simple and powerful querying via Grafana
- **prometheus** support to scrape metrics
- **journald** support to automatically ingest logs

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

// clickhouse
prometheus {
  url      = "http://localhost:9363/metrics"
  interval = "60s"
}

// yamon-server
prometheus {
  url      = "http://localhost:6691/metrics"
  interval = "15s"
}

// read some log off disk
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
