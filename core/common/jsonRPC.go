package common

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pon-network/mev-plus/common"
)

type JsonRPCMessage struct {
	Version         string          `json:"jsonrpc,omitempty"`
	ID              json.RawMessage `json:"id,omitempty"`
	Method          string          `json:"method,omitempty"`
	Params          json.RawMessage `json:"params,omitempty"`
	Error           *jsonError      `json:"error,omitempty"`
	Result          json.RawMessage `json:"result,omitempty"`
	NotifyAll       bool            `json:"notifyAll"`
	NotifyExclusion []string        `json:"notifyExclusion,omitempty"`
	Origin          string          `json:"origin,omitempty"`
}

func (msg *JsonRPCMessage) IsNotification() bool {
	return msg.HasValidVersion() && msg.ID == nil && msg.Method != ""
}

func (msg *JsonRPCMessage) IsCall() bool {
	return msg.HasValidVersion() && msg.HasValidID() && msg.Method != ""
}

func (msg *JsonRPCMessage) IsResponse() bool {
	return msg.HasValidVersion() && msg.HasValidID() && msg.Params == nil && strings.HasSuffix(msg.Method, common.ResponseMethodSuffix)
}

func (msg *JsonRPCMessage) HasValidID() bool {
	return len(msg.ID) > 0 && msg.ID[0] != '{' && msg.ID[0] != '['
}

func (msg *JsonRPCMessage) HasValidVersion() bool {
	return msg.Version == common.Vsn
}

func (msg *JsonRPCMessage) Namespace() string {
	elem := strings.SplitN(msg.Method, common.ServiceMethodSeparator, 2)
	return elem[0]
}

func (msg *JsonRPCMessage) MethodName() string {
	elem := strings.SplitN(msg.Method, common.ServiceMethodSeparator, 3)
	return elem[1]
}

func (msg *JsonRPCMessage) String() string {
	b, _ := json.Marshal(msg)
	return string(b)
}

func (msg *JsonRPCMessage) ErrorResponse(err error) *JsonRPCMessage {
	resp := ErrorMessage(err)
	resp.ID = msg.ID
	if !strings.HasSuffix(msg.Method, common.ResponseMethodSuffix) {
		resp.Method = msg.Method + common.ResponseMethodSuffix
	}
	resp.Origin = msg.Origin
	return resp
}

func (msg *JsonRPCMessage) Response(result interface{}) *JsonRPCMessage {
	enc, err := json.Marshal(result)
	if err != nil {
		return msg.ErrorResponse(&common.InternalServerError{Code: common.RPCUnmarshalErrorCode, Message: err.Error()})
	}
	resp := &JsonRPCMessage{Version: common.Vsn, ID: msg.ID, Result: enc, Origin: msg.Origin}

	if !strings.HasSuffix(msg.Method, common.ResponseMethodSuffix) {
		resp.Method = msg.Method + common.ResponseMethodSuffix
	}
	return resp
}

func ErrorMessage(err error) *JsonRPCMessage {
	msg := &JsonRPCMessage{Version: common.Vsn, ID: null, Error: &jsonError{
		Code:    common.RPCDefaultErrorCode,
		Message: err.Error(),
	}}
	ec, ok := err.(common.Error)
	if ok {
		msg.Error.Code = ec.ErrorCode()
	}
	de, ok := err.(common.DataError)
	if ok {
		msg.Error.Data = de.ErrorData()
	}
	return msg
}

type jsonError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (err *jsonError) Error() string {
	if err.Message == "" {
		return fmt.Sprintf("json-rpc error %d", err.Code)
	}
	return err.Message
}

func (err *jsonError) ErrorCode() int {
	return err.Code
}

func (err *jsonError) ErrorData() interface{} {
	return err.Data
}
