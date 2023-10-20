package relay

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	commonType "github.com/bsn-eng/mev-plus/common"
	coreCommon "github.com/bsn-eng/mev-plus/core/common"
	"github.com/bsn-eng/mev-plus/modules/relay/common"
	"github.com/bsn-eng/mev-plus/modules/relay/signing"
	"github.com/bsn-eng/mev-plus/modules/relay/config"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	errServerAlreadyRunning = errors.New("server already running")
)

const (
	HeaderKeySlotUID = "X-MEVPLUS-SlotID"
	HeaderKeyVersion = "X-MEVPLUS-Version"
)

type slotUID struct {
	slot uint64
	uid  uuid.UUID
}

var (
	nilHash     = phase0.Hash32{}
	nilResponse = struct{}{}
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

	httpClient           http.Client
	requestMaxRetries    int
	bids                 map[bidRespKey]bidResp // keeping track of bids, to log the originating relay on withholding
	bidsLock             sync.Mutex
}

func NewRelayService() *RelayService {

	log := logrus.NewEntry(logrus.New())

	return &RelayService{
		log:                 log,
		relays:              []RelayEntry{},
		cfg:                 config.RelayConfigDefaults,
		relayCheck:          config.RelayConfigDefaults.RelayCheck,
		relaySignatureCheck: config.RelayConfigDefaults.RelaySignatureCheck,
		requestMaxRetries:   config.RelayConfigDefaults.RequestMaxRetries,
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

	r.log.Info("Starting Relay service")
	r.log.Info("Relay check: ", r.relayCheck)
	r.log.Info("Relay min bid: ", r.relayMinBid.BigInt().String())
	r.log.Info("Configured relays: ", strings.Join(RelayEntriesToStrings(r.relays), ", "))

	if r.relayCheck {
		ok := r.checkRelays()
		if ok <= 0 {
			return fmt.Errorf("failed to connect to any relays")
		}
		r.log.Info("Connected to ", ok, " relays")
	}

	ctx := context.Background()
	err := r.coreClient.Notify(ctx, "blockAggregator_connectBlockSource", false, config.ModuleName)
	if err != nil {
		return err
	}

	return nil
}

func (r *RelayService) Stop() error {

	return nil
}
