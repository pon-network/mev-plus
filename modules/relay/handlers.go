package relay

import (
	"fmt"

	apiv1 "github.com/attestantio/go-builder-client/api/v1"
	"github.com/attestantio/go-builder-client/spec"
	commonTypes "github.com/bsn-eng/pon-golang-types/common"
)

func (r *RelayService) Status() error {

	r.log.Info("Checking Relay service status")
	if r.relayCheck {
		r.log.Info("Checking relays for availability")
		ok := r.checkRelays()
		if ok <= 0 {
			return fmt.Errorf("failed to connect to any relays")
		}
		r.log.Info(ok, " relays running and available")
	}

	// If this call was successful
	// then service is accessible, return nil
	return nil
}

func (r *RelayService) RegisterValidator(payload []apiv1.SignedValidatorRegistration) error {
	return r.processRegistration(payload)
}

func (r *RelayService) GetHeader(slot uint64, parentHash, pubkey string) (res []spec.VersionedSignedBuilderBid, err error) {
	result, err := r.processGetHeader(slot, parentHash, pubkey)
	if err != nil {
		return res, err
	}

	return []spec.VersionedSignedBuilderBid{result.response}, nil
}

func (r *RelayService) GetPayload(VersionedSignedBlindedBeaconBlock *commonTypes.VersionedSignedBlindedBeaconBlock) (versionedExecutionPayload []commonTypes.VersionedExecutionPayloadV2WithVersionName, err error) {
	return r.processGetPayload(*VersionedSignedBlindedBeaconBlock)
}
