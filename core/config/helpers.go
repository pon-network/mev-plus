package config

import (
	"github.com/urfave/cli/v2"
	"github.com/bsn-eng/mev-plus/core/version"
)

// NewApp creates a default MEV+ CLI app.
func NewApp(usage string) *cli.App {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Version = version.Info()
	app.Usage = usage
	app.Copyright = "Copyright 2023 BlockSwap Labs"
	return app
}