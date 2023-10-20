package common

import (
	"context"
	"reflect"
	"runtime"
	"unicode"

	"github.com/bsn-eng/mev-plus/common"
)

// callback is a method callback which was registered in the core
type Callback struct {
	fn          reflect.Value  // the function
	rcvr        reflect.Value  // receiver object of method, set if fn is method
	argTypes    []reflect.Type // input argument types
	hasCtx      bool           // method's first argument is a context (not included in argTypes)
	errPos      int            // err return idx, of -1 when method cannot return error
}

// suitableCallbacks iterates over the methods of the given type. It determines if a method
// satisfies the criteria for a RPC callback and adds it to the
// collection of callbacks.
func suitableCallbacks(receiver reflect.Value) map[string]*Callback {
	typ := receiver.Type()
	callbacks := make(map[string]*Callback)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		if _, ok := ParkedCallbacks[formatName(method.Name)]; ok {
			continue
		}
		if method.PkgPath != "" {
			continue // method not exported
		}
		cb := newCallback(receiver, method.Func)
		if cb == nil {
			continue // function invalid
		}
		name := formatName(method.Name)
		callbacks[name] = cb
	}
	return callbacks
}

// newCallback turns fn (a function) into a callback object. It returns nil if the function
// is unsuitable as an RPC callback.
func newCallback(receiver, fn reflect.Value) *Callback {
	fntype := fn.Type()
	c := &Callback{fn: fn, rcvr: receiver, errPos: -1,}
	// Determine parameter types. They must all be exported or builtin types.
	c.makeArgTypes()

	// Verify return types. The function must return at most one error
	// and/or one other non-error value.
	outs := make([]reflect.Type, fntype.NumOut())
	for i := 0; i < fntype.NumOut(); i++ {
		outs[i] = fntype.Out(i)
	}
	if len(outs) > 2 {
		return nil
	}
	// If an error is returned, it must be the last returned value.
	switch {
	case len(outs) == 1 && isErrorType(outs[0]):
		c.errPos = 0
	case len(outs) == 2:
		if isErrorType(outs[0]) || !isErrorType(outs[1]) {
			return nil
		}
		c.errPos = 1
	}
	return c
}

// makeArgTypes composes the argTypes list.
func (c *Callback) makeArgTypes() {
	fntype := c.fn.Type()
	// Skip receiver and context.Context parameter (if present).
	firstArg := 0
	if c.rcvr.IsValid() {
		firstArg++
	}
	if fntype.NumIn() > firstArg && fntype.In(firstArg) == contextType {
		c.hasCtx = true
		firstArg++
	}
	// Add all remaining parameters.
	c.argTypes = make([]reflect.Type, fntype.NumIn()-firstArg)
	for i := firstArg; i < fntype.NumIn(); i++ {
		c.argTypes[i-firstArg] = fntype.In(i)
	}
}

// call invokes the callback.
func (c *Callback) call(ctx context.Context, method string, args []reflect.Value) (res interface{}, errRes error) {
	// Create the argument slice.
	fullargs := make([]reflect.Value, 0, 2+len(args))
	if c.rcvr.IsValid() {
		fullargs = append(fullargs, c.rcvr)
	}
	if c.hasCtx {
		fullargs = append(fullargs, reflect.ValueOf(ctx))
	}
	fullargs = append(fullargs, args...)

	// Catch panic while running the callback.
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			errRes = &common.InternalServerError{Code: common.RPCInternalErrorCode, Message: "RPC method " + method + " crashed: " + err.(error).Error()}
		}
	}()
	// Run the callback.
	results := c.fn.Call(fullargs)
	if len(results) == 0 {
		return nil, nil
	}
	if c.errPos >= 0 && !results[c.errPos].IsNil() {
		// Method has returned non-nil error value.
		err := results[c.errPos].Interface().(error)
		return reflect.Value{}, err
	}
	return results[0].Interface(), nil
}

// Does t satisfy the error interface?
func isErrorType(t reflect.Type) bool {
	return t.Implements(errorType)
}


// formatName converts to first character of name to lowercase.
func formatName(name string) string {
	ret := []rune(name)
	if len(ret) > 0 {
		ret[0] = unicode.ToLower(ret[0])
	}
	return string(ret)
}
