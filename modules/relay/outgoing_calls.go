package relay

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/attestantio/go-builder-client/spec"
	commonTypes "github.com/bsn-eng/pon-golang-types/common"
	"github.com/sirupsen/logrus"
)

func (r *RelayService) requestRelayHeader(slot uint64, parentHashHex, pubkey string, relay RelayEntry, log *logrus.Entry, mu *sync.Mutex, result *bidResp, relaysMap map[string][]RelayEntry) {
	path := fmt.Sprintf("/eth/v1/builder/header/%d/%s/%s", slot, parentHashHex, pubkey)
	url := relay.GetURI(path)
	log = log.WithField("url", url)
	responsePayload := new(spec.VersionedSignedBuilderBid)

	code, err := SendHTTPRequest(context.Background(), r.httpClient, http.MethodGet, url, nil, responsePayload)
	if err != nil {
		log.WithError(err).Warn("error making request to relay")
		return
	}

	if code == http.StatusNoContent {
		log.Debug("no-content response")
		return
	}

	// Skip if payload is empty
	if responsePayload.IsEmpty() {
		return
	}

	// Getting the bid info will check if there are missing fields in the response
	bidInfo, err := parseBidInfo(responsePayload)
	if err != nil {
		log.WithError(err).Warn("error parsing bid info")
		return
	}

	if bidInfo.blockHash == nilHash {
		log.Warn("relay responded with an empty block hash")
		return
	}

	log = log.WithFields(logrus.Fields{
		"blockHash": bidInfo.blockHash.String(),
		"txRoot":    bidInfo.txRoot.String(),
		"value":     bidInfo.value.String(),
	})

	if relay.PublicKey.String() != bidInfo.pubkey.String() {
		log.Errorf("bid pubkey mismatch. expected: %s - got: %s", relay.PublicKey.String(), bidInfo.pubkey.String())
		return
	}

	// Verify the relay signature in the relay response
	if r.relaySignatureCheck {
		ok, err := checkRelaySignature(responsePayload, relay.SigningDomain, relay.PublicKey)
		if err != nil {
			log.WithError(err).Error("error verifying relay signature")
			return
		}
		if !ok {
			log.Error("failed to verify relay signature")
			return
		}
	}

	// Verify response coherence with proposer's input data
	if bidInfo.parentHash.String() != parentHashHex {
		log.WithFields(logrus.Fields{
			"originalParentHash": parentHashHex,
			"responseParentHash": bidInfo.parentHash.String(),
		}).Error("proposer and relay parent hashes are not the same")
		return
	}

	if bidInfo.value.IsZero() {
		log.Warn("ignoring bid with 0 value")
		return
	}
	log.Debug("bid received")

	if bidInfo.value.CmpBig(r.relayMinBid.BigInt()) == -1 {
		log.Debug("ignoring bid below min-bid value")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Remember which relays delivered which bids (multiple relays might deliver the top bid)
	relaysMap[bidInfo.blockHash.String()] = append(relaysMap[bidInfo.blockHash.String()], relay)

	// Compare the bid with already known top bid (if any)
	if !result.response.IsEmpty() {
		valueDiff := bidInfo.value.Cmp(result.bidInfo.value)
		if valueDiff == -1 { // current bid is less profitable than already known one
			return
		} else if valueDiff == 0 { // current bid is equally profitable as already known one. Use hash as tiebreaker
			previousBidBlockHash := result.bidInfo.blockHash
			if bidInfo.blockHash.String() >= previousBidBlockHash.String() {
				return
			}
		}
	}

	// Use this relay's response as mev-plus response because it's most profitable
	log.Debug("new best bid")
	result.response = *responsePayload
	result.bidInfo = bidInfo
	result.t = time.Now()
}


func (r *RelayService) requestRelayPayload(relay RelayEntry, logger *logrus.Entry, block *commonTypes.VersionedSignedBlindedBeaconBlock, result *commonTypes.VersionedExecutionPayloadV2WithVersionName, mu *sync.Mutex, requestCtx context.Context, requestCtxCancel context.CancelFunc) {
	url := relay.GetURI(pathGetPayload)
	logger = logger.WithField("url", url)
	logger.Debug("calling getPayload")

	responsePayload := new(commonTypes.VersionedExecutionPayloadV2WithVersionName)
	_, err := SendHTTPRequestWithRetries(requestCtx, r.httpClient, http.MethodPost, url, block, responsePayload, r.cfg.RequestMaxRetries, logger)

	if err != nil {
		if errors.Is(requestCtx.Err(), context.Canceled) {
			logger.Info("Request was canceled")
		} else {
			logger.WithError(err).Error("Error making request to relay")
		}
		return
	}

	responsePayloadBase, err := responsePayload.VersionedExecutionPayload.ToBaseExecutionPayload()
	if err != nil {
		logger.WithError(err).Error("Error converting response payload to base")
		return
	}

	blockBase, err := block.ToBaseSignedBlindedBeaconBlock()
	if err != nil {
		logger.WithError(err).Error("Error converting block to base")
		return
	}

	// Ensure the response blockhash matches the request
	if blockBase.Message.Body.ExecutionPayloadHeader.BlockHash != responsePayloadBase.BlockHash {
		logger.WithFields(logrus.Fields{
			"responseBlockHash": responsePayloadBase.BlockHash.String(),
		}).Error("RequestBlockHash does not equal ResponseBlockHash")
		return
	}

	// In the case of transaction tampering, check that the response txRoot matches the request
	computedTxRoot, err := commonTypes.ComputeTransactionsRoot(responsePayloadBase.Transactions)
	if err != nil {
		logger.WithError(err).Error("Error computing txRoot")
		return
	}
	if computedTxRoot != blockBase.Message.Body.ExecutionPayloadHeader.TransactionsRoot {
		logger.WithFields(logrus.Fields{
			"computedTxRoot": computedTxRoot.String(),
			"responseTxRoot": blockBase.Message.Body.ExecutionPayloadHeader.TransactionsRoot.String(),
		}).Error("Expected TxRoot does not equal Response TxRoot")
		return
	}

	computedWithdrawalRoot, err := commonTypes.ComputeWithdrawalsRoot(responsePayloadBase.Withdrawals)
	if err != nil {
		logger.WithError(err).Error("Error computing withdrawalRoot")
		return
	}
	if computedWithdrawalRoot != blockBase.Message.Body.ExecutionPayloadHeader.WithdrawalsRoot {
		logger.WithFields(logrus.Fields{
			"computedWithdrawalRoot": computedWithdrawalRoot.String(),
			"responseWithdrawalRoot": blockBase.Message.Body.ExecutionPayloadHeader.WithdrawalsRoot.String(),
		}).Error("Expected WithdrawalRoot does not equal Response WithdrawalRoot")
		return
	}

	// ensure all the commitments, proofs and blobs are present
	blobsBundle := responsePayloadBase.BlobsBundle
	if len(blobsBundle.Commitments) != len(blobsBundle.Blobs) || len(blobsBundle.Commitments) != len(blobsBundle.Proofs) {
		err := fmt.Errorf("commitments, proofs and blobs are not of the same length")
		logger.WithFields(
			logrus.Fields{
				"commitments": len(blobsBundle.Commitments),
				"proofs":      len(blobsBundle.Proofs),
				"blobs":       len(blobsBundle.Blobs),
			}).WithError(
			err,
		).Errorf(
			"Wrong commitments, proofs and blobs returned from relay's get payload endpoint: %s", relay.String(),
		)
		return
	}
	for i, commitment := range blobsBundle.Commitments {
		if commitment != blockBase.Message.Body.BlobKZGCommitments[i] {
			err := fmt.Errorf("commitment returned from relay's get payload endpoint %s does not match the one provided at index %s", commitment.String(), blockBase.Message.Body.BlobKZGCommitments[i].String())
			logger.WithFields(
				logrus.Fields{
					"commitment": commitment.String(),
					"provided":   blockBase.Message.Body.BlobKZGCommitments[i].String(),
				}).WithError(
				err,
			).Errorf(
				"Wrong commitment returned from relay's get payload endpoint: %s", relay.String(),
			)
			continue
		}
	}

	mu.Lock()
	defer mu.Unlock()

	if requestCtx.Err() != nil { // Request has been canceled (or deadline exceeded)
		return
	}

	requestCtxCancel()

	if *result != (commonTypes.VersionedExecutionPayloadV2WithVersionName{}) {
		logger.Warn("Received payload from multiple relays. Ignoring subsequent ones")
		return
	}

	*result = *responsePayload
	logger.Info("Received payload from relay")
}