package config

import (
	"github.com/pon-network/mev-plus/core/version"
	"github.com/urfave/cli/v2"
)

// NewApp creates a default MEV Plus CLI app.
func NewApp(usage string) *cli.App {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Version = version.Info()
	app.Usage = usage
	app.Copyright = "Copyright (c) 2023 Abstract Systems SEZC"
	app.Authors = []*cli.Author{}
	return app
}
