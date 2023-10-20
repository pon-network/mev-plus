package relay

import (
	"errors"
	"fmt"
)

var (
	ErrMissingRelayPubkey        = fmt.Errorf("missing relay public key")
	ErrHTTPErrorResponse         = errors.New("HTTP error response")
	ErrNoBidReceived             = errors.New("No bid received")
	ErrInvalidInput              = errors.New("Invalid input")
	ErrIncompletePayload         = errors.New("Missing parts of the request payload from the beacon-node")
	ErrNoPayloadReceived         = errors.New("No payload received from relay")
	ErrMaxRetriesExceeded        = errors.New("max retries exceeded")
	ErrInvalidTransaction        = errors.New("invalid transaction")
	ErrPointAtInfinityPubkey     = fmt.Errorf("relay public key cannot be the point-at-infinity")
	ErrLength                    = errors.New("invalid length")
	ErrNoRelays                  = errors.New("no relays")
	ErrInvalidSlot               = errors.New("invalid slot")
	ErrInvalidHash               = errors.New("invalid hash")
	ErrInvalidPubkey             = errors.New("invalid pubkey")
	ErrNoSuccessfulRelayResponse = errors.New("no successful relay response")
	ErrUseLastResponse           = errors.New("net/http: use last response")
)
