package config

import (
	"github.com/bsn-eng/mev-plus/cmd/utils"
	cli "github.com/urfave/cli/v2"
)

var (
	LoggerLevelFlag = &cli.StringFlag{
		Name:     ModuleName + "." + "logger-level",
		Usage:    "Set the logger level",
		Category: utils.RelayModuleCategory,
		Value:    "info",
		EnvVars:  []string{"RELAY_LOGGER_LEVEL"},
	}

	LoggerFormatFlag = &cli.StringFlag{
		Name:     ModuleName + "." + "logger-format",
		Usage:    "Set the logger format",
		Category: utils.RelayModuleCategory,
		Value:    "text",
		EnvVars:  []string{"RELAY_LOGGER_FORMAT"},
	}

	RelayEntriesFlag = &cli.StringFlag{
		Name:     ModuleName + "." + "relay-entries",
		Usage:    "Set the relay entries",
		Category: utils.RelayModuleCategory,
		Value:    "text",
	}

	RelayCheckFlag = &cli.BoolFlag{
		Name:     ModuleName + "." + "relay-check",
		Usage:    "On status check, check the status of all relays",
		Category: utils.RelayModuleCategory,
		Value:    true,
	}

	SkipRelaySignatureCheck = &cli.BoolFlag{
		Name:     ModuleName + "." + "skip-relay-signature-check",
		Usage:    "Skip the relay signature check",
		Category: utils.RelayModuleCategory,
		Value:    true,
	}

	MainnetFlag = &cli.BoolFlag{
		Name:     ModuleName + "." + "mainnet",
		Usage:    "Set the network to mainnet",
		Category: utils.RelayModuleCategory,
		Value:    false,
	}

	SepoliaFlag = &cli.BoolFlag{
		Name:     ModuleName + "." + "sepolia",
		Usage:    "Set the network to sepolia",
		Category: utils.RelayModuleCategory,
		Value:    false,
	}

	GoerliFlag = &cli.BoolFlag{
		Name:     ModuleName + "." + "goerli",
		Usage:    "Set the network to goerli",
		Category: utils.RelayModuleCategory,
		Value:    false,
	}

	MinBidFlag = &cli.StringFlag{
		Name:     ModuleName + "." + "min-bid",
		Usage:    "Set the minimum bid",
		Category: utils.RelayModuleCategory,
		Value:    "0",
	}

	GenesisForkVersionFlag = &cli.StringFlag{
		Name:     ModuleName + "." + "genesis-fork-version",
		Usage:    "Set a custom fork version",
		Category: utils.RelayModuleCategory,
		Value:    RelayConfigDefaults.GenesisForkVersion,
	}

	GenesisValidatorsRootFlag = &cli.StringFlag{
		Name:     ModuleName + "." + "genesis-validators-root",
		Usage:    "Set a custom genesis validators root",
		Category: utils.RelayModuleCategory,
		Value:    RelayConfigDefaults.GenesisValidatorsRoot,
	}

	GenesisTimeFlag = &cli.IntFlag{
		Name:     ModuleName + "." + "genesis-time",
		Usage:    "Set a custom genesis time",
		Category: utils.RelayModuleCategory,
		Value:    int(RelayConfigDefaults.GenesisTime),
	}

	RequestTimeoutMsFlag = &cli.IntFlag{
		Name:     ModuleName + "." + "request-timeout-ms",
		Usage:    "Set the request timeout in milliseconds",
		Category: utils.RelayModuleCategory,
		Value:    RelayConfigDefaults.RequestTimeoutMs,
	}

	RequestMaxRetriesFlag = &cli.IntFlag{
		Name:     ModuleName + "." + "request-max-retries",
		Usage:    "Set the request max retries",
		Category: utils.RelayModuleCategory,
		Value:    RelayConfigDefaults.RequestMaxRetries,
	}
)
