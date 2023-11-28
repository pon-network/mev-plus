package config

import (
	"net/url"
)

type ProxyConfig struct {
	LoggerLevel       string
	LoggerFormat      string
	Address           *url.URL
	RequestTimeoutMs  int
	RequestMaxRetries int
}

var ProxyConfigDefaults = ProxyConfig{
	LoggerLevel:       "info",
	LoggerFormat:      "text",
	Address:           nil, // Default to nil so we can check if it's set
	RequestTimeoutMs:  5000,
	RequestMaxRetries: 3,
}
