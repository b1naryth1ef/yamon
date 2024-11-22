package common

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

type ServerConfig struct {
	Bind       string                 `hcl:"bind"`
	Clickhouse ServerClickhouseConfig `hcl:"clickhouse,block"`

	Keys map[string]string `hcl:"keys,optional"`
}

type ServerClickhouseConfig struct {
	Targets  []string `hcl:"targets"`
	Database string   `hcl:"database,optional"`
	Username string   `hcl:"username,optional"`
	Password string   `hcl:"password,optional"`
}

type DaemonConfig struct {
	Target             string                    `hcl:"target"`
	Collectors         []string                  `hcl:"collectors,optional"`
	DisabledCollectors []string                  `hcl:"disabled_collectors,optional"`
	Prometheus         []PrometheusScraperConfig `hcl:"prometheus,block"`
	LogFile            []LogFileBlock            `hcl:"log_file,block"`
	Scripts            []DaemonScriptConfig      `hcl:"script,block"`
	Journal            *DaemonJournalConfig      `hcl:"journal,block"`
	HTTP               *DaemonHTTPConfig         `hcl:"http,block"`
}

type DaemonScriptConfig struct {
	Path     string            `hcl:"path,label"`
	Interval string            `hcl:"interval,optional"`
	Args     []string          `hcl:"args,optional"`
	Env      map[string]string `hcl:"env,optional"`
}

type DaemonHTTPConfig struct {
	Bind string `hcl:"bind"`
}

type DaemonJournalConfig struct {
	Enabled         bool     `hcl:"enabled"`
	CursorPath      string   `hcl:"cursor_path,optional"`
	CursorSync      int      `hcl:"cursor_sync,optional"`
	IgnoredServices []string `hcl:"ignored_services,optional"`
}

type LogFileBlock struct {
	Path    string `hcl:"path,label"`
	Service string `hcl:"service,optional"`
	Level   string `hcl:"level,optional"`
	Format  string `hcl:"format,optional"`
}

type PrometheusScraperConfig struct {
	URL      string            `hcl:"url"`
	Interval string            `hcl:"interval"`
	Timeout  string            `hcl:"timeout,optional"`
	Prefix   string            `hcl:"prefix,optional"`
	Tags     map[string]string `hcl:"tags,optional"`
}

func newHCLEvalContext() *hcl.EvalContext {
	return &hcl.EvalContext{
		Variables: map[string]cty.Value{},
		Functions: map[string]function.Function{},
	}
}

func LoadDaemonConfig(path string) (*DaemonConfig, error) {
	var cfg DaemonConfig
	evalCtx := newHCLEvalContext()
	err := hclsimple.DecodeFile(path, evalCtx, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func LoadServerConfig(path string) (*ServerConfig, error) {
	var cfg ServerConfig
	evalCtx := newHCLEvalContext()
	err := hclsimple.DecodeFile(path, evalCtx, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
