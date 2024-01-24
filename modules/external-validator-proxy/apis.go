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

	p.log.Info("Checking External Validator Proxy service status")

	var respErr error
	proxyRespCh := make(chan error, len(p.cfg.Addresses))

	for _, proxy := range p.cfg.Addresses {
		go func(proxy string) {
			url := proxy + pathStatus
			code, err := SendHTTPRequest(context.Background(), p.httpClient, http.MethodGet, url, nil, nil)
			if err != nil {
				p.log.WithError(err).Warnf("Error while calling proxy's status endpoint: %s", proxy)
				proxyRespCh <- err
				return
			} else if code == http.StatusNoContent {
				proxyRespCh <- nil
				return
			} else if code != http.StatusOK {
				p.log.WithError(err).Warnf("Error while calling proxy's status endpoint: %s", proxy)
				proxyRespCh <- fmt.Errorf("status code %d", code)
				return
			}
			proxyRespCh <- nil
		}(proxy.String())
	}

	for i := 0; i < len(p.cfg.Addresses); i++ {
		respErr = <-proxyRespCh
		if respErr == nil {
			// If any of the proxies is up,
			return nil
		}
	}

	return respErr

}

func (p *ExternalValidatorProxyService) RegisterValidator(payload []apiv1.SignedValidatorRegistration) error {

	p.log.Info("Registering validator with External Validator Proxy service")

	var respErr error
	proxyRespCh := make(chan error, len(p.cfg.Addresses))

	for _, proxy := range p.cfg.Addresses {
		go func(proxy string) {
			url := proxy + pathRegisterValidator
			code, err := SendHTTPRequest(context.Background(), p.httpClient, http.MethodPost, url, payload, nil)
			if err != nil {
				p.log.WithError(err).Warnf("Error while calling proxy's validator registration endpoint: %s", proxy)
				proxyRespCh <- err
				return
			} else if code == http.StatusNoContent {
				proxyRespCh <- nil
				return
			} else if code != http.StatusOK {
				p.log.WithError(err).Warnf("Error while calling proxy's validator registration endpoint: %s", proxy)
				proxyRespCh <- fmt.Errorf("status code %d", code)
				return
			}
			proxyRespCh <- nil
		}(proxy.String())
	}

	for i := 0; i < len(p.cfg.Addresses); i++ {
		respErr = <-proxyRespCh
		if respErr == nil {
			// If any of the proxies is up,
			return nil
		}
	}

	return respErr

}

func (p *ExternalValidatorProxyService) GetHeader(slot uint64, parentHash, pubkey string) (res []spec.VersionedSignedBuilderBid, err error) {

	p.log.Info("Getting header from External Validator Proxy service")

	var respErr error
	proxyErrRespCh := make(chan error, len(p.cfg.Addresses))
	proxyRespCh := make(chan spec.VersionedSignedBuilderBid, len(p.cfg.Addresses))

	for _, proxy := range p.cfg.Addresses {
		go func(proxy string) {
			url := proxy + fmt.Sprintf(pathGetHeader, slot, parentHash, pubkey)
			response := new(spec.VersionedSignedBuilderBid)
			code, err := SendHTTPRequest(context.Background(), p.httpClient, http.MethodGet, url, nil, response)
			if err != nil {
				p.log.WithError(err).Warnf("Error while calling proxy's get header endpoint: %s", proxy)
				proxyErrRespCh <- err
				proxyRespCh <- spec.VersionedSignedBuilderBid{}
				return
			} else if code != http.StatusOK {
				p.log.WithError(err).Warnf("Error while calling proxy's get header endpoint: %s", proxy)
				proxyErrRespCh <- fmt.Errorf("status code %d", code)
				proxyRespCh <- spec.VersionedSignedBuilderBid{}
				return
			}
			proxyRespCh <- *response
			proxyErrRespCh <- nil
		}(proxy.String())
	}

	for i := 0; i < len(p.cfg.Addresses); i++ {
		respErr = <-proxyErrRespCh
		resp := <-proxyRespCh
		if respErr == nil {
			// Check if the response is empty
			if resp.IsEmpty() {
				continue
			}

			// If any of the proxies returns a header append it to the response
			res = append(res, resp)
		}
	}

	if len(res) == 0 {
		// If none of the proxies returns a header, return an error that may have occured
		// If no error occured, return a generic error
		if respErr == nil {
			return res, fmt.Errorf("no header returned")
		}
		return res, respErr
	}

	return res, nil

}

