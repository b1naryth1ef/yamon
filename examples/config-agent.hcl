target = "http://name:key@hostname:6691"

// systemd journal support
journal {
  enabled = true

  // drop log messages from any services in this array
  ignored_services = ["audit"]

  // cursors reduce the duplicate log entries that may get sent if the yamon agent crashes 
  cursor_path = "/var/opt/yamon-journal-cursor.txt"
  cursor_sync = 128
}

// we can completely disable unwanted collectors
collector "gpu" {
  disabled = true
}

// we can configure the interval at which collectors run
collector "apt" {
  interval = "5m"
}

// the http server provides access to the agent api
http {
  bind = "localhost:9877"
}

// we can use the log_file directive to include log lines from regular files
log_file "/var/log/nginx/access.log" {
  service = "nginx"
  level   = "info"
}
log_file "/var/log/nginx/error.log" {
  service = "nginx"
  level   = "error"
}
log_file "/var/log/postgresql/postgresql-12-main.log" {
  service = "postgres"
  level   = "info"
}

// we can also scrape prometheus endpoints, in this example we're scraping the yamon server
prometheus {
  url      = "http://localhost:6691/metrics"
  interval = "15s"
  tags = {
    service = "yamon"
  }
}

// we can also just run scripts on disk
script "/etc/yamon/qbittorrent.ts" {
  env = { "QBITTORRENT_HOST" : "my-host:9989" }

  // we can pass whatever we need as args
  // args = ["--example", "--argument", "passing"]

  // set an interval for how often we want to run the script to collect data
  interval = "30s"

  // config a timeout (should generally be lower than your interval)
  timeout = "20s"

  // we could set STREAMING=1 above and enable this mode to have the script run and stream data from stdout
  // streaming = true
}