package common

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"sync"
	"time"

	"github.com/bsn-eng/mev-plus/common"

	log "github.com/sirupsen/logrus"
)

type handler struct {
	idgen      func() string
	respWait   map[string]*requestOp // active client requests
	callWG     sync.WaitGroup        // pending call goroutines
	rootCtx    context.Context       // canceled by close()
	cancelRoot func()                // cancel function for rootCtx
	conn       chan JsonRPCMessage

	serviceCallBacks map[string]*Callback

	subLock sync.Mutex

	log log.Logger
}

type callProc struct {
	ctx context.Context
}

func newHandler(
	connCtx context.Context,
	conn chan JsonRPCMessage,
	idgen func() string,
	serviceCallBacks map[string]*Callback,
) *handler {
	rootCtx, cancelRoot := context.WithCancel(connCtx)
	h := &handler{
		idgen:            idgen,
		conn:             conn,
		respWait:         make(map[string]*requestOp),
		rootCtx:          rootCtx,
		cancelRoot:       cancelRoot,
		serviceCallBacks: serviceCallBacks,
		log:              *log.New(),
	}
	return h
}

// handleMsg handles a single non-batch message.
func (h *handler) handleMsg(msg *JsonRPCMessage) {
	h.handleResponse(msg, func(msg *JsonRPCMessage) {
		h.processAsync(func(cp *callProc) {
			h.startCallHandling(cp, msg)
		})
	})
}

func (h *handler) startCallHandling(cp *callProc, msg *JsonRPCMessage) {
	var (
		responded sync.Once
		timer     *time.Timer
		cancel    context.CancelFunc
	)
	cp.ctx, cancel = context.WithCancel(cp.ctx)
	defer cancel()

	// Cancel the request context after timeout and send an error response. Since the
	// running method might not return immediately on timeout, we must wait for the
	// timeout concurrently with processing the request.
	if timeout, ok := ContextRequestTimeout(cp.ctx); ok {
		timer = time.AfterFunc(timeout, func() {
			cancel()
			responded.Do(func() {
				resp := msg.ErrorResponse(&common.InternalServerError{
					Code:    common.RPCTimeoutErrorCode,
					Message: "timeout",
				})
				h.conn <- *resp
			})
		})
	}

	answer := h.handleCallMsg(cp, msg)
	if timer != nil {
		timer.Stop()
	}
	if answer != nil {
		responded.Do(func() {
			h.conn <- *answer
		})
	}
}

// ContextRequestTimeout returns the request timeout derived from the given context.
func ContextRequestTimeout(ctx context.Context) (time.Duration, bool) {
	timeout := time.Duration(math.MaxInt64)
	hasTimeout := false
	setTimeout := func(d time.Duration) {
		if d < timeout {
			timeout = d
			hasTimeout = true
		}
	}

	if deadline, ok := ctx.Deadline(); ok {
		setTimeout(time.Until(deadline))
	}

	return timeout, hasTimeout
}

// close cancels all requests except for inflightReq and waits for
// call goroutines to shut down.
func (h *handler) close(err error, inflightReq *requestOp) {
	h.callWG.Wait()
	h.cancelRoot()
}

// addRequestOp registers a request operation.
func (h *handler) addRequestOp(op *requestOp) {
	h.respWait[string(op.id)] = op
}

// removeRequestOps stops waiting for the given request IDs.
func (h *handler) removeRequestOp(op *requestOp) {
	delete(h.respWait, string(op.id))
}

// processAsync runs fn in a new goroutine and starts tracking it in the h.calls wait group.
func (h *handler) processAsync(fn func(*callProc)) {
	h.callWG.Add(1)
	go func() {
		ctx, cancel := context.WithCancel(h.rootCtx)
		defer h.callWG.Done()
		defer cancel()
		fn(&callProc{ctx: ctx})
	}()
}

// handleResponse processes method call responses.
func (h *handler) handleResponse(msg *JsonRPCMessage, handleCall func(*JsonRPCMessage)) {
	var resolvedop *requestOp
	handleResp := func(msg *JsonRPCMessage) {
		op := h.respWait[string(msg.ID)]
		if op == nil {
			// RPC response that did not originate from a call made by this client.
			h.log.Debug("Unsolicited RPC response", "reqid", msg.ID)
			return
		}
		resolvedop = op
		delete(h.respWait, string(msg.ID))

		if !op.hadResponse {
			op.hadResponse = true
			op.resp <- msg
		}
	}

	switch {
	case msg.IsResponse():
		handleResp(msg)
	default:
		handleCall(msg)
	}

	if resolvedop != nil {
		h.removeRequestOp(resolvedop)
	}
}

// handleCallMsg executes a call message and returns the answer.
func (h *handler) handleCallMsg(ctx *callProc, msg *JsonRPCMessage) *JsonRPCMessage {
	start := time.Now()
	switch {
	case msg.IsNotification():
		h.handleCall(ctx, msg)
		h.log.Debug("Served "+msg.Method, "duration", time.Since(start))
		// handle call but no need to send back response as it is a notification with nil resp pointer
		return nil

	case msg.IsCall():
		resp := h.handleCall(ctx, msg)
		logMsg := fmt.Sprintf("Served %s reqid: %s duration: %s", msg.Method, string(msg.ID), time.Since(start).String())
		if resp.Error != nil {
			logMsg += " error: " + resp.Error.Message
			if resp.Error.Data != nil {
				logMsg += " errdata: " + resp.Error.Data.(string)
			}
			// h.log.Warn(logMsg)
		} else {
			// h.log.Debug(logMsg)
		}
		return resp

	case msg.HasValidID():
		return msg.ErrorResponse(&common.InvalidRequestError{Message: "invalid request"})

	default:
		return ErrorMessage(&common.InvalidRequestError{Message: "invalid request"})
	}
}

// handleCall processes method calls.
func (h *handler) handleCall(cp *callProc, msg *JsonRPCMessage) *JsonRPCMessage {
	var callb *Callback
	methodToCall := msg.MethodName()
	callb, ok := h.serviceCallBacks[methodToCall]
	if !ok {
		return msg.ErrorResponse(&common.MethodNotFoundError{Method: msg.Method})
	}

	if callb == nil {
		return msg.ErrorResponse(&common.MethodNotFoundError{Method: msg.Method})
	}

	args, err := parsePositionalArguments(msg.Params, callb.argTypes)
	if err != nil {
		return msg.ErrorResponse(&common.InvalidParamsError{Message: err.Error()})
	}

	answer := h.runMethod(cp.ctx, msg, callb, args)

	return answer
}

// runMethod runs the Go callback for an RPC method.
func (h *handler) runMethod(ctx context.Context, msg *JsonRPCMessage, callb *Callback, args []reflect.Value) *JsonRPCMessage {
	result, err := callb.call(ctx, msg.Method, args)
	if err != nil {
		return msg.ErrorResponse(err)
	}
	return msg.Response(result)
}
