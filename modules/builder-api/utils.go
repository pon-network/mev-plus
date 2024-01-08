package builderapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"
	"strings"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/sirupsen/logrus"
)

const (
	HeaderKeySlotUID = "X-MEVPlusID"
	HeaderKeyVersion = "X-MEVPlus-Version"
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
		fmt.Println("*******************************")
	})
}
