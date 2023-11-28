package blockaggregator

import (
	"fmt"
	"math/big"
	"strconv"
	"sync"

	"github.com/attestantio/go-builder-client/spec"
	"github.com/pon-network/mev-plus/common"
	coreCommon "github.com/pon-network/mev-plus/core/common"
	"github.com/pon-network/mev-plus/modules/block-aggregator/config"
	"github.com/pon-network/mev-plus/modules/block-aggregator/data"

	commonTypes "github.com/bsn-eng/pon-golang-types/common"
	params "github.com/ethereum/go-ethereum/params"

	apiv1 "github.com/attestantio/go-builder-client/api/v1"

	"github.com/sirupsen/logrus"
)

type BlockAggregatorService struct {
	coreClient            *coreCommon.Client
	log                   *logrus.Entry
	Data                  *data.AggregatorData
	ConnectedBLockSources []string
	lock                  sync.Mutex

	cfg config.BlockAggregatorConfig
}

func NewBlockAggregatorService() *BlockAggregatorService {
	return &BlockAggregatorService{
		log:  logrus.NewEntry(logrus.New()).WithField("moduleExecution", config.ModuleName),
		Data: data.NewAggregatorData(),
		cfg:  config.BlockAggregatorConfigDefaults,
	}
}

func (b *BlockAggregatorService) Name() string {
	return config.ModuleName
}

func (b *BlockAggregatorService) Start() error {
	return nil
}

func (b *BlockAggregatorService) Stop() error {
	return nil
}

func (b *BlockAggregatorService) ConnectCore(coreClient *coreCommon.Client, pingId string) error {

	// this is the first and only time the client is set and doesnt need a mutex
	b.coreClient = coreClient

	// test a ping to the core server
	err := b.coreClient.Ping(pingId)
	if err != nil {
		return err
	}

	return nil
}

func (b *BlockAggregatorService) Configure(moduleFlags common.ModuleFlags) error {

	for flagName, flagValue := range moduleFlags {
		switch flagName {
		case config.AuctionDurationFlag.Name:
			flagValint, err := strconv.Atoi(flagValue)
			if err != nil {
				return err
			}
			b.cfg.AuctionDuration = uint64(flagValint)
		case config.SlotDurationFlag.Name:
			flagValint, err := strconv.Atoi(flagValue)
			if err != nil {
				return err
			}
			b.cfg.SlotDuration = uint64(flagValint)
		case config.GenesisTimeFlag.Name:
			flagValint, err := strconv.Atoi(flagValue)
			if err != nil {
				return err
			}
			b.cfg.GenesisTime = uint64(flagValint)
		}
	}

	return nil
}

func (b *BlockAggregatorService) ConnectBlockSource(moduleName string) error {

	if len(moduleName) == 0 {
		return fmt.Errorf("invalid module name")
	}

	b.lock.Lock()
	defer b.lock.Unlock()
	// check if the module is already connected
	for _, module := range b.ConnectedBLockSources {
		if module == moduleName {
			b.log.Infof("Block source [%v] already connected to block aggregator", moduleName)
			return nil
		}
	}

	// check for the existence of the module by calling the module_name method
	var moduleNameCheck string
	err := b.coreClient.Call(&moduleNameCheck, moduleName+"_name", false, nil)
	if err != nil {
		b.log.WithError(err).WithField("module", moduleName).Warn("Error calling module")
		return err
	}
	if moduleNameCheck != moduleName {
		return fmt.Errorf("Could not identify module %s", moduleName)
	}

	// add the module to the list of connected modules
	b.ConnectedBLockSources = append(b.ConnectedBLockSources, moduleName)

	b.log.Infof("Connected block source [%v] to block aggregator", moduleName)

	return nil
}

func (b *BlockAggregatorService) Status() error {
	b.log.Info("Checking status of block aggregator and connected block sources")
	return b.checkBlockSources()
}

