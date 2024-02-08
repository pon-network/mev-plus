package builderapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	commonTypes "github.com/bsn-eng/pon-golang-types/common"

	"github.com/attestantio/go-builder-client/spec"

	apiv1 "github.com/attestantio/go-builder-client/api/v1"
)

const (
	// Router paths
	pathRoot              = "/"
	pathStatus            = "/eth/v1/builder/status"
	pathRegisterValidator = "/eth/v1/builder/validators"
	pathGetHeader         = "/eth/v1/builder/header/{slot:[0-9]+}/{parent_hash:0x[a-fA-F0-9]+}/{pubkey:0x[a-fA-F0-9]+}"
	pathGetPayload        = "/eth/v1/builder/blinded_blocks"
)

type httpErrorResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var (
	nilResponse = struct{}{}
)

func (b *BuilderApiService) handleRoot(w http.ResponseWriter, _ *http.Request) {
	// Get call.
	b.respondOK(w, nilResponse)
}

func (b *BuilderApiService) handleStatus(w http.ResponseWriter, _ *http.Request) {
	// Get call.

	err := b.coreClient.Call(nil, "blockAggregator_status", false, nil)
	if err != nil {
		b.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	b.respondOK(w, nilResponse)
}

func (b *BuilderApiService) handleRegisterValidator(w http.ResponseWriter, req *http.Request) {
	// Post call.

	payload := []apiv1.SignedValidatorRegistration{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&payload)
	if err != nil {
		b.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = b.coreClient.Call(nil, "blockAggregator_registerValidator", false, nil, payload)
	if err != nil {
		b.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	b.respondOK(w, nilResponse)
}

func (b *BuilderApiService) handleGetHeader(w http.ResponseWriter, req *http.Request) {
	// Get call.

	vars := mux.Vars(req)

	slotStr := vars["slot"]
	slot, err := strconv.ParseUint(slotStr, 10, 64)
	if err != nil {
		b.respondError(w, http.StatusBadRequest, errInvalidSlotNumber.Error())
		return
	}

	pubkey := vars["pubkey"]
	if len(pubkey) != 98 {
		b.respondError(w, http.StatusBadRequest, errInvalidPubkey.Error())
		return
	}

	parentHash := vars["parent_hash"]
	if len(parentHash) != 66 {
		b.respondError(w, http.StatusBadRequest, errInvalidHash.Error())
		return
	}

	result := []spec.VersionedSignedBuilderBid{}
	err = b.coreClient.Call(&result, "blockAggregator_getHeader", false, nil, slot, parentHash, pubkey)
	if err != nil {
		b.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(result) == 0 {
		b.respondError(w, http.StatusNoContent, "")
	}

	if result[0].IsEmpty() {
		b.respondError(w, http.StatusNoContent, "")
		return
	}

	b.respondOK(w, &(result[0]))
}

func (b *BuilderApiService) handleGetPayload(w http.ResponseWriter, req *http.Request) {
	// Post call.
	payload := new(commonTypes.VersionedSignedBlindedBeaconBlock)
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		b.respondError(w, http.StatusBadRequest, fmt.Sprintf("invalid payload: %v", err))
		return
	}

	result := []commonTypes.VersionedExecutionPayloadWithVersionName{}
	err := b.coreClient.Call(&result, "blockAggregator_getPayload", false, nil, payload)
	if err != nil {
		b.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(result) == 0 {
		b.respondError(w, http.StatusInternalServerError, "blockAggregator returned no payload")
	}

	b.respondOK(w, &(result[0]))
}
