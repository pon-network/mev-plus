package common

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"strconv"

	"github.com/pon-network/mev-plus/common"
)

func ClientFromContext(ctx context.Context) (*Client, bool) {
	client, ok := ctx.Value(clientContextKey{}).(*Client)
	return client, ok
}

func (c *Client) nextID() json.RawMessage {
	id := c.idCounter.Add(1)
	return strconv.AppendUint(nil, uint64(id), 10)
}

func (c *Client) newMessage(method string, notifyAll bool, notificationExclusion []string, paramsIn ...interface{}) (*JsonRPCMessage, error) {
	if _, ok := c.knownCallbacks[method]; !ok {
		// if its not a notif message to all through core throw unknown
		if !(strings.HasPrefix(method, "core_") && notifyAll) {
			return nil, fmt.Errorf("unknown method: %s", method)
		}
	}
	msg := &JsonRPCMessage{Version: common.Vsn, ID: c.nextID(), Method: method, NotifyAll: notifyAll, NotifyExclusion: notificationExclusion}
	if paramsIn != nil { // prevent sending "params":null
		var err error
		if msg.Params, err = json.Marshal(paramsIn); err != nil {
			return nil, err
		}
	}
	return msg, nil
}