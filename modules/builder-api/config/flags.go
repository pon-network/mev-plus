package config

import (
	"github.com/pon-network/mev-plus/cmd/utils"
	cli "github.com/urfave/cli/v2"
)

var (
	LoggerLevelFlag = &cli.StringFlag{
		Name:     ModuleName + "." + "logger-level",
		Usage:    "Set the logger level",
		Category: utils.BuilderAPICategory,
		Value:    "info",
		EnvVars:  []string{"BUILDERAPI_LOGGER_LEVEL"},
	}
	ListenAddressFlag = &cli.StringFlag{
		Name:     ModuleName + "." + "listen-address",
		Usage:    "Set the listen address",
		Category: utils.BuilderAPICategory,
		Value:    "",
		EnvVars:  []string{"BUILDERAPI_LISTEN_ADDRESS"},
	}
	LoggerFormatFlag = &cli.StringFlag{
		Name:     ModuleName + "." + "logger-format",
		Usage:    "Set the logger format",
		Category: utils.BuilderAPICategory,
		Value:    "text",
		EnvVars:  []string{"BUILDERAPI_LOGGER_FORMAT"},
	}

	ServerReadHeaderTimeoutMsFlag = &cli.IntFlag{
		Name:     ModuleName + "." + "server-read-header-timeout-ms",
		Usage:    "Set the server read header timeout in milliseconds",
		Category: utils.BuilderAPICategory,
		Value:    1000,
		EnvVars:  []string{"BUILDERAPI_SERVER_READ_HEADER_TIMEOUT_MS"},
	}

	ServerReadTimeoutMsFlag = &cli.IntFlag{
		Name:     ModuleName + "." + "server-read-timeout-ms",
		Usage:    "Set the server read timeout in milliseconds",
		Category: utils.BuilderAPICategory,
		Value:    1000,
		EnvVars:  []string{"BUILDERAPI_SERVER_READ_TIMEOUT_MS"},
	}

	ServerWriteTimeoutMsFlag = &cli.IntFlag{
		Name:     ModuleName + "." + "server-write-timeout-ms",
		Usage:    "Set the server write timeout in milliseconds",
		Category: utils.BuilderAPICategory,
		Value:    0,
		EnvVars:  []string{"BUILDERAPI_SERVER_WRITE_TIMEOUT_MS"},
	}

	ServerIdleTimeoutMsFlag = &cli.IntFlag{
		Name:     ModuleName + "." + "server-idle-timeout-ms",
		Usage:    "Set the server idle timeout in milliseconds",
		Category: utils.BuilderAPICategory,
		Value:    0,
		EnvVars:  []string{"BUILDERAPI_SERVER_IDLE_TIMEOUT_MS"},
	}

	ServerMaxHeaderBytesFlag = &cli.IntFlag{
		Name:     ModuleName + "." + "server-max-header-bytes",
		Usage:    "Set the server max header bytes",
		Category: utils.BuilderAPICategory,
		Value:    4000,
		EnvVars:  []string{"BUILDERAPI_SERVER_MAX_HEADER_BYTES"},
	}
)
