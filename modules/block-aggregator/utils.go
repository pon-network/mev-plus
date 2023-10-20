package blockaggregator

import (
	"fmt"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/sirupsen/logrus"
)

var log = logrus.NewEntry(logrus.New())

// _HexToHash converts a hexadecimal string to an Ethereum hash
func _HexToHash(s string) (ret phase0.Hash32) {
	ret, err := HexToHash(s)
	if err != nil {
		log.Error(err, " _HexToHash: ", s)
		panic(err)
	}
	return ret
}

// HexToHash takes a hex string and returns a Hash
func HexToHash(s string) (ret phase0.Hash32, err error) {
	bytes, err := hexutil.Decode(s)
	if len(bytes) != len(ret) {
		return phase0.Hash32{}, fmt.Errorf("Invalid length")
	}
	copy(ret[:], bytes)
	return
}
