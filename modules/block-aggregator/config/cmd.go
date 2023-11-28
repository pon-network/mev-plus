package config

import (
	"github.com/pon-network/mev-plus/cmd/utils"
	cli "github.com/urfave/cli/v2"
)

const ModuleName = "blockAggregator"

func NewCommand() *cli.Command {
	return &cli.Command{
		Name:      ModuleName,
		Usage:     "Start the block aggregator module",
		UsageText: "The block aggregator module is a service that provides handler functions for builder APIs",
		Category:  utils.BlockAggregatorCategory,
		Flags:     blockAggregatorFlags(),
	}
}

func blockAggregatorFlags() []cli.Flag {
	return []cli.Flag{
		GenesisTimeFlag,
		AuctionDurationFlag,
		SlotDurationFlag,
	}
}
