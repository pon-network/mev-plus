package externalvalidatorproxy

import (
	"errors"
)

var (

	ErrHTTPErrorResponse         = errors.New("HTTP error response")
	ErrMaxRetriesExceeded        = errors.New("max retries exceeded")
)
