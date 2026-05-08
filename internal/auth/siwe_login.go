package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// SIWESession orchestrates the gamma-api SIWE login flow and holds the
// resulting session cookie in an in-memory jar. Mirrors the Polymarket
// frontend's `useSignIn` mutation.
//
// Wire flow:
//  1. GET  {gammaURL}/nonce              → { nonce: "..." }
//  2. Build SIWEMessage, personal_sign with [PrivateKeySigner]
//  3. GET  {gammaURL}/login              → Authorization: Bearer <token>,
//     withCredentials → Set-Cookie
//  4. Cookies persist in the jar — pass [HTTPClient] to downstream callers
//     (relayer auth mint) so the cookies ride along.
type SIWESession struct {
	client   *http.Client
	signer   *PrivateKeySigner
	gammaURL string
	now      func() time.Time
}

// SIWESessionOption configures a SIWESession at construction time.
type SIWESessionOption func(*SIWESession)

// WithSIWEHTTPClient lets the caller supply a pre-configured *http.Client
// (with custom Timeout, Transport, etc). The cookie jar is preserved.
func WithSIWEHTTPClient(client *http.Client) SIWESessionOption {
	return func(s *SIWESession) {
		if client.Jar == nil {
			jar, _ := cookiejar.New(nil)
			client.Jar = jar
		}
		s.client = client
	}
}

// WithSIWEClock sets the clock used for `issuedAt` / `expirationTime`.
// Tests pass a fixed time so the SIWE message is deterministic.
func WithSIWEClock(now func() time.Time) SIWESessionOption {
	return func(s *SIWESession) { s.now = now }
}

// NewSIWESession builds a session with a fresh in-memory cookie jar.
// gammaURL is typically https://gamma-api.polymarket.com (no trailing slash).
func NewSIWESession(signer *PrivateKeySigner, gammaURL string, opts ...SIWESessionOption) (*SIWESession, error) {
	if signer == nil {
		return nil, fmt.Errorf("signer is required")
	}
	if strings.TrimSpace(gammaURL) == "" {
		return nil, fmt.Errorf("gammaURL is required")
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create cookie jar: %w", err)
	}
	s := &SIWESession{
		client:   &http.Client{Jar: jar, Timeout: 30 * time.Second},
		signer:   signer,
		gammaURL: strings.TrimRight(gammaURL, "/"),
		now:      time.Now,
	}
	for _, opt := range opts {
		opt(s)
	}
	if s.client.Jar == nil {
		s.client.Jar = jar
	}
	return s, nil
}

// HTTPClient returns the underlying http.Client (with the populated jar).
// Pass this to downstream calls — e.g. the relayer auth mint — so cookies
// captured by [Login] ride along automatically.
func (s *SIWESession) HTTPClient() *http.Client { return s.client }

// CookiesFor returns the cookies the jar would attach to a request to
// rawURL. Useful for diagnostics and for callers that need to forward
// cookies to a different transport.
func (s *SIWESession) CookiesFor(rawURL string) []*http.Cookie {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil
	}
	return s.client.Jar.Cookies(u)
}

type siweNonceResponse struct {
	Nonce string `json:"nonce"`
}

// Login runs the full SIWE flow. On success the polymarket session cookie
// is in the session's jar.
func (s *SIWESession) Login(ctx context.Context) error {
	nonce, err := s.fetchNonce(ctx)
	if err != nil {
		return fmt.Errorf("fetch nonce: %w", err)
	}
	msg := NewPolymarketSIWE(s.signer.Address(), nonce, s.signer.ChainID(), s.now())
	sig, err := s.signer.SignPersonalMessage([]byte(msg.String()))
	if err != nil {
		return fmt.Errorf("sign siwe: %w", err)
	}
	token, err := BuildSIWEBearerToken(msg, sig)
	if err != nil {
		return fmt.Errorf("build bearer token: %w", err)
	}
	if err := s.callLogin(ctx, token); err != nil {
		return fmt.Errorf("login: %w", err)
	}
	return nil
}

func (s *SIWESession) fetchNonce(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.gammaURL+"/nonce", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("HTTP %d %s: %s", resp.StatusCode, req.URL, string(body))
	}
	var parsed siweNonceResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("decode nonce: %w", err)
	}
	if strings.TrimSpace(parsed.Nonce) == "" {
		return "", fmt.Errorf("server returned empty nonce")
	}
	return parsed.Nonce, nil
}

func (s *SIWESession) callLogin(ctx context.Context, bearerToken string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.gammaURL+"/login", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("HTTP %d %s: %s", resp.StatusCode, req.URL, string(body))
	}
	return nil
}
