package blockaggregator

import (

	"github.com/attestantio/go-builder-client/spec"
	"github.com/bsn-eng/mev-plus/modules/block-aggregator/data"
)

func (b *BlockAggregatorService) processNewBid(name string, slot uint64, bid spec.VersionedSignedBuilderBid) error {
	
	value, err := bid.Value()
	if err != nil {
		return err
	}

	blockHash, err := bid.BlockHash()
	if err != nil {
		return err
	}
	
	processedHeader := data.SlotHeader{
		ModuleName: name,
		Slot:       slot,
		Bid:        &bid,
		Value:      value.ToBig(),
		BlockHash:  blockHash.String(),
	}

	err = b.Data.AddSlotHeader(processedHeader)
	if err != nil {
		return err
	}

	return nil
}
