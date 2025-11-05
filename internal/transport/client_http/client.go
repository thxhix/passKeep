// Package client_http provides a thin HTTP client used by CLI and API wrappers.
//
// It contains a small wrapper around net/http.Client with JSON request/response helpers,
// multipart upload helper and basic response/error normalization.
package client_http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/thxhix/passKeeper/internal/client/token"
	"github.com/thxhix/passKeeper/internal/transport/http/dto"
	"go.uber.org/zap"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

// Client is a small HTTP client wrapper.
// It holds baseURL, http.Client and optional logger.
// Use NewHttpClient to construct a new instance.
type Client struct {
	baseURL      string
	http         *http.Client
	accessToken  string
	refreshToken string
	logger       *zap.Logger
	mu           sync.RWMutex
}

// NewHttpClient creates a new Client pointed to baseURL.
// logger may be nil; if nil the client will use zap.NewNop internally (recommended).
// Returns ErrEmptyBaseURL when baseURL is empty.
func NewHttpClient(baseURL string, logger *zap.Logger) (*Client, error) {
	if baseURL == "" {
		return nil, ErrEmptyBaseURL
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       60 * time.Second,
		MaxIdleConns:          100,
		MaxConnsPerHost:       0,
		MaxIdleConnsPerHost:   10,
	}
	return &Client{
		logger:  logger,
		baseURL: baseURL,
		http: &http.Client{
			Timeout:   10 * time.Second,
			Transport: tr,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return ErrTooManyRedirects
				}
				if len(via) > 0 && via[0].Header.Get("Authorization") != "" {
					req.Header.Set("Authorization", via[0].Header.Get("Authorization"))
				}
				return nil
			},
		},
	}, nil
}

// Do send a JSON request (if body != nil, it will be encoded as JSON) and decodes a JSON response into result.
// - body can be any value that encoding/json can encode.
// - result must be a pointer where JSON will be unmarshaled, or nil if empty body expected.
// Returns *HTTPError for non-2xx responses (so callers can inspect StatusCode and Body).
func (c *Client) Do(ctx context.Context, method, path string, body any, result any) error {
	if path == "" {
		return ErrPathIsEmpty
	}
	if method == "" {
		return ErrEmptyMethod
	}

	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			c.logger.Error("marshal request body failed", zap.Error(err))
			return fmt.Errorf("%w: %v", ErrCantMarshalBody, err)
		}
		bodyReader = buf
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		c.logger.Error("create request failed", zap.Error(err))
		return fmt.Errorf("%w: %v", ErrClientRequest, err)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	keyRingTokensStorage, err := token.LoadTokens()
	if err != nil {
		c.logger.Error("load tokens failed", zap.Error(err))
	}

	if keyRingTokensStorage.Access != "" {
		req.Header.Set("Authorization", "Bearer "+keyRingTokensStorage.Access)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		c.logger.Error("http do failed", zap.Error(err))
		return fmt.Errorf("%w: %v", ErrClientRequestFailed, err)
	}
	defer resp.Body.Close()

	const maxBodySize = 10 << 20
	limited := io.LimitReader(resp.Body, maxBodySize+1)
	respBytes, err := io.ReadAll(limited)
	if err != nil {
		c.logger.Error("read response failed", zap.Error(err))
		return fmt.Errorf("%w: %v", ErrCantReadBody, err)
	}
	if int64(len(respBytes)) > maxBodySize {
		c.logger.Error("response too large", zap.Int("read", len(respBytes)))
		return ErrServerResponseTooLarge
	}

	if len(respBytes) == 0 {
		return nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var srvErr dto.ErrorResponse
		if err := json.Unmarshal(respBytes, &srvErr); err == nil && srvErr.ErrorText != "" {
			c.logger.Warn("server returned error", zap.Int("code", resp.StatusCode), zap.Any("error", srvErr))
			return &HTTPError{StatusCode: resp.StatusCode, Body: srvErr.ErrorText}
		}

		c.logger.Warn("server returned non-json error", zap.Int("code", resp.StatusCode), zap.String("body", string(respBytes)))
		return &HTTPError{StatusCode: resp.StatusCode, Body: string(respBytes)}
	}

	if result == nil {
		return nil
	}

	if err := json.Unmarshal(respBytes, result); err != nil {
		c.logger.Error("failed to decode response", zap.Error(err), zap.String("resp_snippet", string(respBytes)))
		return fmt.Errorf("%w: %v", ErrServerResponseUnmarshal, err)
	}

	return nil
}

// DoMultiPart sends a request with a streaming body, typically used for multipart/form-data uploads.
// - body is an io.Reader (often a pipe created with io.Pipe and multipart.Writer).
// - contentType should be the full value returned by multipart.Writer.FormDataContentType().
// - result is pointer to struct to unmarshal response or nil.
func (c *Client) DoMultiPart(ctx context.Context, method, path string, body io.Reader, contentType string, result any) error {
	if path == "" {
		return ErrPathIsEmpty
	}
	if method == "" {
		return ErrEmptyMethod
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("create upload request failed", zap.Error(err))
		}
		return fmt.Errorf("%w: %v", ErrClientRequest, err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", contentType)

	keyRingTokensStorage, err := token.LoadTokens()
	if err != nil {
		c.logger.Error("load tokens failed", zap.Error(err))
	}

	if keyRingTokensStorage.Access != "" {
		req.Header.Set("Authorization", "Bearer "+keyRingTokensStorage.Access)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("upload http do failed", zap.Error(err))
		}
		return fmt.Errorf("%w: %v", ErrClientRequestFailed, err)
	}
	defer resp.Body.Close()

	// reuse same response handling as Do: limit, parse errors, decode result
	const maxBodySize = 10 << 20
	limited := io.LimitReader(resp.Body, maxBodySize+1)
	respBytes, err := io.ReadAll(limited)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("read upload response failed", zap.Error(err))
		}
		return fmt.Errorf("%w: %v", ErrCantReadBody, err)
	}
	if int64(len(respBytes)) > maxBodySize {
		if c.logger != nil {
			c.logger.Error("upload response too large", zap.Int("read", len(respBytes)))
		}
		return ErrServerResponseTooLarge
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var srvErr dto.ErrorResponse
		if err := json.Unmarshal(respBytes, &srvErr); err == nil && srvErr.ErrorText != "" {
			return &HTTPError{StatusCode: resp.StatusCode, Body: string(respBytes)}
		}
		return &HTTPError{StatusCode: resp.StatusCode, Body: string(respBytes)}
	}

	if result == nil || len(respBytes) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBytes, result); err != nil {
		if c.logger != nil {
			c.logger.Error("failed to unmarshal upload response", zap.Error(err))
		}
		return fmt.Errorf("%w: %v", ErrServerResponseUnmarshal, err)
	}
	return nil
}
