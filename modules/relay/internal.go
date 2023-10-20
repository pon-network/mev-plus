package relay

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	apiv1 "github.com/attestantio/go-builder-client/api/v1"
	commonTypes "github.com/bsn-eng/pon-golang-types/common"
	"github.com/sirupsen/logrus"
)

func (r *RelayService) processRegistration(payload []apiv1.SignedValidatorRegistration) error {
	log := r.log.WithField("method", "ProcessRegistration")
	log.Debug("Handling Validator Registration")

	log = log.WithFields(logrus.Fields{
		"numRegistrations": len(payload),
	})

	relayRespCh := make(chan error, len(r.relays))

	for _, relay := range r.relays {
		go func(relayEntry RelayEntry) {
			url := relayEntry.GetURI(pathRegisterValidator)
			log := log.WithField("url", url)

			_, err := SendHTTPRequest(context.Background(), r.httpClient, http.MethodPost, url, payload, nil)
			relayRespCh <- err
			if err != nil {
				log.WithError(err).Warn("Error while calling relay's registration endpoint")
				return
			}
		}(relay)
	}

	for i := 0; i < len(r.relays); i++ {
		respErr := <-relayRespCh
		if respErr == nil {
			return nil
		}
	}
	return nil
}

func (r *RelayService) processGetHeader(slot uint64, parentHashHex, pubkey string) (bidResp, error) {
	log := r.log.WithFields(logrus.Fields{
		"method":     "ProcessHeader",
		"slot":       slot,
		"parentHash": parentHashHex,
		"pubkey":     pubkey,
	})

	// Check pubkey and parentHashHex lengths
	if len(pubkey) != 98 || len(parentHashHex) != 66 {
		return bidResp{}, ErrInvalidInput
	}

	log.Info("Relay handling GetHeader request")

	result := bidResp{}                     // the final response, containing the highest bid (if any)
	relays := make(map[string][]RelayEntry) // relays that sent the bid for a specific blockHash

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, relay := range r.relays {
		wg.Add(1)
		go func(relay RelayEntry) {
			defer wg.Done()
			r.requestRelayHeader(slot, parentHashHex, pubkey, relay, log, &mu, &result, relays)
		}(relay)
	}
	wg.Wait()

	result.relays = relays[result.bidInfo.blockHash.String()]

	if result.response.IsEmpty() {
		log.Info("No valid bid received")
		return bidResp{}, ErrNoBidReceived
	}

	bidKey := bidRespKey{slot: slot, blockHash: result.bidInfo.blockHash.String()}
	r.bidsLock.Lock()
	r.bids[bidKey] = result

	// Remove old bids from the map
	for k := range r.bids {
		if k.slot < slot-64 { // clean bids older than 2 epochs
			delete(r.bids, k)
		}
	}

	r.bidsLock.Unlock()
	return result, nil
}


func (r *RelayService) processGetPayload(block commonTypes.VersionedSignedBlindedBeaconBlock) (versionedExecutionPayload []commonTypes.VersionedExecutionPayloadWithVersionName, err error) {
	log := r.log.WithField("method", "getPayload")

	blockBase, err := block.ToBaseSignedBlindedBeaconBlock()
	if err != nil {
		return nil, err
	}

	if err := validatePayloadBlock(blockBase, log); err != nil {
		return nil, err
	}

	logger := log.WithFields(logrus.Fields{
		"slot":       blockBase.Message.Slot,
		"blockHash":  blockBase.Message.Body.ExecutionPayloadHeader.BlockHash.String(),
		"parentHash": blockBase.Message.Body.ExecutionPayloadHeader.ParentHash.String(),
	})

	r.bidsLock.Lock()
	originalBid := r.bids[bidRespKey{slot: uint64(blockBase.Message.Slot), blockHash: blockBase.Message.Body.ExecutionPayloadHeader.BlockHash.String()}]
	r.bidsLock.Unlock()

	if err := validateOriginalBid(logger, originalBid); err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var result commonTypes.VersionedExecutionPayloadWithVersionName

	requestCtx, requestCtxCancel := context.WithCancel(context.Background())
	defer requestCtxCancel()

	// Get the payload from each relay
	for _, relay := range originalBid.relays {
		wg.Add(1)
		go func(relay RelayEntry) {
			defer wg.Done()
			r.requestRelayPayload(relay, logger, &block, &result, &mu, requestCtx, requestCtxCancel)
		}(relay)
	}

	wg.Wait()

	if result == (commonTypes.VersionedExecutionPayloadWithVersionName{}) {
		originRelays := RelayEntriesToStrings(originalBid.relays)
		logger.WithField("relaysWithBid", strings.Join(originRelays, ", ")).Error("No payload received from any relay!")
		return nil, ErrNoPayloadReceived
	}

	res := []commonTypes.VersionedExecutionPayloadWithVersionName{result}

	return res, nil
}


// CheckRelays sends a request to each one of the relays previously registered to get their status
func (r *RelayService) checkRelays() int {
	var wg sync.WaitGroup
	var numSuccessRequestsToRelay uint32

	for _, relay := range r.relays {
		wg.Add(1)

		go func(relay RelayEntry) {
			defer wg.Done()
			url := relay.GetURI(pathStatus)
			log := r.log.WithField("url", url)
			log.Debug("checking relay status")

			code, err := SendHTTPRequest(context.Background(), r.httpClient, http.MethodGet, url, nil, nil)
			if err != nil {
				log.WithError(err).Error("relay status error - request failed")
				return
			}
			if code == http.StatusOK {
				log.Debug("relay status OK")
			} else {
				log.Errorf("relay status error - unexpected status code %d", code)
				return
			}

			atomic.AddUint32(&numSuccessRequestsToRelay, 1)
		}(relay)
	}

	wg.Wait()
	return int(numSuccessRequestsToRelay)
}
