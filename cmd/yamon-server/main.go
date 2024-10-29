package main

import (
	"github.com/alexflint/go-arg"
	"github.com/b1naryth1ef/yamon"
	"github.com/b1naryth1ef/yamon/clickhouse"
	"github.com/b1naryth1ef/yamon/common"
)

func main() {
	var args struct {
		ConfigPath string `arg:"env:CONFIG_PATH,-c,--config-path" default:"config.hcl"`
		LogLevel   string `arg:"env:LOG_LEVEL,-l,--log-level" default:"info"`
	}
	arg.MustParse(&args)

	yamon.SetupLogging(args.LogLevel)

	config, err := common.LoadServerConfig(args.ConfigPath)
	if err != nil {
		panic(err)
	}

	destination := clickhouse.NewClickhouseWriter(config.Clickhouse)
	go destination.Run()

	server := yamon.NewForwardServer(destination, config.Keys)
	server.Run(config.Bind)
}
