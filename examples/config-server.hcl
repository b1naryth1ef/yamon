bind = "0.0.0.0:6691"
keys = { "client" : "some-secure-key" }

clickhouse {
  targets  = ["clickhouse-host.local:9000"]
  database = "yamon"
}