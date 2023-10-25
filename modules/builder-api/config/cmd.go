package config

import (
	"github.com/pon-pbs/mev-plus/cmd/utils"
	cli "github.com/urfave/cli/v2"
)

const ModuleName = "builderApi"

func NewCommand() *cli.Command {

	return &cli.Command{
		Name:      ModuleName,
		Usage:     "Start the Builder API",
		UsageText: "The Builder API is a service that provides a REST API to build blocks",
		Category:  utils.BuilderAPICategory,
		Flags:     builderApiFlags(),
	}
}

func builderApiFlags() []cli.Flag {
	return []cli.Flag{
		LoggerLevelFlag,
		LoggerFormatFlag,
		ListenAddressFlag,
		ServerReadTimeoutMsFlag,
		ServerReadHeaderTimeoutMsFlag,
		ServerWriteTimeoutMsFlag,
		ServerIdleTimeoutMsFlag,
		ServerMaxHeaderBytesFlag,
	}
}
