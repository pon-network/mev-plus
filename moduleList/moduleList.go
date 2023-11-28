package modulelist

import (
	"github.com/pon-network/mev-plus/common"
	coreCommon "github.com/pon-network/mev-plus/core/common"
	"github.com/urfave/cli/v2"

	proxyModule "github.com/pon-network/mev-plus/modules/external-validator-proxy"
	k2 "github.com/restaking-cloud/native-delegation-for-plus"
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
		k2.NewK2Service(),
	}
	///////////////////////////////////////////////////

	///////////////////////////////////////////////////
	// Import and append your command  here          //
	///////////////////////////////////////////////////
	commandList := []*cli.Command{

		proxyModule.NewCommand(),
		k2.NewCommand(),
	}
	////////////////////////////////////////////////

	var err error
	CommandList, err = common.FormatCommands(commandList)
	if err != nil {
		panic(err)
	}

}
