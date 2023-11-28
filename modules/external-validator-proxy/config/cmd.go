package config

import (
	"github.com/pon-network/mev-plus/cmd/utils"
	cli "github.com/urfave/cli/v2"
)

const ModuleName = "externalValidatorProxy"

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      ModuleName,
		Usage:     "Start the external validator proxy module",
		UsageText: "The external validator proxy module is a service that forwards Builder API requests to and from an attached external proxy",
		Category:  utils.ExternalValidatorProxyCategory,
		Flags:     proxyModuleFlags(),
	}
}

func proxyModuleFlags() []cli.Flag {
	return []cli.Flag{
		LoggerLevelFlag,
		LoggerFormatFlag,
		AddressFlag,
		RequestTimeoutMsFlag,
		RequestMaxRetriesFlag,
	}
}
