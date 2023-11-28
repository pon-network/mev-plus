package utils

import "github.com/urfave/cli/v2"

const (
	MiscCategory = "MISC"
	CoreCategory = "CORE"

	BuilderAPICategory      = "BUILDER API"
	RelayModuleCategory     = "Relay Module"
	BlockAggregatorCategory = "Block Aggregator"
	ExternalValidatorProxyCategory = "External Validator Proxy"
)

func init() {
	// Set the default categories for the help and version flags
	cli.HelpFlag.(*cli.BoolFlag).Category = MiscCategory
	cli.VersionFlag.(*cli.BoolFlag).Category = MiscCategory
}
