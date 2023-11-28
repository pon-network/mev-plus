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
	ServerReadTimeoutMs:       1000,
	ServerReadHeaderTimeoutMs: 1000,
	ServerWriteTimeoutMs:      0,
	ServerIdleTimeoutMs:       0,
	ServerMaxHeaderBytes:      4000,
}
