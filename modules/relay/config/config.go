package config

import "github.com/bsn-eng/mev-plus/modules/relay/common"

type RelayConfig struct {
	LoggerLevel        string
	LoggerFormat       string
	RelayCheck         bool
	RelaySignatureCheck bool
	MinBid             common.U256Str
	GenesisTime        uint64
	RequestTimeoutMs   int
	RequestMaxRetries  int
	GenesisForkVersion string
	GenesisValidatorsRoot string
}

var RelayConfigDefaults = RelayConfig{
	LoggerLevel:        "info",
	LoggerFormat:       "text",
	RelayCheck:         false,
	RelaySignatureCheck: true,
	GenesisTime:        0,
	RequestTimeoutMs:   5000,
	RequestMaxRetries:  3,
	GenesisForkVersion: "0x00000000",
	GenesisValidatorsRoot: "0x00000000000000000000000000000000",
}
