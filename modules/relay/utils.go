package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	commonTypes "github.com/bsn-eng/pon-golang-types/common"
	commonType "github.com/pon-network/mev-plus/common"
	relayCommon "github.com/pon-network/mev-plus/modules/relay/common"
	"github.com/pon-network/mev-plus/modules/relay/config"

	"github.com/attestantio/go-builder-client/spec"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381"
	"github.com/crate-crypto/go-ipa/bandersnatch/fr"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/holiman/uint256"
	"github.com/sirupsen/logrus"
)

// bidResp are entries in the bids cache
type bidResp struct {
	t        time.Time
	response spec.VersionedSignedBuilderBid
	bidInfo  bidInfo
	relays   []RelayEntry
}

// bidRespKey is used as key for the bids cache
type bidRespKey struct {
	slot      uint64
	blockHash string
}

type (
	PublicKey = bls12381.G1Affine
	SecretKey = fr.Element
	Signature = bls12381.G2Affine
)

const (
	PublicKeyLength = bls12381.SizeOfG1AffineCompressed
	SecretKeyLength = fr.Bytes
	SignatureLength = bls12381.SizeOfG2AffineCompressed
)

var (
	_, _, g1Aff, _     = bls12381.Generators()
	domain                    = []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_")
	ErrInvalidPubkeyLength    = errors.New("invalid public key length")
	ErrInvalidSecretKeyLength = errors.New("invalid secret key length")
	ErrInvalidSignatureLength = errors.New("invalid signature length")
	ErrSecretKeyIsZero        = errors.New("invalid secret key is zero")
)

// bidInfo is used to store bid response fields for logging and validation
type bidInfo struct {
	blockHash  phase0.Hash32
	parentHash phase0.Hash32
	pubkey     phase0.BLSPubKey
	txRoot     phase0.Root
	value      *uint256.Int
}

// GetURI returns the full request URI with scheme, host, path and args.
func GetURI(url *url.URL, path string) string {
	u2 := *url
	u2.User = nil
	u2.Path = path
	return u2.String()
}

// HexToPubkey takes a hex string and returns a PublicKey
func HexToPubkey(s string) (ret phase0.BLSPubKey, err error) {
	bytes, err := hexutil.Decode(s)
	if len(bytes) != len(ret) {
		return phase0.BLSPubKey{}, ErrLength
	}
	copy(ret[:], bytes)
	return
}

func parseBidInfo(bid *spec.VersionedSignedBuilderBid) (bidInfo, error) {
	blockHash, err := bid.BlockHash()
	if err != nil {
		return bidInfo{}, err
	}
	parentHash, err := bid.ParentHash()
	if err != nil {
		return bidInfo{}, err
	}
	pubkey, err := bid.Builder()
	if err != nil {
		return bidInfo{}, err
	}
	txRoot, err := bid.TransactionsRoot()
	if err != nil {
		return bidInfo{}, err
	}
	value, err := bid.Value()
	if err != nil {
		return bidInfo{}, err
	}
	bidInfo := bidInfo{
		blockHash:  blockHash,
		parentHash: parentHash,
		pubkey:     pubkey,
		txRoot:     txRoot,
		value:      value,
	}
	return bidInfo, nil
}

// DecodeJSON reads JSON from io.Reader and decodes it into a struct
func DecodeJSON(r io.Reader, dst any) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

// SendHTTPRequest - prepare and send HTTP request, marshaling the payload if any, and decoding the response if dst is set
func SendHTTPRequest(ctx context.Context, client http.Client, method, url string, payload, dst any) (code int, err error) {
	var req *http.Request

	if payload == nil {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	} else {
		payloadBytes, err2 := json.Marshal(payload)
		if err2 != nil {
			return 0, fmt.Errorf("could not marshal request: %w", err2)
		}
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewReader(payloadBytes))
	}
	if err != nil {
		return 0, fmt.Errorf("could not prepare request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return resp.StatusCode, nil
	}

	if resp.StatusCode > 299 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp.StatusCode, fmt.Errorf("could not read error response body for status code %d: %w", resp.StatusCode, err)
		}
		return resp.StatusCode, fmt.Errorf("%w: %d / %s", ErrHTTPErrorResponse, resp.StatusCode, string(bodyBytes))
	}

	if dst != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp.StatusCode, fmt.Errorf("could not read response body: %w", err)
		}

		if err := json.Unmarshal(bodyBytes, dst); err != nil {
			return resp.StatusCode, fmt.Errorf("could not unmarshal response %s: %w", string(bodyBytes), err)
		}
	}

	return resp.StatusCode, nil
}