// Move this to a different module later for validator management
func (b *BlockAggregatorService) RegisterValidator(payload []apiv1.SignedValidatorRegistration) error {
	var proposers []string
	for _, reg := range payload {
		proposers = append(proposers, reg.Message.Pubkey.String())
	}
	b.log.Infof("Processing %v validator registrations through block aggregator", len(payload))
	return b.processValidatorRegistrations(payload)
}

func (b *BlockAggregatorService) GetHeader(slot uint64, parentHash, proposerPubkey string) (res []spec.VersionedSignedBuilderBid, err error) {
	b.log.Info("Processing get header request through block aggregator")
	if len(proposerPubkey) != 98 || len(parentHash) != 66 {
		b.log.WithFields(logrus.Fields{
			"proposerPubkey": proposerPubkey,
			"parentHash":     parentHash,
		}).Error("invalid proposerPubkey or parentHash")
		return res, fmt.Errorf("invalid proposerPubkey or parentHash")
	}

	slotHeader, err := b.processHeaderReq(slot, parentHash, proposerPubkey)
	if err != nil {
		b.log.WithError(err).WithFields(logrus.Fields{
			"slot":           slot,
			"parentHash":     parentHash,
			"proposerPubkey": proposerPubkey,
		}).Error("error processing header request")
		return res, err
	}

	b.log.WithFields(logrus.Fields{
		"slot":           slot,
		"parentHash":     parentHash,
		"proposerPubkey": proposerPubkey,
		"blockHash":      slotHeader.BlockHash,
		"value":          big.NewFloat(0).SetInt(slotHeader.Value).Quo(big.NewFloat(0).SetInt(slotHeader.Value), big.NewFloat(0).SetInt(big.NewInt(params.Ether))).String() + " ETH",
		"fromModule":     slotHeader.ModuleName,
	}).Info("block aggregator selected slot header")

	res = append(res, *slotHeader.Bid)

	return res, nil
}

func (b *BlockAggregatorService) GetPayload(VersionedSignedBlindedBeaconBlock *commonTypes.VersionedSignedBlindedBeaconBlock) (versionedExecutionPayload []commonTypes.VersionedExecutionPayloadWithVersionName, err error) {
	b.log.Info("Processing get payload request through block aggregator")

	base, err := VersionedSignedBlindedBeaconBlock.ToBaseSignedBlindedBeaconBlock()
	if err != nil {
		b.log.WithError(err).Error("error processing payload request, invalid VersionedSignedBlindedBeaconBlock")
		return versionedExecutionPayload, err
	}

	result, slotHeader, err := b.processPayloadReq(*VersionedSignedBlindedBeaconBlock)
	if err != nil {
		b.log.WithError(err).WithFields(logrus.Fields{
			"slot":          base.Message.Slot,
			"parentHash":    base.Message.ParentRoot.String(),
			"proposerIndex": base.Message.ProposerIndex,
			"blockHash":     base.Message.Body.ExecutionPayloadHeader.BlockHash.String(),
		}).Error("error processing payload request")
		return versionedExecutionPayload, err
	}

	if len(result) == 0 {

	}

	baseExecutionPayload, err := result[0].VersionedExecutionPayload.ToBaseExecutionPayload()
	if err != nil {
		b.log.WithError(err).Error("error processing payload request, invalid VersionedExecutionPayload returned")
		return versionedExecutionPayload, err
	}

	b.log.WithFields(logrus.Fields{
		"slot":                   base.Message.Slot,
		"parentHash":             base.Message.ParentRoot.String(),
		"proposerIndex":          base.Message.ProposerIndex,
		"blockHash_fromProposer": base.Message.Body.ExecutionPayloadHeader.BlockHash.String(),
		"blockHash_fromModule":   baseExecutionPayload.BlockHash.String(),
		"value":                  big.NewFloat(0).SetInt(slotHeader.Value).Quo(big.NewFloat(0).SetInt(slotHeader.Value), big.NewFloat(0).SetInt(big.NewInt(params.Ether))).String() + " ETH",
		"fromModule":             slotHeader.ModuleName,
	}).Info("block aggregator retrieved payload")

	return result, nil
}
