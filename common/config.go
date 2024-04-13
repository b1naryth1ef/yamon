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
	Target     string                   `hcl:"target"`
	Collectors string                   `hcl:"collectors,optional"`
	Prometheus []prometheusScraperBlock `hcl:"prometheus,block"`
	LogFile    []logFileBlock           `hcl:"log_file,block"`
	Journal    *DaemonJournalConfig     `hcl:"journal,block"`
}

type DaemonJournalConfig struct {
	Enabled    bool   `hcl:"enabled"`
	CursorPath string `hcl:"cursor_path,optional"`
	CursorSync int    `hcl:"cursor_sync,optional"`
}

type logFileBlock struct {
	Path    string `hcl:"path,label"`
	Service string `hcl:"service,optional"`
	Level   string `hcl:"level,optional"`
}

type prometheusScraperBlock struct {
	URL      string            `hcl:"url"`
	Interval string            `hcl:"interval"`
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
