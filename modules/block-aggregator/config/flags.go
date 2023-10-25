package config

import (
	"github.com/pon-pbs/mev-plus/cmd/utils"
	cli "github.com/urfave/cli/v2"
)

var (
	GenesisTimeFlag = &cli.IntFlag{
		Name:     ModuleName + "." + "genesis-time",
		Usage:    "Set the genesis time (in seconds)",
		Category: utils.BlockAggregatorCategory,
		Value:    int(BlockAggregatorConfigDefaults.GenesisTime),
	}

	AuctionDurationFlag = &cli.IntFlag{
		Name:     ModuleName + "." + "auction-duration",
		Usage:    "Set the auction duration (in seconds)",
		Category: utils.BlockAggregatorCategory,
		Value:    int(BlockAggregatorConfigDefaults.AuctionDuration),
	}

	SlotDurationFlag = &cli.IntFlag{
		Name:     ModuleName + "." + "slot-duration",
		Usage:    "Set the slot duration (in seconds)",
		Category: utils.BlockAggregatorCategory,
		Value:    int(BlockAggregatorConfigDefaults.SlotDuration),
	}
)
