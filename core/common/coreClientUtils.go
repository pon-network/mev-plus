package common

import (
	"context"
	"encoding/json"
	"strconv"
	"fmt"

	"github.com/bsn-eng/mev-plus/common"
)

func ClientFromContext(ctx context.Context) (*Client, bool) {
	client, ok := ctx.Value(clientContextKey{}).(*Client)
	return client, ok
}

func (c *Client) nextID() json.RawMessage {
	id := c.idCounter.Add(1)
	return strconv.AppendUint(nil, uint64(id), 10)
}

func (c *Client) newMessage(method string, notifyAll bool, paramsIn ...interface{}) (*JsonRPCMessage, error) {
	if _, ok := c.knownCallbacks[method]; !ok {
		return nil, fmt.Errorf("unknown method: %s", method)
	}
	msg := &JsonRPCMessage{Version: common.Vsn, ID: c.nextID(), Method: method, NotifyAll: notifyAll}
	if paramsIn != nil { // prevent sending "params":null
		var err error
		if msg.Params, err = json.Marshal(paramsIn); err != nil {
			return nil, err
		}
	}
	return msg, nil
}