func (p *ExternalValidatorProxyService) GetPayload(VersionedSignedBlindedBeaconBlock *commonTypes.VersionedSignedBlindedBeaconBlock) (versionedExecutionPayload []commonTypes.VersionedExecutionPayloadWithVersionName, err error) {

	p.log.Info("Getting payload from External Validator Proxy service")

	baseSignedBlindedBeaconBlock, err := VersionedSignedBlindedBeaconBlock.ToBaseSignedBlindedBeaconBlock()
	if err != nil {
		return versionedExecutionPayload, err
	}

	var respErr error
	proxyErrRespCh := make(chan error, len(p.cfg.Addresses))
	proxyRespCh := make(chan commonTypes.VersionedExecutionPayloadWithVersionName, len(p.cfg.Addresses))

	for _, proxy := range p.cfg.Addresses {
		go func(proxy string) {
			url := proxy + pathGetPayload
			response := new(commonTypes.VersionedExecutionPayloadWithVersionName)
			code, err := SendHTTPRequestWithRetries(context.Background(), p.httpClient, http.MethodPost, url, VersionedSignedBlindedBeaconBlock, response, p.cfg.RequestMaxRetries, p.log)
			if err != nil {
				p.log.WithError(err).Debugf("Error while calling proxy's get payload endpoint: %s", proxy)
				proxyErrRespCh <- err
				proxyRespCh <- commonTypes.VersionedExecutionPayloadWithVersionName{}
				return
			} else if code != http.StatusOK {
				// Do not log an error here as the proxy software may throw an error about
				// an invalid payload header that was not provided by that proxy since not tracking
				// headers accross proxies like relays, as two proxies may get the same header from being
				// connected to the relay or not depending on users configurations
				proxyErrRespCh <- fmt.Errorf("status code %d", code)
				proxyRespCh <- commonTypes.VersionedExecutionPayloadWithVersionName{}
				return
			}
			proxyRespCh <- *response
			proxyErrRespCh <- nil
		}(proxy.String())
	}

	for i := 0; i < len(p.cfg.Addresses); i++ {
		respErr = <-proxyErrRespCh
		resp := <-proxyRespCh
		if respErr == nil {
			// Check if the response is empty
			if resp.VersionName == "" {
				continue
			}

			baseExecutionPayload, err := resp.VersionedExecutionPayload.ToBaseExecutionPayload()
			if err != nil {
				p.log.WithError(err).Debugf("Error while extracting baseExecutionPayload from proxy's get payload endpoint: %s", p.cfg.Addresses[i].String())
				continue
			}
			// check if hash matches
			if baseExecutionPayload.BlockHash.String() != baseSignedBlindedBeaconBlock.Message.Body.ExecutionPayloadHeader.BlockHash.String() {
				err := fmt.Errorf("blockHash returned from proxy's get payload endpoint %s does not match the one provided %s", baseExecutionPayload.BlockHash.String(), baseSignedBlindedBeaconBlock.Message.Body.ExecutionPayloadHeader.BlockHash.String())
				p.log.WithError(err).Debugf("Wrong blockHash returned from proxy's get payload endpoint: %s", p.cfg.Addresses[i].String())
				continue
			}

			// If any of the proxies returns a payload append it to the response
			versionedExecutionPayload = append(versionedExecutionPayload, resp)
		}
	}

	if len(versionedExecutionPayload) == 0 {
		// If none of the proxies returns a payload, return an error that may have occured
		// If no error occured, return a generic error
		if respErr == nil {
			return versionedExecutionPayload, fmt.Errorf("no payload returned")
		}
		return versionedExecutionPayload, respErr
	}

	return versionedExecutionPayload, nil

}