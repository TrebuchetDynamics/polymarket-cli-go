// Package transport provides HTTP transport with retry, redaction, and rate limiting.
// Patterns stolen from polymarket-go-sdk/pkg/transport.
package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/TrebuchetDynamics/polygolem/internal/errors"
)

// Config holds transport configuration.
type Config struct {
	BaseURL     string
	Timeout     time.Duration
	UserAgent   string
	RetryMax    int
	RetryDelay  time.Duration
	RateLimiter    *RateLimiter
	CircuitBreaker *CircuitBreaker
}

// DefaultConfig returns a sensible default config.
func DefaultConfig(baseURL string) Config {
	return Config{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		Timeout:    30 * time.Second,
		UserAgent:  "polygolem/0.1",
		RetryMax:   3,
		RetryDelay: 100 * time.Millisecond,
	}
}

// Client wraps an http.Client with retry, error normalization, and redaction.
type Client struct {
	http   *http.Client
	config Config
	mu     sync.RWMutex
}

// New creates a new transport Client.
func New(httpClient *http.Client, config Config) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: config.Timeout}
	}
	return &Client{http: httpClient, config: config}
}

// Get performs a GET request with retry for idempotent reads.
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	return c.do(ctx, http.MethodGet, path, nil, result)
}

// GetRaw performs a GET request and returns raw bytes.
func (c *Client) GetRaw(ctx context.Context, path string) ([]byte, error) {
	var raw json.RawMessage
	if err := c.do(ctx, http.MethodGet, path, nil, &raw); err != nil {
		return nil, err
	}
	return []byte(raw), nil
}

// Post performs a POST request. POSTs are never retried.
func (c *Client) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.do(ctx, http.MethodPost, path, body, result)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}, result interface{}) (err error) {
	if c.config.CircuitBreaker != nil {
		defer func() {
			c.config.CircuitBreaker.RecordResult(err)
		}()
	}
	if c.config.CircuitBreaker != nil {
		if cerr := c.config.CircuitBreaker.beforeRequest(); cerr != nil {
			return cerr
		}
	}

	if c.config.RateLimiter != nil {
		if err := c.config.RateLimiter.Wait(ctx); err != nil {
			return err
		}
	}

	url := c.config.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return errors.Wrap(errors.CodeInvalidValue, "failed to marshal request body", err)
		}
		bodyReader = strings.NewReader(string(b))
	}

	var lastErr error
	for attempt := 0; attempt <= c.retryMax(method); attempt++ {
		if attempt > 0 {
			delay := c.config.RetryDelay * time.Duration(1<<(attempt-1))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return errors.Wrap(errors.CodeConnectionFailed, "failed to create request", err)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", c.config.UserAgent)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = errors.Wrap(errors.CodeConnectionFailed, "request failed", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = errors.WithHTTP(errors.CodeRateLimited, "rate limited", resp.StatusCode)
			continue
		}

		if resp.StatusCode >= 500 {
			respBody, _ := io.ReadAll(resp.Body)
			lastErr = errors.WithHTTP(errors.CodeConnectionFailed, fmt.Sprintf("server error: %s", string(respBody)), resp.StatusCode)
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			respBody, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("HTTP %d %s: %s", resp.StatusCode, url, string(respBody))
		}

		if result != nil {
			if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
				return fmt.Errorf("decode response from %s: %w", url, err)
			}
		}
		return nil
	}

	return lastErr
}

func (c *Client) retryMax(method string) int {
	if method == http.MethodGet {
		return c.config.RetryMax
	}
	return 0
}

// RedactSecret replaces secret values with "[REDACTED]" for safe logging.
func RedactSecret(v string) string {
	if v == "" {
		return ""
	}
	if len(v) <= 8 {
		return "[REDACTED]"
	}
	return v[:4] + "..." + v[len(v)-4:]
}

// RedactMap returns a copy of headers with sensitive values redacted.
func RedactMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	sensitive := map[string]bool{
		"POLY_PRIVATE_KEY":         true,
		"POLY_API_KEY":             true,
		"POLY_SECRET":              true,
		"POLY_PASSPHRASE":          true,
		"POLY_SIGNATURE":           true,
		"POLY_BUILDER_API_KEY":     true,
		"POLY_BUILDER_SECRET":      true,
		"POLY_BUILDER_PASSPHRASE":  true,
		"POLY_BUILDER_SIGNATURE":   true,
	}
	for k, v := range m {
		if sensitive[strings.ToUpper(k)] {
			out[k] = RedactSecret(v)
		} else {
			out[k] = v
		}
	}
	return out
}

// BaseURL returns the configured base URL.
func (c *Client) BaseURL() string {
	return c.config.BaseURL
}
