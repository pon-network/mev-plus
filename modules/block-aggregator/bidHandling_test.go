package blockaggregator

import (
	"testing"

	"github.com/attestantio/go-builder-client/api/capella"
	"github.com/attestantio/go-builder-client/spec"
	consensusspec "github.com/attestantio/go-eth2-client/spec"
	capella2 "github.com/attestantio/go-eth2-client/spec/capella"
	"github.com/bsn-eng/mev-plus/modules/block-aggregator/data"
	"github.com/holiman/uint256"
)

func TestProcessNewBid(t *testing.T) {
	t.Run("ProcessNewBid", func(t *testing.T) {
		aggregator := data.NewAggregatorData()
		b := &BlockAggregatorService{
			Data: aggregator,
		}
		blockHash := _HexToHash("0x534809bd2b6832edff8d8ce4cb0e50068804fd1ef432c8362ad708a74fdc0e46")
		bid := &spec.VersionedSignedBuilderBid{
			Version: consensusspec.DataVersionCapella,
			Capella: &capella.SignedBuilderBid{
				Message: &capella.BuilderBid{
					Value: &uint256.Int{23},
					Header: &capella2.ExecutionPayloadHeader{
						BlockHash: blockHash,
					},
				},
			},
		}

		err := b.processNewBid("NewBid", 12, *bid)
		if err != nil {
			t.Errorf("Error in processing new bid %v", err)
		}
	})

	t.Run("FailProcessNewBid", func(t *testing.T) {
		aggregator := data.NewAggregatorData()
		b := &BlockAggregatorService{
			Data: aggregator,
		}
		blockHash := _HexToHash("0x534809bd2b6832edff8d8ce4cb0e50068804fd1ef432c8362ad708a74fdc0e46")
		bid := &spec.VersionedSignedBuilderBid{
			Version: consensusspec.DataVersionCapella,
			Capella: &capella.SignedBuilderBid{
				Message: &capella.BuilderBid{
					Value: &uint256.Int{23},
					Header: &capella2.ExecutionPayloadHeader{
						BlockHash: blockHash,
					},
				},
			},
		}
		// Set the last slot to 13
		aggregator.SetLastSlot(13)
		err := b.processNewBid("NewBid", 12, *bid)
		if err == nil {
			t.Errorf("Expected error in processing bid")
		}
	})

	t.Run("FailProcessNewBidWithNoHeader", func(t *testing.T) {
		aggregator := data.NewAggregatorData()
		b := &BlockAggregatorService{
			Data: aggregator,
		}
		bid := &spec.VersionedSignedBuilderBid{
			Version: consensusspec.DataVersionCapella,
			Capella: &capella.SignedBuilderBid{
				Message: &capella.BuilderBid{
					Value: &uint256.Int{23},
				},
			},
		}

		err := b.processNewBid("NewBid", 12, *bid)
		if err == nil {
			t.Errorf("Expected error in processing bid with no header")
		}
	})
}
