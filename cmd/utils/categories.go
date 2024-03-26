package utils

import "github.com/urfave/cli/v2"

const (
	MiscCategory = "MISC"
	CoreCategory = "CORE"

	BuilderAPICategory      = "BUILDER API"
	RelayModuleCategory     = "RELAY MODULE"
	BlockAggregatorCategory = "BLOCK AGGREGATOR"
	ExternalValidatorProxyCategory = "EXTERNAL VALIDATOR PROXY"
)

func init() {
	// Set the default categories for the help and version flags
	cli.HelpFlag.(*cli.BoolFlag).Category = MiscCategory
	cli.VersionFlag.(*cli.BoolFlag).Category = MiscCategory
}
