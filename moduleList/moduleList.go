package modulelist

import (
	"github.com/pon-network/mev-plus/common"
	coreCommon "github.com/pon-network/mev-plus/core/common"

	k2 "github.com/restaking-cloud/native-delegation-for-plus"
	"github.com/urfave/cli/v2"
)

var ServiceList []coreCommon.Service
var CommandList []*cli.Command

func init() {

	///////////////////////////////////////////////////
	// To attach your module to the MEV Plus application //
	// you must import your service and command      //
	// Import and append your service struct here    //
	///////////////////////////////////////////////////
	ServiceList = []coreCommon.Service{k2.NewK2Service()}

	var commandList []*cli.Command
	for _, service := range ServiceList {
		commandList = append(commandList, service.CliCommand())
	}
	var err error
	CommandList, err = common.FormatCommands(commandList)
	if err != nil {
		panic(err)
	}

}
