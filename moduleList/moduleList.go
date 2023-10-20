package modulelist

import (
	"github.com/bsn-eng/mev-plus/common"
	"github.com/urfave/cli/v2"
	coreCommon "github.com/bsn-eng/mev-plus/core/common"
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

		// Test service
		// NewTestService(),
	}
	///////////////////////////////////////////////////

	///////////////////////////////////////////////////
	// Import and append your command  here          //
	///////////////////////////////////////////////////
	commandList := []*cli.Command{
		// NewTestCommand(),
	}
	////////////////////////////////////////////////

	var err error
	CommandList, err = common.FormatCommands(commandList)
	if err != nil {
		panic(err)
	}

}
