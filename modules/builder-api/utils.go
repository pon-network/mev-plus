package builderapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"
	"strings"

	"github.com/attestantio/go-builder-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/holiman/uint256"
	"github.com/sirupsen/logrus"
)

// bidRespKey is used as key for the bids cache
type bidRespKey struct {
	slot      uint64
	blockHash string
}

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

// bidInfo is used to store bid response fields for logging and validation
type bidInfo struct {
	blockHash   phase0.Hash32
	parentHash  phase0.Hash32
	pubkey      phase0.BLSPubKey
	blockNumber uint64
	txRoot      phase0.Root
	value       *uint256.Int
}

// bidResp are entries in the bids cache
type bidResp struct {
	t        time.Time
	response spec.VersionedSignedBuilderBid
	bidInfo  bidInfo
}

func httpClientDisallowRedirects(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}

const (
	HeaderKeySlotUID = "X-MEVPlusID"
	HeaderKeyVersion = "X-MEVPlus-Version"
)

var (
	errHTTPErrorResponse  = errors.New("HTTP error response")
	errInvalidForkVersion = errors.New("invalid fork version")
	errInvalidTransaction = errors.New("invalid transaction")
	errMaxRetriesExceeded = errors.New("max retries exceeded")
)

func createUrl(urlString string) (*url.URL, error) {
	if urlString == "" {
		return nil, nil
	}
	if !strings.HasPrefix(urlString, "http") {
		urlString = "http://" + urlString
	}

	return url.ParseRequestURI(urlString)
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

func (b *BuilderApiService) respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	resp := httpErrorResp{code, message}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		b.log.WithField("response", resp).WithError(err).Error("Couldn't write error response")
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func (b *BuilderApiService) respondOK(w http.ResponseWriter, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		b.log.WithField("response", response).WithError(err).Error("Couldn't write OK response")
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func LoggingMiddleware(logger *logrus.Entry, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)

				method := ""
				url := ""
				if r != nil {
					method = r.Method
					url = r.URL.EscapedPath()
				}

				logger.WithFields(logrus.Fields{
					"err":    err,
					"trace":  string(debug.Stack()),
					"method": r.Method,
				}).Error(fmt.Sprintf("http request panic: %s %s", method, url))
			}
		}()
		start := time.Now()
		fmt.Println("\n*******************************")
		fmt.Println("*******************************")
		logger.Info(r.RequestURI)
		next.ServeHTTP(w, r)
		// wrapped := wrapResponseWriter(w)
		// next.ServeHTTP(wrapped, r)
		logger.WithFields(logrus.Fields{
			"method":   r.Method,
			"path":     r.URL.EscapedPath(),
			"duration": fmt.Sprintf("%f", time.Since(start).Seconds()),
		}).Info(fmt.Sprintf("http: %s %s", r.Method, r.URL.EscapedPath()))
		fmt.Println("*******************************")
		fmt.Println("*******************************\n")
	})
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}
