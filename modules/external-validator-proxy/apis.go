package externalvalidatorproxy

import (
	"context"
	"fmt"
	"net/http"

	commonTypes "github.com/bsn-eng/pon-golang-types/common"

	"github.com/attestantio/go-builder-client/spec"

	apiv1 "github.com/attestantio/go-builder-client/api/v1"
)

const (
	// Builder API paths
	pathStatus            = "/eth/v1/builder/status"
	pathRegisterValidator = "/eth/v1/builder/validators"
	pathGetHeader         = "/eth/v1/builder/header/%d/%s/%s"
	pathGetPayload        = "/eth/v1/builder/blinded_blocks"
)

func (p *ExternalValidatorProxyService) Status() error {
	path := pathStatus
	url := p.cfg.Address.String()
	url += path

	p.log.Info("Checking External Validator Proxy service status")

	// Send GET request to the proxy.
	code, err := SendHTTPRequest(context.Background(), p.httpClient, http.MethodGet, url, nil, nil)
	if err != nil {
		return err
	}

	if code == http.StatusNoContent {
		return nil
	}

	if code != http.StatusOK {
		return fmt.Errorf("status code %d", code)
	}

	return nil
}

func (p *ExternalValidatorProxyService) RegisterValidator(payload []apiv1.SignedValidatorRegistration) error {
	path := pathRegisterValidator
	url := p.cfg.Address.String()
	url += path

	p.log.Info("Registering validator with External Validator Proxy service")

	// Send POST request to the proxy.
	code, err := SendHTTPRequest(context.Background(), p.httpClient, http.MethodPost, url, payload, nil)
	if err != nil {
		return err
	}

	if code == http.StatusNoContent {
		return nil
	}

	if code != http.StatusOK {
		return fmt.Errorf("status code %d", code)
	}

	return nil
}

func (p *ExternalValidatorProxyService) GetHeader(slot uint64, parentHash, pubkey string) (res []spec.VersionedSignedBuilderBid, err error) {

	path := fmt.Sprintf(pathGetHeader, slot, parentHash, pubkey)
	url := p.cfg.Address.String()
	url += path

	p.log.Info("Getting header from External Validator Proxy service")

	// Send GET request to the proxy.
	response := new(spec.VersionedSignedBuilderBid)
	code, err := SendHTTPRequest(context.Background(), p.httpClient, http.MethodGet, url, nil, response)
	if err != nil {
		return res, err
	}

	if code != http.StatusOK {
		return res, fmt.Errorf("status code %d", code)
	}

	res = append(res, *response)

	return res, nil
}

func (p *ExternalValidatorProxyService) GetPayload(VersionedSignedBlindedBeaconBlock *commonTypes.VersionedSignedBlindedBeaconBlock) (versionedExecutionPayload []commonTypes.VersionedExecutionPayloadWithVersionName, err error) {

	path := pathGetPayload
	url := p.cfg.Address.String()
	url += path

	p.log.Info("Getting payload from External Validator Proxy service")

	// Send GET request to the proxy.
	response := new(commonTypes.VersionedExecutionPayloadWithVersionName)
	code, err := SendHTTPRequestWithRetries(context.Background(), p.httpClient, http.MethodGet, url, VersionedSignedBlindedBeaconBlock, response, p.cfg.RequestMaxRetries, p.log)
	if err != nil {
		return versionedExecutionPayload, err
	}

	if code != http.StatusOK {
		return versionedExecutionPayload, fmt.Errorf("status code %d", code)
	}

	versionedExecutionPayload = append(versionedExecutionPayload, *response)

	return versionedExecutionPayload, nil
}
