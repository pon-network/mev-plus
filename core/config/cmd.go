package config

import (
	"github.com/urfave/cli/v2"

	coreCommon "github.com/pon-network/mev-plus/core/common"

	aggregator "github.com/pon-network/mev-plus/modules/block-aggregator"
	builderApi "github.com/pon-network/mev-plus/modules/builder-api"
	proxyModule "github.com/pon-network/mev-plus/modules/external-validator-proxy"
	relay "github.com/pon-network/mev-plus/modules/relay"

	// Additional MEV Plus Functionalities
	"github.com/pon-network/mev-plus/core/additionalFeatures/moduleManagement"
)

var DefaultModules []coreCommon.Service

// AdditionalFunctionalities is a list of commands that cause the software to behave differently
var AdditionalFunctionalities = []*cli.Command{}

func init() {
	DefaultModules = []coreCommon.Service{
		builderApi.NewBuilderApiService(),
		relay.NewRelayService(),
		aggregator.NewBlockAggregatorService(),
		proxyModule.NewExternalValidatorProxyService(),
	}

	AdditionalFunctionalities = []*cli.Command{
		moduleManagement.NewCommand(),
	}

}