func checkRelaySignature(bid *spec.VersionedSignedBuilderBid, domain phase0.Domain, pubKey phase0.BLSPubKey) (bool, error) {
	root, err := bid.MessageHashTreeRoot()
	if err != nil {
		return false, err
	}
	sig, err := bid.Signature()
	if err != nil {
		return false, err
	}
	signingData := phase0.SigningData{ObjectRoot: root, Domain: domain}
	msg, err := signingData.HashTreeRoot()
	if err != nil {
		return false, err
	}

	return VerifySignatureBytes(msg[:], sig[:], pubKey[:])
}

// SendHTTPRequestWithRetries - prepare and send HTTP request, retrying the request if within the client timeout
func SendHTTPRequestWithRetries(ctx context.Context, client http.Client, method, url string, payload, dst any, maxRetries int, log *logrus.Entry) (code int, err error) {
	// Create a context with a timeout as configured in the HTTP client
	requestCtx, cancel := context.WithTimeout(ctx, client.Timeout)
	defer cancel()

	for attempts := 1; attempts <= maxRetries; attempts++ {
		if requestCtx.Err() != nil {
			return 0, fmt.Errorf("request context error after %d attempts: %w", attempts, requestCtx.Err())
		}

		code, err = SendHTTPRequest(ctx, client, method, url, payload, dst)
		if err == nil {
			return code, nil
		}

		log.WithError(err).Warn("Error making request to relay, retrying")
		time.Sleep(100 * time.Millisecond)
	}

	return 0, ErrMaxRetriesExceeded
}

func httpClientDisallowRedirects(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}

func VerifySignatureBytes(msg, sigBytes, pkBytes []byte) (bool, error) {
	pk, err := PublicKeyFromBytes(pkBytes)
	if err != nil {
		return false, err
	}
	sig, err := SignatureFromBytes(sigBytes)
	if err != nil {
		return false, err
	}
	return VerifySignature(sig, pk, msg)
}

func PublicKeyFromBytes(pkBytes []byte) (*PublicKey, error) {
	if len(pkBytes) != PublicKeyLength {
		return nil, ErrInvalidPubkeyLength
	}
	pk := new(PublicKey)
	err := pk.Unmarshal(pkBytes)
	return pk, err
}

func SignatureFromBytes(sigBytes []byte) (*Signature, error) {
	if len(sigBytes) != SignatureLength {
		return nil, ErrInvalidSignatureLength
	}
	sig := new(Signature)
	err := sig.Unmarshal(sigBytes)
	return sig, err
}

func VerifySignature(sig *Signature, pk *PublicKey, msg []byte) (bool, error) {
	Q, err := bls12381.HashToG2(msg, domain)
	if err != nil {
		return false, err
	}
	var negP bls12381.G1Affine
	negP.Neg(&g1Aff)
	return bls12381.PairingCheck(
		[]bls12381.G1Affine{*pk, negP},
		[]bls12381.G2Affine{Q, *sig},
	)
}

