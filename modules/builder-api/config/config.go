package config

type BuilderApiConfig struct {
	LoggerLevel string
	LoggerFormat string
	ListenAddress string
	ServerReadTimeoutMs int
	ServerReadHeaderTimeoutMs int
	ServerWriteTimeoutMs int
	ServerIdleTimeoutMs int
	ServerMaxHeaderBytes int
}

var BuilderApiConfigDefaults = BuilderApiConfig{
	LoggerLevel: "info",
	LoggerFormat: "text",
	ListenAddress: "localhost:18550",
	ServerReadTimeoutMs: 1000,
	ServerReadHeaderTimeoutMs: 1000,
	ServerWriteTimeoutMs: 0,
	ServerIdleTimeoutMs: 0,
	ServerMaxHeaderBytes: 4000,
}
