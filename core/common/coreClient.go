package common

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/pon-network/mev-plus/common"
)

const (
	defaultCommsChanBufferSize = 20000
)

var (
	ErrBadResult  = errors.New("bad result in JSON-RPC response")
	ErrClientQuit = errors.New("client is closed")
	ErrNoResult   = errors.New("JSON-RPC response has no result")
)

type Client struct {
	id          string
	serviceName string

	idgen func() string

	idCounter    atomic.Uint32 // for request IDs
	mu           sync.Mutex
	commChannels *commChannels

	knownCallbacks map[string]bool

	handler *handler

	close  chan struct{} // close is closed when the client is closed (receives first signal to close)
	closed chan struct{} // closed is closed when the client is closed and all requests have been handled (receives last signal after close completes)

	// for dispatch
	readOp     chan readOp     // read messages
	reqInit    chan *requestOp // register response IDs, takes write lock
	reqSent    chan error      // signals write completion, releases write lock
	reqTimeout chan *requestOp // removes response IDs when call timeout expires
}

type clientContextKey struct{}

type clientConn struct {
	Channel chan JsonRPCMessage
	Handler *handler
}

type commChannels struct {
	Incoming chan JsonRPCMessage
	Outgoing chan JsonRPCMessage
}

func NewClient(
	initctx context.Context,
	serviceName string,
	serviceCallbacks map[string]*Callback,
	knownCallbacks map[string]bool,
) (string, *Client, *commChannels, error) {

	c := &Client{
		id:             common.NewID(),
		idgen:          common.NewID,
		close:          make(chan struct{}),
		closed:         make(chan struct{}),
		readOp:         make(chan readOp),
		reqInit:        make(chan *requestOp),
		reqSent:        make(chan error, 1),
		reqTimeout:     make(chan *requestOp),
		knownCallbacks: knownCallbacks,
	}

	c.serviceName = serviceName

	c.commChannels = &commChannels{
		Incoming: make(chan JsonRPCMessage, defaultCommsChanBufferSize),
		Outgoing: make(chan JsonRPCMessage, defaultCommsChanBufferSize),
	}

	ctx := context.WithValue(initctx, clientContextKey{}, c)
	// create the handler for the client that would handle the incoming messages, and write the outgoing responses if any.
	c.handler = newHandler(ctx, c.commChannels.Outgoing, c.idgen, serviceCallbacks)

	go c.dispatch()

	return c.id, c, c.commChannels, nil
}

func (c *Client) Ping(message string) error {
	err := c.Notify(context.Background(), "core_ping", false, nil, message)
	if err != nil {
		return err
	}
	return nil
}

// Close closes the client, aborting any in-flight requests.
func (c *Client) Close() {
	select {
	case c.close <- struct{}{}:
		<-c.closed
	case <-c.closed:
	}
}

// Main listener loop. Handles messages from the
// incomming channel reader and dispatches new requests to
// the outgoing channel, tracking the response.
func (c *Client) dispatch() {

	var (
		lastOp      *requestOp
		reqInitLock = c.reqInit // nil while the send lock is held
	)

	defer func() {
		close(c.close)
		c.handler.close(ErrClientQuit, nil)
		close(c.commChannels.Incoming)
		close(c.commChannels.Outgoing)
		close(c.closed)
	}()

	// Start the read loop.
	go c.read(c.commChannels.Incoming)

	for {
		select {
		case <-c.close:
			return

		// Read path:
		case op := <-c.readOp:
			c.handler.handleMsg(op.msg)

		// Send path:
		case op := <-reqInitLock:
			// Stop listening for further requests until the current one has been sent.
			reqInitLock = nil
			lastOp = op
			c.handler.addRequestOp(op)

		case err := <-c.reqSent:
			if err != nil {
				// Remove response handlers for the last send. When the read loop
				// goes down, it will signal all other current operations.
				c.handler.removeRequestOp(lastOp)
			}
			// Let the next request in, as the last was successfully sent, no need to track it.
			reqInitLock = c.reqInit

		case op := <-c.reqTimeout:
			c.handler.removeRequestOp(op)
		}
	}

}

// read decodes RPC messages from the incoming connection and
// sends them to the dispatch loop for handling.
func (c *Client) read(incomingChan chan JsonRPCMessage) {
	for {
		select {
		case <-c.close:
			return
		default:
			msg := <-incomingChan
			c.readOp <- readOp{msg: &msg}
		}

	}

}