func ParseConfigFLags(r *RelayService, moduleFlags commonType.ModuleFlags) error {

	var forkVersionFlagNameSet string
	var customGenesisTime bool
	var customForkVersion bool

	for flagName, flagValue := range moduleFlags {
		switch flagName {
		case config.LoggerLevelFlag.Name:
			logLevel, err := logrus.ParseLevel(flagValue)
			if err != nil {
				return err
			}
			r.log.Logger.SetLevel(logLevel)
		case config.LoggerFormatFlag.Name:
			switch flagValue {
			case "json":
				r.log.Logger.SetFormatter(&logrus.JSONFormatter{})
			case "text":
				r.log.Logger.SetFormatter(&logrus.TextFormatter{})
			default:
				return fmt.Errorf("invalid logger format %s", flagValue)
			}
		case config.RelayEntriesFlag.Name:
			relayList := strings.Split(flagValue, ",")
			for _, relay := range relayList {
				relayEntry, err := NewRelayEntry(relay)
				if err != nil {
					return err
				}
				r.relays = append(r.relays, relayEntry)
			}
		case config.RelayCheckFlag.Name:
			r.relayCheck = true
		case config.SkipRelaySignatureCheck.Name:
			r.relaySignatureCheck = false
		case config.MinBidFlag.Name:
			minBidBigInt := new(big.Int)
			minBidBigInt, ok := minBidBigInt.SetString(flagValue, 10)
			if !ok {
				return fmt.Errorf("invalid min bid %s", flagValue)
			}
			minBid := relayCommon.U256Str{}
			err := minBid.FromBig(minBidBigInt)
			if err != nil {
				return err
			}
			r.relayMinBid = minBid
		case config.MainnetFlag.Name:
			if forkVersionFlagNameSet != "" || customForkVersion {
				return fmt.Errorf("cannot set %s and %s", config.MainnetFlag.Name, forkVersionFlagNameSet)
			}
			forkVersionFlagNameSet = config.MainnetFlag.Name
			r.cfg.GenesisForkVersion = relayCommon.GenesisForkVersionMainnet
			r.genesisTime = relayCommon.GenesisTimeMainnet
		case config.SepoliaFlag.Name:
			if forkVersionFlagNameSet != "" || customForkVersion {
				return fmt.Errorf("cannot set %s and %s", config.SepoliaFlag.Name, forkVersionFlagNameSet)
			}
			forkVersionFlagNameSet = config.SepoliaFlag.Name
			r.cfg.GenesisForkVersion = relayCommon.GenesisForkVersionSepolia
			r.genesisTime = relayCommon.GenesisTimeSepolia
		case config.GoerliFlag.Name:
			if forkVersionFlagNameSet != "" || customForkVersion {
				return fmt.Errorf("cannot set %s and %s", config.GoerliFlag.Name, forkVersionFlagNameSet)
			}
			forkVersionFlagNameSet = config.GoerliFlag.Name
			r.cfg.GenesisForkVersion = relayCommon.GenesisForkVersionGoerli
			r.genesisTime = relayCommon.GenesisTimeGoerli
		case config.GenesisForkVersionFlag.Name:
			if forkVersionFlagNameSet != "" && forkVersionFlagNameSet != "custom "+config.GenesisForkVersionFlag.Name {
				return fmt.Errorf("cannot set custom fork-version flag and %s", forkVersionFlagNameSet)
			}
			forkVersionFlagNameSet = "custom " + config.GenesisForkVersionFlag.Name
			r.cfg.GenesisForkVersion = flagValue
			customForkVersion = true
		case config.GenesisTimeFlag.Name:
			if forkVersionFlagNameSet != "" && forkVersionFlagNameSet != "custom "+config.GenesisForkVersionFlag.Name {
				return fmt.Errorf("cannot set custom genesis time flag and %s", forkVersionFlagNameSet)
			}
			genesisTime, err := strconv.ParseInt(flagValue, 10, 64)
			if err != nil {
				return err
			}
			r.genesisTime = uint64(genesisTime)
			customGenesisTime = true
		case config.GenesisValidatorsRootFlag.Name:
			r.cfg.GenesisValidatorsRoot = flagValue
		case config.RequestTimeoutMsFlag.Name:
			requestTimeoutMs, err := strconv.ParseInt(flagValue, 10, 64)
			if err != nil {
				return err
			}
			r.cfg.RequestTimeoutMs = int(requestTimeoutMs)
			r.httpClient.Timeout = time.Duration(requestTimeoutMs) * time.Millisecond
		case config.RequestMaxRetriesFlag.Name:
			requestMaxRetries, err := strconv.ParseInt(flagValue, 10, 64)
			if err != nil {
				return err
			}
			r.cfg.RequestMaxRetries = int(requestMaxRetries)
		default:
			return fmt.Errorf("invalid flag %s", flagName)
		}
	}

	if customGenesisTime && forkVersionFlagNameSet != "custom "+config.GenesisTimeFlag.Name && !customForkVersion {
		return fmt.Errorf("cannot set custom genesis-time flag without custom fork-version flag")
	}

	return nil
}

func validatePayloadBlock(blockBase commonTypes.BaseSignedBlindedBeaconBlock, log *logrus.Entry) error {
	if blockBase.Message == nil || blockBase.Message.Body == nil || blockBase.Message.Body.ExecutionPayloadHeader == nil {
		return ErrIncompletePayload
	}
	return nil
}

func validateOriginalBid(logger *logrus.Entry, originalBid bidResp) error {
	if originalBid.response.IsEmpty() {
		logger.Error("No bid for this getPayload payload found. Was getHeader called before?")
		return ErrNoBidReceived
	} else if len(originalBid.relays) == 0 {
		logger.Warn("Bid found but no associated relays")
	}
	return nil
}
