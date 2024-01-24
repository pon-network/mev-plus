package relay

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	commonType "github.com/pon-network/mev-plus/common"
	coreCommon "github.com/pon-network/mev-plus/core/common"
	"github.com/pon-network/mev-plus/modules/relay/common"
	"github.com/pon-network/mev-plus/modules/relay/config"
	"github.com/pon-network/mev-plus/modules/relay/signing"
	"github.com/sirupsen/logrus"
)

const (
	HeaderKeySlotUID = "X-MEVPLUS-SlotID"
	HeaderKeyVersion = "X-MEVPLUS-Version"
)

var (
	nilHash = phase0.Hash32{}
)

const (
	// Router paths
	pathStatus            = "/eth/v1/builder/status"
	pathRegisterValidator = "/eth/v1/builder/validators"
	pathGetHeader         = "/eth/v1/builder/header/{slot:[0-9]+}/{parent_hash:0x[a-fA-F0-9]+}/{pubkey:0x[a-fA-F0-9]+}"
	pathGetPayload        = "/eth/v1/builder/blinded_blocks"
)

type RelayService struct {
	relays     []RelayEntry
	coreClient *coreCommon.Client
	cfg        config.RelayConfig

	log                 *logrus.Entry
	relayCheck          bool
	relaySignatureCheck bool
	relayMinBid         common.U256Str
	genesisTime         uint64

	httpClient http.Client
	bids       map[bidRespKey]bidResp // keeping track of bids, to log the originating relay on withholding
	bidsLock   sync.Mutex
}

func NewRelayService() *RelayService {

	log := logrus.NewEntry(logrus.New()).WithField("moduleExecution", config.ModuleName)

	return &RelayService{
		log:                 log,
		relays:              []RelayEntry{},
		cfg:                 config.RelayConfigDefaults,
		relayCheck:          config.RelayConfigDefaults.RelayCheck,
		relaySignatureCheck: config.RelayConfigDefaults.RelaySignatureCheck,
		bids:                make(map[bidRespKey]bidResp),
		httpClient:          http.Client{Timeout: time.Duration(config.RelayConfigDefaults.RequestTimeoutMs) * time.Millisecond},
	}
}

func (r *RelayService) Name() string {
	return config.ModuleName
}

func (r *RelayService) ConnectCore(coreClient *coreCommon.Client, pingId string) error {

	r.coreClient = coreClient
	err := r.coreClient.Ping(pingId)
	if err != nil {
		return err
	}

	return nil
}

func (r *RelayService) Configure(moduleFlags commonType.ModuleFlags) error {

	// Load custom set config flags to replace any initialized config defaults
	err := ParseConfigFLags(r, moduleFlags)
	if err != nil {
		return err
	}

	if r.relaySignatureCheck {
		domain, err := signing.ComputeDomain(signing.DomainTypeAppBuilder, r.cfg.GenesisForkVersion, r.cfg.GenesisValidatorsRoot)
		if err != nil {
			return err
		}
		var domainPhase0 phase0.Domain
		copy(domainPhase0[:], domain[:])
		for _, relayEntry := range r.relays {
			relayEntry.SigningDomain = domainPhase0
		}
	}

	r.httpClient.CheckRedirect = httpClientDisallowRedirects

	if len(r.relays) == 0 {
		return fmt.Errorf("no relay entries provided")
	}

	return nil
}

func (r *RelayService) Start() error {

	if len(r.relays) == 0 {
		// No relays configured do not connect and start the service
		ctx := context.Background()
		err := r.coreClient.Notify(ctx, "blockAggregator_excludeFromNotifications", false, nil, config.ModuleName)
		if err != nil {
			return err
		}
		return nil
	}

	var operationalRelays int
	if r.relayCheck {
		ok := r.checkRelays()
		if ok <= 0 {
			return fmt.Errorf("failed to connect to any relays")
		}
		operationalRelays = ok
	}

	var builderApiAddress string
	err := r.coreClient.Call(&builderApiAddress, "builderApi_listenAddress", false, nil)
	if err != nil {
		return err
	}

	for _, relayEntry := range r.relays {
		if relayEntry.URL.String() == builderApiAddress {
			return fmt.Errorf("relay address %s is the same as the builder api address %s", relayEntry.URL.String(), builderApiAddress)
		}
	}

	ctx := context.Background()
	err = r.coreClient.Notify(ctx, "blockAggregator_connectBlockSource", false, nil, config.ModuleName)
	if err != nil {
		return err
	}

	r.log.Info("Started Relay service")
	r.log.Info("Relay min bid: ", r.relayMinBid.BigInt().String())
	r.log.Info("Configured relays: ", strings.Join(RelayEntriesToStrings(r.relays), ", "))
	r.log.Infof("Using %d operational relays", operationalRelays)

	return nil
}

func (r *RelayService) Stop() error {

	return nil
}
