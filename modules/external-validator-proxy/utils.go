package externalvalidatorproxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"context"
	"io"
	"time"
	"bytes"
	"strings"

	"github.com/sirupsen/logrus"
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
