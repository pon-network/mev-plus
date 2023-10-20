package common

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
)

// Call performs a JSON-RPC call with the given arguments and unmarshals into
// result if no error occurred.
//
// The result must be a pointer so that package json can unmarshal into it. You
// can also pass nil, in which case the result is ignored.
func (c *Client) Call(result interface{}, method string, notifyAll bool, args ...interface{}) error {
	ctx := context.Background()
	return c.CallContext(ctx, result, method, notifyAll, args...)
}

// CallContext performs a JSON-RPC call with the given arguments. If the context is
// canceled before the call has successfully returned, CallContext returns immediately.
//
// The result must be a pointer so that package json can unmarshal into it. You
// can also pass nil, in which case the result is ignored.
func (c *Client) CallContext(ctx context.Context, result interface{}, method string, notifyAll bool, args ...interface{}) error {
	if result != nil && reflect.TypeOf(result).Kind() != reflect.Ptr {
		return fmt.Errorf("call result parameter must be pointer or nil interface: %v", result)
	}
	msg, err := c.newMessage(method, notifyAll, args...)
	if err != nil {
		return err
	}
	op := &requestOp{
		id:   msg.ID,
		resp: make(chan *JsonRPCMessage, 1),
	}

	err = c.send(ctx, op, msg)
	if err != nil {
		return err
	}

	resp, err := op.wait(ctx, c)
	if err != nil {
		return err
	}
	switch {
	case resp.Error != nil:
		return resp.Error
	case len(resp.Result) == 0:
		return ErrNoResult
	default:
		if result == nil {
			return nil
		}
		return json.Unmarshal(resp.Result, result)
	}
}

// Notify sends a notification, i.e. a method call that doesn't expect a response.
func (c *Client) Notify(ctx context.Context, method string, notifyAll bool, args ...interface{}) error {
	op := new(requestOp)
	msg, err := c.newMessage(method, notifyAll, args...)
	if err != nil {
		return err
	}
	msg.ID = nil

	return c.send(ctx, op, msg)
}

// send registers op with the dispatch loop, then sends msg on the connection.
// if sending fails, op is deregistered.
func (c *Client) send(ctx context.Context, op *requestOp, msg interface{}) error {
	select {
	case c.reqInit <- op:
		err := c.write(ctx, msg)
		c.reqSent <- err
		return err
	case <-ctx.Done():
		// This can happen if the client is overloaded.
		return ctx.Err()
	case <-c.close:
		return ErrClientQuit
	}
}

func (c *Client) write(ctx context.Context, msg interface{}) error {
	c.commChannels.Outgoing <- *msg.(*JsonRPCMessage)
	return nil
}

func (op *requestOp) wait(ctx context.Context, c *Client) (*JsonRPCMessage, error) {

	select {
	case <-ctx.Done():
		// Context timeout or cancelation.
		select {
		case c.reqTimeout <- op:
			// Put and wait for request to be removed from the handler.
		case <-c.close:
			// If client is called to close, stop waiting and return.
		}
		return nil, ctx.Err()
	case resp := <-op.resp:
		// Response received.
		return resp, op.err
	}
}
