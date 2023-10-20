package builderapi

import (
	"errors"
)

var (
	errServerAlreadyRunning = errors.New("server already running")
	ErrLength               = errors.New("invalid length")

	errInvalidSlotNumber = errors.New("invalid slot number")
	errInvalidPubkey	 = errors.New("invalid pubkey")
	errInvalidHash		 = errors.New("invalid hash")
)
