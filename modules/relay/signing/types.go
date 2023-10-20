package signing

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type (
	Domain      [32]byte
	DomainType  [4]byte
	ForkVersion [4]byte
	Root        [32]byte
)

func (h Root) String() string {
	return hexutil.Bytes(h[:]).String()
}