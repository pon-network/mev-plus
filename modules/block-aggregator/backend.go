package blockaggregator

import (
	"context"
	"fmt"
	"sync"
	"time"

	apiv1 "github.com/attestantio/go-builder-client/api/v1"
	"github.com/attestantio/go-builder-client/spec"
	"github.com/pon-network/mev-plus/modules/block-aggregator/data"

	commonTypes "github.com/bsn-eng/pon-golang-types/common"
)

// This is used by builder api to check the status of any block producers or relays
func (b *BlockAggregatorService) checkBlockSources() error {

	var err error
	var wg sync.WaitGroup
	var sourcesUp []string
	var sourcesDown []string
	var mu sync.Mutex

	handleStatusCheck := func(module string) {

		defer wg.Done()

		err := b.coreClient.Call(nil, module+"_status", false, nil)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			b.log.WithError(err).WithField("module", module).Warn("error calling module")
			sourcesDown = append(sourcesDown, module)
			return
		}

		b.log.WithField("module", module).Info("Module is up")
		sourcesUp = append(sourcesUp, module)

	}

	for _, module := range b.ConnectedBLockSources {
		wg.Add(1)
		go handleStatusCheck(module)
	}

	wg.Wait()

	if len(b.ConnectedBLockSources) == 0 {
		err = nil
		b.log.Info("no block sources are configured")
	} else if len(sourcesUp) > 0 {
		b.log.WithField("sources", sourcesUp).Info("block sources are up")
	} else {
		err = fmt.Errorf("no module block sources are up")
		b.log.WithError(err).Error("no block sources are up")
	}

	return err
}

func (b *BlockAggregatorService) processValidatorRegistrations(payload []apiv1.SignedValidatorRegistration) error {

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error
	var successfulRegistrations []string

	// Notify all modules of the new validator registrations
	_ = b.coreClient.Notify(context.Background(), "core_registerValidator", true, b.ConnectedBLockSources, payload)

	handleRegistration := func(module string) {

		defer wg.Done()

		err := b.coreClient.Call(nil, module+"_registerValidator", false, b.ConnectedBLockSources, payload)
		if err != nil {
			b.log.WithError(err).WithField("module", module).Warn("error calling module")
			mu.Lock()
			defer mu.Unlock()
			errors = append(errors, err)
			return
		}
		successfulRegistrations = append(successfulRegistrations, module)
		b.log.WithField("module", module).Info("successfully registered validator")
	}

	for _, module := range b.ConnectedBLockSources {
		wg.Add(1)
		go handleRegistration(module)
	}

	wg.Wait()

	if len(successfulRegistrations) == 0 {
		return fmt.Errorf("failed to process validator registrations: %v", errors)
	}

	return nil
}

func (b *BlockAggregatorService) processHeaderReq(slot uint64, parentHash, proposerPubkey string) (data.SlotHeader, error) {

	slotTime := b.cfg.GenesisTime + (slot * b.cfg.SlotDuration)
	auctionDeadline := slotTime + b.cfg.AuctionDuration

	// Calculate the time remaining until the auction deadline
	timeUntilDeadline := time.Until(time.Unix(int64(auctionDeadline), 0))
	if timeUntilDeadline > 0 {
		// Sleep until the auction deadline has passed
		time.Sleep(timeUntilDeadline)
	}

	// Notify all modules of the new slot header request
	_ = b.coreClient.Notify(context.Background(), "core_getHeader", true, b.ConnectedBLockSources, slot, parentHash, proposerPubkey)

	var wg sync.WaitGroup
	type resultData struct {
		module   string
		response spec.VersionedSignedBuilderBid
	}
	resultChan := make(chan resultData, len(b.ConnectedBLockSources))

	handleModule := func(module string) {
		defer wg.Done()

		var result []spec.VersionedSignedBuilderBid
		err := b.coreClient.Call(&result, module+"_getHeader", false, b.ConnectedBLockSources, slot, parentHash, proposerPubkey)
		if err != nil {
			b.log.WithError(err).WithField("module", module).Warn("error calling module")
			return
		}

		if len(result) == 0 {
			b.log.WithField("module", module).Warn("module returned no header response")
			return
		}

		if result[0].IsEmpty() {
			b.log.WithField("module", module).Warn("module returned empty header response")
			return
		}

		b.log.WithField("module", module).Info("module returned header response")
		resultChan <- resultData{module, result[0]}
	}

	for _, module := range b.ConnectedBLockSources {
		wg.Add(1)
		go handleModule(module)
	}

	wg.Wait()
	close(resultChan)

	for result := range resultChan {
		err := b.processNewBid(result.module, slot, result.response)
		if err != nil {
			return data.SlotHeader{}, err
		}
	}

	slotHeader, err := b.Data.GetSelectedSlotHeaders(slot)
	if err != nil {
		return data.SlotHeader{}, err
	}

	return slotHeader, nil
}

func (b *BlockAggregatorService) processPayloadReq(VersionedSignedBlindedBeaconBlock commonTypes.VersionedSignedBlindedBeaconBlock) (versionedExecutionPayload []commonTypes.VersionedExecutionPayloadWithVersionName, slotHeader data.SlotHeader, err error) {

	baseSignedBlindedBeaconBlock, err := VersionedSignedBlindedBeaconBlock.ToBaseSignedBlindedBeaconBlock()
	if err != nil {
		return versionedExecutionPayload, slotHeader, err
	}

	slotHeader, err = b.Data.GetSlotHeaderByHash(baseSignedBlindedBeaconBlock.Message.Body.ExecutionPayloadHeader.BlockHash.String())
	if err != nil {
		return versionedExecutionPayload, slotHeader, err
	}
	var result []commonTypes.VersionedExecutionPayloadWithVersionName
	err = b.coreClient.Call(&result, slotHeader.ModuleName+"_getPayload", true, b.ConnectedBLockSources, &VersionedSignedBlindedBeaconBlock)
	if err != nil {
		return versionedExecutionPayload, slotHeader, err
	}

	return result, slotHeader, nil

}
