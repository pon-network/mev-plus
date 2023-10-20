package data

import (
	"math/big"

	"github.com/attestantio/go-builder-client/spec"
)

type SlotHeader struct {
	ModuleName string
	Slot       uint64
	Value      *big.Int
	BlockHash  string
	Bid        *spec.VersionedSignedBuilderBid
}
