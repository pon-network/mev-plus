package config

import (
	"github.com/urfave/cli/v2"

	aggregator "github.com/pon-network/mev-plus/modules/block-aggregator/config"
	builderApi "github.com/pon-network/mev-plus/modules/builder-api/config"
	proxyModule "github.com/pon-network/mev-plus/modules/external-validator-proxy"
	relay "github.com/pon-network/mev-plus/modules/relay/config"
)

// Core does not have any extra commands,
// this is to load default module commands that come
// pre=pacakged with the core

// CoreFlags are the flags that are used by the core in flags.go

var DefaulModulesCommands []*cli.Command

func init() {
	DefaulModulesCommands = []*cli.Command{
		builderApi.NewCommand(),
		relay.NewCommand(),
		aggregator.NewCommand(),
		proxyModule.NewCommand(),
	}
}
