package common

const (
	// parked default modules names that should not be
	builderApiModuleName = "builderApi"
	relayModuleName      = "relay"
	blockAggregatorName  = "blockAggregator"
	proxyModuleName      = "externalValidatorProxy"
)

var DefaultModuleNames = []string{
	builderApiModuleName,
	relayModuleName,
	blockAggregatorName,
	proxyModuleName,
}

// JSON RPC constants
const (
	Vsn                    = "1.0"
	ServiceMethodSeparator = "_"
	ResponseMethodSuffix   = "_response"
)
