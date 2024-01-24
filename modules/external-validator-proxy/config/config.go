package config

import (
	"net/url"
)

type ProxyConfig struct {
	LoggerLevel       string
	LoggerFormat      string
	Addresses           []*url.URL
	RequestTimeoutMs  int
	RequestMaxRetries int
}

var ProxyConfigDefaults = ProxyConfig{
	LoggerLevel:       "info",
	LoggerFormat:      "text",
	Addresses: 		 []*url.URL{}, // Default to nil so we can check if it's set
	RequestTimeoutMs:  5000,
	RequestMaxRetries: 3,
}
