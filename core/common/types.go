package common

import (
	"encoding/json"

	"github.com/pon-pbs/mev-plus/common"
)

// Service represents a service that can handle events.
type Service interface {
	// Any attached service must implement these method.
	Name() string
	Start() error
	Stop() error
	ConnectCore(coreClient *Client, pingId string) error
	Configure(moduleFlags common.ModuleFlags) error
}

// Should not be accessible over communication channels
var ParkedCallbacks map[string]bool = map[string]bool{
	"start":       true,
	"stop":        true,
	"connectCore": true,
	"configure":   true,
}

type Module struct {
	Name         string
	Service      Service
	ServiceAlive bool
	Callbacks    map[string]*Callback
}

type ModuleCommChannels struct {
	Incoming chan JsonRPCMessage
	Outgoing chan JsonRPCMessage
}

type requestOp struct {
	id          json.RawMessage
	err         error
	resp        chan *JsonRPCMessage // the response goes here
	hadResponse bool                 // true when the request was responded to
}

type readOp struct {
	msg *JsonRPCMessage
}
