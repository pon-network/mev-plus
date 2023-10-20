package common

const (
	// parked default modules names that should not be
	builderApiModuleName = "builderApi"
	relayModuleName      = "relay"
	blockAggregatorName  = "blockAggregator"
)

var DefaultModuleNames = []string{
	builderApiModuleName,
	relayModuleName,
	blockAggregatorName,
}

// JSON RPC constants
const (
	Vsn                    = "1.0"
	ServiceMethodSeparator = "_"
	ResponseMethodSuffix   = "_response"
)
