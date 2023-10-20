package common

import "fmt"


// CoreServices errors

var (
	ErrCoreRunning = fmt.Errorf("core is already running")
	ErrCoreStopped = fmt.Errorf("core is stopped")
)

// Error wraps RPC errors, which contain an error code in addition to the message.
type Error interface {
	Error() string  // returns the message
	ErrorCode() int // returns the code
}

// A DataError contains some data in addition to the error message.
type DataError interface {
	Error() string          // returns the message
	ErrorData() interface{} // returns the error data
}

// Error types defined below are the built-in JSON-RPC errors.

var (
	_ Error = new(MethodNotFoundError)
	_ Error = new(ParseError)
	_ Error = new(InvalidRequestError)
	_ Error = new(InvalidMessageError)
	_ Error = new(InvalidParamsError)
	_ Error = new(InternalServerError)
)

type MethodNotFoundError struct{ Method string }

func (e *MethodNotFoundError) ErrorCode() int { return -32601 }

func (e *MethodNotFoundError) Error() string {
	return fmt.Sprintf("the method %s does not exist/is not available", e.Method)
}

type NotificationsUnsupportedError struct{}

func (e NotificationsUnsupportedError) Error() string {
	return "notifications not supported"
}

func (e NotificationsUnsupportedError) ErrorCode() int { return -32601 }


func (e NotificationsUnsupportedError) Is(other error) bool {
	if other == (NotificationsUnsupportedError{}) {
		return true
	}
	rpcErr, ok := other.(Error)
	if ok {
		code := rpcErr.ErrorCode()
		return code == -32601
	}
	return false
}

// Invalid JSON was received by the server.
type ParseError struct{ Message string }

func (e *ParseError) ErrorCode() int { return -32700 }

func (e *ParseError) Error() string { return e.Message }

// received message isn't a valid request
type InvalidRequestError struct{ Message string }

func (e *InvalidRequestError) ErrorCode() int { return -32600 }

func (e *InvalidRequestError) Error() string { return e.Message }

// received message is invalid
type InvalidMessageError struct{ Message string }

func (e *InvalidMessageError) ErrorCode() int { return -32700 }

func (e *InvalidMessageError) Error() string { return e.Message }

// unable to decode supplied params, or an invalid number of parameters
type InvalidParamsError struct{ Message string }

func (e *InvalidParamsError) ErrorCode() int { return -32602 }

func (e *InvalidParamsError) Error() string { return e.Message }

// InternalServerError is used for server errors during request processing.
type InternalServerError struct {
	Code    int
	Message string
}

func (e *InternalServerError) ErrorCode() int { return e.Code }

func (e *InternalServerError) Error() string { return e.Message }
