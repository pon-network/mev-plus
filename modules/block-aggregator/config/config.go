package config

type BlockAggregatorConfig struct {
	GenesisTime        uint64
	AuctionDuration	uint64 // in seconds
	SlotDuration		uint64 // in seconds
}

var BlockAggregatorConfigDefaults = BlockAggregatorConfig{
	GenesisTime:        0,
	AuctionDuration:	0,
	SlotDuration:		12,
}