package main

import (
	"os"

	"github.com/b1naryth1ef/yamon"
	"github.com/b1naryth1ef/yamon/common"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.hcl"
	}

	config, err := common.LoadServerConfig(configPath)
	if err != nil {
		panic(err)
	}

	destination := yamon.NewClickhouseWriter(config.Clickhouse)
	go destination.Run()

	server := yamon.NewForwardServer(destination, config.Keys)
	server.Run(config.Bind)
}
