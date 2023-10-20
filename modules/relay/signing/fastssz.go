package signing

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

)

func (f *ForkVersion) FromSlice(s []byte) error {
	if len(s) != 4 {
		return errors.New("invalid fork version length")
	}
	copy(f[:], s)
	return nil
}

var (
	DomainBuilder Domain

	DomainTypeAppBuilder     = DomainType{0x00, 0x00, 0x00, 0x01}
)

type ForkData struct {
	CurrentVersion        ForkVersion `ssz-size:"4"`
	GenesisValidatorsRoot Root        `ssz-size:"32"`
}

func ComputeDomain(domainType DomainType, forkVersionHex, genesisValidatorsRootHex string) (domain Domain, err error) {
	genesisValidatorsRoot := Root(common.HexToHash(genesisValidatorsRootHex))
	forkVersionBytes, err := hexutil.Decode(forkVersionHex)
	if err != nil || len(forkVersionBytes) != 4 {
		return domain, errors.New("Wrong Fork Version")
	}
	var forkVersion [4]byte
	copy(forkVersion[:], forkVersionBytes[:4])
	return ComputeSSZDomain(domainType, forkVersion, genesisValidatorsRoot), nil
}

func ComputeSSZDomain(dt DomainType, forkVersion ForkVersion, genesisValidatorsRoot Root) [32]byte {
	forkDataRoot, _ := (&ForkData{
		CurrentVersion:        forkVersion,
		GenesisValidatorsRoot: genesisValidatorsRoot,
	}).HashTreeRoot()

	var domain [32]byte
	copy(domain[0:4], dt[:])
	copy(domain[4:], forkDataRoot[0:28])

	return domain
}
