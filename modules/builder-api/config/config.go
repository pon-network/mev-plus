package config

import (
	"net/url"
)

type BuilderApiConfig struct {
	LoggerLevel               string
	LoggerFormat              string
	ListenAddress             *url.URL
	ServerReadTimeoutMs       int
	ServerReadHeaderTimeoutMs int
	ServerWriteTimeoutMs      int
	ServerIdleTimeoutMs       int
	ServerMaxHeaderBytes      int
}

var BuilderApiConfigDefaults = BuilderApiConfig{
	LoggerLevel:               "info",
	LoggerFormat:              "text",
	ListenAddress:             &url.URL{Scheme: "http", Host: "localhost:18551"},
	ServerReadTimeoutMs:       12000,
	ServerReadHeaderTimeoutMs: 12000,
	ServerWriteTimeoutMs:      12000,
	ServerIdleTimeoutMs:       12000,
	ServerMaxHeaderBytes:      100000,
}
