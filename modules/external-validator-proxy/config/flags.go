package config

import (
	"github.com/pon-network/mev-plus/cmd/utils"
	cli "github.com/urfave/cli/v2"
)

var (
	LoggerLevelFlag = &cli.StringFlag{
		Name:     ModuleName + "." + "logger-level",
		Usage:    "Set the logger level",
		Category: utils.ExternalValidatorProxyCategory,
		Value:    "info",
		EnvVars:  []string{"PROXY_LOGGER_LEVEL"},
	}

	LoggerFormatFlag = &cli.StringFlag{
		Name:     ModuleName + "." + "logger-format",
		Usage:    "Set the logger format",
		Category: utils.ExternalValidatorProxyCategory,
		Value:    "text",
		EnvVars:  []string{"PROXY_LOGGER_FORMAT"},
	}

	AddressFlag = &cli.StringFlag{
		Name:     ModuleName + "." + "address",
		Usage:    "Set the listen addresses [comma separated to a max of 2] for the external validator proxies to connect to",
		Category: utils.ExternalValidatorProxyCategory,
	}

	RequestTimeoutMsFlag = &cli.IntFlag{
		Name:     ModuleName + "." + "request-timeout-ms",
		Usage:    "Set the request timeout in milliseconds",
		Category: utils.ExternalValidatorProxyCategory,
		Value:    ProxyConfigDefaults.RequestTimeoutMs,
	}

	RequestMaxRetriesFlag = &cli.IntFlag{
		Name:     ModuleName + "." + "request-max-retries",
		Usage:    "Set the request max retries",
		Category: utils.ExternalValidatorProxyCategory,
		Value:    ProxyConfigDefaults.RequestMaxRetries,
	}
)
