package modulelist

import (
	"github.com/pon-network/mev-plus/common"
	coreCommon "github.com/pon-network/mev-plus/core/common"
	"github.com/urfave/cli/v2"

	proxyModule "github.com/pon-network/mev-plus/modules/external-validator-proxy"
)

var ServiceList []coreCommon.Service
var CommandList []*cli.Command

func init() {

	///////////////////////////////////////////////////
	// To attach your module to the MEV+ application //
	// you must import your service and command      //
	// Import and append your service struct here    //
	///////////////////////////////////////////////////
	ServiceList = []coreCommon.Service{

		proxyModule.NewExternalValidatorProxyService(),
	}
	///////////////////////////////////////////////////

	///////////////////////////////////////////////////
	// Import and append your command  here          //
	///////////////////////////////////////////////////
	commandList := []*cli.Command{

		proxyModule.NewCommand(),
	}
	////////////////////////////////////////////////

	var err error
	CommandList, err = common.FormatCommands(commandList)
	if err != nil {
		panic(err)
	}

}
