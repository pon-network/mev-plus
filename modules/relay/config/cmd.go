package config

import (
	"github.com/bsn-eng/mev-plus/cmd/utils"
	cli "github.com/urfave/cli/v2"
)

const ModuleName = "relay"

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      ModuleName,
		Usage:     "Start the relay module",
		UsageText: "The relay module is a service that connects the builder api to external relays",
		Category:  utils.RelayModuleCategory,
		Flags:     relayModulFlags(),
	}
}

func relayModulFlags() []cli.Flag {
	return []cli.Flag{
		LoggerLevelFlag,
		LoggerFormatFlag,
		RelayEntriesFlag,
		RelayCheckFlag,
		SkipRelaySignatureCheck,
		MainnetFlag,
		SepoliaFlag,
		GoerliFlag,
		MinBidFlag,
		GenesisForkVersionFlag,
		GenesisValidatorsRootFlag,
		GenesisTimeFlag,
		RequestTimeoutMsFlag,
		RequestMaxRetriesFlag,
	}
}
